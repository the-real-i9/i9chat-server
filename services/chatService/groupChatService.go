package chatservice

import (
	"fmt"
	"log"
	"model/chatmodel"
	"time"
	"utils/appglobals"
	"utils/apptypes"
	"utils/helpers"
)

func broadcastNewGroup(initUsers [][]string, newGroupData map[string]any) {
	for _, newMember := range initUsers[1:] { // the first user is the creator, hence, they're excluded
		newMemberId := newMember[0]
		go appglobals.ChatObserver{}.Send(fmt.Sprintf("user-%s", newMemberId), newGroupData, "new")
	}
}

func uploadGroupPicture(groupChatId int, picture []byte) {
	if len(picture) < 1 {
		return
	}
	// upload picture data and SET picture url
	picPath := fmt.Sprintf("chat_pictures/group_chat_%d_pic.jpg", groupChatId)

	picUrl, err := helpers.UploadFile(picPath, picture)

	if err != nil {
		log.Println(err)
		return
	}

	helpers.QueryRowField[bool]("UPDATE group_chat SET picture = $1 WHERE id = $2 RETURNING true", picUrl, groupChatId)
}

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

	go uploadGroupPicture(respData.Crd["new_group_chat_id"].(int), picture)

	go broadcastNewGroup(initUsers, respData.Nmrd)

	return respData.Crd, nil
}

type GroupChat struct {
	Id int
}

func (gpc GroupChat) broadcastMessage(memberIds []int, msgData map[string]any) {
	for _, mId := range memberIds {
		memberId := mId
		go appglobals.GroupChatMessageObserver{}.Send(fmt.Sprintf("user-%d--groupchat-%d", memberId, gpc.Id), msgData, "new")
		// go appglobals.ChatObserver{}.Send(fmt.Sprintf("user-%d", memberId), map[string]any{"groupChatId": gpc.Id, "event": "new message", "senderId": senderId}, "update")
	}
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

	go gpc.broadcastMessage(respData.MemberIds, respData.Mrd)

	return respData.Srd, nil
}

func (gpc GroupChat) GetChatHistory(offset int) ([]*map[string]any, error) {
	return chatmodel.GroupChat{Id: gpc.Id}.GetChatHistory(offset)
}

func (gpc GroupChat) broadcastMessageDeliveryStatusUpdate(clientId int, delivDatas []*apptypes.GroupChatMsgDeliveryData, status string) {
	memberIds, err := helpers.QueryRowsField[int]("SELECT member_id FROM group_chat_membership WHERE group_chat_id = $1 AND member_id != $2 AND deleted = false", gpc.Id, clientId)
	if err == nil {
		for _, mId := range memberIds {
			mId := *mId
			for _, data := range delivDatas {
				msgId := data.MsgId
				go appglobals.DMChatMessageObserver{}.Send(
					fmt.Sprintf("user-%d--groupchat-%d", mId, gpc.Id),
					map[string]any{"msgId": msgId, "event": "delivery", "status": status},
					"update",
				)
			}
		}
	}
}

func (gpc GroupChat) BatchUpdateGroupChatMessageDeliveryStatus(receiverId int, status string, delivDatas []*apptypes.GroupChatMsgDeliveryData) {
	if overallDeliveryStatus, err := (chatmodel.GroupChat{Id: gpc.Id}).BatchUpdateGroupChatMessageDeliveryStatus(receiverId, status, delivDatas); err == nil {
		go gpc.broadcastMessageDeliveryStatusUpdate(receiverId, delivDatas, overallDeliveryStatus)
	}
}

type GroupChatMessage struct {
	Id          int
	GroupChatId int
	SenderId    int
}

func (gpcm GroupChatMessage) UpdateDeliveryStatus() {

}

func (gpcm GroupChatMessage) React() {

}
