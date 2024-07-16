package chatService

import (
	"fmt"
	"i9chat/models/chatModel"
	"i9chat/services/appObservers"
	"i9chat/utils/appTypes"
	"i9chat/utils/helpers"
	"time"
)

func broadcastNewGroup(initUsers [][]appTypes.String, newGroupData map[string]any) {
	for _, newMember := range initUsers {
		newMemberId := newMember[0]
		go appObservers.GroupChatObserver{}.Send(fmt.Sprintf("user-%s", newMemberId), newGroupData, "new chat")
	}
}

func uploadGroupPicture(pictureData []byte) (string, error) {
	if len(pictureData) < 1 {
		return "", fmt.Errorf("upload error: no picture data")
	}
	picPath := fmt.Sprintf("group_chat_pictures/group_chat_pic_%d.jpg", time.Now().UnixNano())

	picUrl, err := helpers.UploadFile(picPath, pictureData)

	if err != nil {
		return "", err
	}

	return picUrl, nil
}

func NewGroupChat(name string, description string, pictureData []byte, creator []string, initUsers [][]appTypes.String) (map[string]any, error) {
	picUrl, _ := uploadGroupPicture(pictureData)

	data, err := chatModel.NewGroupChat(name, description, picUrl, creator, initUsers)
	if err != nil {
		return nil, err
	}

	var respData struct {
		Crd  map[string]any // client_resp_data
		Nmrd map[string]any // new_members_resp_data
	}

	helpers.MapToStruct(data, &respData)

	go broadcastNewGroup(initUsers, respData.Nmrd)

	return respData.Crd, nil
}

type GroupChat struct {
	Id int
}

func (gpc GroupChat) broadcastActivity(dbResp map[string]any) {
	var respData struct {
		MemberIds    []int
		ActivityData map[string]any
	}

	helpers.MapToStruct(dbResp, &respData)

	for _, mId := range respData.MemberIds {
		go appObservers.GroupChatObserver{}.Send(fmt.Sprintf("user-%d", mId), respData.ActivityData, "new activity")
	}
}

func (gpc GroupChat) ChangeName(admin []string, newName string) error {
	dbResp, err := chatModel.GroupChat{Id: gpc.Id}.ChangeName(admin, newName)
	if err != nil {
		return err
	}

	go gpc.broadcastActivity(dbResp)

	return nil
}

func (gpc GroupChat) ChangeDescription(admin []string, newDescription string) error {
	dbResp, err := chatModel.GroupChat{Id: gpc.Id}.ChangeDescription(admin, newDescription)
	if err != nil {
		return err
	}

	go gpc.broadcastActivity(dbResp)

	return nil
}

func (gpc GroupChat) ChangePicture(admin []string, newPictureData []byte) error {
	newPicUrl, err := uploadGroupPicture(newPictureData)
	if err != nil {
		return err
	}

	dbResp, err := chatModel.GroupChat{Id: gpc.Id}.ChangePicture(admin, newPicUrl)
	if err != nil {
		return err
	}

	go gpc.broadcastActivity(dbResp)

	return nil
}

func (gpc GroupChat) AddUsers(admin []string, newUsers [][]appTypes.String) error {
	dbResp, err := chatModel.GroupChat{Id: gpc.Id}.AddUsers(admin, newUsers)
	if err != nil {
		return err
	}

	go gpc.broadcastActivity(dbResp)

	return nil
}

func (gpc GroupChat) RemoveUser(admin []string, user []appTypes.String) error {
	dbResp, err := chatModel.GroupChat{Id: gpc.Id}.RemoveUser(admin, user)
	if err != nil {
		return err
	}

	go gpc.broadcastActivity(dbResp)

	return nil
}

func (gpc GroupChat) Join(newUser []string) error {
	dbResp, err := chatModel.GroupChat{Id: gpc.Id}.Join(newUser)
	if err != nil {
		return err
	}

	go gpc.broadcastActivity(dbResp)

	return nil
}

func (gpc GroupChat) Leave(user []string) error {
	dbResp, err := chatModel.GroupChat{Id: gpc.Id}.Leave(user)
	if err != nil {
		return err
	}

	go gpc.broadcastActivity(dbResp)

	return nil
}

func (gpc GroupChat) MakeUserAdmin(admin []string, user []appTypes.String) error {
	dbResp, err := chatModel.GroupChat{Id: gpc.Id}.MakeUserAdmin(admin, user)
	if err != nil {
		return err
	}

	go gpc.broadcastActivity(dbResp)

	return nil
}

func (gpc GroupChat) RemoveUserFromAdmins(admin []string, user []appTypes.String) error {
	dbResp, err := chatModel.GroupChat{Id: gpc.Id}.RemoveUserFromAdmins(admin, user)
	if err != nil {
		return err
	}

	go gpc.broadcastActivity(dbResp)

	return nil
}

func (gpc GroupChat) broadcastMessage(memberIds []int, msgData map[string]any) {
	for _, mId := range memberIds {
		memberId := mId
		go appObservers.GroupChatObserver{}.Send(fmt.Sprintf("user-%d", memberId), msgData, "new message")
	}
}

func (gpc GroupChat) SendMessage(senderId int, msgContent map[string]any, createdAt time.Time) (map[string]any, error) {
	var respData struct {
		Srd       map[string]any // sender_resp_data AS srd
		Mrd       map[string]any // members_resp_data AS mrd
		MemberIds []int
	}

	data, err := chatModel.GroupChat{Id: gpc.Id}.SendMessage(senderId, msgContent, createdAt)
	if err != nil {
		return nil, err
	}

	helpers.MapToStruct(data, &respData)

	go gpc.broadcastMessage(respData.MemberIds, respData.Mrd)

	return respData.Srd, nil
}

func (gpc GroupChat) GetChatHistory(offset int) ([]*map[string]any, error) {
	return chatModel.GroupChat{Id: gpc.Id}.GetChatHistory(offset)
}

func (gpc GroupChat) broadcastMessageDeliveryStatusUpdate(clientUserId int, ackDatas []*appTypes.GroupChatMsgAckData, status string) {
	memberIds, err := helpers.QueryRowsField[int]("SELECT member_id FROM group_chat_membership WHERE group_chat_id = $1 AND member_id != $2 AND deleted = false", gpc.Id, clientUserId)
	if err == nil {
		for _, mId := range memberIds {
			mId := *mId
			for _, data := range ackDatas {
				msgId := data.MsgId
				go appObservers.DMChatSessionObserver{}.Send(
					fmt.Sprintf("user-%d--groupchat-%d", mId, gpc.Id),
					map[string]any{"msgId": msgId, "status": status},
					"delivery status update",
				)
			}
		}
	}
}

func (gpc GroupChat) BatchUpdateGroupChatMessageDeliveryStatus(receiverId int, status string, ackDatas []*appTypes.GroupChatMsgAckData) {
	result, err := chatModel.GroupChat{Id: gpc.Id}.BatchUpdateGroupChatMessageDeliveryStatus(receiverId, status, ackDatas)
	if err != nil {
		return
	}

	// The idea is that, the delivery status of a group message changes
	// when all members have acknowledged the message as "delivered" or "seen",
	// this is set in the overall_delivery_status, after a certain number of members acknowledges delivery.

	// should_broadcast tells us if we should broadcast the overall_delivery_status
	// (if it has changed since the last one) or not (if it hasn't changed since the last one)
	// this saves unnecessary data transfer

	var res struct {
		Ods string // overall_delivery_status
		Sb  bool   // should_broadcast
	}

	helpers.MapToStruct(result, &res)

	if res.Sb {
		go gpc.broadcastMessageDeliveryStatusUpdate(receiverId, ackDatas, res.Ods)
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
