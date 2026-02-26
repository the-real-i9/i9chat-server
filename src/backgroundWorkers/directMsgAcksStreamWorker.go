package backgroundWorkers

import (
	"context"
	"fmt"
	"i9chat/src/appTypes"
	"i9chat/src/cache"
	"i9chat/src/helpers"
	"i9chat/src/services/eventStreamService/eventTypes"
	"log"
	"maps"

	"github.com/redis/go-redis/v9"
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
				msg.CHEIds = helpers.FromJson[appTypes.BinableSlice](stmsg.Values["CHEIds"].(string))
				msg.Ack = stmsg.Values["ack"].(string)
				msg.At = helpers.ParseInt(stmsg.Values["at"].(string))
				msg.ChatCursor = helpers.ParseInt(stmsg.Values["chatCursor"].(string))

				msgs = append(msgs, msg)

			}

			ackMessages := [][3]any{}

			userChatUnreadMsgs := make(map[string]map[string][]any)
			userChatReadMsgs := make(map[string]map[string][]any)

			updatedFromUserChats := make(map[string]map[string]float64)

			// batch data for batch processing
			for _, msg := range msgs {

				for _, CHEId := range msg.CHEIds {
					ackMessages = append(ackMessages, [3]any{CHEId, msg.Ack, msg.At})
				}

				if msg.Ack == "delivered" {
					if updatedFromUserChats[msg.FromUser] == nil {
						updatedFromUserChats[msg.FromUser] = make(map[string]float64)
					}

					updatedFromUserChats[msg.FromUser][msg.ToUser] = float64(msg.ChatCursor)

					if userChatUnreadMsgs[msg.FromUser] == nil {
						userChatUnreadMsgs[msg.FromUser] = make(map[string][]any)
					}

					for _, CHEId := range msg.CHEIds {
						userChatUnreadMsgs[msg.FromUser][msg.ToUser] = append(userChatUnreadMsgs[msg.FromUser][msg.ToUser], CHEId)
					}
				}

				if msg.Ack == "read" {
					if userChatReadMsgs[msg.FromUser] == nil {
						userChatReadMsgs[msg.FromUser] = make(map[string][]any)
					}

					for _, CHEId := range msg.CHEIds {
						userChatReadMsgs[msg.FromUser][msg.ToUser] = append(userChatReadMsgs[msg.FromUser][msg.ToUser], CHEId)
					}
				}
			}

			// batch processing
			msgId_updateKVMap_StringCmd := make(map[string][2]any)

			_, err = rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {

				for _, msgId_ack_ackAt := range ackMessages {
					msgId, ack, ackAt := msgId_ack_ackAt[0].(string), msgId_ack_ackAt[1], msgId_ack_ackAt[2]

					msgId_updateKVMap_StringCmd[msgId] = [2]any{map[string]any{
						"delivery_status":         ack,
						fmt.Sprintf("%s_at", ack): ackAt,
					}, pipe.HGet(ctx, "direct_chat_history_entries", msgId)}
				}

				for ownerUser, partnerUser_score_Pairs := range updatedFromUserChats {
					cache.StoreUserChatIdents(pipe, ctx, ownerUser, partnerUser_score_Pairs)
				}

				for ownerUser, partnerUser_unreadMsgs_Map := range userChatUnreadMsgs {
					for partnerUser, unreadMsgs := range partnerUser_unreadMsgs_Map {
						cache.StoreUserChatUnreadMsgs(pipe, ctx, ownerUser, partnerUser, unreadMsgs)
					}
				}

				for ownerUser, partnerUser_readMsgs_Map := range userChatReadMsgs {
					for partnerUser, readMsgs := range partnerUser_readMsgs_Map {
						cache.RemoveUserChatUnreadMsgs(pipe, ctx, ownerUser, partnerUser, readMsgs)
					}
				}

				return nil
			})
			if err != nil {
				helpers.LogError(err)
				return
			}

			msgUpdates := []string{}

			for msgId, updateKVMap_StringCmd := range msgId_updateKVMap_StringCmd {
				updateKVMap, stringCmd := updateKVMap_StringCmd[0].(map[string]any), updateKVMap_StringCmd[1].(*redis.StringCmd)

				msgDataMsgPack, err := stringCmd.Result()
				if err != nil {
					helpers.LogError(err)
					continue
				}

				msgData := helpers.FromMsgPack[map[string]any](msgDataMsgPack)

				// if a client skips the "delivered" ack, and acks "read"
				// it means the message is delivered and read at the same time
				if updateKVMap["read_at"] != nil && msgData["delivered_at"] == nil {
					msgData["delivered_at"] = updateKVMap["read_at"]
				}

				maps.Copy(msgData, updateKVMap)

				msgUpdates = append(msgUpdates, msgId, helpers.ToMsgPack(msgData))
			}

			if len(msgUpdates) != 0 {
				err = rdb.HSet(ctx, "direct_chat_history_entries", msgUpdates).Err()
				if err != nil {
					helpers.LogError(err)
					return
				}
			}

			// acknowledge messages
			if err := rdb.XAck(ctx, streamName, groupName, stmsgIds...).Err(); err != nil {
				helpers.LogError(err)
			}
		}
	}()
}
