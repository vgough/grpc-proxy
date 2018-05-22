# proxy
--
    import "github.com/vgough/grpc-proxy/proxy"

Package proxy provides a gRPC proxy library.

This package exposes a `StreamDirector` API that allows users of the package to
implement arbitrary request routing logic.

The implementation integrates with `grpc.Server`, allowing the StreamDirector to
### connect an incoming ServerStream to an outgoing ClientStream without encoding or
decoding the messages. This allows the construction of forward and reverse gRPC
proxies.

## Usage

#### func  Codec

```go
func Codec() grpc.Codec
```
Codec returns a proxying grpc.Codec with the default protobuf codec as parent.

See CodecWithParent.

#### func  CodecWithParent

```go
func CodecWithParent(fallback grpc.Codec) grpc.Codec
```
CodecWithParent returns a proxying grpc.Codec with a user provided codec as
parent.

This codec is *crucial* to the functioning of the proxy. It allows the proxy
server to be oblivious to the schema of the forwarded messages. It basically
treats a gRPC message frame as raw bytes. However, if the server handler, or the
client caller are not proxy-internal functions it will fall back to trying to
decode the message using a fallback codec.

#### func  RegisterService

```go
func RegisterService(server *grpc.Server, director StreamDirector, serviceName string, methodNames ...string)
```
RegisterService sets up a proxy handler for a particular gRPC service and
method. The behavior is the same as if you were registering a handler method,
e.g. from a codegenerated pb.go file.

This can *only* be used if the `server` also uses proxy.CodecForServer()
ServerOption.

#### func  TransparentHandler

```go
func TransparentHandler(director StreamDirector) grpc.StreamHandler
```
TransparentHandler returns a handler that attempts to proxy all requests that
are not registered in the server. The indented use here is as a transparent
proxy, where the server doesn't know about the services implemented by the
backends. It should be used as a `grpc.UnknownServiceHandler`.

This can *only* be used if the `server` also uses proxy.CodecForServer()
ServerOption.

#### type StreamDirector

```go
type StreamDirector interface {
	// Connect returns a connection to use for the given method,
	// or an error if the call should not be handled.
	//
	// The provided context may be inspected for filtering on request
	// metadata.
	//
	// Method is the gRPC request path, which is in the form "/service/method".
	//
	// The returned context is used as the basis for the outgoing connection.
	Connect(ctx context.Context, method string) (context.Context, *grpc.ClientConn, error)

	// Release is called when a connection is longer being used.  This is called
	// once for every call to Connect that does not return an error.
	//
	// The provided context is the one returned from Connect.
	//
	// This can be used by the director to pool connections or close unused
	// connections.
	Release(ctx context.Context, conn *grpc.ClientConn)
}
```

StreamDirector manages gRPC Client connections for forwarding requests.

The presence of the `Context` allows for rich filtering, e.g. based on Metadata
(headers). If no handling is meant to be done, a `codes.NotImplemented` gRPC
error should be returned.

Connect will be called *after* all server-side stream interceptors are invoked.
So decisions around authorization, monitoring etc. are better to be handled
there.
