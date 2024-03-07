package chatservice

func GetChats(clientId int) map[string]any {
	return nil
}

func NewDMChat() {
}

func NewGroupChat() {
}

type Chat interface {
	SendMessage()
	GetChatHistory()
}

type DMChat struct {
	Id int
}

func (dmc DMChat) SendMessage() {

}

func (dmc DMChat) GetChatHistory() {

}

type GroupChat struct {
	Id int
}

func (gpc GroupChat) SendMessage() {

}

func (gpc GroupChat) GetChatHistory() {

}
