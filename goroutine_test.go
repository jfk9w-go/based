package based_test

import (
	"context"
	"sync"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jfk9w-go/based"
)

func TestGoroutine(t *testing.T) {
	ctx := context.Background()

	t.Run("join", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)
		value := 10
		handle := based.Go(ctx, func(ctx context.Context) {
			wg.Wait()
			value = 20
		})

		assert.Equal(t, 10, value)
		wg.Done()
		err := handle.Join(ctx)
		require.NoError(t, err)
		assert.Equal(t, 20, value)
	})

	t.Run("cancel", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		handle := based.Go(ctx, func(ctx context.Context) {
			wg.Done()
			<-ctx.Done()
		})

		wg.Wait()
		handle.Cancel()

		err := handle.Join(ctx)
		assert.NoError(t, err)
	})

	t.Run("handle panic", func(t *testing.T) {
		handle := based.Go(ctx, func(ctx context.Context) {
			panic(errors.New("test error"))
		})

		err := handle.Join(ctx)
		assert.Error(t, err, "test error")

		handle = based.Go(ctx, func(ctx context.Context) {
			panic("test error")
		})

		err = handle.Join(ctx)
		assert.Error(t, err, "panic: test error")
	})
}
