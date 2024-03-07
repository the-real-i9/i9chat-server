package usermodel

import (
	"time"
	"utils/helpers"
)

func CreateUser(email string, username string, password string) (map[string]any, error) {
	user, err := helpers.QueryRowFields("SELECT * FROM create_user($1, $2, $3)", email, username, password)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func QueryUser(uniqueId string) (map[string]any, error) {
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

type User struct {
	Id int
}

func (user User) Edit(fieldValuePair [][]string) (map[string]any, error) {
	res, err := helpers.QueryRowFields("SELECT * FROM edit_user($1, $2)", user.Id, fieldValuePair)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (user User) SwitchPresence(presence string, last_seen time.Time) error {
	_, err := helpers.QueryRowFields(`SELECT switch_user_presence($1, $2, $3)`, user.Id, presence, last_seen)
	if err != nil {
		return err
	}

	return nil
}
