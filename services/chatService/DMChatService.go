package chatservice

func NewDMChat() {
}

type DMChat struct {
	Id int
}

func (dmc DMChat) SendMessage() {

}

func (dmc DMChat) GetChatHistory() {

}

type DMMessage struct {
	DMChatId int
	Id       int
}

func (dmm DMMessage) React() {

}

func (dmm DMMessage) UpdateDeliveryStatus() {

}
