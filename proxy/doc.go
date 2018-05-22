// Copyright 2018 Valient Gough
// All Rights Reserved.
// See LICENSE for licensing terms.

/*
Package proxy provides a gRPC proxy library.

This package exposes a `StreamDirector` API that allows users of the package to
implement arbitrary request routing logic.

The implementation integrates with `grpc.Server`, allowing the StreamDirector to
connect an incoming ServerStream to an outgoing ClientStream without encoding or
decoding the messages.  This allows the construction of forward and reverse gRPC
proxies.
*/
package proxy
