package appTypes

import (
	"bytes"
	"time"
)

type SignupSessionData struct {
	SessionId string
	Email     string
}

type ClientUser struct {
	Id       int
	Username string
}

type DMChatMsgAckData struct {
	MsgId    int
	DmChatId int
	SenderId int
	At       time.Time
}

type GroupChatMsgAckData struct {
	MsgId int
	At    time.Time
}

type WSResp struct {
	StatusCode int
	Body       any
	Error      string
}

type String string

func (s *String) UnmarshalJSON(b []byte) error {
	nb := bytes.Trim(b, "\"")

	*s = String(nb)

	return nil
}
