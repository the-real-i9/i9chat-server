package dmChat

import (
	"context"
	"i9chat/src/helpers"
	"i9chat/src/models/db"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type NewMessage struct {
	ClientData  map[string]any `json:"client_resp"`
	PartnerData map[string]any `json:"partner_resp"`
}

func SendMessage(ctx context.Context, clientUsername, partnerUsername, msgContent string, at time.Time) (NewMessage, error) {
	var newMsg NewMessage

	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (clientUser:User{ username: $client_username }), (partnerUser:User{ username: $partner_username })
		MERGE (clientUser)-[:HAS_CHAT]->(clientChat:DMChat{ owner_username: $client_username, partner_username: $partner_username })-[:WITH_USER]->(partnerUser)
		MERGE (partnerUser)-[:HAS_CHAT]->(partnerChat:DMChat{ owner_username: $partner_username, partner_username: $client_username })-[:WITH_USER]->(clientUser)

		WITH clientUser, clientChat, partnerUser, partnerChat
		CREATE (message:DMMessage:DMChatEntry{ id: randomUUID(), chat_hist_entry_type: "message", content: $message_content, delivery_status: "sent", created_at: $at }),
			(clientUser)-[:SENDS_MESSAGE]->(message)-[:IN_DM_CHAT]->(clientChat),
			(partnerUser)-[:RECEIVES_MESSAGE]->(message)-[:IN_DM_CHAT]->(partnerChat)
			
		WITH message, message.created_at.epochMillis AS created_at, clientUser { .username, .profile_pic_url, .presence } AS sender
		RETURN { new_msg_id: message.id } AS client_resp,
			message { .*, content: apoc.convert.fromJsonMap(message.content), created_at, sender } AS partner_resp
		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"message_content":  msgContent,
			"at":               at,
		},
	)
	if err != nil {
		log.Println("DMChatModel.go: SendMessage:", err)
		return newMsg, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return newMsg, nil
	}

	helpers.ToStruct(res.Records[0].AsMap(), &newMsg)

	return newMsg, nil
}

func AckMessageDelivered(ctx context.Context, clientUsername, partnerUsername, msgId string, deliveredAt time.Time) (bool, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (clientChat:DMChat{ owner_username: $client_username, partner_username: $partner_username }),
      (clientChat)<-[:IN_DM_CHAT]-(message:DMMessage{ id: $message_id, delivery_status: "sent" })<-[:RECEIVES_MESSAGE]-()
    SET message.delivery_status = "delivered", message.delivered_at = $delivered_at, clientChat.unread_messages_count = coalesce(clientChat.unread_messages_count, 0) + 1

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
		log.Println("DMChatModel.go: AckMessageDelivered", err)
		return false, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return false, nil
	}

	done, _, _ := neo4j.GetRecordValue[bool](res.Records[0], "done")

	return done, nil
}

func AckMessageRead(ctx context.Context, clientUsername, partnerUsername, msgId string, readAt time.Time) (bool, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (clientChat:DMChat{ owner_username: $client_username, partner_username: $partner_username }),
      (clientChat)<-[:IN_DM_CHAT]-(message:DMMessage{ id: $message_id } WHERE message.delivery_status IN ["sent", "delivered"])<-[:RECEIVES_MESSAGE]-()

    WITH clientChat, message, CASE coalesce(clientChat.unread_messages_count, 0) WHEN <> 0 THEN clientChat.unread_messages_count - 1 ELSE 0 END AS unread_messages_count
    SET message.delivery_status = "read", message.read_at = $read_at, clientChat.unread_messages_count = unread_messages_count

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
		log.Println("DMChatModel.go: AckMessageRead", err)
		return false, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return false, nil
	}

	done, _, _ := neo4j.GetRecordValue[bool](res.Records[0], "done")

	return done, nil
}

func ReplyToMessage(ctx context.Context, clientUsername, partnerUsername, targetMsgId, msgContent string, at time.Time) (NewMessage, error) {
	var newMsg NewMessage

	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (clientUser:User{ username: $client_username }), (partnerUser:User{ username: $partner_username })
		MATCH (targetMsg:DMMessage { id: $target_msg_id })<-[:SENDS_MESSAGE]-(targetMsgSender)
		MERGE (clientUser)-[:HAS_CHAT]->(clientChat:DMChat{ owner_username: $client_username, partner_username: $partner_username })-[:WITH_USER]->(partnerUser)
		MERGE (partnerUser)-[:HAS_CHAT]->(partnerChat:DMChat{ owner_username: $partner_username, partner_username: $client_username })-[:WITH_USER]->(clientUser)

		WITH clientUser, clientChat, partnerUser, partnerChat, targetMsg
		CREATE (replyMsg:DMMessage:DMChatEntry{ id: randomUUID(), chat_hist_entry_type: "reply", content: $message_content, delivery_status: "sent", created_at: $at }),
			(clientUser)-[:SENDS_MESSAGE]->(replyMsg)-[:IN_DM_CHAT]->(clientChat),
			(partnerUser)-[:RECEIVES_MESSAGE]->(replyMsg)-[:IN_DM_CHAT]->(partnerChat),
			(replyMsg)-[:REPLIES_TO]->(targetMsg)

		WITH replyMsg, replyMsg.created_at.epochMillis AS created_at,
			clientUser { .username, .profile_pic_url, .presence } AS sender,
			targetMsg { .id, .content, sender_username: targetMsgSender.username } AS replied_to

		RETURN { new_msg_id: replyMsg.id } AS client_resp,
			replyMsg { .*, replyMsg: apoc.convert.fromJsonMap(replyMsg.content), created_at, sender, replied_to } AS partner_resp
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
		log.Println("DMChatModel.go: ReplyToMessage:", err)
		return newMsg, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return newMsg, nil
	}

	helpers.ToStruct(res.Records[0].AsMap(), &newMsg)

	return newMsg, nil
}

type RxnToMessage struct {
	ClientData  map[string]any `json:"client_resp"`
	PartnerData map[string]any `json:"partner_resp"`
}

func ReactToMessage(ctx context.Context, clientUsername, partnerUsername, msgId, reaction string, at time.Time) (RxnToMessage, error) {
	var rxnToMessage RxnToMessage

	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (clientUser)-[:HAS_CHAT]->(clientChat:Chat{ owner_username: $client_username, partner_username: $partner_username })-[:WITH_USER]->(partnerUser),
			(clientChat)<-[:IN_DM_CHAT]-(message:DMMessage{ id: $message_id }),
			(partnerUser)-[:HAS_CHAT]->(partnerChat)-[:WITH_USER]->(clientUser)
		
		WITH clientUser, message, partnerUser, partnerChat
		MERGE (msgrxn:DMMessageReaction:DMChatEntry{ reactor_username: clientUser.username, message_id: message.id })
		SET msgrxn.reaction = $reaction, msgrxn.chat_hist_entry_type = "reaction", msgrxn.created_at = $at

		MERGE (clientUser)-[crxn:REACTS_TO_MESSAGE]->(message)
		SET crxn.reaction = $reaction, crxn.created_at = $at
		
		MERGE (clientUser)-[:SENDS_REACTION]->(msgrxn)-[:IN_DM_CHAT]->(clientChat)
		MERGE (partnerUser)-[:RECEIVES_REACTION]->(msgrxn)-[:IN_DM_CHAT]->(partnerChat)

		WITH clientUser.username AS partner_username, message.id AS msg_id, 
			clientUser { .username, .profile_pic_url } AS reactor, crxn

		RETURN true AS client_resp,
			{ partner_username, msg_id, reactor, reaction: crxn.reaction, at: crxn.created_at.epochMillis } AS partner_resp

		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"message_id":       msgId,
			"reaction":         reaction,
			"at":               at,
		},
	)
	if err != nil {
		log.Println("DMChatModel.go: ReactTs]tyoMessage:", err)
		return rxnToMessage, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return rxnToMessage, nil
	}

	helpers.ToStruct(res.Records[0].AsMap(), &rxnToMessage)

	return rxnToMessage, nil
}

func RemoveReactionToMessage(ctx context.Context, clientUsername, partnerUsername, msgId string) (bool, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (clientUser)-[:HAS_CHAT]->(clientChat:DMChat{ owner_username: $client_username, partner_username: $partner_username })-[:WITH_USER]->(partnerUser),
			(partnerUser)-[:HAS_CHAT]->(partnerChat)-[:WITH_USER]->(clientUser),
			(clientChat)<-[:IN_DM_CHAT]-(message:DMMessage{ id: $message_id }),

			(msgrxn:DMMessageReaction:DMChatEntry{ reactor_username: $client_username, message_id: $message_id }),
			(clientUser)-[crxn:REACTS_TO_MESSAGE]->(message)

		DETACH DELETE msgrxn, crxn
		
		RETURN true AS done
    `,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"message_id":       msgId,
		},
	)
	if err != nil {
		log.Println("DMChatModel.go: RemoveReactionToMessage:", err)
		return false, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return false, nil
	}

	done, _, _ := neo4j.GetRecordValue[bool](res.Records[0], "done")
	return done, nil
}

type ChatHistoryEntry struct {
	EntryType string `json:"chat_hist_entry_type"`
	CreatedAt int64  `json:"created_at"`

	// for message and reply entry
	Id             string           `json:"id,omitempty"`
	Content        map[string]any   `json:"content,omitempty"`
	DeliveryStatus string           `json:"delivery_status,omitempty"`
	Sender         map[string]any   `json:"sender,omitempty"`
	IsOwn          bool             `json:"is_own,omitempty"`
	Reactions      []map[string]any `json:"reactions,omitempty"`

	// for reply entry
	RepliedTo map[string]any `json:"replied_to,omitempty"`

	// for reaction entry
	Reaction string `json:"reaction,omitempty"`
}

func ChatHistory(ctx context.Context, clientUsername, partnerUsername string, limit int, offset time.Time) ([]ChatHistoryEntry, error) {
	var chatHistory []ChatHistoryEntry

	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (clientChat:DMChat{ owner_username: $client_username, partner_username: $partner_username })

		OPTIONAL MATCH (clientChat)<-[:IN_DM_CHAT]-(entry:DMChatEntry WHERE entry.created_at < $offset)
		OPTIONAL MATCH (entry)<-[:SENDS_MESSAGE]-(senderUser)
		OPTIONAL MATCH (entry)<-[rxn:REACTS_TO_MESSAGE]-(reactorUser)
		OPTIONAL MATCH (entry)-[:REPLIES_TO]->(repliedMsg:DMMessage)
		OPTIONAL MATCH (repliedMsg)<-[:SENDS_MESSAGE]-(repliedSender)

		WITH entry, senderUser, repliedMsg, repliedSender,
     collect(CASE WHEN rxn IS NOT NULL 
             THEN { reactor: reactorUser { .username, .profile_pic_url }, reaction: rxn.reaction, at: rxn.created_at.epochMillis }
             ELSE NULL 
             END) AS reaction_list

		WITH entry, entry.created_at.epochMillis AS created_at,
			CASE WHEN senderUser IS NOT NULL
				THEN senderUser { .username, .profile_pic_url } 
				ELSE NULL
			END AS sender,
			CASE WHEN senderUser IS NOT NULL AND senderUser.username = $client_username
				THEN true 
				ELSE false
			END AS is_own,
			CASE WHEN size([r IN reaction_list WHERE r IS NOT NULL]) > 0
         THEN [r IN reaction_list WHERE r IS NOT NULL]
         ELSE NULL
			END AS reactions,
			CASE WHEN repliedMsg IS NOT NULL
				THEN repliedMsg { .id, content: apoc.convert.fromJsonMap(repliedMsg.content), sender_username: repliedSender.username, is_own: repliedSender.username = $client_username }
				ELSE NULL
			END AS replied_to,
			CASE WHEN entry.chat_hist_entry_type = "message" OR entry.chat_hist_entry_type = "reply"
				THEN apoc.convert.fromJsonMap(entry.content)
				ELSE NULL
			END AS content
		ORDER BY entry.created_at
		LIMIT $limit
		
		RETURN collect(entry { .*, content, created_at, sender, is_own, reactions, replied_to }) AS chat_history
		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"limit":            limit,
			"offset":           offset,
		},
	)
	if err != nil {
		log.Println("DMChatModel.go: ChatHistory", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	history, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "chat_history")

	helpers.ToStruct(history, &chatHistory)

	return chatHistory, nil
}
