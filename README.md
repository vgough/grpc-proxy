# gRPC Proxy

[![Travis Build](https://travis-ci.org/vgough/grpc-proxy.svg?branch=master)](https://travis-ci.org/vgough/grpc-proxy)
[![Go Report Card](https://goreportcard.com/badge/github.com/vgough/grpc-proxy)](https://goreportcard.com/report/github.com/vgough/grpc-proxy)
[![GoDoc](http://img.shields.io/badge/GoDoc-Reference-blue.svg)](https://godoc.org/github.com/vgough/grpc-proxy)
[![Apache 2.0 License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

[gRPC Go](https://github.com/grpc/grpc-go) Proxy server

## Project Goal

Makes it easy to build a reverse proxy for gRPC targets.  This allows users to
route gRPC requests based on method names and metadata without any knowledge of
message contents.

The project was forked from the proof of concept work in
github.com/mwitkow/grpc-proxy, and further refined with a real router project
running in a production environment with multiple routing topologies.

## Proxy Package

The package [`proxy`](proxy/) contains a generic gRPC reverse proxy handler that
allows a gRPC server.

A `StreamDirector` implementation is responsible for deciding where (if at all)
to send a request (see example_test.go).  This contrived example demonstrates how
a user could use the path and associated request metadata to route a request:
```go
func (d *ExampleDirector) Connect(ctx context.Context, method string) (context.Context, *grpc.ClientConn, error) {
  // Disable forwarding for all services prefixed with com.example.internal.
  if strings.HasPrefix(method, "/com.example.internal.") {
    return nil, nil, grpc.Errorf(codes.Unimplemented, "Unknown method")
  }
  md, ok := metadata.FromIncomingContext(ctx)
  if ok {
    // Decide on which backend to dial
    if val, exists := md[":authority"]; exists && val[0] == "staging.api.example.com" {
      // Make sure we use DialContext so the dialing can be cancelled/time out together with the context.
      conn, err := grpc.DialContext(ctx, "api-service.staging.svc.local", grpc.WithCodec(proxy.Codec()))
      return ctx, conn, err
    } else if val, exists := md[":authority"]; exists && val[0] == "api.example.com" {
      conn, err := grpc.DialContext(ctx, "api-service.prod.svc.local", grpc.WithCodec(proxy.Codec()))
      return ctx, conn, err
    }
  }
  return nil, nil, grpc.Errorf(codes.Unimplemented, "Unknown method")
}
```
The direct is registered with a `grpc.Server`, along with a special codec which
allows the proxy to handle raw byte frames and pass them along without
any serialization. The server may have other handlers that will be served
locally, and the codec will fall back to the protobuf codec when necessary:

```go
server := grpc.NewServer(
    grpc.CustomCodec(proxy.Codec()),
    grpc.UnknownServiceHandler(proxy.TransparentHandler(director)))
pb_test.RegisterTestServiceServer(server, &testImpl{})
```

## License

`grpc-proxy` is released under the Apache 2.0 license. See [LICENSE.txt](LICENSE.txt).

