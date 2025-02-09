package groupChat

import (
	"context"
	"fmt"
	"i9chat/appTypes"
	"i9chat/helpers"
	"i9chat/models/db"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

type NewGroupChat struct {
	ClientData     map[string]any `json:"client_resp"`
	InitMemberData map[string]any `json:"init_member_resp"`
}

func New(ctx context.Context, clientUsername, name, description, pictureUrl string, initUsers []string, createdAt time.Time) (NewGroupChat, error) {
	var newGroupChat NewGroupChat

	res, err := db.Query(
		ctx,
		`
		CREATE (group:Group{ id: randomUUID(), name: $name, description: $description, picture_url: $picture_url, created_at: $created_at })

		WITH group
		MATCH (clientUser:User{ username: $client_username })
		CREATE (clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group),
			(clientUser)-[:HAS_CHAT]->(clientChat:GroupChat{ owner_username: $client_username, group_id: $group.id, last_activity_type: "group activity", last_group_activity_at: $created_at, updated_at: $created_at })-[:WITH_GROUP]->(group),
			(clientUser)-[:RECEIVES_ACTIVITY]->(:GroupActivity{ info: "You created " + $name, created_at: $created_at })-[:IN_GROUP_CHAT]->(clientChat),
			(clientUser)-[:RECEIVES_ACTIVITY]->(:GroupActivity{ info: "You added " + toString($init_users), created_at: $created_at })-[:IN_GROUP_CHAT]->(clientChat)

		WITH group
		MATCH (initUser:User WHERE initUser.username IN $init_users)
		CREATE (initUser)-[:IS_MEMBER_OF { role: "member" }]->(group),
			(initUser)-[:HAS_CHAT]->(initUserChat:GroupChat{ owner_username: initUser.username, group_id: $group.id, last_activity_type: "group activity", last_group_activity_at: $created_at, updated_at: $created_at })-[:WITH_GROUP]->(group),
			(initUser)-[:RECEIVES_ACTIVITY]->(:GroupActivity{ info: $client_username + " created " + $name, created_at: $created_at })-[:IN_GROUP_CHAT]->(initUserChat),
			(initUser)-[:RECEIVES_ACTIVITY]->(:GroupActivity{ info: "You were added", created_at: $created_at })-[:IN_GROUP_CHAT]->(initUserChat)

		WITH group
		RETURN group { group_chat_id: .id, .name, .description, .picture_url, last_activity: { type: "group activity", info: "You added " + toString($init_users) } } AS client_resp,
			group { group_chat_id: .id, .name, .picture_url, last_activity: { type: "group activity", info: "You were added" } } AS init_member_resp
		`,
		map[string]any{
			"client_username": clientUsername,
			"name":            name,
			"description":     description,
			"picture_url":     pictureUrl,
			"init_users":      initUsers,
			"created_at":      createdAt,
		},
	)
	if err != nil {
		log.Println("groupChatModel.go: New:", err)
		return newGroupChat, fiber.ErrInternalServerError
	}

	helpers.MapToStruct(res.Records[0].AsMap(), &newGroupChat)

	return newGroupChat, nil
}

type NewActivity struct {
	ClientData      any      `json:"client_resp"`
	MemberData      any      `json:"member_resp"`
	MemberUsernames []string `json:"members_usernames"`
}

func ChangeName(ctx context.Context, groupId, clientUsername, newName string) (NewActivity, error) {
	var newActivity NewActivity

	res, err := db.Query(
		ctx,
		`
		MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
			(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group)
		SET clientChat.last_activity_type = "group activity",
			clientChat.last_group_activity_at = $at
		CREATE (clientUser)-[:RECEIVES_ACTIVITY]->(cligact:GroupActivity{ info: "You changed group name from " + group.name + " to " + $new_name, created_at: $at })-[:IN_GROUP_CHAT]->(clientChat)

		WITH group, cligact
		MATCH (group)<-[:IS_MEMBER_OF]-(memberUser:User WHERE memberUser.username <> $client_username),
			(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
		SET memberChat.last_activity_type = "group activity",
			memberChat.last_group_activity_at = $at
		CREATE (memberUser)-[:RECEIVES_ACTIVITY]->(memgact:GroupActivity{ info: $client_username + " changed group name from " + group.name + " to " + $new_name, created_at: $at })-[:IN_GROUP_CHAT]->(memberChat)

		SET group.name = $new_name

		RETURN cligact.info AS client_resp,
			memgact.info AS member_resp,
			collect(memberUser.username) AS member_usernames
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
			"new_name":        newName,
			"at":              time.Now(),
		},
	)
	if err != nil {
		log.Println("groupChatModel.go: ChangeName:", err)
		return newActivity, fiber.ErrInternalServerError
	}

	helpers.MapToStruct(res.Records[0].AsMap(), &newActivity)

	return newActivity, nil
}

func ChangeDescription(ctx context.Context, groupId, clientUsername, newDescription string) (NewActivity, error) {
	var newActivity NewActivity

	res, err := db.Query(
		ctx,
		`
		MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
			(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group)
		SET clientChat.last_activity_type = "group activity",
			clientChat.last_group_activity_at = $at
		CREATE (clientUser)-[:RECEIVES_ACTIVITY]->(cligact:GroupActivity{ info: "You changed group description from " + group.description + " to " + $new_description, created_at: $at })-[:IN_GROUP_CHAT]->(clientChat)

		WITH group, cligact
		MATCH (group)<-[:IS_MEMBER_OF]-(memberUser:User WHERE memberUser.username <> $client_username),
			(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
		SET memberChat.last_activity_type = "group activity",
			memberChat.last_group_activity_at = $at
		CREATE (memberUser)-[:RECEIVES_ACTIVITY]->(memgact:GroupActivity{ info: $client_username + " changed group description from " + group.description + " to " + $new_description, created_at: $at })-[:IN_GROUP_CHAT]->(memberChat)

		SET group.description = $new_description

		RETURN cligact.info AS client_resp,
			memgact.info AS member_resp,
			collect(memberUser.username) AS member_usernames
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
			"new_description": newDescription,
			"at":              time.Now(),
		},
	)
	if err != nil {
		log.Println("groupChatModel.go: ChangeDescription:", err)
		return newActivity, fiber.ErrInternalServerError
	}

	helpers.MapToStruct(res.Records[0].AsMap(), &newActivity)

	return newActivity, nil
}

func ChangePicture(ctx context.Context, groupId, clientUsername, newPictureUrl string) (NewActivity, error) {
	var newActivity NewActivity

	res, err := db.Query(
		ctx,
		`
		MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
			(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group)
		SET clientChat.last_activity_type = "group activity",
			clientChat.last_group_activity_at = $at
		CREATE (clientUser)-[:RECEIVES_ACTIVITY]->(cligact:GroupActivity{ info: "You changed group picture", created_at: $at })-[:IN_GROUP_CHAT]->(clientChat)

		WITH group, cligact
		MATCH (group)<-[:IS_MEMBER_OF]-(memberUser:User WHERE memberUser.username <> $client_username),
			(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
		SET memberChat.last_activity_type = "group activity",
			memberChat.last_group_activity_at = $at
		CREATE (memberUser)-[:RECEIVES_ACTIVITY]->(memgact:GroupActivity{ info: $client_username + " changed group picture", created_at: $at })-[:IN_GROUP_CHAT]->(memberChat)

		SET group.picture_url = $new_pic_url

		RETURN cligact.info AS client_resp,
			memgact.info AS member_resp,
			collect(memberUser.username) AS member_usernames
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
			"new_pic_url":     newPictureUrl,
			"at":              time.Now(),
		},
	)
	if err != nil {
		log.Println("groupChatModel.go: ChangePicture:", err)
		return newActivity, fiber.ErrInternalServerError
	}

	helpers.MapToStruct(res.Records[0].AsMap(), &newActivity)

	return newActivity, nil
}

func AddUsers(ctx context.Context, groupId, clientUsername string, newUsers []string) (NewActivity, any, error) {
	var newActivity NewActivity

	res, err := db.Query(
		ctx,
		`
		MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
			(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group)
		SET clientChat.last_activity_type = "group activity",
			clientChat.last_group_activity_at = $at
		CREATE (clientUser)-[:RECEIVES_ACTIVITY]->(cligact:GroupActivity{ info: "You added " + toString($new_users), created_at: $at })-[:IN_GROUP_CHAT]->(clientChat)

		WITH group, cligact
		MATCH (group)<-[:IS_MEMBER_OF]-(memberUser:User WHERE memberUser.username <> $client_username),
			(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
		SET memberChat.last_activity_type = "group activity",
			memberChat.last_group_activity_at = $at
		CREATE (memberUser)-[:RECEIVES_ACTIVITY]->(memgact:GroupActivity{ info: $client_username + " added " + toString($new_users), created_at: $at })-[:IN_GROUP_CHAT]->(memberChat)

		WITH group, cligact, memgact, collect(memberUser.username) AS member_usernames
		MATCH (newUser:User WHERE newUser.username IN $new_users)
		CREATE (newUser)-[:IS_MEMBER_OF { role: "member" }]->(group),
			(newUser)-[:HAS_CHAT]->(newUserChat:GroupChat{ owner_username: newUser.username, group_id: $group.id, last_activity_type: "group activity", last_group_activity_at: $at, updated_at: $at })-[:WITH_GROUP]->(group),
			(newUser)-[:RECEIVES_ACTIVITY]->(:GroupActivity{ info: "You were added", created_at: $at })-[:IN_GROUP_CHAT]->(newUserChat)

		RETURN cligact.info AS client_resp,
			memgact.info AS member_resp,
			group { .id, .name, .picture_url, last_activity: { type: "group activity", info: "You were added" } } AS new_user_resp,
			member_usernames AS member_usernames
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
			"new_users":       newUsers,
			"at":              time.Now(),
		},
	)
	if err != nil {
		log.Println("groupChatModel.go: AddUsers:", err)
		return newActivity, nil, fiber.ErrInternalServerError
	}

	recMap := res.Records[0].AsMap()

	helpers.MapToStruct(recMap, &newActivity)

	return newActivity, recMap["new_user_resp"], nil
}

func RemoveUser(ctx context.Context, groupId, clientUsername, targetUser string) (NewActivity, any, error) {
	var newActivity NewActivity

	res, err := db.Query(
		ctx,
		`
		MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
			(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group)
		SET clientChat.last_activity_type = "group activity",
			clientChat.last_group_activity_at = $at
		CREATE (clientUser)-[:RECEIVES_ACTIVITY]->(cligact:GroupActivity{ info: "You removed " + $target_user, created_at: $at })-[:IN_GROUP_CHAT]->(clientChat)

		WITH group, cligact
		MATCH (group)<-[mem:IS_MEMBER_OF]-(targetUser:User{ username: $target_user }),
			(targetUserChat:GroupChat{ owner_username: targetUser.username, group_id: $group_id })
		DELETE mem
		SET targetUserChat.last_activity_type = "group activity",
			targetUserChat.last_group_activity_at = $at
		CREATE (targetUser)-[:RECEIVES_ACTIVITY]->(tugact:GroupActivity{ info: $client_username + " removed you", created_at: $at })-[:IN_GROUP_CHAT]-(targetUserChat)

		WITH group, cligact, tugact
		MATCH (group)<-[:IS_MEMBER_OF]-(memberUser:User WHERE memberUser.username <> $client_username),
			(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
		SET memberChat.last_activity_type = "group activity",
			memberChat.last_group_activity_at = $at
		CREATE (memberUser)-[:RECEIVES_ACTIVITY]->(memgact:GroupActivity{ info: $client_username + " removed " + $target_user, created_at: $at })-[:IN_GROUP_CHAT]->(memberChat)

		RETURN cligact.info AS client_resp,
			memgact.info AS member_resp,
			tugact.info AS target_user_resp,
			collect(memberUser.username) AS member_usernames
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
			"target_user":     targetUser,
			"at":              time.Now(),
		},
	)
	if err != nil {
		log.Println("groupChatModel.go: RemoveUser:", err)
		return newActivity, nil, fiber.ErrInternalServerError
	}

	recMap := res.Records[0].AsMap()

	helpers.MapToStruct(recMap, &newActivity)

	return newActivity, recMap["target_user_resp"], nil
}

func Join(ctx context.Context, groupId, clientUsername string) (NewActivity, error) {
	var newActivity NewActivity

	res, err := db.Query(
		ctx,
		`
		MATCH (group:Group{ id: $group_id })<-[:IS_MEMBER_OF]-(memberUser:User),
			(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
		SET memberChat.last_activity_type = "group activity",
			memberChat.last_group_activity_at = $at
		CREATE (memberUser)-[:RECEIVES_ACTIVITY]->(memgact:GroupActivity{ info: $client_username + " joined", created_at: $at })-[:IN_GROUP_CHAT]->(memberChat)

		WITH group, memgact, collect(memberUser.username) AS member_usernames
		MATCH (newUser:User{ username: $client_username })
		CREATE (newUser)-[:IS_MEMBER_OF { role: "member" }]->(group),
			(newUser)-[:HAS_CHAT]->(newUserChat:GroupChat{ owner_username: newUser.username, group_id: $group.id, last_activity_type: "group activity", last_group_activity_at: $at, updated_at: $at })-[:WITH_GROUP]->(group),
			(newUser)-[:RECEIVES_ACTIVITY]->(:GroupActivity{ info: "You joined", created_at: $at })-[:IN_GROUP_CHAT]->(newUserChat)

		RETURN group { .id, .name, .picture_url, last_activity: { type: "group activity", info: "You joined" } } AS client_resp,
			memgact.info AS member_resp,
			member_usernames AS member_usernames
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
			"at":              time.Now(),
		},
	)
	if err != nil {
		log.Println("groupChatModel.go: Join:", err)
		return newActivity, fiber.ErrInternalServerError
	}

	recMap := res.Records[0].AsMap()

	helpers.MapToStruct(recMap, &newActivity)

	return newActivity, nil
}

func Leave(ctx context.Context, groupId, clientUsername string) (NewActivity, error) {
	var newActivity NewActivity

	res, err := db.Query(
		ctx,
		`
		MATCH (group:Group{ id: $group_id })<-[mem:IS_MEMBER_OF]-(clientUser:User{ username: $client_username }),
			(clientChat:GroupChat{ owner_username: clientUser.username, group_id: $group_id })
		DELETE mem
		SET clientChat.last_activity_type = "group activity",
			clientChat.last_group_activity_at = $at
		CREATE (clientUser)-[:RECEIVES_ACTIVITY]->(cligact:GroupActivity{ info: "You left", created_at: $at })-[:IN_GROUP_CHAT]-(clientChat)

		WITH group, cligact
		MATCH (group)<-[:IS_MEMBER_OF]-(memberUser:User),
			(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
		SET memberChat.last_activity_type = "group activity",
			memberChat.last_group_activity_at = $at
		CREATE (memberUser)-[:RECEIVES_ACTIVITY]->(memgact:GroupActivity{ info: $client_username + " left", created_at: $at })-[:IN_GROUP_CHAT]->(memberChat)

		RETURN cligact.info AS client_resp,
			memgact.info AS member_resp,
			collect(memberUser.username) AS member_usernames
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
			"at":              time.Now(),
		},
	)
	if err != nil {
		log.Println("groupChatModel.go: Leave:", err)
		return newActivity, fiber.ErrInternalServerError
	}

	recMap := res.Records[0].AsMap()

	helpers.MapToStruct(recMap, &newActivity)

	return newActivity, nil
}

func MakeUserAdmin(ctx context.Context, groupId, clientUsername, targetUser string) (NewActivity, any, error) {
	var newActivity NewActivity

	res, err := db.Query(
		ctx,
		`
		MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
			(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group)
		SET clientChat.last_activity_type = "group activity",
			clientChat.last_group_activity_at = $at
		CREATE (clientUser)-[:RECEIVES_ACTIVITY]->(cligact:GroupActivity{ info: "You made " + $target_user + " admin", created_at: $at })-[:IN_GROUP_CHAT]->(clientChat)

		WITH group, cligact
		MATCH (group)<-[mem:IS_MEMBER_OF]-(targetUser:User{ username: $target_user }),
			(targetUserChat:GroupChat{ owner_username: targetUser.username, group_id: $group_id })
		SET mem.role = "admin"
			targetUserChat.last_activity_type = "group activity",
			targetUserChat.last_group_activity_at = $at
		CREATE (targetUser)-[:RECEIVES_ACTIVITY]->(tugact:GroupActivity{ info: $client_username + " made you admin", created_at: $at })-[:IN_GROUP_CHAT]-(targetUserChat)

		WITH group, cligact, tugact
		MATCH (group)<-[:IS_MEMBER_OF]-(memberUser:User WHERE memberUser.username NOT IN [$client_username, $target_user]),
			(memberChat:GroupChat{ owner_username: memberUser.username, group_id: $group_id })
		SET memberChat.last_activity_type = "group activity",
			memberChat.last_group_activity_at = $at
		CREATE (memberUser)-[:RECEIVES_ACTIVITY]->(memgact:GroupActivity{ info: $client_username + " made " + $target_user + " admin", created_at: $at })-[:IN_GROUP_CHAT]->(memberChat)

		RETURN cligact.info AS client_resp,
			memgact.info AS member_resp,
			tugact.info AS target_user_resp,
			collect(memberUser.username) AS member_usernames
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
			"target_user":     targetUser,
			"at":              time.Now(),
		},
	)
	if err != nil {
		log.Println("groupChatModel.go: MakeUserAdmin:", err)
		return newActivity, nil, fiber.ErrInternalServerError
	}

	recMap := res.Records[0].AsMap()

	helpers.MapToStruct(recMap, &newActivity)

	return newActivity, recMap["target_user_resp"], nil
}

// I was here before shutting down
func RemoveUserFromAdmins(ctx context.Context, groupChatId int, admin []string, user []appTypes.String) (NewActivity, error) {
	newActivity, err := helpers.QueryRowType[NewActivity](ctx, "SELECT * remove_user_from_group_admins($1, $2, $3)", groupChatId, admin, user)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: RemoveUserFromAdmins: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	return newActivity, nil
}

type NewMessage struct {
	ClientData      map[string]any `json:"client_resp"`
	MemberData      map[string]any `json:"member_resp"`
	MemberUsernames []string       `json:"members_usernames"`
}

func SendMessage(ctx context.Context, groupId, clientUsername string, msgContent []byte, createdAt time.Time) (NewMessage, error) {
	var newMessage NewMessage

	res, err := db.Query(
		ctx,
		`
		CREATE (message:GroupMessage{ id: randomUUID(), content: $message_content, delivery_status: "sent", created_at: $created_at })

		WITH message
		MATCH (clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser)
		SET clientChat.last_activity_type = "message", 
			clientChat.updated_at = $created_at,
			clientChat.last_message_id = message.id
		CREATE (clientUser)-[:SENDS_MESSAGE]->(message)-[:IN_GROUP_CHAT]->(clientChat)

		WITH message, clientUser
		MATCH (memberChat:GroupChat{ group_id: $group_id } WHERE memberChat.owner_username <> $client_username)<-[:HAS_CHAT]-(memberUser)
		SET memberChat.last_activity_type = "message", 
			memberChat.updated_at = $created_at,
			memberChat.last_message_id = message.id
		CREATE (memberUser)-[:RECEIVES_MESSAGE]->(message)-[:IN_GROUP_CHAT]->(memberChat)

		WITH message, toString(message.created_at) AS created_at, clientUser { .username, .profile_pic_url } AS sender, memberUser
		RETURN { new_msg_id: message.id } AS client_resp,
			message { .*, created_at, group_id: $group_id, sender } AS member_resp,
			collect(memberUser.username) AS member_usernames
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
			"message_content": msgContent,
			"created_at":      createdAt,
		},
	)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: SendMessage: %s", err))
		return newMessage, fiber.ErrInternalServerError
	}

	helpers.MapToStruct(res.Records[0].AsMap(), &newMessage)

	return newMessage, nil
}

func AckMessageDelivered(ctx context.Context, clientUsername, groupId, msgId string, deliveredAt time.Time) error {
	_, err := db.Query(
		ctx,
		`
		MATCH (clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
      (clientChat)<-[:IN_GROUP_CHAT]-(message:GroupMessage{ id: $message_id })<-[:RECEIVES_MESSAGE]-()
    SET clientChat.unread_messages_count = coalesce(clientChat.unread_messages_count, 0) + 1
		CREATE (message)-[:DELIVERED_TO { at: $delivered_at, group_id: $group_id }]->(clientUser)
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
			"message_id":      msgId,
			"delivered_at":    deliveredAt,
		},
	)
	if err != nil {
		log.Println("groupChatModel.go: AckMessageDelivered", err)
		return fiber.ErrInternalServerError
	}

	return nil
}

func AckMessageRead(ctx context.Context, clientUsername, groupId, msgId string, readAt time.Time) error {
	_, err := db.Query(
		ctx,
		`
		MATCH (clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
      (clientChat)<-[:IN_GROUP_CHAT]-(message:GroupMessage{ id: $message_id })<-[:RECEIVES_MESSAGE]-()
    WITH clientChat, message, CASE coalesce(clientChat.unread_messages_count, 0) WHEN <> 0 THEN clientChat.unread_messages_count - 1 ELSE 0 END AS unread_messages_count
    SET clientChat.unread_messages_count = unread_messages_count
		CREATE (message)-[:READ_BY { at: $read_at, group_id: $group_id } ]->(clientUser)
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
			"message_id":      msgId,
			"read_at":         readAt,
		},
	)
	if err != nil {
		log.Println("DMChatModel.go: AckMessageRead", err)
		return fiber.ErrInternalServerError
	}

	return nil
}

func ReactToMessage(ctx context.Context, groupChatId, msgId, clientUsername string, reaction rune) error {
	return nil
}

type BatchStatusUpdateResult struct {
	OverallDeliveryStatus string `json:"overall_delivery_status"`
	ShouldBroadcast       bool   `json:"should_broadcast"`
}

// work in progress
func GetChatHistory(ctx context.Context, groupChatId string, limit int, offset time.Time) ([]any, error) {

	return nil, nil
}
