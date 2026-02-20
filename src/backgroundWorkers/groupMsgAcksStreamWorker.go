package backgroundWorkers

import (
	"context"
	"i9chat/src/appTypes"
	"i9chat/src/cache"
	"i9chat/src/helpers"
	"i9chat/src/models/db"
	"i9chat/src/services/eventStreamService/eventTypes"
	"i9chat/src/services/realtimeService"
	"log"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
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

				msgs = append(msgs, msg)

			}

			groupMsgDelvToUsers := make(map[string]map[string][][2]any)
			groupMsgReadByUsers := make(map[string]map[string][][2]any)

			userChatUnreadMsgs := make(map[string]map[string][]any)
			userChatReadMsgs := make(map[string]map[string][]any)

			updatedUserChats := make(map[string]map[string]float64)

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
				}
			}

			// batch processing
			eg, sharedCtx := errgroup.WithContext(ctx)

			for ownerUser, groupId_score_Pairs := range updatedUserChats {
				eg.Go(func() error {
					ownerUser, groupId_score_Pairs := ownerUser, groupId_score_Pairs

					return cache.StoreUserChatIdents(sharedCtx, ownerUser, groupId_score_Pairs)
				})
			}

			for ownerUser, groupId_unreadMsgs_Map := range userChatUnreadMsgs {
				eg.Go(func() error {
					ownerUser, groupId_unreadMsgs_Map := ownerUser, groupId_unreadMsgs_Map

					for groupId, unreadMsgs := range groupId_unreadMsgs_Map {
						if err := cache.StoreUserChatUnreadMsgs(sharedCtx, ownerUser, groupId, unreadMsgs); err != nil {
							return err
						}
					}

					return nil
				})
			}

			for ownerUser, groupId_readMsgs_Map := range userChatReadMsgs {
				eg.Go(func() error {
					ownerUser, groupId_readMsgs_Map := ownerUser, groupId_readMsgs_Map

					for groupId, readMsgs := range groupId_readMsgs_Map {
						if err := cache.RemoveUserChatUnreadMsgs(sharedCtx, ownerUser, groupId, readMsgs); err != nil {
							return err
						}
					}

					return nil
				})
			}

			for groupId, msgId_userDelvAtPairs_Map := range groupMsgDelvToUsers {
				eg.Go(func() error {
					groupId, msgId_userDelvAtPairs_Map := groupId, msgId_userDelvAtPairs_Map

					for msgId, user_delvAt_Pairs := range msgId_userDelvAtPairs_Map {
						if err := cache.StoreGroupMsgDeliveredToUsers(sharedCtx, groupId, msgId, user_delvAt_Pairs); err != nil {
							return err
						}

						go func(groupId, msgId string) {
							ctx := context.Background()

							membersList, err := cache.GetGroupMembersList(ctx, groupId)
							if err != nil {
								return
							}

							delvToUsersCount, err := cache.GetGroupMsgDeliveredToUsersCount(ctx, groupId, msgId)
							if err != nil {
								return
							}

							// delvToUsers cannot include the message sender,
							// therefore, we exempt them from the membersList count
							if len(membersList)-1 == int(delvToUsersCount) {
								go func(membersList []string) {
									for _, mu := range membersList {
										realtimeService.SendEventMsg(mu, appTypes.ServerEventMsg{
											Event: "group chat: message delivered",
											Data: map[string]string{
												"group_id": groupId,
												"msg_id":   msgId,
											},
										})
									}
								}(membersList)

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
						}(groupId, msgId)
					}

					return nil
				})
			}

			for groupId, msgId_userReadAtPairs_Map := range groupMsgReadByUsers {
				eg.Go(func() error {
					groupId, msgId_userReadAtPairs_Map := groupId, msgId_userReadAtPairs_Map

					for msgId, user_readAt_Pairs := range msgId_userReadAtPairs_Map {
						if err := cache.StoreGroupMsgReadByUsers(sharedCtx, groupId, msgId, user_readAt_Pairs); err != nil {
							return err
						}

						go func(groupId, msgId string) {
							ctx := context.Background()

							membersList, err := cache.GetGroupMembersList(ctx, groupId)
							if err != nil {
								return
							}

							readByUsersCount, err := cache.GetGroupMsgReadByUsersCount(ctx, groupId, msgId)
							if err != nil {
								return
							}

							// readByUsers cannot include the message sender,
							// therefore, we exempt them from the membersList count
							if len(membersList)-1 == int(readByUsersCount) {
								go func(membersList []string) {
									for _, mu := range membersList {
										realtimeService.SendEventMsg(mu, appTypes.ServerEventMsg{
											Event: "group chat: message read",
											Data: map[string]string{
												"group_id": groupId,
												"msg_id":   msgId,
											},
										})
									}
								}(membersList)

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
						}(groupId, msgId)
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
