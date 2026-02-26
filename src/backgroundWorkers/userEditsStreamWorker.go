package backgroundWorkers

import (
	"context"
	"i9chat/src/appTypes"
	"i9chat/src/helpers"
	"i9chat/src/services/eventStreamService/eventTypes"
	"log"
	"maps"

	"github.com/redis/go-redis/v9"
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
				msg.UpdateKVMap = helpers.FromJson[appTypes.BinableMap](stmsg.Values["updateKVMap"].(string))

				msgs = append(msgs, msg)

			}

			editUsers := make(map[string]map[string]any, len(msgs))

			// batch data for batch processing
			for _, msg := range msgs {
				editUsers[msg.Username] = map[string]any(msg.UpdateKVMap)
			}

			// batch processing
			user_updateKVMap_StringCmd := make(map[string][2]any)

			_, err = rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
				for user, updateKVMap := range editUsers {
					user_updateKVMap_StringCmd[user] = [2]any{updateKVMap, pipe.HGet(ctx, "users", user)}
				}

				return nil
			})
			if err != nil {
				helpers.LogError(err)
				return
			}

			userUpdates := []string{}

			for user, updateKVMap_StringCmd := range user_updateKVMap_StringCmd {
				updateKVMap, stringCmd := updateKVMap_StringCmd[0].(map[string]any), updateKVMap_StringCmd[1].(*redis.StringCmd)

				userDataMsgPack, err := stringCmd.Result()
				if err != nil {
					helpers.LogError(err)
					continue
				}

				userData := helpers.FromMsgPack[map[string]any](userDataMsgPack)

				maps.Copy(userData, updateKVMap)

				userUpdates = append(userUpdates, user, helpers.ToMsgPack(userData))
			}

			if len(userUpdates) != 0 {
				err = rdb.HSet(ctx, "users", userUpdates).Err()
				if err != nil {
					helpers.LogError(err)
					return
				}
			}

			// acknowledge messages
			if err := rdb.XAck(ctx, streamName, groupName, stmsgIds...).Err(); err != nil {
				helpers.LogError(err)
			}
		}
	}()
}
