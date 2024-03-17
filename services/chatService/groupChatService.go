package chatservice

func NewGroupChat() {
}

type GroupChat struct {
	Id int
}

func (gpc GroupChat) SendMessage() {

}

func (gpc GroupChat) GetChatHistory() {

}

type GroupMessage struct {
	GroupChatId int
	Id          int
}

func (gpcm GroupMessage) React() {

}

func (gpcm GroupMessage) UpdateDeliveryStatus() {

}
