package user

import (
	"context"
	"fmt"
	"i9chat/src/appTypes"
	"i9chat/src/helpers"
	"i9chat/src/models/db"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func Exists(ctx context.Context, emailOrUsername string) (bool, error) {
	res, err := db.Query(ctx,
		`
		RETURN EXISTS {
			MATCH (u:User) WHERE u.username = $emailOrUsername OR u.email = $emailOrUsername
		} AS user_exists
		`,
		map[string]any{
			"emailOrUsername": emailOrUsername,
		},
	)
	if err != nil {
		log.Println("userModel.go: Exists:", err)
		return false, fiber.ErrInternalServerError
	}

	userExists, _, err := neo4j.GetRecordValue[bool](res.Records[0], "user_exists")
	if err != nil {
		log.Println("userModel.go: Exists:", err)
		return false, fiber.ErrInternalServerError
	}

	return userExists, nil
}

func New(ctx context.Context, email, username, password string) (map[string]any, error) {
	res, err := db.Query(
		ctx,
		`
		CREATE (u:User { email: $email, username: $username, password: $password, profile_pic_url: "", presence: "online" })
		WITH u, toString(u.last_seen) AS last_seen
		RETURN u { .username, .email, .profile_pic_url, .presence, last_seen } AS new_user
		`,
		map[string]any{
			"email":    email,
			"username": username,
			"password": password,
		},
	)
	if err != nil {
		log.Println("userModel.go: New:", err)
		return nil, fiber.ErrInternalServerError
	}

	new_user, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "new_user")

	return new_user, nil
}

func SigninFind(ctx context.Context, uniqueIdent string) (map[string]any, error) {
	res, err := db.Query(ctx,
		`
	MATCH (u:User)
	WHERE u.username = $uniqueIdent OR u.email = $uniqueIdent

	WITH u, toString(u.last_seen) AS last_seen
	RETURN u { .username, .email, .profile_pic_url, .presence, last_seen, .password } AS found_user
	`,
		map[string]any{
			"uniqueIdent": uniqueIdent,
		},
	)
	if err != nil {
		log.Println("userModel.go: SigninFind:", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	found_user, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "found_user")

	return found_user, nil
}

func SessionFind(ctx context.Context, username string) (map[string]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (u:User{ username: $username })

		WITH u, toString(u.last_seen) AS last_seen
		RETURN u { .username, .email, .profile_pic_url, .presence, last_seen } AS found_user
		`,
		map[string]any{
			"username": username,
		},
	)
	if err != nil {
		log.Println("userModel.go: SessionFind:", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	found_user, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "found_user")

	return found_user, nil
}

func ChangePassword(ctx context.Context, email, newPassword string) error {
	_, err := db.Query(
		ctx,
		`
		MATCH (user:User{ email: $email })
		SET user.password = $newPassword
		`,
		map[string]any{
			"email":       email,
			"newPassword": newPassword,
		},
	)
	if err != nil {
		log.Println("userModel.go: ChangePassword:", err)
		return fiber.ErrInternalServerError
	}

	return nil
}

func FindNearby(ctx context.Context, clientUsername string, x, y, radius float64) ([]any, error) {
	res, err := db.Query(ctx,
		`
		MATCH (u:User)
		WHERE u.username <> $client_username AND point.distance(point({ x: $live_long, y: $live_lat, crs: "WGS-84-2D" }), u.geolocation) <= $radius

		WITH u, toString(u.last_seen) AS last_seen
		RETURN collect(u { .username, .email, .profile_pic_url, .presence, last_seen }) AS nearby_users
	`,
		map[string]any{
			"client_username": clientUsername,
			"live_long":       x,
			"live_lat":        y,
			"radius":          radius,
		},
	)
	if err != nil {
		log.Println("userModel.go: FindNearbyUsers:", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	nearbyUsers, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "nearby_users")

	return nearbyUsers, nil
}

func FindOne(ctx context.Context, emailUsername string) (map[string]any, error) {
	res, err := db.Query(ctx,
		`
		MATCH (u:User)
		WHERE u.username = $eup OR u.email = $eup

		WITH u, toString(u.last_seen) AS last_seen
		RETURN u { .username, .email, .profile_pic_url, .presence, last_seen } AS found_user
		`,
		map[string]any{
			"eup": emailUsername,
		},
	)
	if err != nil {
		log.Println("userModel.go: FindOne:", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	foundUser, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "found_user")

	return foundUser, nil
}

type Chat struct {
	ChatType  string `json:"chat_type"`
	ChatIdent string `json:"chat_ident"`
	UnreadMC  int    `json:"unread_messages_count"`

	// for dm chat
	Partner map[string]any `json:"partner,omitempty"`

	// for group chat
	GroupInfo map[string]any `json:"group_info,omitempty"`
}

func GetMyChats(ctx context.Context, clientUsername string) ([]Chat, error) {
	var myChats []Chat

	res, err := db.Query(
		ctx,
		`
		CALL () {
			MATCH (clientChat:DMChat{ owner_username: $client_username })-[:WITH_USER]->(partnerUser)

			WITH clientChat, 
				toString(clientChat.updated_at) AS updated_at, 
				partnerUser { .username, .profile_pic_url, .presence, .last_seen } AS partner, 
				partnerUser.username AS chat_ident
				
			RETURN clientChat { chat_ident, partner, .unread_messages_count, chat_type: "DM" } AS chat, clientChat.updated_at AS updated_at
		UNION
			MATCH (clientChat:GroupChat{ owner_username: $client_username })-[:WITH_GROUP]->(group)

			WITH clientChat, 
				toString(clientChat.updated_at) AS updated_at, 
				group { .name, .description, .picture_url } AS group_info, 
				group.id AS chat_ident

			RETURN clientChat { chat_ident, group_info, .unread_messages_count, chat_type: "group" } AS chat, clientChat.updated_at AS updated_at
		}
		WITH chat, updated_at
		ORDER BY updated_at DESC

		RETURN collect(chat) AS my_chats
		`,
		map[string]any{
			"client_username": clientUsername,
		},
	)
	if err != nil {
		log.Println("userModel.go: GetChats:", err)
		return myChats, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return myChats, nil
	}

	mc, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "my_chats")

	helpers.ToStruct(mc, &myChats)

	return myChats, nil
}

func GetMyProfile(ctx context.Context, clientUsername string) (map[string]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (u:User{ username: $client_username })
		WITH u, { x: toFloat(u.geolocation.x), y: toFloat(u.geolocation.y) } AS geolocation, toString(u.last_seen) AS last_seen
		RETURN u { .username, .email, .profile_pic_url, .presence, last_seen, geolocation } AS my_profile
		`,
		map[string]any{
			"client_username": clientUsername,
		},
	)
	if err != nil {
		log.Println("userModel.go: GetChats:", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	mp, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "my_profile")

	return mp, nil
}

func ChangeProfilePicture(ctx context.Context, clientUsername, newPicUrl string) error {
	_, err := db.Query(ctx,
		`
		MATCH (u:User{ username: $client_username })
		SET u.profile_pic_url = $new_pic_url
		`,
		map[string]any{
			"client_username": clientUsername,
			"new_pic_url":     newPicUrl,
		},
	)
	if err != nil {
		log.Println("userModel.go: ChangeProfilePicture:", err)
		return fiber.ErrInternalServerError
	}

	return nil
}

func ChangePresence(ctx context.Context, clientUsername, presence string, lastSeen time.Time) ([]any, error) {
	var lastSeenVal string
	if presence == "online" {
		lastSeenVal = "null"
	} else {
		lastSeenVal = "$last_seen"
	}
	res, err := db.Query(ctx,
		fmt.Sprintf(`
		MATCH (user:User{ username: $client_username })
		SET user.presence = $presence, user.last_seen = %s

		WITH user
		OPTIONAL MATCH (user)-[:HAS_DM_CHAT]->()-[:WITH_USER]->(partnerUser)
		
		RETURN collect(partnerUser.username) AS partner_usernames
		`, lastSeenVal),
		map[string]any{
			"client_username": clientUsername,
			"presence":        presence,
			"last_seen":       lastSeen,
		},
	)
	if err != nil {
		log.Println("userModel.go: ChangePresence:", err)
		return nil, err
	}

	pus, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "partner_usernames")

	return pus, nil
}

func UpdateLocation(ctx context.Context, clientUsername string, newGeolocation appTypes.UserGeolocation) error {
	_, err := db.Query(ctx,
		`
		MATCH (u:User{ username: $client_username })
		SET u.geolocation = point({ x: $x, y: $y, crs: "WGS-84-2D" })
		`,
		map[string]any{
			"client_username": clientUsername,
			"x":               newGeolocation.X,
			"y":               newGeolocation.Y,
		},
	)
	if err != nil {
		log.Println("userModel.go: UpdateLocation:", err)
		return fiber.ErrInternalServerError
	}

	return nil
}
