package userService

import (
	"fmt"
	"i9chat/appGlobals"
	"i9chat/helpers"
	user "i9chat/models/userModel"
	"log"
	"time"
)

func ChangeMyProfilePicture(clientUserId int, clientUsername string, pictureData []byte) error {
	// if any picture size validation error, do it here

	picPath := fmt.Sprintf("profile_pictures/%s/profile_pic_%d.jpg", clientUsername, time.Now().UnixNano())

	newPicUrl, err := helpers.UploadFile(picPath, pictureData)
	if err != nil {
		log.Println(err)
		return err
	}

	_, ed_err := user.EditProfile(clientUserId, [][]string{{"profile_picture_url", newPicUrl}})
	if ed_err != nil {
		log.Println(fmt.Errorf("userService.go: ChangeProfilePicture: %s", ed_err))
		return appGlobals.ErrInternalServerError
	}

	return nil
}
