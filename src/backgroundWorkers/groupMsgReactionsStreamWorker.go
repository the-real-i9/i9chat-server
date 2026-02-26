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

func groupMsgReactionsStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "group_msg_reactions"
		groupName    = "group_msg_reaction_listeners"
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
			var msgs []eventTypes.NewGroupMsgReactionEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.NewGroupMsgReactionEvent

				msg.FromUser = stmsg.Values["fromUser"].(string)
				msg.ToGroup = stmsg.Values["toGroup"].(string)
				msg.CHEId = stmsg.Values["CHEId"].(string)
				msg.RxnData = stmsg.Values["rxnData"].(string)
				msg.ToMsgId = stmsg.Values["toMsgId"].(string)
				msg.Emoji = stmsg.Values["emoji"].(string)
				msg.CHECursor = helpers.ParseInt(stmsg.Values["cheCursor"].(string))

				msgs = append(msgs, msg)

			}

			newMsgReactionEntries := []string{}

			chatMsgReactions := make(map[string][][2]any)

			msgReactions := make(map[string][]string)

			postDBExtrasFuncs := []func(context.Context) error{}

			// batch data for batch processing
			for _, msg := range msgs {
				newMsgReactionEntries = append(newMsgReactionEntries, msg.CHEId, msg.RxnData)

				chatMsgReactions[msg.FromUser+" "+msg.ToGroup] = append(chatMsgReactions[msg.FromUser+" "+msg.ToGroup], [2]any{msg.CHEId, float64(msg.CHECursor)})

				msgReactions[msg.ToMsgId] = append(msgReactions[msg.ToMsgId], msg.FromUser, msg.Emoji)

				postDBExtrasFuncs = append(postDBExtrasFuncs, func(ctx context.Context) error {
					return groupChat.PostReactToMessage(ctx, msg.FromUser, msg.ToGroup, msg.ToMsgId)
				})
			}

			// batch processing
			if err := cache.StoreGroupChatHistoryEntries(ctx, newMsgReactionEntries); err != nil {
				return
			}

			_, err = rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
				for ownerUserGroupId, CHEId_score_Pairs := range chatMsgReactions {
					var ownerUser, groupId string

					fmt.Sscanf(ownerUserGroupId, "%s %s", &ownerUser, &groupId)

					cache.StoreGroupChatHistory(pipe, ctx, ownerUser, groupId, CHEId_score_Pairs)
				}

				for msgId, userWithEmojiPairs := range msgReactions {
					cache.StoreMsgReactions(pipe, ctx, msgId, userWithEmojiPairs)
				}

				return nil
			})
			if err != nil {
				helpers.LogError(err)
				return
			}

			eg, sharedCtx := errgroup.WithContext(ctx)

			for _, fn := range postDBExtrasFuncs {
				eg.Go(func() error {
					fn := fn

					return fn(sharedCtx)
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
