// Copyright (c) 2021 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package pkg

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/allegro/bigcache"
	"github.com/eko/gocache/cache"
	"github.com/eko/gocache/metrics"
	"github.com/eko/gocache/store"
	"github.com/go-redis/redis/v7"
	"golang.org/x/xerrors"
)

var (
	cacheManager *cache.MetricCache
)

type CacheObject struct {
	Header     http.Header
	Body       []byte
	StatusCode int
}

func (cacheObj *CacheObject) ToJson() ([]byte, error) {
	b, err := json.Marshal(cacheObj)
	if err != nil {
		return nil, xerrors.Errorf("cannot convert cache object to JSON: %v", err)
	}
	return b, nil
}

func (cacheObj *CacheObject) FromJson(v []byte) error {
	err := json.Unmarshal(v, cacheObj)
	if err != nil {
		return xerrors.Errorf("cannot convert cache object from JSON: %v", err)
	}
	return nil
}

func SetupCache() error {
	var cacheStore store.StoreInterface

	if cfg.RedisEnabled() {
		redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
		cacheStore = store.NewRedis(redisClient, nil)
	} else {
		bigcacheClient, err := bigcache.NewBigCache(bigcache.DefaultConfig(time.Duration(cfg.CacheDurationBackup)))
		if err != nil {
			return err
		}
		cacheStore = store.NewBigcache(bigcacheClient, nil)
	}

	// TODO(clu): find out how to config metrics or remove metrics
	m := metrics.NewPrometheus("openvsx-proxy")
	cacheManager = cache.NewMetric(m, cache.New(cacheStore))

	return nil
}

func StoreCache(key string, obj *CacheObject) error {
	b, err := obj.ToJson()
	if err != nil {
		return err
	}
	if cfg.RedisEnabled() {
		return cacheManager.Set(key, string(b), nil)
	}
	return cacheManager.Set(key, b, nil)
}

func ReadCache(key string) (obj *CacheObject, ok bool, err error) {
	b, err := cacheManager.Get(key)
	if err == bigcache.ErrEntryNotFound || err == redis.Nil {
		return nil, false, nil
	} else if err != nil || b == nil {
		return
	}
	var bb []byte
	if cfg.RedisEnabled() {
		bb = []byte(b.(string))
	} else {
		bb = b.([]byte)
	}
	obj = &CacheObject{}
	if err = obj.FromJson(bb); err != nil {
		return
	}
	ok = true
	return
}
