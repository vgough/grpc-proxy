package proxy

import (
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metadata "google.golang.org/grpc/metadata"
)

func TestBiDirCopy_ClientEOF(t *testing.T) {
	req := &ServerStream{}  // requestor side
	dest := &ClientStream{} // dest side

	header := metadata.MD{
		"test": []string{"abc"},
	}
	trailer := metadata.MD{
		"test": []string{"xyz"},
	}

	dest.On("Header").Return(header, nil).Once()
	req.On("SendHeader", mock.AnythingOfType("metadata.MD")).Run(func(args mock.Arguments) {
		md := args.Get(0).(metadata.MD)
		assert.EqualValues(t, header, md)
	}).Return(nil).Once()

	// Client immediately sends EOF.
	req.On("RecvMsg", mock.AnythingOfType("*proxy.frame")).Return(io.EOF).Once()
	dest.On("CloseSend").Return(nil).Once()

	// Server message is forwarded to client.
	dest.On("RecvMsg", mock.AnythingOfType("*proxy.frame")).Run(func(args mock.Arguments) {
		frame := args.Get(0).(*frame)
		frame.payload = []byte{0x01, 0x02}
	}).Return(nil).Once()
	req.On("SendMsg", mock.AnythingOfType("*proxy.frame")).Run(func(args mock.Arguments) {
		frame := args.Get(0).(*frame)
		assert.Equal(t, []byte{0x01, 0x02}, frame.payload)
	}).Return(nil).Once()
	dest.On("RecvMsg", mock.AnythingOfType("*proxy.frame")).Return(io.EOF).Once()

	// Trailers will also be sent.
	dest.On("Trailer").Return(trailer, nil).Once()
	req.On("SetTrailer", mock.AnythingOfType("metadata.MD")).Run(func(args mock.Arguments) {
		md := args.Get(0).(metadata.MD)
		assert.EqualValues(t, trailer, md)
	}).Return(nil).Once()

	err := biDirCopy(req, dest)
	require.EqualError(t, err, io.EOF.Error())

	req.AssertExpectations(t)
	dest.AssertExpectations(t)
}

func TestBiDirCopy_ClientFail(t *testing.T) {
	req := &ServerStream{}  // requestor side
	dest := &ClientStream{} // dest side

	header := metadata.MD{}
	trailer := metadata.MD{}

	dest.On("Header").Return(header, nil).Once()
	req.On("SendHeader", mock.AnythingOfType("metadata.MD")).Run(func(args mock.Arguments) {
		md := args.Get(0).(metadata.MD)
		assert.EqualValues(t, header, md)
	}).Return(nil).Once()

	block := make(chan time.Time)

	// Client fails immediately.
	req.On("RecvMsg", mock.AnythingOfType("*proxy.frame")).Return(io.ErrNoProgress).Once()
	dest.On("CloseSend").Run(func(args mock.Arguments) {
		close(block)
	}).Return(nil).Once()

	// Server blocks.
	dest.On("RecvMsg", mock.AnythingOfType("*proxy.frame")).WaitUntil(block).Return(io.EOF).Once()

	// Trailers will also be sent.
	dest.On("Trailer").Return(trailer, nil).Once()
	req.On("SetTrailer", mock.AnythingOfType("metadata.MD")).Return(nil).Once()

	err := biDirCopy(req, dest)
	require.Error(t, err)

	req.AssertExpectations(t)
	dest.AssertExpectations(t)
}
