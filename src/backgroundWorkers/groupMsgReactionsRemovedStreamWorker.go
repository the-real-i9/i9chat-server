package backgroundWorkers

import (
	"context"
	"fmt"
	"i9chat/src/cache"
	"i9chat/src/helpers"
	"i9chat/src/services/eventStreamService/eventTypes"
	"log"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

func groupMsgReactionsRemovedStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "group_msg_reactions_removed"
		groupName    = "group_msg_reaction_removed_listeners"
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
			var msgs []eventTypes.GroupMsgReactionRemovedEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.GroupMsgReactionRemovedEvent

				msg.FromUser = stmsg.Values["fromUser"].(string)
				msg.ToGroup = stmsg.Values["toGroup"].(string)
				msg.ToMsgId = stmsg.Values["toMsgId"].(string)
				msg.CHEId = stmsg.Values["CHEId"].(string)

				msgs = append(msgs, msg)

			}

			msgReactionEntriesRemoved := []string{}

			chatMsgReactionsRemoved := make(map[string][]any)

			msgReactionsRemoved := make(map[string][]string)

			// batch data for batch processing
			for _, msg := range msgs {
				msgReactionEntriesRemoved = append(msgReactionEntriesRemoved, msg.CHEId)

				chatMsgReactionsRemoved[msg.FromUser+" "+msg.ToGroup] = append(chatMsgReactionsRemoved[msg.FromUser+" "+msg.ToGroup], msg.CHEId)

				msgReactionsRemoved[msg.ToMsgId] = append(msgReactionsRemoved[msg.ToMsgId], msg.FromUser)
			}

			// batch processing
			if err := cache.RemoveGroupChatHistoryEntries(ctx, msgReactionEntriesRemoved); err != nil {
				return
			}

			eg, sharedCtx := errgroup.WithContext(ctx)

			for ownerUserGroupId, CHEIds := range chatMsgReactionsRemoved {
				eg.Go(func() error {
					ownerUserGroupId, CHEIds := ownerUserGroupId, CHEIds

					var ownerUser, groupId string

					fmt.Sscanf(ownerUserGroupId, "%s %s", &ownerUser, &groupId)

					return cache.RemoveGroupChatHistory(sharedCtx, ownerUser, groupId, CHEIds)
				})
			}

			for msgId, reactorUsers := range msgReactionsRemoved {
				eg.Go(func() error {
					msgId, reactorUsers := msgId, reactorUsers

					return cache.RemoveMsgReactions(sharedCtx, msgId, reactorUsers)
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
