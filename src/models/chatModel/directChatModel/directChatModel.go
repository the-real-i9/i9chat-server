package directChat

import (
	"context"
	"i9chat/src/helpers"
	"i9chat/src/models/db"
	"i9chat/src/models/modelHelpers"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type NewMessageT struct {
	Id                   string         `json:"id" db:"id"`
	ChatHistoryEntryType string         `json:"che_type" db:"che_type"`
	Content              map[string]any `json:"content" db:"content"`
	DeliveryStatus       string         `json:"delivery_status" db:"delivery_status"`
	CreatedAt            int64          `json:"created_at" db:"created_at"`
	Sender               any            `json:"sender" db:"sender"`
	ReplyTargetMsg       map[string]any `json:"reply_target_msg,omitempty" db:"reply_target_msg"`
}

func SendMessage(ctx context.Context, clientUsername, partnerUsername, msgContent string, at int64) (NewMessageT, error) {
	var newMsg NewMessageT

	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (clientUser:User{ username: $client_username }), (partnerUser:User{ username: $partner_username })
		MERGE (clientUser)-[:HAS_CHAT]->(clientChat:DirectChat{ owner_username: $client_username, partner_username: $partner_username })-[:WITH_USER]->(partnerUser)
		MERGE (partnerUser)-[:HAS_CHAT]->(partnerChat:DirectChat{ owner_username: $partner_username, partner_username: $client_username })-[:WITH_USER]->(clientUser)

		WITH clientUser, clientChat, partnerUser, partnerChat
		CREATE (message:DirectMessage:DirectChatEntry{ id: randomUUID(), che_type: "message", content: $message_content, delivery_status: "sent", created_at: $at }),
			(clientUser)-[:SENDS_MESSAGE]->(message)-[:IN_DIRECT_CHAT]->(clientChat),
			(partnerUser)-[:RECEIVES_MESSAGE]->(message)-[:IN_DIRECT_CHAT]->(partnerChat)
			
		WITH message, clientUser { .username, .profile_pic_url, .presence } AS sender

		RETURN message { .*, content: apoc.convert.fromJsonMap(message.content), sender } AS new_message
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
		return newMsg, fiber.ErrInternalServerError
	}

	newMsg = modelHelpers.RKeyGet[NewMessageT](res.Records, "new_message")

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

func ReplyToMessage(ctx context.Context, clientUsername, partnerUsername, targetMsgId, msgContent string, at int64) (NewMessageT, error) {
	var newMsg NewMessageT

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
			clientUser { .username, .profile_pic_url, .presence } AS sender,
			targetMsg { .id, content: apoc.convert.fromJsonMap(targetMsg.content), sender_username: targetMsgSender.username } AS reply_target_msg

		RETURN replyMsg { .*, content: apoc.convert.fromJsonMap(replyMsg.content), created_at, sender, reply_target_msg } AS new_message
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
		return newMsg, fiber.ErrInternalServerError
	}

	newMsg = modelHelpers.RKeyGet[NewMessageT](res.Records, "new_message")

	return newMsg, nil
}

type RxnToMessageT struct {
	CHEId                string `json:"-" db:"che_id"`
	ChatHistoryEntryType string `json:"che_type" db:"che_type"`
	Emoji                string `json:"emoji" db:"emoji"`
	Reactor              any    `json:"reactor" db:"reactor"`
	ToMsgId              string `json:"-" db:"to_msg_id"`
}

func ReactToMessage(ctx context.Context, clientUsername, partnerUsername, msgId, emoji string, at int64) (RxnToMessageT, error) {
	var rxnToMessage RxnToMessageT

	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (clientUser)-[:HAS_CHAT]->(clientChat:Chat{ owner_username: $client_username, partner_username: $partner_username })-[:WITH_USER]->(partnerUser),
			(clientChat)<-[:IN_DIRECT_CHAT]-(message:DirectMessage{ id: $message_id }),
			(partnerUser)-[:HAS_CHAT]->(partnerChat)-[:WITH_USER]->(clientUser)
		
		WITH clientUser, message, partnerUser, partnerChat
		MERGE (msgrxn:DirectMessageReaction:DirectChatEntry{ reactor_username: clientUser.username, message_id: message.id })
		ON CREATE
			msgrxn.che_id = randomUUID()
			msgrxn.che_type = "reaction"

		SET msgrxn.emoji = $emoji, msgrxn.at = $at

		MERGE (clientUser)-[crxn:REACTS_TO_MESSAGE]->(message)
		SET crxn.emoji = $emoji, crxn.at = $at
		
		MERGE (msgrxn)-[:IN_DIRECT_CHAT]->(clientChat)
		MERGE (msgrxn)-[:IN_DIRECT_CHAT]->(partnerChat)

		WITH msgrxn, clientUser { .username, .profile_pic_url } AS reactor

		RETURN msgrxn { .che_id, .che_type, .emoji, reactor, to_msg_id: msgrxn.message_id  } AS rxn_to_msg

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
		return rxnToMessage, fiber.ErrInternalServerError
	}

	rxnToMessage = modelHelpers.RKeyGet[RxnToMessageT](res.Records, "rxn_to_msg")

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

		WITH msgrxn, crxn, msgrxn.che_id AS msgrxn_che_id

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

type ChatHistoryEntry struct {
	EntryType string `json:"chat_hist_entry_type"`
	CreatedAt int64  `json:"created_at"`

	// for message entry
	Id             string           `json:"id,omitempty"`
	Content        map[string]any   `json:"content,omitempty"`
	DeliveryStatus string           `json:"delivery_status,omitempty"`
	Sender         map[string]any   `json:"sender,omitempty"`
	IsOwn          bool             `json:"is_own"`
	Reactions      []map[string]any `json:"reactions,omitempty"`

	// for a reply message entry
	ReplyTargetMsg map[string]any `json:"reply_target_msg,omitempty"`

	// for reaction entry
	Reaction string `json:"reaction,omitempty"`
}

func ChatHistory(ctx context.Context, clientUsername, partnerUsername string, limit int, offset int64) ([]ChatHistoryEntry, error) {
	var chatHistory []ChatHistoryEntry

	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (clientChat:DirectChat{ owner_username: $client_username, partner_username: $partner_username })

		OPTIONAL MATCH (clientChat)<-[:IN_DIRECT_CHAT]-(entry:DirectChatEntry WHERE entry.created_at < $offset)
		OPTIONAL MATCH (entry)<-[:SENDS_MESSAGE]-(senderUser)
		OPTIONAL MATCH (entry)<-[rxn:REACTS_TO_MESSAGE]-(reactorUser)
		OPTIONAL MATCH (entry)-[:REPLIES_TO]->(replyTargetMsg:DirectMessage)
		OPTIONAL MATCH (replyTargetMsg)<-[:SENDS_MESSAGE]-(replyTargetMsgSender)

		WITH entry, senderUser, replyTargetMsg, replyTargetMsgSender,
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
			CASE WHEN replyTargetMsg IS NOT NULL
				THEN replyTargetMsg { .id, content: apoc.convert.fromJsonMap(replyTargetMsg.content), sender_username: replyTargetMsgSender.username, is_own: replyTargetMsgSender.username = $client_username }
				ELSE NULL
			END AS reply_target_msg,
			CASE WHEN entry.chat_hist_entry_type = "message"
				THEN apoc.convert.fromJsonMap(entry.content)
				ELSE NULL
			END AS content
		ORDER BY entry.created_at
		LIMIT $limit
		
		RETURN collect(entry { .*, content, created_at, sender, is_own, reactions, reply_target_msg }) AS chat_history
		`,
		map[string]any{
			"client_username":  clientUsername,
			"partner_username": partnerUsername,
			"limit":            limit,
			"offset":           offset,
		},
	)
	if err != nil {
		log.Println("directChatModel.go: ChatHistory", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	history, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "chat_history")

	helpers.ToStruct(history, &chatHistory)

	return chatHistory, nil
}
