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

func groupNewAdminsStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "group_new_admins"
		groupName    = "group_new_admin_listeners"
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
			var msgs []eventTypes.GroupMakeUserAdminEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.GroupMakeUserAdminEvent

				msg.GroupId = stmsg.Values["groupId"].(string)
				msg.Admin = stmsg.Values["admin"].(string)
				msg.NewAdmin = stmsg.Values["newAdmin"].(string)
				msg.AdminCHE = helpers.FromMsgPack[appTypes.BinableMap](stmsg.Values["adminCHE"].(string))
				msg.NewAdminCHE = helpers.FromMsgPack[appTypes.BinableMap](stmsg.Values["newAdminCHE"].(string))
				msg.MemInfo = stmsg.Values["memInfo"].(string)

				msgs = append(msgs, msg)

			}

			msgsLen := len(msgs)

			groupNewAdmins := make(map[string][]any, msgsLen)

			newGroupActivityEntries := []string{}

			chatGroupActivities := make(map[string][][2]any)

			// batch data for batch processing
			for i, msg := range msgs {
				groupNewAdmins[msg.GroupId] = append(groupNewAdmins[msg.GroupId], msg.NewAdmin)

				gactche := msg.AdminCHE

				CHEId := gactche["che_id"].(string)
				CHECursor := gactche["cursor"].(int64)

				newGroupActivityEntries = append(newGroupActivityEntries, CHEId, helpers.ToMsgPack(gactche))

				chatGroupActivities[msg.Admin+" "+msg.GroupId] = append(chatGroupActivities[msg.Admin+" "+msg.GroupId], [2]any{CHEId, float64(CHECursor)})

				{
					gactche := msg.NewAdminCHE

					CHEId := gactche["che_id"].(string)
					CHECursor := gactche["cursor"].(int64)

					newGroupActivityEntries = append(newGroupActivityEntries, CHEId, helpers.ToMsgPack(gactche))

					chatGroupActivities[msg.NewAdmin+" "+msg.GroupId] = append(chatGroupActivities[msg.NewAdmin+" "+msg.GroupId], [2]any{CHEId, float64(CHECursor)})
				}

				postActivity, err := groupChat.PostGroupActivityBgDBOper(ctx, msg.GroupId, msg.MemInfo, stmsgIds[i], CHECursor, []any{msg.Admin, msg.NewAdmin})
				if err != nil {
					return
				}

				for _, memUser := range postActivity.MemberUsernames {
					memUser := memUser.(string)

					gactche := postActivity.MemberUsersCHE[memUser].(map[string]any)

					CHEId := gactche["che_id"].(string)
					CHECursor := gactche["cursor"].(int64)

					newGroupActivityEntries = append(newGroupActivityEntries, CHEId, helpers.ToMsgPack(gactche))

					chatGroupActivities[memUser+" "+msg.GroupId] = append(chatGroupActivities[memUser+" "+msg.GroupId], [2]any{CHEId, float64(CHECursor)})
				}
			}

			// batch processing
			eg, sharedCtx := errgroup.WithContext(ctx)

			if err := cache.StoreGroupChatHistoryEntries(ctx, newGroupActivityEntries); err != nil {
				return
			}

			for groupId, newAdmins := range groupNewAdmins {
				eg.Go(func() error {
					groupId, newAdmins := groupId, newAdmins

					return cache.StoreGroupAdmins(sharedCtx, groupId, newAdmins)
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
