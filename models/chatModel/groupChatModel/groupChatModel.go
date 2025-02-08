package groupChat

import (
	"context"
	"fmt"
	"i9chat/appTypes"
	"i9chat/helpers"
	"i9chat/models/db"
	user "i9chat/models/userModel"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

type NewGroupChat struct {
	ClientData     map[string]any `json:"client_resp"`
	InitMemberData map[string]any `json:"member_resp"`
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
			group { group_chat_id: .id, .name, .description, .picture_url, last_activity: { type: "group activity", info: "You were added" } } AS member_resp
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
	ClientData      string   `json:"client_resp"`
	MemberData      string   `json:"member_resp"`
	MemberUsernames []string `json:"members_usernames"`
}

func ChangeName(ctx context.Context, groupId, clientUsername, newName string) (NewActivity, error) {
	var newActivity NewActivity

	res, err := db.Query(
		ctx,
		`
		MATCH (group)<-[:WITH_GROUP]-(clientChat:GroupChat{ owner_username: $client_username, group_id: $group_id })<-[:HAS_CHAT]-(clientUser),
			(clientUser)-[:IS_MEMBER_OF { role: "admin" }]->(group)
		CREATE (clientUser)-[:RECEIVES_ACTIVITY]->(cligact:GroupActivity{ info: "You changed group name from " + group.name + " to " + $new_name, created_at: datetime() })-[:IN_GROUP_CHAT]->(clientChat)

		WITH group, cligact
		MATCH (memberChat:GroupChat{ group_id: $group_id } WHERE memberChat.owner_username <> $client_username)<-[:HAS_CHAT]-(memberUser)
		CREATE (memberUser)-[:RECEIVES_ACTIVITY]->(memgact:GroupActivity{ info: $client_username + " changed group name from " + group.name + " to " + $new_name, created_at: datetime() })-[:IN_GROUP_CHAT]->(clientChat)

		RETURN cligact.info AS client_resp,
			memgact.info AS member_resp,
			collect(memberUser.username) AS member_usernames
		`,
		map[string]any{
			"client_username": clientUsername,
			"group_id":        groupId,
			"new_name":        newName,
		},
	)
	if err != nil {
		log.Println("groupChatModel.go: New:", err)
		return newActivity, fiber.ErrInternalServerError
	}

	helpers.MapToStruct(res.Records[0].AsMap(), &newActivity)

	return newActivity, nil
}

func ChangeDescription(ctx context.Context, groupChatId int, admin []string, newDescription string) (NewActivity, error) {
	newActivity, err := helpers.QueryRowType[NewActivity](ctx, "SELECT * change_group_description($1, $2, $3)", groupChatId, admin, newDescription)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: ChangeDescription: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	return newActivity, nil
}

func ChangePicture(ctx context.Context, groupChatId int, admin []string, newPictureUrl string) (NewActivity, error) {
	newActivity, err := helpers.QueryRowType[NewActivity](ctx, "SELECT * change_group_picture($1, $2, $3)", groupChatId, admin, newPictureUrl)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: ChangePicture: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	return newActivity, nil
}

func AddUsers(ctx context.Context, groupChatId int, admin []string, newUsers [][]appTypes.String) (NewActivity, error) {
	newActivity, err := helpers.QueryRowType[NewActivity](ctx, "SELECT * add_users_to_group($1, $2, $3)", groupChatId, admin, newUsers)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: AddUsers: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	return newActivity, nil
}

func RemoveUser(ctx context.Context, groupChatId int, admin []string, user []appTypes.String) (NewActivity, error) {
	newActivity, err := helpers.QueryRowType[NewActivity](ctx, "SELECT * remove_user_to_group($1, $2, $3)", groupChatId, admin, user)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: RemoveUser: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	return newActivity, nil
}

func Join(ctx context.Context, groupChatId int, newUser []string) (NewActivity, error) {
	newActivity, err := helpers.QueryRowType[NewActivity](ctx, "SELECT * join_group($1, $2)", groupChatId, newUser)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: Join: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	return newActivity, nil
}

func Leave(ctx context.Context, groupChatId int, user []string) (NewActivity, error) {
	newActivity, err := helpers.QueryRowType[NewActivity](ctx, "SELECT * leave_group($1, $2)", groupChatId, user)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: Leave: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	return newActivity, nil
}

func MakeUserAdmin(ctx context.Context, groupChatId int, admin []string, user []appTypes.String) (NewActivity, error) {
	newActivity, err := helpers.QueryRowType[NewActivity](ctx, "SELECT * make_user_group_admin($1, $2, $3)", groupChatId, admin, user)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: MakeUserAdmin: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	return newActivity, nil
}

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
		CREATE (message)-[:DELIVERED_TO { at: $delivered_at }]->(clientUser)
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
		CREATE (message)-[:READ_BY { at: $read_at } ]->(clientUser)
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
	OverallDeliveryStatus string `db:"overall_delivery_status"`
	ShouldBroadcast       bool   `db:"should_broadcast"`
}

// work in progress
func GetChatHistory(ctx context.Context, groupChatId string, limit int, offset time.Time) ([]any, error) {

	return nil, nil
}
