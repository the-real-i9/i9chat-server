package chatService

type Chat interface {
	SendMessage()
	GetChatHistory()
}

type Message interface {
	React()
	UpdateDeliveryStatus()
}
