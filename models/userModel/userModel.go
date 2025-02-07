package user

import (
	"context"
	"fmt"
	"i9chat/appTypes"
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
			MATCH (u:User) WHERE username = $emailOrUsername OR email = $emailOrUsername
		} AS user_exists
		`,
		map[string]any{
			"emailOrUsername": emailOrUsername,
		},
	)
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: Exists: %s", err))
		return false, fiber.ErrInternalServerError
	}

	userExists, _, err := neo4j.GetRecordValue[bool](res.Records[0], "user_exists")
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: Exists: %s", err))
		return false, fiber.ErrInternalServerError
	}

	return userExists, nil
}

func New(ctx context.Context, email, username, password string, geolocation *appTypes.UserGeolocation) (map[string]any, error) {
	res, err := db.Query(
		ctx,
		`
		CREATE (u:User { email: $email, username: $username, password: $password, profile_pic_url: "", geolocation: point({ longitude: $long, latitude: $lat }) })
		WITH u, { longitude: toFloat(u.geolocation.longitude), latitude: toFloat(u.geolocation.latitude) } AS geolocation
		RETURN u { .username, .profile_pic_url, .presence, .last_seen, geolocation } AS new_user
		`,
		map[string]any{
			"email":    email,
			"username": username,
			"password": password,
			"long":     geolocation.Longitude,
			"lat":      geolocation.Latitude,
		},
	)
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: New: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	new_user, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "new_user")

	return new_user, nil
}

func FindOne(ctx context.Context, uniqueIdent string) (map[string]any, error) {
	res, err := db.Query(ctx,
		`
	MATCH (u:User) WHERE u.username = $uniqueIdent OR u.email = $uniqueIdent
	WITH u, { longitude: toFloat(u.geolocation.longitude), latitude: toFloat(u.geolocation.latitude) } AS geolocation
	RETURN u { .username, .profile_pic_url, .presence, .last_seen, .password, geolocation } AS found_user
	`,
		map[string]any{
			"uniqueIdent": uniqueIdent,
		},
	)
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: FindOne: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	found_user, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "found_user")

	return found_user, nil
}

func FindNearby(ctx context.Context, clientUsername string, long, lat, radius float64) ([]any, error) {
	res, err := db.Query(ctx,
		`
		MATCH (u:User)
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
		log.Println(fmt.Errorf("userModel.go: FindNearbyUsers: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	nearbyUsers, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "nearby_users")

	return nearbyUsers, nil
}

func Search(ctx context.Context, clientUsername, searchQuery string) ([]any, error) {
	res, err := db.Query(ctx,
		`
		MATCH (u:User)
		WHERE u.username <> $client_username AND $query <> "" AND lower($query) CONTAINS lower(u.username)
		WITH u, { longitude: toFloat(u.geolocation.longitude), latitude: toFloat(u.geolocation.latitude) } AS geolocation
		RETURN collect(u { .username, .profile_pic_url, .presence, .last_seen, .password, geolocation }) AS match_users
		`,
		map[string]any{
			"client_username": clientUsername,
			"query":           searchQuery,
		},
	)
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: Search: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	matchUsers, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "match_user")

	return matchUsers, nil
}

// work in progress
func GetChats(ctx context.Context, clientUsername string) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (clientChat:Chat{ owner_username: $client_username })-[:WITH_USER]->(partnerUser),
			(clientChat)<-[:IN_CHAT]-(lmsg:Message WHERE lmsg.created_at = clientChat.last_message_at),
			(clientChat)<-[:IN_CHAT]-(:Message)<-[lrxn:REACTS_TO_MESSAGE WHERE lrxn.at = clientChat.last_reaction_at]-(reactor)
		WITH clientChat, toString(clientChat.last_message_at) AS last_message_at, partnerUser { .username, .profile_pic_url, .connection_status } AS partner,
			CASE clientChat.last_activity_type 
				WHEN "message" THEN lmsg { type: "message", .content, .delivery_status }
				WHEN "reaction" THEN lrxn { type: "reaction", .reaction, reactor: reactor.username }
			END AS last_activity
		ORDER BY clientChat.last_message_at DESC
		RETURN collect(clientChat { partner, .unread_messages_count, last_message_at, last_activity }) AS my_chats
		`,
		map[string]any{
			"client_username": clientUsername,
		},
	)
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: GetChats: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	myChats, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "my_chats")

	return myChats, nil
}

func EditProfile(ctx context.Context, username string, fieldValueMap map[string]any) error {
	paramsMap := fieldValueMap

	setArgs := ""

	for k, _ := range paramsMap {
		if setArgs != "" {
			setArgs += ", "
		}

		setArgs = fmt.Sprintf("%s%s = $%[2]s", setArgs, k)
	}

	paramsMap["username"] = username

	_, err := db.Query(ctx,
		fmt.Sprintf(`
		MATCH (u:User{ username: $username })
		SET %s
		`, setArgs),
		paramsMap,
	)
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: EditProfile: %s", err))
		return fiber.ErrInternalServerError
	}

	return nil
}

// work in progress
func ChangePresence(ctx context.Context, clientUsername, presence string, lastSeen time.Time) ([]*int, error) {
	_, err := db.Query(ctx,
		`
		MATCH (user:User{ username: $client_username })
		SET user.presence = $presence, user.last_seen = $last_seen)
		`,
		map[string]any{
			"client_username": clientUsername,
			"presence":        presence,
			"last_seen":       nil,
		},
	)
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: ChangePresence: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	return nil, nil
}

func UpdateLocation(ctx context.Context, username string, newGeolocation *appTypes.UserGeolocation) error {
	_, err := db.Query(ctx,
		`
		MATCH (u:User{ username: $username })
		SET u.geolocation.longitude = $long, u.geolocation.latitude = $lat
		`,
		map[string]any{
			"username": username,
			"long":     newGeolocation.Longitude,
			"lat":      newGeolocation.Latitude,
		},
	)
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: UpdateLocation: %s", err))
		return fiber.ErrInternalServerError
	}

	return nil
}
