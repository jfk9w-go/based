package based_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jfk9w-go/based"
)

func TestLazy(t *testing.T) {
	ctx := context.Background()

	t.Run("calculates on demand", func(t *testing.T) {
		var value atomic.Int64
		ref := based.LazyFuncRef[int64](func(ctx context.Context) (int64, error) {
			return value.Load(), nil
		})

		value.Store(10)

		result, err := ref.Get(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(10), result)
	})
}

func TestFuture(t *testing.T) {
	ctx := context.Background()

	t.Run("calculates immediately", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		var value atomic.Int64
		ref := based.FutureFuncRef[int64](ctx, func(ctx context.Context) (int64, error) {
			defer wg.Done()
			return value.Load(), nil
		})

		wg.Wait()
		value.Store(10)

		result, err := ref.Get(context.Background())
		require.NoError(t, err)
		assert.Equal(t, int64(0), result)
	})
}

func TestRefs(t *testing.T) {
	tests := []struct {
		name           string
		fn             func() based.FuncRef[int]
		numberOfCalls  int
		expectedResult int
		expectedError  string
	}{
		{
			name: "calculates only once",
			fn: func() based.FuncRef[int] {
				value := 1
				return func(ctx context.Context) (int, error) {
					defer func() { value++ }()
					return value, nil
				}
			},
			numberOfCalls:  2,
			expectedResult: 1,
		},
		{
			name: "error is not recalculated",
			fn: func() based.FuncRef[int] {
				wasCalled := false
				return func(ctx context.Context) (int, error) {
					defer func() { wasCalled = true }()
					if wasCalled {
						return 1, nil
					}

					return 0, errors.New("calculation error")
				}
			},
			numberOfCalls: 2,
			expectedError: "calculation error",
		},
		{
			name: "panic is recovered",
			fn: func() based.FuncRef[int] {
				value := 1
				return func(ctx context.Context) (int, error) {
					defer func() { value++ }()
					panic(value)
				}
			},
			numberOfCalls: 2,
			expectedError: "panic: 1",
		},
	}

	for _, fn := range []struct {
		name string
		call func(context.Context, based.Ref[int]) based.Ref[int]
	}{
		{
			name: "lazy",
			call: func(ctx context.Context, ref based.Ref[int]) based.Ref[int] { return based.LazyRef(ref) },
		},
		{
			name: "future",
			call: based.FutureRef[int],
		},
	} {
		for _, tt := range tests {
			t.Run(fmt.Sprintf("%v_%s", fn.name, tt.name), func(t *testing.T) {
				ctx := context.Background()
				ref := fn.call(ctx, tt.fn())
				for i := 0; i < tt.numberOfCalls; i++ {
					result, err := ref.Get(ctx)
					if tt.expectedError != "" {
						assert.ErrorContainsf(t, err, tt.expectedError, "attempt #%d", i+1)
					} else {
						assert.Equalf(t, tt.expectedResult, result, "attempt #%d", i+1)
					}
				}
			})
		}
	}
}
