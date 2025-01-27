package user

import (
	"context"
	"fmt"
	"i9chat/appGlobals"
	"i9chat/appTypes"
	"i9chat/helpers"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func Exists(ctx context.Context, emailOrUsername string) (bool, error) {
	res, err := neo4j.ExecuteQuery(ctx, appGlobals.Neo4jDriver,
		`
		RETURN EXISTS {
			MATCH (u:User) WHERE username = $emailOrUsername OR email = $emailOrUsername
		} AS user_exists
		`,
		map[string]any{
			"emailOrUsername": emailOrUsername,
		},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithReadersRouting(),
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

type User struct {
	Username                  string `json:"username"`
	ProfilePictureUrl         string `json:"profile_picture_url"`
	Presence                  string `json:"presence,omitempty"`
	LastSeen                  string `json:"last_seen,omitempty"`
	Password                  string `json:"_"`
	*appTypes.UserGeolocation `json:"geolocation,omitempty"`
}

func New(ctx context.Context, email, username, password string, geolocation *appTypes.UserGeolocation) (*User, error) {
	res, err := neo4j.ExecuteQuery(ctx, appGlobals.Neo4jDriver,
		`
	CREATE (u:User { email: $email, username: $username, password: $password, profile_pic_url: "", geolocation: point({ longitude: toFloat($long), latitude: toFloat($lat) }) })
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
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithWritersRouting(),
	)
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: New: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	new_user, _, err := neo4j.GetRecordValue[map[string]any](res.Records[0], "new_user")
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: New: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	var newUser User

	helpers.MapToStruct(new_user, &newUser)

	return &newUser, nil
}

func FindOne(ctx context.Context, uniqueIdent string) (*User, error) {
	res, err := neo4j.ExecuteQuery(ctx, appGlobals.Neo4jDriver,
		`
	MATCH (u:User) WHERE u.username = $uniqueIdent OR u.email = $uniqueIdent
	WITH u, { longitude: toFloat(u.geolocation.longitude), latitude: toFloat(u.geolocation.latitude) } AS geolocation
	RETURN u { .username, .profile_pic_url, .presence, .last_seen, .password, geolocation } AS found_user
	
	`,
		map[string]any{
			"uniqueIdent": uniqueIdent,
		},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithWritersRouting(),
	)
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: FindOne: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	found_user, notFound, err := neo4j.GetRecordValue[map[string]any](res.Records[0], "found_user")
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: FindOne: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	if notFound {
		return nil, nil
	}

	var foundUser User

	helpers.MapToStruct(found_user, &foundUser)

	return &foundUser, nil
}

func FindNearby(ctx context.Context, clientUsername string, liveLocation *appTypes.UserGeolocation, radius float64) ([]*User, error) {

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: FindNearbyUsers: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	return nearbyUsers, nil
}

func Search(ctx context.Context, clientUserId int, searchQuery string) ([]*User, error) {

	matchUsers, err := helpers.QueryRowsType[User](ctx, "SELECT * FROM search_user($1, $2)", clientUserId, searchQuery)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: Search: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	return matchUsers, nil
}

func GetAll(ctx context.Context, clientUserId int) ([]*User, error) {

	allUsers, err := helpers.QueryRowsType[User](ctx, "SELECT * FROM get_all_users($1)", clientUserId)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: GetAll: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	return allUsers, nil
}

func GetChats(ctx context.Context, userId int) ([]*map[string]any, error) {
	myChats, err := helpers.QueryRowsField[map[string]any](ctx, "SELECT chat FROM get_my_chats($1)", userId)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: GetChats: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	return myChats, nil
}

func EditProfile(ctx context.Context, userId int, fieldValuePair [][]string) (*User, error) {

	updatedUser, err := helpers.QueryRowType[User](ctx, "SELECT * FROM edit_user($1, $2)", userId, fieldValuePair)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: EditProfile: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	return updatedUser, nil
}

func GetPassword(ctx context.Context, uniqueIdent string) (string, error) {
	hashedPassword, err := helpers.QueryRowField[string](ctx, "SELECT password FROM get_user_password($1)", uniqueIdent)
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: GetPassword: %s", err))
		return "", fiber.ErrInternalServerError
	}

	return *hashedPassword, nil
}

func ChangePresence(ctx context.Context, clientUserId int, presence string, lastSeen time.Time) ([]*int, error) {

	userDMChatPartnersIdList, err := helpers.QueryRowsField[int](ctx, `SELECT * FROM change_user_presence($1, $2, $3)`, clientUserId, presence, lastSeen)
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: ChangePresence: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	return userDMChatPartnersIdList, nil
}

func UpdateLocation(ctx context.Context, userId int, newGeolocation string) error {

	_, err := helpers.QueryRowField[bool](ctx, "SELECT update_user_location($1, $2)", userId, newGeolocation)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: UpdateLocation: %s", err))
		return fiber.ErrInternalServerError
	}

	return nil
}
