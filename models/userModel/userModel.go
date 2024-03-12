package usermodel

import (
	"time"
	"utils/helpers"
)

func NewUser(email string, username string, password string, geolocation string) (map[string]any, error) {
	user, err := helpers.QueryRowFields("SELECT * FROM new_user($1, $2, $34, $)", email, username, password, geolocation)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func GetUser(uniqueId string) (map[string]any, error) {
	user, err := helpers.QueryRowFields("SELECT * FROM get_user($1)", uniqueId)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func AccountExists(emailOrUsername string) (bool, error) {
	exist, err := helpers.QueryRowField[bool]("SELECT exist FROM account_exists($1)", emailOrUsername)
	if err != nil {
		return false, err
	}

	return *exist, nil
}

func FindNearbyUsers(clientId int, liveLocation string) ([]map[string]any, error) {
	nearbyUsers, err := helpers.QueryRowsFields("SELECT * FROM find_nearby_users($1, $2)", clientId, liveLocation)
	if err != nil {
		return nil, err
	}

	return nearbyUsers, nil
}

func SearchUser(clientId int, searchQuery string) ([]map[string]any, error) {
	matchUsers, err := helpers.QueryRowsFields("SELECT * FROM search_user($1, $2)", clientId, searchQuery)
	if err != nil {
		return nil, err
	}

	return matchUsers, nil
}

type User struct {
	Id int
}

func (user User) Edit(fieldValuePair [][]string) (map[string]any, error) {
	updatedUser, err := helpers.QueryRowFields("SELECT * FROM edit_user($1, $2)", user.Id, fieldValuePair)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (user User) SwitchPresence(presence string, last_seen time.Time) error {
	// go helpers.QueryRowField[bool](`SELECT switch_user_presence($1, $2, $3)`, user.Id, presence, last_seen)
	_, err := helpers.QueryRowField[bool](`SELECT switch_user_presence($1, $2, $3)`, user.Id, presence, last_seen)
	if err != nil {
		return err
	}

	return nil
}

func (user User) UpdateLocation(newGeolocation string) error {
	// go helpers.QueryRowField[bool](`SELECT update_user_location($1, $2, $3)`, user.Id, geolocation)
	_, err := helpers.QueryRowField[bool](`SELECT update_user_location($1, $2, $3)`, user.Id, newGeolocation)
	if err != nil {
		return err
	}

	return nil
}
