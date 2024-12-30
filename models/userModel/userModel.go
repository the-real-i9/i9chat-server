package user

import (
	"context"
	"fmt"
	"i9chat/helpers"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgtype"
)

type User struct {
	Id                int               `json:"id"`
	Username          string            `json:"username"`
	ProfilePictureUrl string            `db:"profile_picture_url" json:"profile_picture_url"`
	Presence          string            `json:"presence,omitempty"`
	LastSeen          *pgtype.Timestamp `db:"last_seen" json:"last_seen,omitempty"`
	Location          *pgtype.Circle    `json:"location,omitempty"`
}

func New(ctx context.Context, email string, username string, password string, geolocation string) (*User, error) {

	user, err := helpers.QueryRowType[User](ctx, "SELECT * FROM new_user($1, $2, $3, $4)", email, username, password, geolocation)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: NewUser: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	return user, nil
}

func FindOne(ctx context.Context, uniqueIdent string) (*User, error) {

	user, err := helpers.QueryRowType[User](ctx, "SELECT * FROM get_user($1)", uniqueIdent)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: FindOne: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	return user, nil
}

func FindNearby(ctx context.Context, clientUserId int, liveLocation string) ([]*User, error) {

	nearbyUsers, err := helpers.QueryRowsType[User](ctx, "SELECT * FROM find_nearby_users($1, $2)", clientUserId, liveLocation)

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

func ChangePresence(ctx context.Context, userId int, presence string, lastSeen time.Time) ([]*int, error) {

	userDMChatPartnersIdList, err := helpers.QueryRowsField[int](ctx, `SELECT * FROM change_user_presence($1, $2, $3)`, userId, presence, lastSeen)
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
