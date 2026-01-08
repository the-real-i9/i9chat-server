package cache

import (
	"i9chat/src/appGlobals"

	"github.com/redis/go-redis/v9"
)

func rdb() *redis.Client {
	return appGlobals.RedisClient
}
