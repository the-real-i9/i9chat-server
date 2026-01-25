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

func groupRemovedAdminsStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "group_removed_admins"
		groupName    = "group_removed_admin_listeners"
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
			var msgs []eventTypes.GroupRemoveUserFromAdminsEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.GroupRemoveUserFromAdminsEvent

				msg.GroupId = stmsg.Values["groupId"].(string)
				msg.Admin = stmsg.Values["admin"].(string)
				msg.OldAdmin = stmsg.Values["oldAdmin"].(string)
				msg.AdminCHE = helpers.FromJson[appTypes.BinableMap](stmsg.Values["adminCHE"].(string))
				msg.OldAdminCHE = helpers.FromJson[appTypes.BinableMap](stmsg.Values["oldAdminCHE"].(string))
				msg.MemInfo = stmsg.Values["memInfo"].(string)

				msgs = append(msgs, msg)

			}

			msgsLen := len(msgs)

			groupRemovedAdmins := make(map[string][]any, msgsLen)

			newGroupActivityEntries := []string{}

			chatGroupActivities := make(map[string][][2]string)

			// batch data for batch processing
			for i, msg := range msgs {
				groupRemovedAdmins[msg.GroupId] = append(groupRemovedAdmins[msg.GroupId], msg.OldAdmin)

				gactData := msg.AdminCHE

				CHEId := gactData["che_id"].(string)

				newGroupActivityEntries = append(newGroupActivityEntries, CHEId, helpers.ToJson(gactData))

				chatGroupActivities[msg.Admin+" "+msg.GroupId] = append(chatGroupActivities[msg.Admin+" "+msg.GroupId], [2]string{CHEId, stmsgIds[i]})

				{
					gactData := msg.OldAdminCHE

					CHEId := gactData["che_id"].(string)

					newGroupActivityEntries = append(newGroupActivityEntries, CHEId, helpers.ToJson(gactData))

					chatGroupActivities[msg.OldAdmin+" "+msg.GroupId] = append(chatGroupActivities[msg.OldAdmin+" "+msg.GroupId], [2]string{CHEId, stmsgIds[i]})
				}

				postActivity, err := groupChat.PostGroupActivityBgDBOper(ctx, msg.GroupId, msg.MemInfo, stmsgIds[i], []any{msg.Admin, msg.OldAdmin})
				if err != nil {
					return
				}

				for memui, memUser := range postActivity.MemberUsernames {
					memUser := memUser.(string)

					gactData := postActivity.MemberUsersCHE[memUser].(map[string]any)

					CHEId := gactData["che_id"].(string)

					newGroupActivityEntries = append(newGroupActivityEntries, CHEId, helpers.ToJson(gactData))

					chatGroupActivities[memUser+" "+msg.GroupId] = append(chatGroupActivities[memUser+" "+msg.GroupId], [2]string{CHEId, fmt.Sprintf("%s-%d", stmsgIds[i], memui)})
				}
			}

			// batch processing
			eg, sharedCtx := errgroup.WithContext(ctx)

			if err := cache.StoreGroupChatHistoryEntries(ctx, newGroupActivityEntries); err != nil {
				return
			}

			for groupId, remAdmins := range groupRemovedAdmins {
				eg.Go(func() error {
					groupId, remAdmins := groupId, remAdmins

					return cache.RemoveGroupAdmins(sharedCtx, groupId, remAdmins)
				})
			}

			for ownerUserGroupId, CHEId_stmsgId_Pairs := range chatGroupActivities {
				eg.Go(func() error {
					ownerUserGroupId, CHEId_stmsgId_Pairs := ownerUserGroupId, CHEId_stmsgId_Pairs

					var ownerUser, groupId string

					fmt.Sscanf(ownerUserGroupId, "%s %s", &ownerUser, &groupId)

					return cache.StoreGroupChatHistory(sharedCtx, ownerUser, groupId, CHEId_stmsgId_Pairs)
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
