// Copyright 2018 Valient Gough
// All Rights Reserved.
// See LICENSE for licensing terms.

package proxy

import (
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// biDirCopy connects an incoming ServerStream with an outgoing ClientStream.
// This acts as a middleman, passing messages from streams in both directions.
func biDirCopy(in grpc.ServerStream, out grpc.ClientStream) error {
	done := make(chan error)
	go func() {
		done <- forwardIn(in, out)
	}()
	err := forwardOut(in, out)
	err2 := <-done
	if err != io.EOF {
		return err
	}
	return err2
}

// forward from input to destination.
func forwardOut(in grpc.ServerStream, out grpc.ClientStream) error {
	err := copyStream(in, out)
	err2 := out.CloseSend()

	switch err {
	case io.EOF:
		return err
	case nil:
		return err2
	default:
		return grpc.Errorf(codes.Internal, "failed proxying s2c: %s", err)
	}
}

// forward from output back to caller.
func forwardIn(in grpc.ServerStream, out grpc.ClientStream) error {
	// Forward header first.
	md, err := out.Header()
	if err != nil {
		return err
	}
	if err := in.SendHeader(md); err != nil {
		return err
	}

	err = copyStream(out, in)
	in.SetTrailer(out.Trailer())

	return err
}

func copyStream(src grpc.Stream, dst grpc.Stream) error {
	var f frame
	for {
		if err := src.RecvMsg(&f); err != nil {
			return err
		}
		if err := dst.SendMsg(&f); err != nil {
			return err
		}
	}
}
