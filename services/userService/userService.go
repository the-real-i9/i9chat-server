package userService

import (
	"fmt"
	"i9chat/models/userModel"
	"i9chat/utils/helpers"
	"log"
	"time"
)

type User struct {
	Id int
}

func (user User) GetAllUsers() ([]map[string]any, error) {
	return userModel.GetAllUsers(user.Id)
}

func (user User) GetMyChats() ([]*map[string]any, error) {
	return userModel.User{Id: user.Id}.GetMyChats()
}

func (user User) ChangeProfilePicture(pictureData []byte) error {
	// if any picture size validation error, do it here

	picPath := fmt.Sprintf("profile_pictures/profile_pic_%d.jpg", time.Now().UnixNano())

	newPicUrl, err := helpers.UploadFile(picPath, pictureData)
	if err != nil {
		log.Println(err)
		return err
	}

	_, ed_err := userModel.User{Id: user.Id}.Edit([][]string{{"profile_picture_url", newPicUrl}})
	if ed_err != nil {
		log.Println(fmt.Errorf("userService.go: ChangeProfilePicture: %s", ed_err))
		return helpers.ErrInternalServerError
	}

	return nil
}

func (user User) GetDMChatEventsPendingReceipt() ([]*map[string]any, error) {
	return userModel.User{Id: user.Id}.GetDMChatEventsPendingReceipt()
}

func (user User) GetGroupChatEventsPendingReceipt() ([]*map[string]any, error) {
	return userModel.User{Id: user.Id}.GetGroupChatEventsPendingReceipt()
}

func (user User) GetDMChatMessageEventsPendingReceipt(dmChatid int) ([]*map[string]any, error) {
	return userModel.User{Id: user.Id}.GetDMChatMessageEventsPendingReceipt(dmChatid)
}

func (user User) GetGroupChatMessageEventsPendingReceipt(groupChatId int) ([]*map[string]any, error) {
	return userModel.User{Id: user.Id}.GetGroupChatMessageEventsPendingReceipt(groupChatId)
}
