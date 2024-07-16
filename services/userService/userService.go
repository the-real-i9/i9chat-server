package userService

import (
	"fmt"
	user "i9chat/models/userModel"
	"i9chat/services/appObservers"
	"i9chat/utils/helpers"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func SwitchMyPresence(clientUserId int, presence string, lastSeen pgtype.Timestamp) error {
	if err := user.SwitchPresence(clientUserId, presence, lastSeen); err != nil {
		return err
	}

	go func() {
		// "recepients" are: all users to whom I am a DMChat partner
		recepientIds, err := helpers.QueryRowsField[int]("SELECT user_id FROM user_dm_chat WHERE partner_id = $1", clientUserId)
		if err != nil {
			return
		}

		for _, rId := range recepientIds {
			rId := *rId
			go appObservers.DMChatObserver{}.SendPresenceUpdate(
				fmt.Sprintf("user-%d", rId), map[string]any{
					"userId":   clientUserId,
					"presence": presence,
					"lastSeen": lastSeen,
				}, "user presence update",
			)
		}
	}()

	return nil
}

func ChangeMyProfilePicture(clientUserId int, pictureData []byte) error {
	// if any picture size validation error, do it here

	picPath := fmt.Sprintf("profile_pictures/profile_pic_%d.jpg", time.Now().UnixNano())

	newPicUrl, err := helpers.UploadFile(picPath, pictureData)
	if err != nil {
		log.Println(err)
		return err
	}

	_, ed_err := user.EditProfile(clientUserId, [][]string{{"profile_picture_url", newPicUrl}})
	if ed_err != nil {
		log.Println(fmt.Errorf("userService.go: ChangeProfilePicture: %s", ed_err))
		return helpers.ErrInternalServerError
	}

	return nil
}
