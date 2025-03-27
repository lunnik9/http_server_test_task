package cache

// DummyCache is a dummy implementation of Cache. used for testing purposes
type DummyCache struct{}

func (d *DummyCache) Set(key string, value string) error {
	return nil
}

func (d *DummyCache) Get(key string) (string, error) {
	return "", ErrorNotFoundCacheItem
}
