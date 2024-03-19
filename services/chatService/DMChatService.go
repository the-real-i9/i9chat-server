package chatservice

import (
	"fmt"
	"model/chatmodel"
	"time"
	"utils/appglobals"
	"utils/helpers"
)

func NewDMChat(initiatorId int, partnerId int, initMsgContent map[string]any, createdAt time.Time) (map[string]any, error) {
	var respData struct {
		Ird map[string]any // initiator_resp_data AS ird
		Prd map[string]any // partner_resp_data AS prd
	}

	data, err := chatmodel.NewDMChat(initiatorId, partnerId, initMsgContent, createdAt)
	if err != nil {
		return nil, err
	}

	helpers.MapToStruct(data, &respData)

	go appglobals.ChatObserver{}.Send(fmt.Sprintf("user-%d", partnerId), respData.Prd, "new")

	return respData.Ird, nil
}

type DMChat struct {
	Id int
}

func (dmc DMChat) SendMessage(senderId int, msgContent map[string]any, createdAt time.Time) (map[string]any, error) {
	var respData struct {
		Srd        map[string]any // sender_resp_data AS srd
		Rrd        map[string]any // receiver_resp_data AS rrd
		ReceiverId int
	}

	data, err := chatmodel.DMChat{Id: dmc.Id}.SendMessage(senderId, msgContent, createdAt)
	if err != nil {
		return nil, err
	}

	helpers.MapToStruct(data, &respData)

	go appglobals.DMChatMessageObserver{}.Send(
		fmt.Sprintf("user-%d--dmchat-%d", respData.ReceiverId, dmc.Id), respData.Rrd, "new",
	)
	go appglobals.ChatObserver{}.Send(
		fmt.Sprintf("user-%d", respData.ReceiverId), map[string]any{"dmChatId": dmc.Id, "event": "new message"}, "update",
	)

	return respData.Srd, nil
}

func (dmc DMChat) GetChatHistory(offset int) ([]*map[string]any, error) {
	return chatmodel.DMChat{Id: dmc.Id}.GetChatHistory(offset)
}

type DMMessage struct {
	DMChatId int
	Id       int
}

func (dmm DMMessage) React() {

}

func (dmm DMMessage) UpdateDeliveryStatus() {

}
