package dmChat

import (
	"context"
	"fmt"
	"i9chat/helpers"
	"i9chat/models/db"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type NewMessage struct {
	ClientNewMsgData  map[string]any `json:"client_res"`
	PartnerNewMsgData map[string]any `json:"partner_res"`
}

func SendMessage(ctx context.Context, clientUsername, partnerUsername string, msgContentJson []byte, createdAt time.Time) (NewMessage, error) {
	var newMsg NewMessage

	res, err := db.Query(
		ctx,
		`
		MATCH (clientUser:User{ username: $client_username }), (partnerUser:User{ username: $partner_username })
		MERGE (clientUser)-[:HAS_CHAT]->(clientChat:DMChat{ owner_username: $client_username, partner_username: $partner_username })-[:WITH_USER]->(partnerUser)
		MERGE (partnerUser)-[:HAS_CHAT]->(partnerChat:DMChat{ owner_username: $partner_username, partner_username: $client_username })-[:WITH_USER]->(clientUser)
		SET clientChat.last_activity_type = "message", 
			partnerChat.last_activity_type = "message",
			clientChat.updated_at = $created_at, 
			partnerChat.updated_at = $created_at
		WITH clientUser, clientChat, partnerUser, partnerChat
		CREATE (message:Message{ id: randomUUID(), content: $message_content, delivery_status: "sent", created_at: $created_at }),
			(clientUser)-[:SENDS_MESSAGE]->(message)-[:IN_DM_CHAT]->(clientChat),
			(partnerUser)-[:RECEIVES_MESSAGE]->(message)-[:IN_DM_CHAT]->(partnerChat)
		SET clientChat.last_message_id = message.id,
			partnerChat.last_message_id = message.id
		WITH message, toString(message.created_at) AS created_at, clientUser { .username, .profile_pic_url, .connection_status } AS sender
		RETURN { new_msg_id: message.id } AS client_res,
			message { .*, created_at, sender } AS partner_res
		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"message_content":  msgContentJson,
			"created_at":       createdAt,
		},
	)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: SendMessage: %s", err))
		return newMsg, fiber.ErrInternalServerError
	}

	helpers.MapToStruct(res.Records[0].AsMap(), &newMsg)

	return newMsg, nil
}

func ReactToMessage(ctx context.Context, clientDMChatId string, msgId, clientUserId int, reaction rune) error {

	return nil
}

func GetChatHistory(ctx context.Context, clientUsername, partnerUsername string, limit int, offset time.Time) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (clientChat:DMChat{ owner_username: $client_username, partner_username: $partner_username })<-[:IN_DM_CHAT]-(message:Message)<-[rxn:REACTS_TO_MESSAGE]-(reactor)
		WHERE message.created_at >= $offset
		WITH message, toString(message.created_at) AS created_at, collect({ user: reactor { .username, .profile_pic_url }, reaction: rxn.reaction }) AS reactions
		ORDER BY message.created_at DESC
		LIMIT $limit
		RETURN collect(message { .*, created_at, reactions }) AS chat_history
		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"limit":            limit,
			"offset":           offset,
		},
	)
	if err != nil {
		log.Println("DMChatModel.go: GetChatHistory", err)
		return nil, fiber.ErrInternalServerError
	}

	messages, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "chat_history")

	return messages, nil
}

func AckMessageDelivered(ctx context.Context, clientUsername, partnerUsername, msgId string, deliveredAt time.Time) error {
	_, err := db.Query(
		ctx,
		`
		MATCH (clientChat:DMChat{ owner_username: $client_username, partner_username: $partner_username }),
      (clientChat)<-[:IN_DM_CHAT]-(message:Message{ id: $message_id, delivery_status: "sent" })<-[:RECEIVES_MESSAGE]-()
    SET message.delivery_status = "delivered", message.delivered_at = datetime($delivered_at), clientChat.unread_messages_count = coalesce(clientChat.unread_messages_count, 0) + 1
		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"message_id":       msgId,
			"delivered_at":     deliveredAt,
		},
	)
	if err != nil {
		log.Println("DMChatModel.go: AckMessageDelivered", err)
		return fiber.ErrInternalServerError
	}

	return nil
}

func AckMessageRead(ctx context.Context, clientUsername, partnerUsername, msgId string, readAt time.Time) error {
	_, err := db.Query(
		ctx,
		`
		MATCH (clientChat:DMChat{ owner_username: $client_username, partner_username: $partner_username }),
      (clientChat)<-[:IN_DM_CHAT]-(message:Message{ id: $message_id } WHERE message.delivery_status IN ["sent", "delivered"])<-[:RECEIVES_MESSAGE]-()
    WITH clientChat, message, CASE coalesce(clientChat.unread_messages_count, 0) WHEN <> 0 THEN clientChat.unread_messages_count - 1 ELSE 0 END AS unread_messages_count
    SET message.delivery_status = "read", message.read_at = datetime($read_at), clientChat.unread_messages_count = unread_messages_count
		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"message_id":       msgId,
			"read_at":          readAt,
		},
	)
	if err != nil {
		log.Println("DMChatModel.go: AckMessageDelivered", err)
		return fiber.ErrInternalServerError
	}

	return nil
}
