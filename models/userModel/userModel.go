package user

import (
	"context"
	"i9chat/appTypes"
	"i9chat/helpers"
	"i9chat/models/db"
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

func New(ctx context.Context, email, username, phone, password string, geolocation appTypes.UserGeolocation) (map[string]any, error) {
	res, err := db.Query(
		ctx,
		`
		CREATE (u:User { email: $email, username: $username, phone: $phone, password: $password, profile_pic_url: "", geolocation: point({ longitude: $long, latitude: $lat }), presence: "offline" })
		WITH u, { longitude: toFloat(u.geolocation.longitude), latitude: toFloat(u.geolocation.latitude) } AS geolocation
		RETURN u { .username, .profile_pic_url, .presence, .last_seen, geolocation } AS new_user
		`,
		map[string]any{
			"email":    email,
			"username": username,
			"phone":    phone,
			"password": password,
			"long":     geolocation.Longitude,
			"lat":      geolocation.Latitude,
		},
	)
	if err != nil {
		log.Println("userModel.go: New:", err)
		return nil, fiber.ErrInternalServerError
	}

	new_user, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "new_user")

	return new_user, nil
}

func FindOne(ctx context.Context, uniqueIdent string) (map[string]any, error) {
	res, err := db.Query(ctx,
		`
	OPTIONAL MATCH (u:User) WHERE u.username = $uniqueIdent OR u.email = $uniqueIdent OR  u.phone = $uniqueIdent
	WITH u, { longitude: toFloat(u.geolocation.longitude), latitude: toFloat(u.geolocation.latitude) } AS geolocation
	RETURN u { .username, .profile_pic_url, .presence, .last_seen, .password, geolocation } AS found_user
	`,
		map[string]any{
			"uniqueIdent": uniqueIdent,
		},
	)
	if err != nil {
		log.Println("userModel.go: FindOne:", err)
		return nil, fiber.ErrInternalServerError
	}

	found_user, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "found_user")

	return found_user, nil
}

func FindNearby(ctx context.Context, clientUsername string, long, lat, radius float64) ([]any, error) {
	res, err := db.Query(ctx,
		`
		OPTIONAL MATCH (u:User)
		WHERE u.username <> $client_username AND point.distance(point({ longitude: $live_long, latitude: $live_lat }), u.geolocation) <= $radius
		WITH u, { longitude: toFloat(u.geolocation.longitude), latitude: toFloat(u.geolocation.latitude) } AS geolocation
		RETURN collect(u { .username, .profile_pic_url, .presence, .last_seen, .password, geolocation }) AS nearby_users
	`,
		map[string]any{
			"client_username": clientUsername,
			"live_long":       long,
			"live_lat":        lat,
			"radius":          radius,
		},
	)
	if err != nil {
		log.Println("userModel.go: FindNearbyUsers:", err)
		return nil, fiber.ErrInternalServerError
	}

	nearbyUsers, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "nearby_users")

	return nearbyUsers, nil
}

func Find(ctx context.Context, emailUsernamePhone string) (map[string]any, error) {
	res, err := db.Query(ctx,
		`
		OPTIONAL MATCH (u:User)
		WHERE u.username = $eup OR u.email = $eup OR u.phone = $eup
		WITH u, { longitude: toFloat(u.geolocation.longitude), latitude: toFloat(u.geolocation.latitude) } AS geolocation
		RETURN u { .username, .email, .phone, .profile_pic_url, .presence, .last_seen, geolocation } AS found_user
		`,
		map[string]any{
			"eup": emailUsernamePhone,
		},
	)
	if err != nil {
		log.Println("userModel.go: Find:", err)
		return nil, fiber.ErrInternalServerError
	}

	foundUser, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "found_user")

	return foundUser, nil
}

type ChatItem struct {
	ChatItemType string         `json:"chat_item_type"`
	UnreadMC     int            `json:"unread_messages_count"`
	UpdatedAt    string         `json:"updated_at"`
	LastActivity map[string]any `json:"last_activity"`

	// for dm chat
	Partner map[string]any `json:"partner,omitempty"`

	// for group chat
	GroupInfo map[string]any `json:"group_info,omitempty"`
}

func GetMyChats(ctx context.Context, clientUsername string) ([]ChatItem, error) {
	var myChats []ChatItem

	res, err := db.Query(
		ctx,
		`
		CALL () {
			MATCH (clientChat:DMChat{ owner_username: $client_username })-[:WITH_USER]->(partnerUser),
				(clientChat)<-[:IN_DM_CHAT]-(lmsg:DMMessage WHERE lmsg.id = clientChat.last_message_id)
			OPTIONAL MATCH (clientChat)<-[:IN_DM_CHAT]-(:DMMessage)<-[lrxn:REACTS_TO_MESSAGE WHERE lrxn.at = clientChat.last_reaction_at]-(reactor)

			WITH clientChat, toString(clientChat.updated_at) AS updated_at, partnerUser { .username, .profile_pic_url, .connection_status } AS partner,
				CASE clientChat.last_activity_type 
					WHEN "message" THEN lmsg { type: "message", .content, .delivery_status }
					WHEN "reaction" THEN lrxn { type: "reaction", .reaction, reactor: reactor.username }
				END AS last_activity
			RETURN clientChat { partner, .unread_messages_count, updated_at, last_activity, chat_item_type: "dm" } AS chat_item, clientChat.updated_at AS updated_at
		UNION
			MATCH (clientChat:GroupChat{ owner_username: $client_username })-[:WITH_GROUP]->(group)
			OPTIONAL MATCH (clientChat)<-[:IN_GROUP_CHAT]-(lmsg:GroupMessage WHERE lmsg.id = clientChat.last_message_id),
				(clientChat)<-[:IN_GROUP_CHAT]-(:GroupMessage)<-[lrxn:REACTS_TO_MESSAGE WHERE lrxn.at = clientChat.last_reaction_at]-(reactor),
				(clientChat)<-[:IN_GROUP_CHAT]-(lgact:GroupActivity WHERE lgact.created_at = clientChat.last_group_activity_at)
			WITH clientChat, toString(clientChat.updated_at) AS updated_at, group { .name, .picture_url } AS group_info,
				CASE clientChat.last_activity_type
					WHEN "message" THEN lmsg { type: "message", .content, .delivery_status }
					WHEN "reaction" THEN lrxn { type: "reaction", .reaction, reactor: reactor.username }
					WHEN "group activity" THEN lgact { type: "group activity", .info }
				END AS last_activity
			RETURN clientChat { group_info, .unread_messages_count, updated_at, last_activity, chat_item_type: "group" } AS chat_item, clientChat.updated_at AS updated_at
		}
		WITH chat_item, updated_at
		ORDER BY updated_at DESC

		RETURN collect(chat_item) AS my_chats
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

	helpers.AnyToStruct(mc, &myChats)

	return myChats, nil
}

func GetMyProfile(ctx context.Context, clientUsername string) (map[string]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (u:User{ username: $client_username })
		WITH u, { longitude: toFloat(u.geolocation.longitude), latitude: toFloat(u.geolocation.latitude) } AS geolocation
		RETURN u { .username, .email, .phone, .profile_pic_url, .presence, .last_seen, geolocation } AS my_profile
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

func ChangePhone(ctx context.Context, clientUsername, newPhone string) error {
	_, err := db.Query(ctx,
		`
		MATCH (u:User{ username: $client_username })
		SET u.phone = $new_phone
		`,
		map[string]any{
			"client_username": clientUsername,
			"new_phone":       newPhone,
		},
	)
	if err != nil {
		log.Println("userModel.go: ChangePhone:", err)
		return fiber.ErrInternalServerError
	}

	return nil
}

func ChangePresence(ctx context.Context, clientUsername, presence string, lastSeen time.Time) []any {
	res, err := db.Query(ctx,
		`
		MATCH (user:User{ username: $client_username })
		SET user.presence = $presence, user.last_seen = $last_seen)

		OPTIONAL MATCH (user)-[:HAS_DM_CHAT]->()-[:WITH_USER]->(partnerUser)
		RETURN collect(partnerUser.username) AS partner_usernames
		`,
		map[string]any{
			"client_username": clientUsername,
			"presence":        presence,
			"last_seen":       nil,
		},
	)
	if err != nil {
		log.Println("userModel.go: ChangePresence:", err)
	}

	pus, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "partner_usernames")

	return pus
}

func UpdateLocation(ctx context.Context, clientUsername string, newGeolocation appTypes.UserGeolocation) error {
	_, err := db.Query(ctx,
		`
		MATCH (u:User{ username: $client_username })
		SET u.geolocation = point({ longitude: $long, latitude: $lat })
		`,
		map[string]any{
			"client_username": clientUsername,
			"long":            newGeolocation.Longitude,
			"lat":             newGeolocation.Latitude,
		},
	)
	if err != nil {
		log.Println("userModel.go: UpdateLocation:", err)
		return fiber.ErrInternalServerError
	}

	return nil
}
