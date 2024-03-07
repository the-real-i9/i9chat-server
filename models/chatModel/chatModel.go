package chatmodel

func QueryChats(clientId int) {

}

type Chat interface {
	CreateMessage()
	QueryChatHistory()
}

type DMChat struct {
	Id int
}

func CreateDMChat() {
}

func (dmc DMChat) CreateMessage() {
}

func (dmc DMChat) QueryChatHistory() {
}

type GroupChat struct {
	Id int
}

func CreateGroupChat() {
}

func (gpc GroupChat) CreateMessage() {

}

func (gpc GroupChat) QueryChatHistory() {

}
