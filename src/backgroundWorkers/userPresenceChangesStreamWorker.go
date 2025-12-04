package backgroundWorkers

import (
	"context"
	"i9chat/src/cache"
	"i9chat/src/helpers"
	"i9chat/src/services/eventStreamService/eventTypes"
	"log"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

func userPresenceChangesStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "user_presence_changes"
		groupName    = "user_presence_change_listeners"
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
			var msgs []eventTypes.UserPresenceChangeEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.UserPresenceChangeEvent
				msg.Username = stmsg.Values["username"].(string)
				msg.Presence = stmsg.Values["presence"].(string)
				msg.LastSeen = helpers.FromJson[int64](stmsg.Values["lastSeen"].(string))

				msgs = append(msgs, msg)
			}

			onlineUsers := []any{}
			offlineUsers := make(map[string]int64)

			// batch data for batch processing
			for _, msg := range msgs {
				if msg.Presence == "online" {
					onlineUsers = append(onlineUsers, msg.Username)
				} else {
					offlineUsers[msg.Username] = msg.LastSeen
				}
			}

			// batch processing
			eg, sharedCtx := errgroup.WithContext(ctx)

			eg.Go(func() error {
				return cache.StoreOfflineUsers(sharedCtx, offlineUsers)
			})

			eg.Go(func() error {
				return cache.RemoveOfflineUsers(sharedCtx, onlineUsers)
			})

			if eg.Wait() != nil {
				return
			}

			// acknowledge messages
			if err := rdb.XAck(ctx, streamName, groupName, stmsgIds...).Err(); err != nil {
				helpers.LogError(err)
			}
		}
	}()
}
