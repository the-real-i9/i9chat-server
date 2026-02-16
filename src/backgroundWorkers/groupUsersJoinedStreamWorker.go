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

func groupUsersJoinedStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "group_users_joined"
		groupName    = "group_user_joined_listeners"
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
			var msgs []eventTypes.GroupUserJoinedEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.GroupUserJoinedEvent

				msg.GroupId = stmsg.Values["groupId"].(string)
				msg.NewMember = stmsg.Values["newMember"].(string)
				msg.NewMemberCHE = helpers.FromMsgPack[appTypes.BinableMap](stmsg.Values["newMemberCHE"].(string))
				msg.MemInfo = stmsg.Values["memInfo"].(string)
				msg.ChatCursor = helpers.FromMsgPack[int64](stmsg.Values["chatCursor"].(string))

				msgs = append(msgs, msg)

			}

			msgsLen := len(msgs)

			groupNewMembers := make(map[string][]any, msgsLen)

			newUserChats := make(map[string][]string)

			userChats := make(map[string]map[string]float64)

			newGroupActivityEntries := []string{}

			chatGroupActivities := make(map[string][][2]any)

			// batch data for batch processing
			for i, msg := range msgs {
				groupNewMembers[msg.GroupId] = append(groupNewMembers[msg.GroupId], msg.NewMember)

				newMem := msg.NewMember

				newUserChats[newMem] = append(newUserChats[newMem], msg.GroupId, helpers.ToMsgPack(map[string]any{"type": "group", "group": msg.GroupId, "cursor": msg.ChatCursor}))

				if userChats[newMem] == nil {
					userChats[newMem] = make(map[string]float64)
				}

				userChats[newMem][msg.GroupId] = float64(msg.ChatCursor)

				gactche := msg.NewMemberCHE

				CHEId := gactche["che_id"].(string)
				CHECursor := gactche["cursor"].(int64)

				newGroupActivityEntries = append(newGroupActivityEntries, CHEId, helpers.ToMsgPack(gactche))

				chatGroupActivities[newMem+" "+msg.GroupId] = append(chatGroupActivities[newMem+" "+msg.GroupId], [2]any{CHEId, float64(CHECursor)})

				postActivity, err := groupChat.PostGroupActivityBgDBOper(ctx, msg.GroupId, msg.MemInfo, stmsgIds[i], CHECursor, []any{msg.NewMember})
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

			for ownerUser, groupId_score_Pairs := range userChats {
				eg.Go(func() error {
					ownerUser, groupId_score_Pairs := ownerUser, groupId_score_Pairs

					return cache.StoreUserChatIdents(sharedCtx, ownerUser, groupId_score_Pairs)
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
