package connector

import (
	"context"
	"errors"
	"sync"

	"golang.org/x/sync/singleflight"
	"google.golang.org/grpc"
)

// Opt is an option to NewCachingConnector.
type Opt func(*CachingConnector)

// WithDialer specifies a customer dialer for the caching connector.
func WithDialer(dialer func(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error)) Opt {
	return func(cc *CachingConnector) {
		cc.dialer = dialer
	}
}

// NewCachingConnector returns a new connection cache instance.
// Connections will be cached and reused between calls.
//
// The provided dialer is used to create new gRPC connections.  This may be
// a grpc.Dialer.
//
// To use, use CachingConnector.Dial instead of grpc.Dial when creating
// connections to remote endpoints.  CachingConnector.Release must be called
// once for each successful Dial call.
//
// Note that CachingConnector.Expire must be called periodically in order to
// free unused resources.  Any connection which has been inactive for 2
// consecutive Expire calls will be closed.
func NewCachingConnector(opts ...Opt) *CachingConnector {
	cc := &CachingConnector{
		entries: make(map[string]*cachedEntry),
		dialer:  grpc.DialContext,
	}
	for _, o := range opts {
		o(cc)
	}
	return cc
}

// CachingConnector is a caching creator of rpc connections, keyed by address.
type CachingConnector struct {
	mu        sync.Mutex
	sf        singleflight.Group
	entries   map[string]*cachedEntry // cached entries by address
	cleanup   []*cachedEntry          // entries awaiting cleanup
	dialer    func(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error)
	skipClose bool // for use in testing
	openCount int  // number of open connections

	// OnConnect is an optional callback when a connection request is received.
	// Use for metrics integration.
	OnConnect func(addr string)
	// OnCacheMiss is an optional callback when a new connection will be
	// established.  Use for metrics integration.
	OnCacheMiss func(addr string)
	// OnConnectionCountUpdate is an optional callback which provides the
	// total active connection count.  Use for metrics integration.
	OnConnectionCountUpdate func(count int)
}

// cachedEntry tracks usage and age of a connection.
// When an entry is looked up, it's reference count is incremented.
// Connector.Release() must be called when the connection is no longer in
// use, in order to release unused resources.
type cachedEntry struct {
	// Cfg is the immutable configuration for the endpoint.
	Conn     *grpc.ClientConn
	refCount int
	age      int
	addr     string
}

// Dial establishes a connection or returns a cached entry.  Close must be
// called on the returned value, if not nil.
//
// Dial options are not supported, since they would be ignored if a cached
// connection is used.
func (c *CachingConnector) Dial(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	if c.OnConnect != nil {
		c.OnConnect(addr)
	}

	// singleflight serializes lookups by address, allowing parallism between addresses.
	v, err, _ := c.sf.Do(addr, func() (interface{}, error) {
		ent := c.lookup(addr)
		if ent != nil {
			return ent, nil
		}
		if c.OnCacheMiss != nil {
			c.OnCacheMiss(addr)
		}

		// Try connecting.  This may take some time, but the singleflight wrapper
		// ensures that we only have 1 ongoing connection attempt per address.
		conn, err := c.dialer(ctx, addr)
		if err != nil {
			return nil, err
		}

		// Store in cache.
		return c.store(addr, conn), nil
	})

	if err != nil {
		return nil, err
	}
	return c.reserve(v.(*cachedEntry))
}

// Release allows unused resources to be freed.
// Must be called once per successful call to Connect.
func (c *CachingConnector) Release(addr string, conn *grpc.ClientConn) {
	c.mu.Lock()
	ent := c.entries[addr]

	if ent != nil && ent.Conn == conn {
		ent.refCount--
	}
	c.mu.Unlock()
}

// Expire cleans up old connections.
//
// This must be called periodically for proper functioning of CachedConnection.
func (c *CachingConnector) Expire() []string {
	old := c.unlinkOldConnections()
	var removed []string
	for _, ent := range old {
		if !c.skipClose {
			ent.Conn.Close()
		}
		removed = append(removed, ent.addr)
	}
	return removed
}

// CloseOnRelease moves any connection associated with the given address to the
// cleanup list, where it will be closed by Expire as soon as the reference
// count reaches zero.
//
// It is not normally necessary to call this function, as unused connections are
// closed automatically over time.
func (c *CachingConnector) CloseOnRelease(addr string) bool {
	c.mu.Lock()

	ent, ok := c.entries[addr]
	if ok {
		delete(c.entries, addr)
		c.cleanup = append(c.cleanup, ent)
	}

	c.mu.Unlock()
	return ok
}

func (c *CachingConnector) lookup(addr string) *cachedEntry {
	c.mu.Lock()
	ent, ok := c.entries[addr]
	if ok {
		ent.age = 0
	}
	c.mu.Unlock()
	return ent
}

func (c *CachingConnector) store(addr string, conn *grpc.ClientConn) *cachedEntry {
	c.mu.Lock()
	ent := &cachedEntry{
		Conn: conn,
		addr: addr,
	}
	c.entries[addr] = ent
	c.openCount++
	if c.OnConnectionCountUpdate != nil {
		c.OnConnectionCountUpdate(c.openCount)
	}
	c.mu.Unlock()

	return ent
}

func (c *CachingConnector) reserve(ent *cachedEntry) (ret *grpc.ClientConn, err error) {
	c.mu.Lock()
	if ent.age > 1 {
		// Should never happen.
		err = errors.New("connection reservation timeout")
	} else {
		ent.refCount++
		ret = ent.Conn
	}
	c.mu.Unlock()
	return ret, err
}

func (c *CachingConnector) unlinkOldConnections() []*cachedEntry {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Move old entries w/ 0 refCount to cleanup list.
	for addr, ent := range c.entries {
		ent.age++
		if ent.age > 1 && ent.refCount == 0 {
			delete(c.entries, addr)
			c.cleanup = append(c.cleanup, ent)
		}
	}

	var old []*cachedEntry
	filtered := c.cleanup[:0] // filter in-place
	for _, ent := range c.cleanup {
		if ent.refCount > 0 {
			// Reference exists, retain for review next time.
			filtered = append(filtered, ent)
			continue
		}

		c.openCount--
		old = append(old, ent)
	}
	c.cleanup = filtered
	return old
}
