package chatservice

import "model/chatmodel"

func NewDMChat() {
}

type DMChat struct {
	Id int
}

func (dmc DMChat) SendMessage() {

}

func (dmc DMChat) GetChatHistory(offset int) ([]*map[string]any, error) {
	dmChat := chatmodel.DMChat{Id: dmc.Id}
	return dmChat.GetChatHistory(offset)
}

type DMMessage struct {
	DMChatId int
	Id       int
}

func (dmm DMMessage) React() {

}

func (dmm DMMessage) UpdateDeliveryStatus() {

}
