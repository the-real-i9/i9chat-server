package backgroundWorkers

import (
	"context"
	"fmt"
	"i9chat/src/cache"
	"i9chat/src/helpers"
	groupChat "i9chat/src/models/chatModel/groupChatModel"
	"i9chat/src/services/eventStreamService/eventTypes"
	"log"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

func newGroupMessagesStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "new_group_messages"
		groupName    = "new_group_message_listeners"
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
			var msgs []eventTypes.NewGroupMessageEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.NewGroupMessageEvent

				msg.FromUser = stmsg.Values["fromUser"].(string)
				msg.ToGroup = stmsg.Values["toGroup"].(string)
				msg.CHEId = stmsg.Values["CHEId"].(string)
				msg.MsgData = stmsg.Values["msgData"].(string)
				msg.CHECursor = helpers.FromMsgPack[int64](stmsg.Values["cheCursor"].(string))

				msgs = append(msgs, msg)
			}

			newMessageEntries := []string{}

			updatedUserChats := make(map[string]map[string]float64)

			chatMessages := make(map[string][][2]any)

			// batch data for batch processing
			for _, msg := range msgs {
				newMessageEntries = append(newMessageEntries, msg.CHEId, msg.MsgData)

				if updatedUserChats[msg.FromUser] == nil {
					updatedUserChats[msg.FromUser] = make(map[string]float64)
				}

				updatedUserChats[msg.FromUser][msg.ToGroup] = float64(msg.CHECursor)

				chatMessages[msg.FromUser+" "+msg.ToGroup] = append(chatMessages[msg.FromUser+" "+msg.ToGroup], [2]any{msg.CHEId, float64(msg.CHECursor)})

				postNewMessage, err := groupChat.PostSendMessage(ctx, msg.FromUser, msg.ToGroup, msg.CHEId)
				if err != nil {
					return
				}

				for _, memUser := range postNewMessage.MemberUsernames {
					memUser := memUser.(string)

					chatMessages[memUser+" "+msg.ToGroup] = append(chatMessages[memUser+" "+msg.ToGroup], [2]any{msg.CHEId, float64(msg.CHECursor)})
				}
			}

			// batch processing
			if err := cache.StoreGroupChatHistoryEntries(ctx, newMessageEntries); err != nil {
				return
			}

			eg, sharedCtx := errgroup.WithContext(ctx)

			for ownerUser, groupId_score_Pairs := range updatedUserChats {
				eg.Go(func() error {
					ownerUser, groupId_score_Pairs := ownerUser, groupId_score_Pairs

					return cache.StoreUserChatIdents(sharedCtx, ownerUser, groupId_score_Pairs)
				})
			}

			for ownerUserGroupId, CHEId_score_Pairs := range chatMessages {
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
