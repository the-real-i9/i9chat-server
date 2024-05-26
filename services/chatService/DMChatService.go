package chatService

import (
	"fmt"
	"i9chat/models/chatModel"
	"i9chat/services/appObservers"
	"i9chat/utils/appTypes"
	"i9chat/utils/helpers"
	"time"
)

func NewDMChat(initiatorId int, partnerId int, initMsgContent map[string]any, createdAt time.Time) (map[string]any, error) {
	var respData struct {
		Ird map[string]any // initiator_resp_data AS ird
		Prd map[string]any // partner_resp_data AS prd
	}

	data, err := chatModel.NewDMChat(initiatorId, partnerId, initMsgContent, createdAt)
	if err != nil {
		return nil, err
	}

	helpers.MapToStruct(data, &respData)

	go appObservers.DMChatObserver{}.Send(fmt.Sprintf("user-%d", partnerId), respData.Prd, "new chat")

	return respData.Ird, nil
}

func BatchUpdateDMChatMessageDeliveryStatus(receiverId int, status string, ackDatas []*appTypes.DMChatMsgAckData) {
	if err := chatModel.BatchUpdateDMChatMessageDeliveryStatus(receiverId, status, ackDatas); err == nil {
		for _, data := range ackDatas {
			data := data
			go func() {
				appObservers.DMChatSessionObserver{}.Send(
					fmt.Sprintf("user-%d--dmchat-%d", data.SenderId, data.DmChatId),
					map[string]any{"msgId": data.MsgId, "key": "delivery_status", "value": status},
					"message update",
				)
			}()
		}
	}
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

	data, err := chatModel.DMChat{Id: dmc.Id}.SendMessage(senderId, msgContent, createdAt)
	if err != nil {
		return nil, err
	}

	helpers.MapToStruct(data, &respData)

	go appObservers.DMChatObserver{}.Send(
		fmt.Sprintf("user-%d", respData.ReceiverId), respData.Rrd, "new message",
	)

	return respData.Srd, nil
}

func (dmc DMChat) GetChatHistory(offset int) ([]*map[string]any, error) {
	return chatModel.DMChat{Id: dmc.Id}.GetChatHistory(offset)
}

type DMChatMessage struct {
	Id       int
	DmChatId int
	SenderId int
}

func (dmcm DMChatMessage) UpdateDeliveryStatus(receiverId int, status string, updatedAt time.Time) {
	if err := (chatModel.DMChatMessage{Id: dmcm.Id, DmChatId: dmcm.DmChatId}).UpdateDeliveryStatus(receiverId, status, updatedAt); err == nil {

		go appObservers.DMChatSessionObserver{}.Send(
			fmt.Sprintf("user-%d--dmchat-%d", dmcm.SenderId, dmcm.DmChatId),
			map[string]any{"msgId": dmcm.Id, "key": "delivery_status", "value": status},
			"message update",
		)
	}

}

func (dmcm DMChatMessage) React() {

}
