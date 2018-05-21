// Copyright 2018 Valient Gough
// Copyright 2017 Michal Witkowski
// All Rights Reserved.
// See LICENSE for licensing terms.

/*
Package proxy provides a reverse proxy handler for gRPC.

The implementation allows a `grpc.Server` to pass a received ServerStream to a
ClientStream without understanding the semantics of the messages exchanged. It
basically provides a transparent reverse-proxy.

This package exposes a `StreamDirector` API that allows users of this package to
implement whatever logic of backend-picking, dialing and service verification to
perform.

See examples on documented functions.
*/
package proxy
