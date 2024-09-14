package utils

import (
	"strings"

	"github.com/coocood/freecache"
)

const (
	CACHESIZE                = 1 * 1024 * 1024 // Note value size larger than 1/1024 of total cache size will not be cached
	EXPIRE_GRPC_DNS          = 60 * 60 * 2
	EXPIRE_GRPC_SMARTSENCE   = 60 * 60 * 2
	EXPIRE_LINKER_LEASE      = 60 * 5
	EXPIRE_LOG_TIMEGAP       = 5 // Interval for logging output, using seconds as the unit.
	EXPIRE_LINKER_PEER       = 60
	EXPIRE_PEER_POLICY_IP    = 10
	EXPIRE_USER_FORBIDDEN_IP = 10
	EXPIRE                   = 0
)

var cacheStore *freecache.Cache

func init() {
	cacheStore = freecache.NewCache(CACHESIZE)
}

func FormatCacheKey(strs ...string) string {
	if len(strs) == 0 {
		return ""
	}

	return strings.Join(strs, "/")
}

func CacheWriteValue(key, value string, timeout int) error {
	err := cacheStore.Set([]byte(key), []byte(value), timeout)
	return err
}

func CacheReadValue(key string) string {
	res, err := cacheStore.Get([]byte(key))
	if err != nil || len(res) == 0 {
		return ""
	}

	return string(res)
}

func CacheDeleteValue(key string) bool {
	ok := cacheStore.Del([]byte(key))
	return ok
}
