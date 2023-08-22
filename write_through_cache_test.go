package based_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jfk9w-go/based"
)

func TestWriteThroughCache(t *testing.T) {
	values := map[string]string{
		"call2": "value2",
	}

	storage := &based.WriteThroughCacheStorageFunc[string, string]{
		LoadFn: func(ctx context.Context, key string) (string, error) {
			return values[key], nil
		},
		UpdateFn: func(ctx context.Context, key string, value string) error {
			values[key] = value
			return nil
		},
	}

	ctx := context.Background()
	cache := based.NewWriteThroughCache[string, string](storage)

	value, err := cache.Get(ctx, "call1")
	require.NoError(t, err)
	assert.Equal(t, "", value)

	err = cache.Update(ctx, "call1", "value1")
	require.NoError(t, err)

	value, err = cache.Get(ctx, "call1")
	require.NoError(t, err)
	assert.Equal(t, "value1", value)

	value, err = cache.Get(ctx, "call2")
	require.NoError(t, err)
	assert.Equal(t, "value2", value)

	err = cache.Update(ctx, "call2", "value3")
	require.NoError(t, err)

	value, err = cache.Get(ctx, "call2")
	require.NoError(t, err)
	assert.Equal(t, "value3", value)

	assert.Equal(t, map[string]string{
		"call1": "value1",
		"call2": "value3",
	}, values)
}

func TestWriteThroughCached(t *testing.T) {
	values := map[string]string{}
	storage := &based.WriteThroughCacheStorageFunc[string, string]{
		LoadFn: func(ctx context.Context, key string) (string, error) {
			return values[key], nil
		},
		UpdateFn: func(ctx context.Context, key string, value string) error {
			values[key] = value
			return nil
		},
	}

	ctx := context.Background()
	cache := based.NewWriteThroughCached[string, string](storage, "call")

	value, err := cache.Get(ctx)
	require.NoError(t, err)
	assert.Equal(t, "", value)

	err = cache.Update(ctx, "value")
	require.NoError(t, err)

	value, err = cache.Get(ctx)
	require.NoError(t, err)
	assert.Equal(t, "value", value)

	assert.Equal(t, map[string]string{
		"call": "value",
	}, values)
}
