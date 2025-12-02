package backgroundWorkers

import (
	"context"
	"i9chat/src/cache"
	"i9chat/src/helpers"
	"i9chat/src/services/eventStreamService/eventTypes"
	"log"

	"github.com/redis/go-redis/v9"
)

func newUsersStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "new_users"
		groupName    = "new_user_listeners"
		consumerName = "worker-1"
	)

	ctx := context.Background()

	err := rdb.XGroupCreateMkStream(ctx, streamName, groupName, "$").Err()
	if err != nil && (err.Error() != "BUSYGROUP Consumer Group name already exists") {
		helpers.LogError(err)
		log.Fatal()
	}

	go func() {
		for {
			streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    groupName,
				Consumer: consumerName,
				Streams:  []string{streamName, ">"},
				Count:    500,
				Block:    0,
			}).Result()

			if err != nil {
				helpers.LogError(err)
				continue
			}

			var stmsgIds []string
			var msgs []eventTypes.NewUserEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.NewUserEvent
				msg.Username = stmsg.Values["username"].(string)
				msg.UserData = stmsg.Values["userData"].(string)

				msgs = append(msgs, msg)
			}

			newUsers := []string{}

			// batch data for batch processing
			for _, msg := range msgs {
				newUsers = append(newUsers, msg.Username, msg.UserData)

			}

			// batch processing
			if err := cache.StoreNewUsers(ctx, newUsers); err != nil {
				return
			}

			// acknowledge messages
			if err := rdb.XAck(ctx, streamName, groupName, stmsgIds...).Err(); err != nil {
				helpers.LogError(err)
			}
		}
	}()
}
