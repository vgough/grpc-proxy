// Copyright 2018 Valient Gough
// Copyright 2017 Michal Witkowski
// All Rights Reserved.
// See LICENSE for licensing terms.

/*
Package proxy provides a reverse proxy handler for gRPC.

This package exposes a `StreamDirector` API that allows users of this package to
implement arbitrary request routing logic.

The implementation integrates with `grpc.Server`, allowing the StreamDirector
to connect an incoming ServerStream to an outgoing ClientStream without
understanding the semantics of the messages exchanged.  This is a building block
for creation of forward and reverse gRPC proxies.
*/
package proxy
