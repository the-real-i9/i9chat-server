package backgroundWorkers

import (
	"context"
	"fmt"
	"i9chat/src/appTypes"
	"i9chat/src/cache"
	"i9chat/src/helpers"
	"i9chat/src/models/db"
	"i9chat/src/services/eventStreamService/eventTypes"
	"i9chat/src/services/realtimeService"
	"log"

	"github.com/redis/go-redis/v9"
)

func groupMsgAcksStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "group_msg_acks"
		groupName    = "group_msg_ack_listeners"
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
			var msgs []eventTypes.GroupMsgAckEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.GroupMsgAckEvent

				msg.FromUser = stmsg.Values["fromUser"].(string)
				msg.ToGroup = stmsg.Values["toGroup"].(string)
				msg.CHEIds = helpers.FromJson[appTypes.BinableSlice](stmsg.Values["CHEIds"].(string))
				msg.Ack = stmsg.Values["ack"].(string)
				msg.At = helpers.ParseInt(stmsg.Values["at"].(string))
				msg.ChatCursor = helpers.ParseInt(stmsg.Values["chatCursor"].(string))
				msg.MsgIdtoSender = helpers.FromJson[appTypes.BinableSlice](stmsg.Values["msgIdtoSender"].(string))

				msgs = append(msgs, msg)

			}

			groupMsgDelvToUsers := make(map[string]map[string][][2]any)
			groupMsgReadByUsers := make(map[string]map[string][][2]any)

			userChatUnreadMsgs := make(map[string]map[string][]any)
			userChatReadMsgs := make(map[string]map[string][]any)

			updatedUserChats := make(map[string]map[string]float64)

			delv_groupMsgtoSender := [][2]any{}
			read_groupMsgtoSender := [][2]any{}

			// batch data for batch processing
			for _, msg := range msgs {
				if msg.Ack == "delivered" {
					if updatedUserChats[msg.FromUser] == nil {
						updatedUserChats[msg.FromUser] = make(map[string]float64)
					}

					updatedUserChats[msg.FromUser][msg.ToGroup] = float64(msg.ChatCursor)

					if userChatUnreadMsgs[msg.FromUser] == nil {
						userChatUnreadMsgs[msg.FromUser] = make(map[string][]any)
					}

					for _, CHEId := range msg.CHEIds {
						userChatUnreadMsgs[msg.FromUser][msg.ToGroup] = append(userChatUnreadMsgs[msg.FromUser][msg.ToGroup], CHEId)
					}

					if groupMsgDelvToUsers[msg.ToGroup] == nil {
						groupMsgDelvToUsers[msg.ToGroup] = make(map[string][][2]any)
					}

					for _, CHEId := range msg.CHEIds {
						CHEId := CHEId.(string)
						groupMsgDelvToUsers[msg.ToGroup][CHEId] = append(groupMsgDelvToUsers[msg.ToGroup][CHEId], [2]any{msg.FromUser, msg.At})
					}

					delv_groupMsgtoSender = append(delv_groupMsgtoSender, [2]any{msg.ToGroup, msg.MsgIdtoSender})
				}

				if msg.Ack == "read" {
					if userChatReadMsgs[msg.FromUser] == nil {
						userChatReadMsgs[msg.FromUser] = make(map[string][]any)
					}

					for _, CHEId := range msg.CHEIds {
						userChatReadMsgs[msg.FromUser][msg.ToGroup] = append(userChatReadMsgs[msg.FromUser][msg.ToGroup], CHEId)
					}

					if groupMsgReadByUsers[msg.ToGroup] == nil {
						groupMsgReadByUsers[msg.ToGroup] = make(map[string][][2]any)
					}

					for _, CHEId := range msg.CHEIds {
						CHEId := CHEId.(string)
						groupMsgReadByUsers[msg.ToGroup][CHEId] = append(groupMsgReadByUsers[msg.ToGroup][CHEId], [2]any{msg.FromUser, msg.At})
					}

					read_groupMsgtoSender = append(read_groupMsgtoSender, [2]any{msg.ToGroup, msg.MsgIdtoSender})
				}
			}

			// batch processing
			_, err = rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
				for ownerUser, groupId_score_Pairs := range updatedUserChats {
					cache.StoreUserChatIdents(pipe, ctx, ownerUser, groupId_score_Pairs)
				}

				for ownerUser, groupId_unreadMsgs_Map := range userChatUnreadMsgs {
					for groupId, unreadMsgs := range groupId_unreadMsgs_Map {
						cache.StoreUserChatUnreadMsgs(pipe, ctx, ownerUser, groupId, unreadMsgs)
					}
				}

				for ownerUser, groupId_readMsgs_Map := range userChatReadMsgs {
					for groupId, readMsgs := range groupId_readMsgs_Map {
						cache.RemoveUserChatUnreadMsgs(pipe, ctx, ownerUser, groupId, readMsgs)
					}
				}

				for groupId, msgId_userDelvAtPairs_Map := range groupMsgDelvToUsers {
					for msgId, user_delvAt_Pairs := range msgId_userDelvAtPairs_Map {
						cache.StoreGroupMsgDeliveredToUsers(pipe, ctx, groupId, msgId, user_delvAt_Pairs)
					}
				}

				for groupId, msgId_userReadAtPairs_Map := range groupMsgReadByUsers {
					for msgId, user_readAt_Pairs := range msgId_userReadAtPairs_Map {
						cache.StoreGroupMsgReadByUsers(pipe, ctx, groupId, msgId, user_readAt_Pairs)
					}
				}

				return nil
			})
			if err != nil {
				helpers.LogError(err)
				return
			}

			for _, groupId_msgIdtoSender := range delv_groupMsgtoSender {
				groupId, msgIdtoSender := groupId_msgIdtoSender[0].(string), groupId_msgIdtoSender[1].(appTypes.BinableSlice)
				for _, msgId_Sender := range msgIdtoSender {
					msgId_Sender := msgId_Sender.([]any)
					msgId, senderUser := msgId_Sender[0].(string), msgId_Sender[1].(string)

					go func(groupId, msgId, senderUser string) {
						ctx := context.Background()

						var membersCountIntCmd, delvToUsersCountIntCmd *redis.IntCmd

						_, err := rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
							membersCountIntCmd = pipe.SCard(ctx, fmt.Sprintf("group:%s:members", groupId))
							delvToUsersCountIntCmd = pipe.ZCard(ctx, fmt.Sprintf("group:%s:msg:%s:delivered_to_users", groupId, msgId))

							return nil
						})
						if err != nil {
							helpers.LogError(err)
							return
						}

						membersCount, delvToUsersCount := membersCountIntCmd.Val(), delvToUsersCountIntCmd.Val()

						// delvToUsers cannot include the message sender,
						// therefore, we exempt them from the membersList count
						if membersCount-1 == delvToUsersCount {
							go realtimeService.SendEventMsg(senderUser, appTypes.ServerEventMsg{
								Event: "group chat: message delivered",
								Data: map[string]string{
									"group_id": groupId,
									"msg_id":   msgId,
								},
							})

							go cache.UpdateGroupMessageDelivery(ctx, msgId, map[string]any{"delivery_status": "delivered"})

							go func() {
								_, err := db.Query(
									ctx,
									`/* cypher */
											MATCH (gm:GroupMessage{ id: $msg_id })
											SET gm.delivery_status = "delivered"
											`,
									map[string]any{
										"msg_id": msgId,
									},
								)
								if err != nil {
									helpers.LogError(err)
								}
							}()
						}
					}(groupId, msgId, senderUser)
				}
			}

			for _, groupId_msgIdtoSender := range read_groupMsgtoSender {
				groupId, msgIdtoSender := groupId_msgIdtoSender[0].(string), groupId_msgIdtoSender[1].(appTypes.BinableSlice)
				for _, msgId_Sender := range msgIdtoSender {
					msgId_Sender := msgId_Sender.([]any)
					msgId, senderUser := msgId_Sender[0].(string), msgId_Sender[1].(string)
					go func(groupId, msgId, senderUser string) {
						ctx := context.Background()

						var membersCountIntCmd, readByUsersCountIntCmd *redis.IntCmd

						_, err := rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
							membersCountIntCmd = pipe.SCard(ctx, fmt.Sprintf("group:%s:members", groupId))
							readByUsersCountIntCmd = pipe.ZCard(ctx, fmt.Sprintf("group:%s:msg:%s:read_by_users", groupId, msgId))

							return nil
						})
						if err != nil {
							helpers.LogError(err)
							return
						}

						membersCount, readByUsersCount := membersCountIntCmd.Val(), readByUsersCountIntCmd.Val()

						// readByUsers cannot include the message sender,
						// therefore, we exempt them from the membersList count
						if membersCount-1 == readByUsersCount {
							go realtimeService.SendEventMsg(senderUser, appTypes.ServerEventMsg{
								Event: "group chat: message read",
								Data: map[string]string{
									"group_id": groupId,
									"msg_id":   msgId,
								},
							})

							go cache.UpdateGroupMessageDelivery(ctx, msgId, map[string]any{"delivery_status": "read"})

							go func() {
								_, err := db.Query(
									ctx,
									`/* cypher */
											MATCH (gm:GroupMessage{ id: $msg_id })
											SET gm.delivery_status = "read"
											`,
									map[string]any{
										"msg_id": msgId,
									},
								)
								if err != nil {
									helpers.LogError(err)
								}
							}()
						}
					}(groupId, msgId, senderUser)
				}
			}

			// acknowledge messages
			if err := rdb.XAck(ctx, streamName, groupName, stmsgIds...).Err(); err != nil {
				helpers.LogError(err)
			}
		}
	}()
}
