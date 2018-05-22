# connector
--
    import "github.com/vgough/grpc-proxy/connector"

Package connector provides connection management strategies for gRPC proxies.

## Usage

#### type CachedEntry

```go
type CachedEntry struct {
	// Cfg is the immutable configuration for the endpoint.
	Conn *grpc.ClientConn
}
```

CachedEntry tracks usage and age of a connection. When an entry is looked up,
it's reference count is incremented. Connector.Release() must be called when the
connection is no longer in use, in order to release unused resources.

#### type CachingConnector

```go
type CachingConnector struct {

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
```

CachingConnector is a caching creator of rpc connections, keyed by address.

#### func  NewCachingConnector

```go
func NewCachingConnector(opts ...Opt) *CachingConnector
```
NewCachingConnector returns a new connection cache instance. Connections will be
cached and reused between calls.

The provided dialer is used to create new gRPC connections. This may be a
grpc.Dialer.

#### func (*CachingConnector) Connect

```go
func (c *CachingConnector) Connect(ctx context.Context, addr string) (*grpc.ClientConn, error)
```
Connect establishes a connection or returns a cached entry. Close must be called
on the returned value, if not nil.

#### func (*CachingConnector) Expire

```go
func (c *CachingConnector) Expire() []string
```
Expire cleans up old connections.

#### func (*CachingConnector) Release

```go
func (c *CachingConnector) Release(addr string, conn *grpc.ClientConn)
```
Release allows unused resources to be freed. Must be called once per successful
call to Connect.

#### func (*CachingConnector) Remove

```go
func (c *CachingConnector) Remove(addr string) bool
```
Remove moves the connection to the cleanup list, where it will be closed on the
next call to Expire when the reference count reaches zero.

#### type Opt

```go
type Opt func(*CachingConnector)
```

Opt is an option to NewCachingConnector.

#### func  WithDialer

```go
func WithDialer(dialer func(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error)) Opt
```
WithDialer specifies a customer dialer for the caching connector.
