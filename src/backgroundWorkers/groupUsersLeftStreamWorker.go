package backgroundWorkers

import (
	"context"
	"fmt"
	"i9chat/src/appTypes"
	"i9chat/src/cache"
	"i9chat/src/helpers"
	"i9chat/src/services/eventStreamService/eventTypes"
	"log"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

func groupUsersLeftStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "group_users_left"
		groupName    = "group_user_left_listeners"
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
			var msgs []eventTypes.GroupUserLeftEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.GroupUserLeftEvent

				msg.GroupId = stmsg.Values["groupId"].(string)
				msg.OldMember = stmsg.Values["oldMember"].(string)
				msg.MemberUsers = helpers.FromJson[appTypes.BinableSlice](stmsg.Values["memberUsers"].(string))
				msg.OldMemberCHE = helpers.FromJson[appTypes.BinableMap](stmsg.Values["oldMemberCHE"].(string))
				msg.MemberUsersCHE = helpers.FromJson[appTypes.BinableMap](stmsg.Values["memberUsersCHE"].(string))

				msgs = append(msgs, msg)

			}

			msgsLen := len(msgs)

			groupOldMembers := make(map[string][]any, msgsLen)

			newGroupActivityEntries := []string{}

			chatGroupActivities := make(map[string][][2]string)

			// batch data for batch processing
			for i, msg := range msgs {
				groupOldMembers[msg.GroupId] = append(groupOldMembers[msg.GroupId], msg.OldMember)

				gactData := msg.OldMemberCHE

				CHEId := gactData["che_id"].(string)

				newGroupActivityEntries = append(newGroupActivityEntries, CHEId, helpers.ToJson(gactData))

				chatGroupActivities[msg.OldMember+" "+msg.GroupId] = append(chatGroupActivities[msg.OldMember+" "+msg.GroupId], [2]string{CHEId, stmsgIds[i]})

				for memui, memUser := range msg.MemberUsers {
					memUser := memUser.(string)

					gactData := msg.MemberUsersCHE[memUser].(map[string]any)

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

			for groupId, oldMembers := range groupOldMembers {
				eg.Go(func() error {
					groupId, oldMembers := groupId, oldMembers

					return cache.RemoveGroupMembers(sharedCtx, groupId, oldMembers)
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
