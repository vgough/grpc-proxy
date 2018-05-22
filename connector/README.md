# connector
--
    import "github.com/vgough/grpc-proxy/connector"

Package connector provides connection management strategies for gRPC proxies.

## Usage

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

To use, use CachingConnector.Dial instead of grpc.Dial when creating connections
to remote endpoints. CachingConnector.Release must be called once for each
successful Dial call.

Note that CachingConnector.Expire must be called periodically in order to free
unused resources. Any connection which has been inactive for 2 consecutive
Expire calls will be closed.

#### func (*CachingConnector) CloseOnRelease

```go
func (c *CachingConnector) CloseOnRelease(addr string) bool
```
CloseOnRelease moves any connection associated with the given address to the
cleanup list, where it will be closed by Expire as soon as the reference count
reaches zero.

It is not normally necessary to call this function, as unused connections are
closed automatically over time.

#### func (*CachingConnector) Dial

```go
func (c *CachingConnector) Dial(ctx context.Context, addr string) (*grpc.ClientConn, error)
```
Dial establishes a connection or returns a cached entry. Close must be called on
the returned value, if not nil.

Dial options are not supported, since they would be ignored if a cached
connection is used.

#### func (*CachingConnector) Expire

```go
func (c *CachingConnector) Expire() []string
```
Expire cleans up old connections.

This must be called periodically for proper functioning of CachedConnection.

#### func (*CachingConnector) Release

```go
func (c *CachingConnector) Release(addr string, conn *grpc.ClientConn)
```
Release allows unused resources to be freed. Must be called once per successful
call to Connect.

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
