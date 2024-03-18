package chatservice

import (
	"fmt"
	"model/chatmodel"
	"utils/appglobals"
	"utils/helpers"
)

func NewGroupChat(name string, description string, picture string, creator [2]string, initUsers [][2]string) (map[string]any, error) {
	var respData struct {
		Crd  map[string]any
		Nmrd map[string]any
	}

	data, err := chatmodel.NewGroupChat(name, description, picture, creator, initUsers)
	if err != nil {
		return nil, err
	}

	helpers.MapToStruct(data, &respData)

	for _, newMember := range initUsers[1:] { // the first user is the creator, hence, they're excluded
		newMemberId := newMember[0]
		go appglobals.SendNewChatUpdate(fmt.Sprintf("user-%s", newMemberId), respData.Nmrd)
	}

	return respData.Crd, nil
}

type GroupChat struct {
	Id int
}

func (gpc GroupChat) SendMessage(senderId int, msgContent map[string]any) (map[string]any, error) {
	var respData struct {
		Srd map[string]any // sender_resp_data AS srd
		Mrd map[string]any // members_resp_data AS mrd
	}

	groupChat := chatmodel.GroupChat{Id: gpc.Id}

	data, err := groupChat.SendMessage(senderId, msgContent)
	if err != nil {
		return nil, err
	}

	helpers.MapToStruct(data, &respData)

	go appglobals.SendNewGroupChatMessageUpdate(gpc.Id, senderId, respData.Mrd)

	return respData.Srd, nil
}

func (gpc GroupChat) GetChatHistory(offset int) ([]*map[string]any, error) {
	groupChat := chatmodel.GroupChat{Id: gpc.Id}
	return groupChat.GetChatHistory(offset)
}

type GroupMessage struct {
	GroupChatId int
	Id          int
}

func (gpcm GroupMessage) React() {

}

func (gpcm GroupMessage) UpdateDeliveryStatus() {

}
