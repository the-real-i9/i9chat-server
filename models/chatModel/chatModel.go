package chatmodel

import "utils/helpers"

func GetMyChats(clientId int) ([]*map[string]any, error) {
	myChats, err := helpers.QueryRowsField[map[string]any]("SELECT chat FROM get_my_chats($1)", clientId)
	if err != nil {
		return nil, err
	}

	return myChats, nil
}

type Chat interface {
	SendMessage()
	GetChatHistory()
}

type Message interface {
	React()
}
