package user

import (
	"context"
	"fmt"
	"i9chat/src/appTypes"
	"i9chat/src/helpers"
	"i9chat/src/models/db"
	"i9chat/src/models/modelHelpers"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func Exists(ctx context.Context, emailOrUsername string) (bool, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		RETURN EXISTS {
			MATCH (u:User) WHERE u.username = $emailOrUsername OR u.email = $emailOrUsername
		} AS user_exists
		`,
		map[string]any{
			"emailOrUsername": emailOrUsername,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return false, fiber.ErrInternalServerError
	}

	userExists := modelHelpers.RKeyGet[bool](res.Records, "user_exists")

	return userExists, nil
}

type NewUserT struct {
	Email         string `json:"email"`
	Username      string `json:"username"`
	ProfilePicUrl string `json:"profile_pic_url" db:"profile_pic_url"`
	Bio           string `json:"bio"`
}

func New(ctx context.Context, email, username, password string) (newUser NewUserT, err error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		CREATE (u:User { email: $email, username: $username, password: $password, profile_pic_url: "{notset}", presence: "online", bio: "i9chat is Awesome!" })
		RETURN u { .username, .email, .profile_pic_url, .bio } AS new_user
		`,
		map[string]any{
			"email":    email,
			"username": username,
			"password": password,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return newUser, fiber.ErrInternalServerError
	}

	newUser = modelHelpers.RKeyGet[NewUserT](res.Records, "new_user")

	return newUser, nil
}

type ToAuthUserT struct {
	Email         string `json:"email"`
	Username      string `json:"username"`
	ProfilePicUrl string `json:"profile_pic_url" db:"profile_pic_url"`
	Password      string `json:"-"`
}

func AuthFind(ctx context.Context, uniqueIdent string) (user ToAuthUserT, err error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
	MATCH (u:User)
	WHERE u.username = $uniqueIdent OR u.email = $uniqueIdent

	RETURN u { .email, .username, .profile_pic_url, .password } AS found_user
	`,
		map[string]any{
			"uniqueIdent": uniqueIdent,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return user, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return user, nil
	}

	user = modelHelpers.RKeyGet[ToAuthUserT](res.Records, "found_user")

	return user, nil
}

func ChangePassword(ctx context.Context, email, newPassword string) (bool, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (user:User{ email: $email })
		SET user.password = $newPassword

		RETURN true AS done
		`,
		map[string]any{
			"email":       email,
			"newPassword": newPassword,
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

func FindNearby(ctx context.Context, clientUsername string, x, y, radius float64) ([]any, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (u:User)
		WHERE u.username <> $client_username AND point.distance(point({ x: $live_long, y: $live_lat, crs: "WGS-84" }), u.geolocation) <= $radius

		RETURN collect(u { .username, .profile_pic_url, .bio, .presence, last_seen }) AS nearby_users
	`,
		map[string]any{
			"client_username": clientUsername,
			"live_long":       x,
			"live_lat":        y,
			"radius":          radius,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	nearbyUsers, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "nearby_users")

	return nearbyUsers, nil
}

type Chat struct {
	ChatType  string `json:"chat_type"`
	ChatIdent string `json:"chat_ident"`
	UnreadMC  int    `json:"unread_messages_count"`

	// for direct chat
	Partner map[string]any `json:"partner,omitempty"`

	// for group chat
	GroupInfo map[string]any `json:"group_info,omitempty"`
}

func GetMyChats(ctx context.Context, clientUsername string) ([]Chat, error) {
	// serve from cache
	var myChats []Chat

	res, err := db.Query(
		ctx,
		`/*cypher*/
		CALL () {
			MATCH (clientChat:DirectChat{ owner_username: $client_username })-[:WITH_USER]->(partnerUser)

			WITH clientChat, 
				partnerUser { .username, .profile_pic_url, .presence, .last_seen } AS partner, 
				partnerUser.username AS chat_ident
				
			RETURN clientChat { chat_ident, partner, .unread_messages_count, chat_type: "direct" } AS chat
		UNION
			MATCH (clientChat:GroupChat{ owner_username: $client_username })-[:WITH_GROUP]->(group)

			WITH clientChat, 
				group { .id, .name, .description, .picture_url } AS group_info, 
				group.id AS chat_ident

			RETURN clientChat { chat_ident, group_info, .unread_messages_count, chat_type: "group" } AS chat
		}
		WITH chat

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
	// serve from redis instead
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (u:User{ username: $client_username })
		WITH u, { x: toFloat(u.geolocation.x), y: toFloat(u.geolocation.y) } AS geolocation, coalesce(u.last_seen.epochMillis, null) AS last_seen
		RETURN u { .username, .email, .profile_pic_url, .bio, .presence, last_seen, geolocation } AS my_profile
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

func ChangeProfilePicture(ctx context.Context, clientUsername, newPicUrl string) (bool, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (u:User{ username: $client_username })
		SET u.profile_pic_url = $new_pic_url

		RETURN true AS done
		`,
		map[string]any{
			"client_username": clientUsername,
			"new_pic_url":     newPicUrl,
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

func ChangeBio(ctx context.Context, clientUsername, newBio string) (bool, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (u:User{ username: $client_username })
		SET u.bio = $new_bio

		RETURN true AS done
		`,
		map[string]any{
			"client_username": clientUsername,
			"new_bio":         newBio,
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

func ChangePresence(ctx context.Context, clientUsername, presence string, lastSeen int64) bool {
	var lastSeenVal string
	if presence == "online" {
		lastSeenVal = "null"
	} else {
		lastSeenVal = "$last_seen"
	}
	res, err := db.Query(
		ctx,
		fmt.Sprintf(`/*cypher*/
		MATCH (user:User{ username: $client_username })
		SET user.presence = $presence, user.last_seen = %s

		RETURN true AS done
		`, lastSeenVal),
		map[string]any{
			"client_username": clientUsername,
			"presence":        presence,
			"last_seen":       lastSeen,
		},
	)
	if err != nil {
		helpers.LogError(err)
		return false
	}

	if len(res.Records) == 0 {
		return false
	}

	return true
}

func SetLocation(ctx context.Context, clientUsername string, newGeolocation appTypes.UserGeolocation) (bool, error) {
	res, err := db.Query(
		ctx,
		`/*cypher*/
		MATCH (u:User{ username: $client_username })
		SET u.geolocation = point({ x: $x, y: $y, crs: "WGS-84" })

		RETURN true AS done
		`,
		map[string]any{
			"client_username": clientUsername,
			"x":               newGeolocation.X,
			"y":               newGeolocation.Y,
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
