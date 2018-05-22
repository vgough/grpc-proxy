package connector

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type MockDialer struct {
	mu    sync.Mutex
	count map[string]int
}

func (md *MockDialer) DialContext(ctx context.Context, addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	md.mu.Lock()
	defer md.mu.Unlock()
	md.count[addr]++
	return &grpc.ClientConn{}, nil
}

func TestConnectorDialFail(t *testing.T) {
	dialer := func(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
		return nil, errors.New("internal error")
	}
	c := NewCachingConnector(WithDialer(dialer))

	conn, err := c.Dial(context.Background(), "test")
	require.Error(t, err)
	require.Nil(t, conn)

	// Other failures..
	c.Release("test", nil)
}

func TestConnector(t *testing.T) {
	const addr = "https://localhost:1234"
	md := &MockDialer{
		count: make(map[string]int),
	}
	c := NewCachingConnector(WithDialer(md.DialContext))
	c.skipClose = true

	var connectCount int
	var cacheMiss int
	c.OnConnectionCountUpdate = func(count int) {
		connectCount = count
	}
	c.OnCacheMiss = func(addr string) {
		cacheMiss++
	}

	t.Run("Parallel_Connect", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				conn, err := c.Dial(context.Background(), addr)
				require.NoError(t, err)
				time.Sleep(50 * time.Millisecond)
				c.Release(addr, conn)
			}()
		}
		wg.Wait()

		require.Len(t, md.count, 1)
		require.Equal(t, 1, md.count[addr])
		require.Equal(t, 0, c.entries[addr].refCount)

		require.NotZero(t, connectCount)
		require.NotZero(t, cacheMiss)
	})

	t.Run("Remove", func(t *testing.T) {
		conn, err := c.Dial(context.Background(), addr)
		require.NoError(t, err)
		c.Release(addr, conn)

		// Check that there's a cached connection and nothing to cleanup.
		require.NotNil(t, c.entries[addr])
		require.Len(t, c.cleanup, 0)

		// Remove.
		ok := c.CloseOnRelease(addr)
		require.True(t, ok)
		require.Len(t, c.cleanup, 1)

		expired := c.Expire()
		require.EqualValues(t, []string{addr}, expired)
	})
}

func BenchmarkConnector(b *testing.B) {
	md := &MockDialer{
		count: make(map[string]int),
	}
	c := NewCachingConnector(WithDialer(md.DialContext))
	c.skipClose = true

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		addr := "https://localhost:1234"
		conn, err := c.Dial(context.Background(), addr)
		assert.NoError(b, err)
		c.Release(addr, conn)
	}
}
