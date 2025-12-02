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

func directMsgReactionsRemovedStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "direct_msg_reactions_removed"
		groupName    = "direct_msg_reaction_removed_listeners"
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
			var msgs []eventTypes.DirectMsgReactionRemovedEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.DirectMsgReactionRemovedEvent

				msg.FromUser = stmsg.Values["fromUser"].(string)
				msg.ToUser = stmsg.Values["toUser"].(string)
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

				chatMsgReactionsRemoved[msg.FromUser+" "+msg.ToUser] = append(chatMsgReactionsRemoved[msg.FromUser+" "+msg.ToUser], msg.CHEId)

				msgReactionsRemoved[msg.ToMsgId] = append(msgReactionsRemoved[msg.ToMsgId], msg.FromUser)
			}

			// batch processing
			if err := cache.RemoveDirectChatHistoryEntries(ctx, msgReactionEntriesRemoved); err != nil {
				return
			}

			eg, sharedCtx := errgroup.WithContext(ctx)

			for ownerUserPartnerUser, CHEIds := range chatMsgReactionsRemoved {
				eg.Go(func() error {
					ownerUserPartnerUser, CHEIds := ownerUserPartnerUser, CHEIds

					var ownerUser, partnerUser string

					fmt.Sscanf(ownerUserPartnerUser, "%s %s", &ownerUser, &partnerUser)

					return cache.RemoveDirectChatHistory(sharedCtx, ownerUser, partnerUser, CHEIds)
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
