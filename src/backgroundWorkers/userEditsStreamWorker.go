package backgroundWorkers

import (
	"context"
	"i9chat/src/appTypes"
	"i9chat/src/cache"
	"i9chat/src/helpers"
	"i9chat/src/services/eventStreamService/eventTypes"
	"log"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

func userEditsStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "user_edits"
		groupName    = "user_edit_listeners"
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
			var msgs []eventTypes.EditUserEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.EditUserEvent

				msg.Username = stmsg.Values["username"].(string)
				msg.UpdateKVMap = helpers.FromMsgPack[appTypes.BinableMap](stmsg.Values["updateKVMap"].(string))

				msgs = append(msgs, msg)

			}

			userEdits := make(map[string]map[string]any, len(msgs))

			// batch data for batch processing
			for _, msg := range msgs {
				userEdits[msg.Username] = map[string]any(msg.UpdateKVMap)
			}

			// batch processing
			eg, sharedCtx := errgroup.WithContext(ctx)

			for user, updateKVMap := range userEdits {

				eg.Go(func() error {
					user, updateKVMap := user, updateKVMap

					return cache.UpdateUser(sharedCtx, user, updateKVMap)
				})
			}

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
