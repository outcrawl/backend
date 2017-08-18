package backend

import (
	"encoding/json"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/memcache"
)

func cacheItemJSON(ctx context.Context, key string, data []byte) {
	item := &memcache.Item{
		Key:        key,
		Value:      data,
		Expiration: 30 * time.Minute,
	}
	memcache.Add(ctx, item)
}

func cacheItem(ctx context.Context, key string, item interface{}) {
	if data, err := json.Marshal(item); err == nil {
		cacheItemJSON(ctx, key, data)
	}
}

func getCachedItem(ctx context.Context, key string, v interface{}) error {
	if item, err := memcache.Get(ctx, key); err != nil {
		return err
	} else {
		return json.Unmarshal(item.Value, &v)
	}
}

func clearCachedItem(ctx context.Context, key string) error {
	return memcache.Delete(ctx, key)
}
