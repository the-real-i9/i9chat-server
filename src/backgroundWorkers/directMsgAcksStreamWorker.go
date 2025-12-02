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

func directMsgAcksStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "direct_msg_acks"
		groupName    = "direct_msg_ack_listeners"
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
			var msgs []eventTypes.DirectMsgAckEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.DirectMsgAckEvent

				msg.FromUser = stmsg.Values["fromUser"].(string)
				msg.ToUser = stmsg.Values["toUser"].(string)
				msg.CHEId = stmsg.Values["CHEId"].(string)
				msg.Ack = stmsg.Values["ack"].(string)
				msg.At = helpers.FromJson[int64](stmsg.Values["at"].(string))

				msgs = append(msgs, msg)

			}

			ackMessages := [][3]any{}

			userChatUnreadMsgs := make(map[string]map[string][]any)
			userChatReadMsgs := make(map[string]map[string][]any)

			updatedFromUserChats := make(map[string]map[string]string)

			// batch data for batch processing
			for i, msg := range msgs {

				ackMessages = append(ackMessages, [3]any{msg.CHEId, msg.Ack, msg.At})

				if msg.Ack == "delivered" {
					if updatedFromUserChats[msg.FromUser] == nil {
						updatedFromUserChats[msg.FromUser] = make(map[string]string)
					}

					updatedFromUserChats[msg.FromUser][msg.ToUser] = stmsgIds[i]

					if userChatUnreadMsgs[msg.FromUser] == nil {
						userChatUnreadMsgs[msg.FromUser] = make(map[string][]any)
					}

					userChatUnreadMsgs[msg.FromUser][msg.ToUser] = append(userChatUnreadMsgs[msg.FromUser][msg.ToUser], msg.CHEId)
				}

				if msg.Ack == "read" {
					if userChatReadMsgs[msg.FromUser] == nil {
						userChatReadMsgs[msg.FromUser] = make(map[string][]any)
					}

					userChatReadMsgs[msg.FromUser][msg.ToUser] = append(userChatReadMsgs[msg.FromUser][msg.ToUser], msg.CHEId)
				}
			}

			// batch processing
			eg, sharedCtx := errgroup.WithContext(ctx)

			for _, CHEId_ack_ackAt_stmsgId := range ackMessages {

				eg.Go(func() error {
					CHEId, ack, ackAt := CHEId_ack_ackAt_stmsgId[0], CHEId_ack_ackAt_stmsgId[1], CHEId_ack_ackAt_stmsgId[2]

					return cache.UpdateDirectMessage(sharedCtx, CHEId.(string), map[string]any{
						"delivery_status":         ack,
						fmt.Sprintf("%s_at", ack): ackAt.(int64),
					})
				})
			}

			for ownerUser, partnerUser_stmsgId_Pairs := range updatedFromUserChats {
				eg.Go(func() error {
					ownerUser, partnerUser_stmsgId_Pairs := ownerUser, partnerUser_stmsgId_Pairs

					return cache.StoreUserChatsSorted(sharedCtx, ownerUser, partnerUser_stmsgId_Pairs)
				})
			}

			for ownerUser, partnerUser_unreadMsgs_Map := range userChatUnreadMsgs {
				eg.Go(func() error {
					ownerUser, partnerUser_unreadMsgs_Map := ownerUser, partnerUser_unreadMsgs_Map

					for partnerUser, unreadMsgs := range partnerUser_unreadMsgs_Map {
						if err := cache.StoreUserChatUnreadMsgs(sharedCtx, ownerUser, partnerUser, unreadMsgs); err != nil {
							return err
						}
					}

					return nil
				})
			}

			for ownerUser, partnerUser_readMsgs_Map := range userChatReadMsgs {
				eg.Go(func() error {
					ownerUser, partnerUser_readMsgs_Map := ownerUser, partnerUser_readMsgs_Map

					for partnerUser, readMsgs := range partnerUser_readMsgs_Map {
						if err := cache.RemoveUserChatUnreadMsgs(sharedCtx, ownerUser, partnerUser, readMsgs); err != nil {
							return err
						}
					}

					return nil
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
