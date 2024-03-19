package chatservice

import (
	"fmt"
	"log"
	"model/chatmodel"
	"time"
	"utils/appglobals"
	"utils/helpers"
)

func NewGroupChat(name string, description string, picture []byte, creator []string, initUsers [][]string) (map[string]any, error) {
	data, err := chatmodel.NewGroupChat(name, description, "", creator, initUsers)
	if err != nil {
		return nil, err
	}

	var respData struct {
		Crd  map[string]any
		Nmrd map[string]any
	}

	helpers.MapToStruct(data, &respData)

	go func() {
		for _, newMember := range initUsers[1:] { // the first user is the creator, hence, they're excluded
			newMemberId := newMember[0]
			go appglobals.ChatObserver{}.Send(fmt.Sprintf("user-%s", newMemberId), respData.Nmrd, "new")
		}
	}()

	go func() {
		if len(picture) < 1 {
			return
		}
		// upload picture and get back url
		groupChatId := respData.Crd["new_group_chat_id"].(int)

		picPath := fmt.Sprintf("chat_pictures/group_chat_%d_pic.jpg", groupChatId)

		picUrl, err := helpers.UploadFile(picPath, picture)

		if err != nil {
			log.Println(err)
			return
		}

		helpers.QueryRowField[bool]("UPDATE group_chat SET picture = $1 WHERE id = $2 RETURNING true", picUrl, groupChatId)
	}()

	return respData.Crd, nil
}

type GroupChat struct {
	Id int
}

func (gpc GroupChat) SendMessage(senderId int, msgContent map[string]any, createdAt time.Time) (map[string]any, error) {
	var respData struct {
		Srd       map[string]any // sender_resp_data AS srd
		Mrd       map[string]any // members_resp_data AS mrd
		MemberIds []int
	}

	data, err := chatmodel.GroupChat{Id: gpc.Id}.SendMessage(senderId, msgContent, createdAt)
	if err != nil {
		return nil, err
	}

	helpers.MapToStruct(data, &respData)

	go func() {
		for _, mId := range respData.MemberIds {
			memberId := mId
			go appglobals.GroupChatMessageObserver{}.Send(fmt.Sprintf("user-%d--groupchat-%d", memberId, gpc.Id), respData.Mrd, "new")
			go appglobals.ChatObserver{}.Send(fmt.Sprintf("user-%d", memberId), map[string]any{"groupChatId": gpc.Id, "event": "new message"}, "update")
		}
	}()

	return respData.Srd, nil
}

func (gpc GroupChat) GetChatHistory(offset int) ([]*map[string]any, error) {
	return chatmodel.GroupChat{Id: gpc.Id}.GetChatHistory(offset)
}

type GroupMessage struct {
	GroupChatId int
	Id          int
}

func (gpcm GroupMessage) React() {

}

func (gpcm GroupMessage) UpdateDeliveryStatus() {

}
