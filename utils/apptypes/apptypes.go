package apptypes

import "time"

type SignupSessionData struct {
	SessionId string `json:"sessionId"`
	Email     string `json:"email"`
}

type JWTUserData struct {
	UserId   int    `json:"userId"`
	Username string `json:"username"`
}

type DMChatMsgDeliveryData struct {
	MsgId    int       `json:"msgId"`
	DmChatId int       `json:"dmChatId"`
	SenderId int       `json:"senderId"`
	At       time.Time `json:"at"`
}

type GroupChatMsgDeliveryData struct {
	MsgId int       `json:"msgId"`
	At    time.Time `json:"at"`
}
