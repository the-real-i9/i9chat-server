package chatservice

import (
	"fmt"
	"model/chatmodel"
	"utils/appglobals"
	"utils/helpers"
)

func NewDMChat(initiatorId int, partnerId int, initMsgContent map[string]any) (map[string]any, error) {
	var respData struct {
		Ird map[string]any // initiator_resp_data AS ird
		Prd map[string]any // partner_resp_data AS prd
	}

	data, err := chatmodel.NewDMChat(initiatorId, partnerId, initMsgContent)
	if err != nil {
		return nil, err
	}

	helpers.MapToStruct(data, &respData)

	go appglobals.SendNewChatUpdate(fmt.Sprintf("user-%d", partnerId), respData.Prd)

	return respData.Ird, nil
}

type DMChat struct {
	Id int
}

func (dmc DMChat) SendMessage(senderId int, msgContent map[string]any) (map[string]any, error) {
	var respData struct {
		Srd map[string]any // sender_resp_data AS srd
		Rrd map[string]any // receiver_resp_data AS rrd
	}

	dmChat := chatmodel.DMChat{Id: dmc.Id}

	data, err := dmChat.SendMessage(senderId, msgContent)
	if err != nil {
		return nil, err
	}

	helpers.MapToStruct(data, &respData)

	go appglobals.SendNewDMChatMessageUpdate(dmc.Id, senderId, respData.Rrd)

	return respData.Srd, nil
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
