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

func groupUsersAddedStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "group_users_added"
		groupName    = "group_user_added_listeners"
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
			var msgs []eventTypes.GroupUsersAddedEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.GroupUsersAddedEvent

				msg.GroupId = stmsg.Values["groupId"].(string)
				msg.Admin = stmsg.Values["admin"].(string)
				msg.NewMembers = helpers.FromJson[appTypes.BinableSlice](stmsg.Values["newMembers"].(string))
				msg.AdminCHE = helpers.FromJson[appTypes.BinableMap](stmsg.Values["adminCHE"].(string))
				msg.NewMembersCHE = helpers.FromJson[appTypes.BinableMap](stmsg.Values["newMembersCHE"].(string))
				msg.MemInfo = stmsg.Values["memInfo"].(string)

				msgs = append(msgs, msg)

			}

			msgsLen := len(msgs)

			groupNewMembers := make(map[string][]any, msgsLen)

			newUserChats := make(map[string][]string)

			userChats := make(map[string]map[string]string)

			newGroupActivityEntries := []string{}

			chatGroupActivities := make(map[string][][2]string)

			// batch data for batch processing
			for i, msg := range msgs {
				groupNewMembers[msg.GroupId] = append(groupNewMembers[msg.GroupId], msg.NewMembers...)

				gactData := msg.AdminCHE

				CHEId := gactData["che_id"].(string)

				newGroupActivityEntries = append(newGroupActivityEntries, CHEId, helpers.ToJson(gactData))

				chatGroupActivities[msg.Admin+" "+msg.GroupId] = append(chatGroupActivities[msg.Admin+" "+msg.GroupId], [2]string{CHEId, stmsgIds[i]})

				for nmemi, newMem := range msg.NewMembers {
					newMem := newMem.(string)

					newUserChats[newMem] = append(newUserChats[newMem], msg.GroupId, helpers.ToJson(map[string]any{"type": "group", "group": msg.GroupId}))

					if userChats[newMem] == nil {
						userChats[newMem] = make(map[string]string)
					}

					userChats[newMem][msg.GroupId] = fmt.Sprintf("%s-%d", stmsgIds[i], nmemi)

					gactData := msg.NewMembersCHE[newMem].(map[string]any)

					CHEId := gactData["che_id"].(string)

					newGroupActivityEntries = append(newGroupActivityEntries, CHEId, helpers.ToJson(gactData))

					chatGroupActivities[newMem+" "+msg.GroupId] = append(chatGroupActivities[newMem+" "+msg.GroupId], [2]string{CHEId, fmt.Sprintf("%s-%d", stmsgIds[i], nmemi)})
				}

				postActivity, err := groupChat.PostGroupActivityBgDBOper(ctx, msg.GroupId, msg.MemInfo, stmsgIds[i], append(msg.NewMembers, msg.Admin))
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

			for groupId, newMembers := range groupNewMembers {
				eg.Go(func() error {
					groupId, newMembers := groupId, newMembers

					return cache.StoreGroupMembers(sharedCtx, groupId, newMembers)
				})
			}

			for ownerUser, groupIdWithChatInfoPairs := range newUserChats {
				eg.Go(func() error {
					ownerUser, groupIdWithChatInfoPairs := ownerUser, groupIdWithChatInfoPairs

					return cache.StoreNewUserChats(sharedCtx, ownerUser, groupIdWithChatInfoPairs)
				})
			}

			for ownerUser, groupId_stmsgId_Pairs := range userChats {
				eg.Go(func() error {
					ownerUser, groupId_stmsgId_Pairs := ownerUser, groupId_stmsgId_Pairs

					return cache.StoreUserChatIdents(sharedCtx, ownerUser, groupId_stmsgId_Pairs)
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
