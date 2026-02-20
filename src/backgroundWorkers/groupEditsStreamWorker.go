package backgroundWorkers

import (
	"context"
	"fmt"
	"i9chat/src/appTypes"
	"i9chat/src/cache"
	"i9chat/src/helpers"
	groupChat "i9chat/src/models/chatModel/groupChatModel"
	"i9chat/src/services/eventStreamService/eventTypes"
	"log"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

func groupEditsStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "group_edits"
		groupName    = "group_edit_listeners"
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
			var msgs []eventTypes.GroupEditEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.GroupEditEvent

				msg.GroupId = stmsg.Values["groupId"].(string)
				msg.EditorUser = stmsg.Values["editorUser"].(string)
				msg.UpdateKVMap = helpers.FromJson[appTypes.BinableMap](stmsg.Values["updateKVMap"].(string))
				msg.EditorUserCHE = helpers.FromJson[appTypes.BinableMap](stmsg.Values["editorUserCHE"].(string))
				msg.MemInfo = stmsg.Values["memInfo"].(string)

				msgs = append(msgs, msg)

			}

			groupEdits := make(map[string]map[string]any, len(msgs))

			newGroupActivityEntries := []string{}

			chatGroupActivities := make(map[string][][2]any)

			// batch data for batch processing
			for i, msg := range msgs {
				groupEdits[msg.GroupId] = map[string]any(msg.UpdateKVMap)

				gactche := msg.EditorUserCHE

				CHEId := gactche["che_id"].(string)
				CHECursor := gactche["cursor"].(float64)

				newGroupActivityEntries = append(newGroupActivityEntries, CHEId, helpers.ToMsgPack(gactche))

				chatGroupActivities[msg.EditorUser+" "+msg.GroupId] = append(chatGroupActivities[msg.EditorUser+" "+msg.GroupId], [2]any{CHEId, CHECursor})

				postActivity, err := groupChat.PostGroupActivityBgDBOper(ctx, msg.GroupId, msg.MemInfo, stmsgIds[i] /* for uniquness, for idempotency */, CHECursor, []any{msg.EditorUser})
				if err != nil {
					return
				}

				for _, memUser := range postActivity.MemberUsernames {
					memUser := memUser.(string)

					gactche := postActivity.MemberUsersCHE[memUser].(map[string]any)

					CHEId := gactche["che_id"].(string)
					CHECursor := gactche["cursor"].(float64)

					newGroupActivityEntries = append(newGroupActivityEntries, CHEId, helpers.ToMsgPack(gactche))

					chatGroupActivities[memUser+" "+msg.GroupId] = append(chatGroupActivities[memUser+" "+msg.GroupId], [2]any{CHEId, CHECursor})
				}
			}

			// batch processing
			eg, sharedCtx := errgroup.WithContext(ctx)

			if err := cache.StoreGroupChatHistoryEntries(ctx, newGroupActivityEntries); err != nil {
				return
			}

			for groupId, updateKVMap := range groupEdits {

				eg.Go(func() error {
					groupId, updateKVMap := groupId, updateKVMap

					return cache.UpdateGroup(sharedCtx, groupId, updateKVMap)
				})
			}

			for ownerUserGroupId, CHEId_score_Pairs := range chatGroupActivities {
				eg.Go(func() error {
					ownerUserGroupId, CHEId_score_Pairs := ownerUserGroupId, CHEId_score_Pairs

					var ownerUser, groupId string

					fmt.Sscanf(ownerUserGroupId, "%s %s", &ownerUser, &groupId)

					return cache.StoreGroupChatHistory(sharedCtx, ownerUser, groupId, CHEId_score_Pairs)
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
