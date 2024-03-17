package userservice

import "model/usermodel"

func GetMyChats(clientId int) ([]*map[string]any, error) {
	return usermodel.GetMyChats(clientId)
}
