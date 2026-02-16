package directChat

import (
	"context"
	"fmt"
	"i9chat/src/appGlobals"
	"i9chat/src/appTypes/UITypes"
	"i9chat/src/helpers"
	"i9chat/src/models/db"
	"i9chat/src/models/modelHelpers"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func redisDB() *redis.Client {
	return appGlobals.RedisClient
}

type NewMessage struct {
	Id             string         `json:"id" db:"id"`
	CHEType        string         `json:"che_type" db:"che_type"`
	Content        map[string]any `json:"content" db:"content"`
	DeliveryStatus string         `json:"delivery_status" db:"delivery_status"`
	CreatedAt      int64          `json:"created_at" db:"created_at"`
	Sender         any            `json:"sender" db:"sender"`
	ReplyTargetMsg map[string]any `json:"reply_target_msg,omitempty" db:"reply_target_msg"`
	Cursor         int64          `json:"cursor" db:"cursor"`
	FirstFromUser  bool           `json:"-" db:"ffu"`
	FirstToUser    bool           `json:"-" db:"ftu"`
}

func SendMessage(ctx context.Context, clientUsername, partnerUsername, msgContent string, at int64) (NewMessage, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (clientUser:User{ username: $client_username }), (partnerUser:User{ username: $partner_username })

		WITH clientUser, partnerUser,
			NOT EXISTS { (clientUser)-[:HAS_CHAT]->(:DirectChat)-[:WITH_USER]->(partnerUser) } AS ffu,
			NOT EXISTS { (partnerUser)-[:HAS_CHAT]->(:DirectChat)-[:WITH_USER]->(clientUser) } AS ftu

		MERGE (serialCounter:DirectCHESerialCounter{ name: $direct_che_serial_counter })
		ON CREATE SET serialCounter.value = 0

		LET dummy = 0

		CALL apoc.atomic.add(serialCounter, 'value', 1) YIELD cheNextVal

		MERGE (clientUser)-[:HAS_CHAT]->(clientChat:DirectChat{ owner_username: $client_username, partner_username: $partner_username })-[:WITH_USER]->(partnerUser)
		MERGE (partnerUser)-[:HAS_CHAT]->(partnerChat:DirectChat{ owner_username: $partner_username, partner_username: $client_username })-[:WITH_USER]->(clientUser)

		WITH clientUser, clientChat, partnerUser, partnerChat, cheNextVal, ffu, ftu
		CREATE (message:DirectMessage:DirectChatEntry{ id: randomUUID(), che_type: "message", content: $message_content, delivery_status: "sent", created_at: $at, cursor: cheNextVal }),
			(clientUser)-[:SENDS_MESSAGE]->(message)-[:IN_DIRECT_CHAT]->(clientChat),
			(message)-[:IN_DIRECT_CHAT { receipt: "received" }]->(partnerChat)
		
		SET clientChat.cursor = cheNextVal

		RETURN message { .*, content: apoc.convert.fromJsonMap(message.content), sender: $client_username, ffu: ffu, ftu: ftu } AS new_message
		`,
		map[string]any{
			"client_username":           clientUsername,
			"partner_username":          partnerUsername,
			"message_content":           msgContent,
			"at":                        at,
			"direct_che_serial_counter": "$directCHESC$",
		},
	)
	if err != nil {
		helpers.LogError(err)
		return NewMessage{}, fiber.ErrInternalServerError
	}

	newMsg := modelHelpers.RKeyGet[NewMessage](res.Records, "new_message")

	return newMsg, nil
}

func AckMessageDelivered(ctx context.Context, clientUsername, partnerUsername string, msgIds []any, deliveredAt int64) (int64, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (clientChat:DirectChat{ owner_username: $client_username, partner_username: $partner_username }),
      (clientChat)<-[:IN_DIRECT_CHAT { receipt: "received" }]-(message:DirectMessage{ delivery_status: "sent" } WHERE message.id IN $message_ids)

    SET message.delivery_status = "delivered", message.delivered_at = $delivered_at, clientChat.cursor = message.cursor

		RETURN DISTINCT clientChat.cursor AS cursor
		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"message_ids":      msgIds,
			"delivered_at":     deliveredAt,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return 0, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return 0, nil
	}

	cursor := modelHelpers.RKeyGet[int64](res.Records, "cursor")

	return cursor, nil
}

func AckMessageRead(ctx context.Context, clientUsername, partnerUsername string, msgIds []any, readAt int64) (bool, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (clientChat:DirectChat{ owner_username: $client_username, partner_username: $partner_username }),
      (clientChat)<-[:IN_DIRECT_CHAT { receipt: "received" }]-(message:DirectMessage WHERE message.delivery_status <> "read" AND message.id IN $message_ids)

    SET message.delivery_status = "read", 
			message.read_at = $read_at,
			// if a client skips the "delivered" ack, and acks "read"
			// it means the message is delivered and read at the same time
			// so if delivered_at is NULL, then use read_at
			message.delivered_at = coalesce(message.delivered_at, message.read_at)

		RETURN true AS done
		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"message_ids":      msgIds,
			"read_at":          readAt,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return false, nil
	}

	return true, nil
}

func ReplyToMessage(ctx context.Context, clientUsername, partnerUsername, targetMsgId, msgContent string, at int64) (NewMessage, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (clientUser)-[:HAS_CHAT]->(clientChat:Chat{ owner_username: $client_username, partner_username: $partner_username })-[:WITH_USER]->(partnerUser),
			(clientChat)<-[:IN_DIRECT_CHAT]-(targetMsg:DirectMessage { id: $target_msg_id })
			
		MATCH (targetMsg)<-[:SENDS_MESSAGE]-(targetMsgSender)

		WITH clientUser, partnerUser, clientChat, targetMsgSender,
			NOT EXISTS { (clientUser)-[:HAS_CHAT]->(:DirectChat)-[:WITH_USER]->(partnerUser) } AS ffu,
			NOT EXISTS { (partnerUser)-[:HAS_CHAT]->(:DirectChat)-[:WITH_USER]->(clientUser) } AS ftu

		MERGE (serialCounter:DirectCHESerialCounter{ name: $direct_che_serial_counter })
		ON CREATE SET serialCounter.value = 0

		LET dummy = 0

		CALL apoc.atomic.add(serialCounter, 'value', 1) YIELD cheNextVal

		MERGE (partnerUser)-[:HAS_CHAT]->(partnerChat:DirectChat{ owner_username: $partner_username, partner_username: $client_username })-[:WITH_USER]->(clientUser)

		WITH clientUser, clientChat, partnerUser, partnerChat, targetMsg, targetMsgSender, cheNextVal, ffu, ftu
		CREATE (replyMsg:DirectMessage:DirectChatEntry{ id: randomUUID(), che_type: "message", content: $message_content, delivery_status: "sent", created_at: $at, cursor: cheNextVal }),
			(clientUser)-[:SENDS_MESSAGE]->(replyMsg)-[:IN_DIRECT_CHAT]->(clientChat),
			(replyMsg)-[:IN_DIRECT_CHAT { receipt: "received" }]->(partnerChat),
			(replyMsg)-[:REPLIES_TO]->(targetMsg)

		SET clientChat.cursor = cheNextVal

		WITH replyMsg,
			targetMsg { .id, content: apoc.convert.fromJsonMap(targetMsg.content), sender_user: targetMsgSender.username } AS reply_target_msg

		RETURN replyMsg { .*, content: apoc.convert.fromJsonMap(replyMsg.content), created_at, sender: $client_username, reply_target_msg, ffu: ffu, ftu: ftu } AS new_message
		`,
		map[string]any{
			"client_username":           clientUsername,
			"partner_username":          partnerUsername,
			"message_content":           msgContent,
			"target_msg_id":             targetMsgId,
			"at":                        at,
			"direct_che_serial_counter": "$directCHESC$",
		},
	)
	if err != nil {
		helpers.LogError(err)
		return NewMessage{}, fiber.ErrInternalServerError
	}

	newMsg := modelHelpers.RKeyGet[NewMessage](res.Records, "new_message")

	return newMsg, nil
}

type RxnToMessage struct {
	CHEId   string `json:"-" db:"che_id"`
	CHEType string `json:"che_type" db:"che_type"`
	Emoji   string `json:"emoji" db:"emoji"`
	Reactor any    `json:"reactor" db:"reactor"`
	Cursor  int64  `json:"cursor" db:"cursor"`
	ToMsgId string `json:"to_msg_id" db:"to_msg_id"`
}

func ReactToMessage(ctx context.Context, clientUsername, partnerUsername, msgId, emoji string, at int64) (RxnToMessage, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (clientUser)-[:HAS_CHAT]->(clientChat:Chat{ owner_username: $client_username, partner_username: $partner_username })-[:WITH_USER]->(partnerUser),
			(clientChat)<-[:IN_DIRECT_CHAT]-(message:DirectMessage{ id: $message_id }),
			(partnerUser)-[:HAS_CHAT]->(partnerChat)-[:WITH_USER]->(clientUser)

		MERGE (serialCounter:DirectCHESerialCounter{ name: $direct_che_serial_counter })
		ON CREATE SET serialCounter.value = 0

		LET dummy = 0
		
		CALL apoc.atomic.add(serialCounter, 'value', 1) YIELD cheNextVal

		WITH clientUser, message, partnerUser, cheNextVal
		MERGE (msgrxn:DirectMessageReaction:DirectChatEntry{ reactor_username: clientUser.username, message_id: $message_id })
		ON CREATE
			SET msgrxn.che_id = randomUUID(),
				msgrxn.che_type = "reaction"

		SET msgrxn.emoji = $emoji, msgrxn.at = $at, msgrxn.cursor = cheNextVal

		MERGE (clientUser)-[crxn:REACTS_TO_MESSAGE]->(message)
		SET crxn.emoji = $emoji, crxn.at = $at

		MERGE (msgrxn)-[:IN_DIRECT_CHAT]->(clientChat)
		MERGE (msgrxn)-[:IN_DIRECT_CHAT]->(partnerChat)

		RETURN msgrxn { .che_id, .che_type, .emoji, to_msg_id: $message_id, reactor: $client_username, .cursor } AS rxn_to_msg

		`,
		map[string]any{
			"client_username":           clientUsername,
			"partner_username":          partnerUsername,
			"message_id":                msgId,
			"emoji":                     emoji,
			"at":                        at,
			"direct_che_serial_counter": "$directCHESC$",
		},
	)
	if err != nil {
		helpers.LogError(err)
		return RxnToMessage{}, fiber.ErrInternalServerError
	}

	rxnToMessage := modelHelpers.RKeyGet[RxnToMessage](res.Records, "rxn_to_msg")

	return rxnToMessage, nil
}

func RemoveReactionToMessage(ctx context.Context, clientUsername, partnerUsername, msgId string) (string, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (clientUser)-[:HAS_CHAT]->(clientChat:DirectChat{ owner_username: $client_username, partner_username: $partner_username })-[:WITH_USER]->(partnerUser),
			(clientChat)<-[:IN_DIRECT_CHAT]-(message:DirectMessage{ id: $message_id }),
			(partnerUser)-[:HAS_CHAT]->(partnerChat)-[:WITH_USER]->(clientUser),

			(msgrxn:DirectMessageReaction:DirectChatEntry{ reactor_username: $client_username, message_id: $message_id }),
			(clientUser)-[crxn:REACTS_TO_MESSAGE]->(message)

		LET msgrxn_che_id = msgrxn.che_id

		DETACH DELETE msgrxn, crxn
		
		RETURN msgrxn_che_id
    `,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"message_id":       msgId,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return "", fiber.ErrInternalServerError
	}

	CHEId := modelHelpers.RKeyGet[string](res.Records, "msgrxn_che_id")

	return CHEId, nil
}

func ChatHistory(ctx context.Context, clientUsername, partnerUsername string, limit int, cursor float64) ([]UITypes.ChatHistoryEntry, error) {
	cheMembers, err := redisDB().ZRevRangeByScoreWithScores(ctx, fmt.Sprintf("direct_chat:owner:%s:partner:%s:history", clientUsername, partnerUsername), &redis.ZRangeBy{
		Max:   helpers.MaxCursor(cursor),
		Min:   "-inf",
		Count: int64(limit),
	}).Result()
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	history, err := modelHelpers.CHEMembersForUICHEs(ctx, cheMembers, "direct")
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	return history, nil
}
