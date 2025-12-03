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
	Id                   string         `json:"id" db:"id"`
	ChatHistoryEntryType string         `json:"che_type" db:"che_type"`
	Content              map[string]any `json:"content" db:"content"`
	DeliveryStatus       string         `json:"delivery_status" db:"delivery_status"`
	CreatedAt            int64          `json:"created_at" db:"created_at"`
	Sender               any            `json:"sender" db:"sender"`
	ReplyTargetMsg       map[string]any `json:"reply_target_msg,omitempty" db:"reply_target_msg"`
	FirstFromUser        bool           `json:"-" db:"ffu"`
	FirstToUser          bool           `json:"-" db:"ftu"`
}

func SendMessage(ctx context.Context, clientUsername, partnerUsername, msgContent string, at int64) (NewMessage, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (clientUser:User{ username: $client_username }), (partnerUser:User{ username: $partner_username })

		WITH clientUser, partnerUser,
			NOT EXISTS { (clientUser)-[:HAS_CHAT]->(:DirectChat)-[:WITH_USER]->(partnerUser) } AS ffu
			NOT EXISTS { (partnerUser)-[:HAS_CHAT]->(:DirectChat)-[:WITH_USER]->(clientUser) } AS ftu

		MERGE (clientUser)-[:HAS_CHAT]->(clientChat:DirectChat{ owner_username: $client_username, partner_username: $partner_username })-[:WITH_USER]->(partnerUser)
		MERGE (partnerUser)-[:HAS_CHAT]->(partnerChat:DirectChat{ owner_username: $partner_username, partner_username: $client_username })-[:WITH_USER]->(clientUser)

		WITH clientUser, clientChat, partnerUser, partnerChat, ffu, ftu
		CREATE (message:DirectMessage:DirectChatEntry{ id: randomUUID(), che_type: "message", content: $message_content, delivery_status: "sent", created_at: $at }),
			(clientUser)-[:SENDS_MESSAGE]->(message)-[:IN_DIRECT_CHAT]->(clientChat),
			(partnerUser)-[:RECEIVES_MESSAGE]->(message)-[:IN_DIRECT_CHAT]->(partnerChat)

		RETURN message { .*, content: apoc.convert.fromJsonMap(message.content), sender: $client_username, ffu, ftu } AS new_message
		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"message_content":  msgContent,
			"at":               at,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return NewMessage{}, fiber.ErrInternalServerError
	}

	newMsg := modelHelpers.RKeyGet[NewMessage](res.Records, "new_message")

	return newMsg, nil
}

func AckMessageDelivered(ctx context.Context, clientUsername, partnerUsername, msgId string, deliveredAt int64) (bool, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (clientChat:DirectChat{ owner_username: $client_username, partner_username: $partner_username }),
      (clientChat)<-[:IN_DIRECT_CHAT]-(message:DirectMessage{ id: $message_id, delivery_status: "sent" })<-[:RECEIVES_MESSAGE]-()

    SET message.delivery_status = "delivered", message.delivered_at = $delivered_at

		RETURN true AS done
		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"message_id":       msgId,
			"delivered_at":     deliveredAt,
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

func AckMessageRead(ctx context.Context, clientUsername, partnerUsername, msgId string, readAt int64) (bool, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (clientChat:DirectChat{ owner_username: $client_username, partner_username: $partner_username }),
      (clientChat)<-[:IN_DIRECT_CHAT]-(message:DirectMessage{ id: $message_id } WHERE message.delivery_status IN ["sent", "delivered"])<-[:RECEIVES_MESSAGE]-()

    SET message.delivery_status = "read", message.read_at = $read_at

		RETURN true AS done
		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"message_id":       msgId,
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
		MATCH (clientUser:User{ username: $client_username }), (partnerUser:User{ username: $partner_username })
		MATCH (targetMsg:DirectMessage { id: $target_msg_id })<-[:SENDS_MESSAGE]-(targetMsgSender)
		MERGE (clientUser)-[:HAS_CHAT]->(clientChat:DirectChat{ owner_username: $client_username, partner_username: $partner_username })-[:WITH_USER]->(partnerUser)
		MERGE (partnerUser)-[:HAS_CHAT]->(partnerChat:DirectChat{ owner_username: $partner_username, partner_username: $client_username })-[:WITH_USER]->(clientUser)

		WITH clientUser, clientChat, partnerUser, partnerChat, targetMsg, targetMsgSender
		CREATE (replyMsg:DirectMessage:DirectChatEntry{ id: randomUUID(), che_type: "message", content: $message_content, delivery_status: "sent", created_at: $at }),
			(clientUser)-[:SENDS_MESSAGE]->(replyMsg)-[:IN_DIRECT_CHAT]->(clientChat),
			(partnerUser)-[:RECEIVES_MESSAGE]->(replyMsg)-[:IN_DIRECT_CHAT]->(partnerChat),
			(replyMsg)-[:REPLIES_TO]->(targetMsg)

		WITH replyMsg,
			targetMsg { .id, content: apoc.convert.fromJsonMap(targetMsg.content), sender_user: targetMsgSender.username } AS reply_target_msg

		RETURN replyMsg { .*, content: apoc.convert.fromJsonMap(replyMsg.content), created_at, sender: $client_username, reply_target_msg } AS new_message
		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"message_content":  msgContent,
			"target_msg_id":    targetMsgId,
			"at":               at,
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
	CHEId                string `json:"-" db:"che_id"`
	ChatHistoryEntryType string `json:"che_type" db:"che_type"`
	Emoji                string `json:"emoji" db:"emoji"`
	Reactor              any    `json:"reactor" db:"reactor"`
	ToMsgId              string `json:"to_msg_id" db:"to_msg_id"`
}

func ReactToMessage(ctx context.Context, clientUsername, partnerUsername, msgId, emoji string, at int64) (RxnToMessage, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (clientUser)-[:HAS_CHAT]->(clientChat:Chat{ owner_username: $client_username, partner_username: $partner_username })-[:WITH_USER]->(partnerUser),
			(clientChat)<-[:IN_DIRECT_CHAT]-(message:DirectMessage{ id: $message_id }),
			(partnerUser)-[:HAS_CHAT]->(partnerChat)-[:WITH_USER]->(clientUser)

		WITH clientUser, message, partnerUser, partnerChat
		MERGE (msgrxn:DirectMessageReaction:DirectChatEntry{ reactor_username: clientUser.username, message_id: $message_id })
		ON CREATE
			SET msgrxn.che_id = randomUUID(),
				msgrxn.che_type = "reaction"

		SET msgrxn.emoji = $emoji, msgrxn.at = $at

		MERGE (clientUser)-[crxn:REACTS_TO_MESSAGE]->(message)
		SET crxn.emoji = $emoji, crxn.at = $at
		
		MERGE (msgrxn)-[:IN_DIRECT_CHAT]->(clientChat)
		MERGE (msgrxn)-[:IN_DIRECT_CHAT]->(partnerChat)

		RETURN msgrxn { .che_id, .che_type, .emoji, to_msg_id: $message_id, reactor: $client_username } AS rxn_to_msg

		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"message_id":       msgId,
			"emoji":            emoji,
			"at":               at,
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
			(partnerUser)-[:HAS_CHAT]->(partnerChat)-[:WITH_USER]->(clientUser),
			(clientChat)<-[:IN_DIRECT_CHAT]-(message:DirectMessage{ id: $message_id }),

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
