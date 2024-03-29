package usermodel

import (
	"fmt"
	"log"
	"time"
	"utils/helpers"
)

func NewUser(email string, username string, password string, geolocation string) (map[string]any, error) {

	user, err := helpers.QueryRowFields("SELECT * FROM new_user($1, $2, $3, $4)", email, username, password, geolocation)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: NewUser: %s", err))
		return nil, helpers.ErrInternalServerError
	}

	return user, nil
}

func GetUser(uniqueId string) (map[string]any, error) {

	user, err := helpers.QueryRowFields("SELECT * FROM get_user($1)", uniqueId)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: GetUser: %s", err))
		return nil, helpers.ErrInternalServerError
	}

	return user, nil
}

func FindNearbyUsers(clientId int, liveLocation string) ([]map[string]any, error) {

	nearbyUsers, err := helpers.QueryRowsFields("SELECT * FROM find_nearby_users($1, $2)", clientId, liveLocation)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: FindNearbyUsers: %s", err))
		return nil, helpers.ErrInternalServerError
	}

	return nearbyUsers, nil
}

func SearchUser(clientId int, searchQuery string) ([]map[string]any, error) {

	matchUsers, err := helpers.QueryRowsFields("SELECT * FROM search_user($1, $2)", clientId, searchQuery)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: SearchUser: %s", err))
		return nil, helpers.ErrInternalServerError
	}

	return matchUsers, nil
}

func GetMyChats(clientId int) ([]*map[string]any, error) {

	myChats, err := helpers.QueryRowsField[map[string]any]("SELECT chat FROM get_my_chats($1)", clientId)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: GetMyChats: %s", err))
		return nil, helpers.ErrInternalServerError
	}

	return myChats, nil
}

type User struct {
	Id int
}

func (user User) Edit(fieldValuePair [][]string) (map[string]any, error) {

	updatedUser, err := helpers.QueryRowFields("SELECT * FROM edit_user($1, $2)", user.Id, fieldValuePair)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: User_Edit: %s", err))
		return nil, err
	}

	return updatedUser, nil
}

func (user User) SwitchPresence(presence string, last_seen time.Time) error {

	_, err := helpers.QueryRowField[bool](`SELECT switch_user_presence($1, $2, $3)`, user.Id, presence, last_seen)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: User_SwitchPresence: %s", err))
		return helpers.ErrInternalServerError
	}

	return nil
}

func (user User) UpdateLocation(newGeolocation string) error {

	_, err := helpers.QueryRowField[bool]("SELECT update_user_location($1, $2, $3)", user.Id, newGeolocation)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: User_UpdateLocation: %s", err))
		return helpers.ErrInternalServerError
	}

	return nil
}

func (user User) GetDMChatEventsPendingDispatch() ([]*map[string]any, error) {

	data, err := helpers.QueryRowsField[map[string]any]("SELECT event_data_kvp FROM get_dm_chat_events_pending_dispatch($1)", user.Id)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: User_GetDMChatEventsPendingDispatch: %s", err))
		return nil, helpers.ErrInternalServerError
	}

	return data, nil
}

func (user User) GetGroupChatEventsPendingDispatch() ([]*map[string]any, error) {

	data, err := helpers.QueryRowsField[map[string]any]("SELECT event_data_kvp FROM get_group_chat_events_pending_dispatch($1)", user.Id)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: User_GetGroupChatEventsPendingDispatch: %s", err))
		return nil, helpers.ErrInternalServerError
	}

	return data, nil
}

func (user User) GetDMChatMessageEventsPendingDispatch(dmChatId int) ([]*map[string]any, error) {

	data, err := helpers.QueryRowsField[map[string]any]("SELECT event_data_kvp FROM get_dm_chat_message_events_pending_dispatch($1, $2)", user.Id, dmChatId)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: User_GetDMChatEventsPendingDispatch: %s", err))
		return nil, helpers.ErrInternalServerError
	}

	return data, nil
}

func (user User) GetGroupChatMessageEventsPendingDispatch(groupChatId int) ([]*map[string]any, error) {

	data, err := helpers.QueryRowsField[map[string]any]("SELECT event_data_kvp FROM get_group_chat_message_events_pending_dispatch($1, $2)", user.Id, groupChatId)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: User_GetGroupChatEventsPendingDispatch: %s", err))
		return nil, helpers.ErrInternalServerError
	}

	return data, nil
}
