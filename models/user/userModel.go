package user

import (
	"utils/mytypes"
)

type User struct {
	Id              int            `json:"user_id"`
	Username        string         `json:"username"`
	Password        string         `json:"password"`
	Email           string         `json:"email"`
	Profile_picture string         `json:"profile_picture"`
	Presence        string         `json:"presence"`
	Last_seen       mytypes.PgTime `json:"last_seen"`
}

func NewUser(username string, password string, email string, profile_picture string) (*User, error) {
	// create a new user with RETURNING data
	var newUser User

	// UnMarshal the data into newUser
	return &newUser, nil
}

func GetUserById(id int) (*User, error) {

	var user User

	// find user by id
	// UnMarshalJSON into user and return it
	return &user, nil
}

func GetUserByEmail(email string) (*User, error) {

	var user User

	// find user by email
	// UnMarshalJSON into user and return it
	return &user, nil
}
