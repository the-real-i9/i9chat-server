package user

import (
	"fmt"
	"i9chat/appGlobals"
	"i9chat/helpers"
	"log"

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

func New(email string, username string, password string, geolocation string) (*User, error) {

	user, err := helpers.QueryRowType[User]("SELECT * FROM new_user($1, $2, $3, $4)", email, username, password, geolocation)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: NewUser: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return user, nil
}

func FindOne(uniqueIdent string) (*User, error) {

	user, err := helpers.QueryRowType[User]("SELECT * FROM get_user($1)", uniqueIdent)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: FindOne: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return user, nil
}

func FindNearby(clientUserId int, liveLocation string) ([]*User, error) {

	nearbyUsers, err := helpers.QueryRowsType[User]("SELECT * FROM find_nearby_users($1, $2)", clientUserId, liveLocation)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: FindNearbyUsers: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return nearbyUsers, nil
}

func Search(clientUserId int, searchQuery string) ([]*User, error) {

	matchUsers, err := helpers.QueryRowsType[User]("SELECT * FROM search_user($1, $2)", clientUserId, searchQuery)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: Search: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return matchUsers, nil
}

func GetAll(clientUserId int) ([]*User, error) {

	allUsers, err := helpers.QueryRowsType[User]("SELECT * FROM get_all_users($1)", clientUserId)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: GetAll: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return allUsers, nil
}

func GetChats(userId int) ([]*map[string]any, error) {
	myChats, err := helpers.QueryRowsField[map[string]any]("SELECT chat FROM get_my_chats($1)", userId)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: GetChats: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return myChats, nil
}

func EditProfile(userId int, fieldValuePair [][]string) (*User, error) {

	updatedUser, err := helpers.QueryRowType[User]("SELECT * FROM edit_user($1, $2)", userId, fieldValuePair)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: EditProfile: %s", err))
		return nil, err
	}

	return updatedUser, nil
}

func GetPassword(uniqueIdent string) (string, error) {
	hashedPassword, err := helpers.QueryRowField[string]("SELECT password FROM get_user_password($1)", uniqueIdent)
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: GetPassword: %s", err))
		return "", appGlobals.ErrInternalServerError
	}

	return *hashedPassword, nil
}

func SwitchPresence(userId int, presence string, lastSeen pgtype.Timestamp) ([]*int, error) {

	userDMChatPartnersIdList, err := helpers.QueryRowsField[int](`SELECT * FROM switch_user_presence($1, $2, $3)`, userId, presence, lastSeen)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: SwitchPresence: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return userDMChatPartnersIdList, nil
}

func UpdateLocation(userId int, newGeolocation string) error {

	_, err := helpers.QueryRowField[bool]("SELECT update_user_location($1, $2)", userId, newGeolocation)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: UpdateLocation: %s", err))
		return appGlobals.ErrInternalServerError
	}

	return nil
}

func GetDMChatEventsPendingReceipt(userId int) ([]*map[string]any, error) {

	data, err := helpers.QueryRowsField[map[string]any]("SELECT event_data_kvp FROM get_dm_chat_events_pending_receipt($1)", userId)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: GetDMChatEventsPendingReceipt: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return data, nil
}

func GetGroupChatEventsPendingReceipt(userId int) ([]*map[string]any, error) {

	data, err := helpers.QueryRowsField[map[string]any]("SELECT event_data_kvp FROM get_group_chat_events_pending_receipt($1)", userId)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: GetGroupChatEventsPendingReceipt: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return data, nil
}

func GetDMChatMessageEventsPendingReceipt(userId int, dmChatId int) ([]*map[string]any, error) {

	data, err := helpers.QueryRowsField[map[string]any]("SELECT event_data_kvp FROM get_dm_chat_message_events_pending_receipt($1, $2)", userId, dmChatId)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: GetDMChatEventsPendingReceipt: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return data, nil
}

func GetGroupChatMessageEventsPendingReceipt(userId int, groupChatId int) ([]*map[string]any, error) {

	data, err := helpers.QueryRowsField[map[string]any]("SELECT event_data_kvp FROM get_group_chat_message_events_pending_receipt($1, $2)", userId, groupChatId)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: GetGroupChatEventsPendingReceipt: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return data, nil
}
