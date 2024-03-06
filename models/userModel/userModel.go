package usermodel

import (
	"database/sql"
	"time"
	"utils/helpers"
)

type User struct {
	Id              int
	Username        string
	Password        string
	Email           string
	Profile_picture string
	Presence        string
	Last_seen       sql.NullTime
	Created_at      time.Time `json:"-"`
	Deleted         bool      `json:"-"`
}

func UpdateUser(id int, fieldValuePair [][]string) (*User, error) {
	user, err := helpers.QueryRow[User]("SELECT * FROM edit_user($1, $2)", id, fieldValuePair)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func NewUser(email string, username string, password string) (*User, error) {
	user, err := helpers.QueryRow[User]("SELECT * FROM create_user($1, $2, $3)", email, username, password)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func GetUser(uniqueId string) (*User, error) {
	user, err := helpers.QueryRow[User]("SELECT * FROM get_user($1)", uniqueId)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func UserExists(emailOrUsername string) (bool, error) {
	type exist_t struct {
		Exist bool
	}

	row, err := helpers.QueryRow[exist_t]("SELECT exist FROM account_exists($1)", emailOrUsername)
	if err != nil {
		return false, err
	}

	return row.Exist, nil
}

type user_pres struct {
	Presence  string       `db:"out_presence"`
	Last_seen sql.NullTime `db:"out_last_seen"`
}

func SwitchUserPresence(userid int) (*user_pres, error) {
	row, err := helpers.QueryRow[user_pres]("SELECT * FROM switch_user_presence($1)", userid)
	if err != nil {
		return nil, err
	}

	return row, nil
}
