package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCacheSetGet(t *testing.T) {
	c := NewCache(context.Background(), 5)

	err := c.Set("salam", "popolam")
	assert.Nil(t, err)

	res, err := c.Get("salam")
	assert.Nil(t, err)
	assert.Equal(t, "popolam", res)
}

func TestCacheTTL(t *testing.T) {
	c := NewCache(context.Background(), 1)

	err := c.Set("salam", "popolam")
	assert.Nil(t, err)

	time.Sleep(2 * time.Second)

	res, err := c.Get("salam")
	assert.ErrorIs(t, err, ErrorNotFoundCacheItem)
	assert.Equal(t, "", res)
}

func TestCacheInaccessible(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)

	c := NewCache(ctx, 10)

	time.Sleep(2 * time.Second)

	err := c.Set("salam", "popolam")
	assert.ErrorIs(t, err, ErrorCacheUnavailable)

	res, err := c.Get("salam")
	assert.ErrorIs(t, err, ErrorCacheUnavailable)
	assert.Equal(t, res, "")
}
