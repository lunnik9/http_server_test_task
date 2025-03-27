package cache

import "errors"

var (
	ErrorSettingCacheItem  = errors.New("failed to set cache item")
	ErrorNotFoundCacheItem = errors.New("cache item not found")
	ErrorCacheUnavailable  = errors.New("cache is unavailable")
)
