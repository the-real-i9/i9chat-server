package userservice

import (
	"fmt"
	"log"
	"model/usermodel"
	"time"
	"utils/appglobals"
	"utils/helpers"
)

func GetMyChats(clientId int) ([]*map[string]any, error) {
	return usermodel.GetMyChats(clientId)
}

func ChangeProfilePicture(clientId int, picture []byte) error {
	// if any picture size validation error, do it here

	picPath := fmt.Sprintf("profile_pictures/user_%d_pic_%d.jpg", clientId, time.Now().UnixMilli())

	picUrl, err := helpers.UploadFile(picPath, picture)
	if err != nil {
		log.Println(err)
		return err
	}

	_, ed_err := usermodel.User{Id: clientId}.Edit([][]string{{"profile_picture", picUrl}})
	if ed_err != nil {
		log.Println(fmt.Errorf("userService.go: ChangeProfilePicture: %s", ed_err))
		return appglobals.ErrInternalServerError
	}

	return nil
}
