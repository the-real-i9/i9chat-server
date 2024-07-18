package models_test

import (
	"encoding/json"
	"fmt"
	user "i9chat/models/userModel"
	"i9chat/utils/helpers"
	"testing"
)

func TestCreateNewUser(t *testing.T) {
	user, err := user.New("kenny@gmail.com", "i9x", "rubbishPassword", "(2, 5), 2")

	defer helpers.QueryRowFields("DELETE FROM i9c_user WHERE username = $1", "i9x")

	if err != nil {
		t.Error(err)
		return
	}

	d, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("%s\n", d)

}
