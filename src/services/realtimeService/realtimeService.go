package realtimeService

import (
	"i9chat/src/appGlobals"
	"sync"

	"github.com/redis/go-redis/v9"
)

func rdb() *redis.Client {
	return appGlobals.RedisClient
}

var AllClientSockets = &sync.Map{}
