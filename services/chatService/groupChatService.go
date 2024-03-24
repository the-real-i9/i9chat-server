package chatservice

import (
	"fmt"
	"model/chatmodel"
	"time"
	"utils/appglobals"
	"utils/apptypes"
	"utils/helpers"
)

func broadcastNewGroup(initUsers [][]string, newGroupData map[string]any) {
	for _, newMember := range initUsers[1:] { // the first user is the creator, hence, they're excluded
		newMemberId := newMember[0]
		go appglobals.GroupChatObserver{}.Send(fmt.Sprintf("user-%s", newMemberId), newGroupData, "new chat")
	}
}

func uploadGroupPicture(groupChatId int, picture []byte) string {
	if len(picture) < 1 {
		return ""
	}
	// upload picture data and SET picture url
	picPath := fmt.Sprintf("chat_pictures/group_chat_%d_pic_%d.jpg", groupChatId, time.Now().UnixMilli())

	picUrl, err := helpers.UploadFile(picPath, picture)

	if err != nil {
		return ""
	}

	return picUrl
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

	helpers.ParseToStruct(data, &respData)

	go func() {
		groupChatId := respData.Crd["new_group_chat_id"].(int)
		if picUrl := uploadGroupPicture(groupChatId, picture); picUrl != "" {
			helpers.QueryRowField[bool]("UPDATE group_chat SET picture = $1 WHERE id = $2 RETURNING true", picUrl, groupChatId)
		}
	}()

	go broadcastNewGroup(initUsers, respData.Nmrd)

	return respData.Crd, nil
}

type GroupChat struct {
	Id int
}

func (gpc GroupChat) broadcastActivity(modelResp map[string]any) {
	var respData struct {
		MemberIds    []int
		ActivityData map[string]any
	}

	helpers.ParseToStruct(modelResp, &respData)

	for _, mId := range respData.MemberIds {
		go appglobals.GroupChatObserver{}.Send(fmt.Sprintf("user-%d", mId), respData.ActivityData, "new activity")
	}
}

func (gpc GroupChat) ChangeName(admin []string, newName string) {
	if resp, err := (chatmodel.GroupChat{Id: gpc.Id}).ChangeName(admin, newName); err == nil {
		go gpc.broadcastActivity(resp)
	}
}

func (gpc GroupChat) ChangeDescription(admin []string, newDescription string) {
	if resp, err := (chatmodel.GroupChat{Id: gpc.Id}).ChangeDescription(admin, newDescription); err == nil {
		go gpc.broadcastActivity(resp)
	}
}

func (gpc GroupChat) ChangePicture(admin []string, newPicture []byte) {
	picUrl := uploadGroupPicture(gpc.Id, newPicture)
	if picUrl == "" {
		return
	}

	if resp, err := (chatmodel.GroupChat{Id: gpc.Id}).ChangePicture(admin, picUrl); err == nil {
		go gpc.broadcastActivity(resp)
	}
}

func (gpc GroupChat) AddUsers(admin []string, newUsers [][]string) {
	if resp, err := (chatmodel.GroupChat{Id: gpc.Id}).AddUsers(admin, newUsers); err == nil {
		go gpc.broadcastActivity(resp)
	}
}

func (gpc GroupChat) RemoveUser(admin []string, user []string) {
	if resp, err := (chatmodel.GroupChat{Id: gpc.Id}).RemoveUser(admin, user); err == nil {
		go gpc.broadcastActivity(resp)
	}
}

func (gpc GroupChat) Join(newUser []string) {
	if resp, err := (chatmodel.GroupChat{Id: gpc.Id}).Join(newUser); err == nil {
		go gpc.broadcastActivity(resp)
	}
}

func (gpc GroupChat) Leave(user []string) {
	if resp, err := (chatmodel.GroupChat{Id: gpc.Id}).Leave(user); err == nil {
		go gpc.broadcastActivity(resp)
	}
}

func (gpc GroupChat) MakeUserAdmin(admin []string, user []string) {
	if resp, err := (chatmodel.GroupChat{Id: gpc.Id}).MakeUserAdmin(admin, user); err == nil {
		go gpc.broadcastActivity(resp)
	}
}

func (gpc GroupChat) RemoveUserFromAdmins(admin []string, user []string) {
	if resp, err := (chatmodel.GroupChat{Id: gpc.Id}).RemoveUserFromAdmins(admin, user); err == nil {
		go gpc.broadcastActivity(resp)
	}
}

func (gpc GroupChat) broadcastMessage(memberIds []int, msgData map[string]any) {
	for _, mId := range memberIds {
		memberId := mId
		go appglobals.GroupChatObserver{}.Send(fmt.Sprintf("user-%d", memberId), msgData, "new message")
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

	helpers.ParseToStruct(data, &respData)

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
				go appglobals.DMChatSessionObserver{}.Send(
					fmt.Sprintf("user-%d--groupchat-%d", mId, gpc.Id),
					map[string]any{"msgId": msgId, "key": "delivery_status", "value": status},
					"message update",
				)
			}
		}
	}
}

func (gpc GroupChat) BatchUpdateGroupChatMessageDeliveryStatus(receiverId int, status string, delivDatas []*apptypes.GroupChatMsgDeliveryData) {
	result, err := (chatmodel.GroupChat{Id: gpc.Id}).BatchUpdateGroupChatMessageDeliveryStatus(receiverId, status, delivDatas)
	if err != nil {
		return
	}

	// The idea is that, the delivery status of a group message changes
	// if all members have acknowledged the message as "delivered" or "seen",
	// this is set in the overall_delivery_status, after a certain number of members acknowledges delivery.

	// should_broadcast tells us if we should broadcast the overall_delivery_status
	// (if it has changed since the last one) or not (if it hasn't changed since the last one)
	// this saves unnecessary data transfer

	var res struct {
		Ods string // overall_delivery_status
		Sb  bool   // should_broadcast
	}

	helpers.ParseToStruct(result, &res)

	if res.Sb {
		go gpc.broadcastMessageDeliveryStatusUpdate(receiverId, delivDatas, res.Ods)
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
