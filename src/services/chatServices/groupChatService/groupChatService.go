package groupChatService

import (
	"context"
	"i9chat/src/appTypes"
	"i9chat/src/helpers"
	groupChat "i9chat/src/models/chatModel/groupChatModel"
	"i9chat/src/services/appServices"
	"i9chat/src/services/eventStreamService"
	"i9chat/src/services/eventStreamService/eventTypes"
	"time"
)

func NewGroup(ctx context.Context, clientUsername, name, description string, pictureData []byte, initUsers []string, createdAt int64) (map[string]any, error) {
	picUrl, err := uploadGroupPicture(ctx, pictureData)
	if err != nil {
		return nil, err
	}

	newGroup, err := groupChat.New(ctx, clientUsername, name, description, picUrl, initUsers, createdAt)
	if err != nil {
		return nil, err
	}

	done := newGroup.Id != ""

	if !done {
		return nil, nil
	}

	go eventStreamService.QueueNewGroupEvent(eventTypes.NewGroupEvent{
		CreatorUser:     clientUsername,
		GroupId:         newGroup.Id,
		GroupData:       helpers.ToJson(newGroup),
		InitMembers:     newGroup.InitUsers,
		CreatorUserCHEs: newGroup.ClientUserCHEs,
		InitMembersCHEs: newGroup.InitUsersCHEs,
	})

	type groupActivityCHE struct {
		CHEType string `json:"che_type"`
		Info    string `json:"info"`
	}

	go func() {
		// history is the same for all init users, we only need separation for the cache
		// so we make use of one user's history
		cheMaps := newGroup.InitUsersCHEs[initUsers[0]]

		var history []groupActivityCHE

		helpers.ToStruct(cheMaps, &history)

		initMemberData := map[string]any{
			"group":   newGroup,
			"history": history,
		}

		broadcastNewGroup(newGroup.InitUsers, initMemberData)
	}()

	var history []groupActivityCHE

	helpers.ToStruct(newGroup.ClientUserCHEs, &history)

	clientData := map[string]any{
		"group":   newGroup,
		"history": history,
	}

	return clientData, nil
}

func ChangeGroupName(ctx context.Context, groupId, clientUsername, newName string) (any, error) {
	newActivity, err := groupChat.ChangeName(ctx, groupId, clientUsername, newName)
	if err != nil {
		return nil, err
	}

	done := newActivity.ClientUserCHE != nil

	if !done {
		return nil, nil
	}

	go eventStreamService.QueueGroupEditEvent(eventTypes.GroupEditEvent{
		GroupId:        groupId,
		UpdateKVMap:    map[string]any{"name": newName},
		MemberUsers:    newActivity.MemberUsernames,
		EditorUserCHE:  newActivity.ClientUserCHE,
		MemberUsersCHE: newActivity.MemberUsersCHE,
	})

	if memus := newActivity.MemberUsernames; len(memus) != 0 {
		activInfo := newActivity.MemberUsersCHE[memus[0].(string)].(map[string]any)["info"]

		go broadcastActivity(memus, activInfo, groupId)
	}

	return newActivity.ClientUserCHE["info"], nil
}

func ChangeGroupDescription(ctx context.Context, groupId, clientUsername, newDescription string) (any, error) {
	newActivity, err := groupChat.ChangeDescription(ctx, groupId, clientUsername, newDescription)
	if err != nil {
		return nil, err
	}

	done := newActivity.ClientUserCHE != nil

	if !done {
		return nil, nil
	}

	go eventStreamService.QueueGroupEditEvent(eventTypes.GroupEditEvent{
		GroupId:        groupId,
		EditorUser:     clientUsername,
		UpdateKVMap:    map[string]any{"description": newDescription},
		MemberUsers:    newActivity.MemberUsernames,
		EditorUserCHE:  newActivity.ClientUserCHE,
		MemberUsersCHE: newActivity.MemberUsersCHE,
	})

	if memus := newActivity.MemberUsernames; len(memus) != 0 {
		activInfo := newActivity.MemberUsersCHE[memus[0].(string)].(map[string]any)["info"]

		go broadcastActivity(memus, activInfo, groupId)
	}

	return newActivity.ClientUserCHE["info"], nil
}

func ChangeGroupPicture(ctx context.Context, groupId, clientUsername string, newPictureData []byte) (any, error) {
	newPicUrl, err := uploadGroupPicture(ctx, newPictureData)
	if err != nil {
		return nil, err
	}

	newActivity, err := groupChat.ChangePicture(ctx, groupId, clientUsername, newPicUrl)
	if err != nil {
		return nil, err
	}

	done := newActivity.ClientUserCHE != nil

	if !done {
		return nil, nil
	}

	go eventStreamService.QueueGroupEditEvent(eventTypes.GroupEditEvent{
		GroupId:        groupId,
		UpdateKVMap:    map[string]any{"picture_url": newPicUrl},
		MemberUsers:    newActivity.MemberUsernames,
		EditorUserCHE:  newActivity.ClientUserCHE,
		MemberUsersCHE: newActivity.MemberUsersCHE,
	})

	if memus := newActivity.MemberUsernames; len(memus) != 0 {
		activInfo := newActivity.MemberUsersCHE[memus[0].(string)].(map[string]any)["info"]

		go broadcastActivity(memus, activInfo, groupId)
	}

	return newActivity.ClientUserCHE["info"], nil
}

func AddUsersToGroup(ctx context.Context, groupId, clientUsername string, newUsers []string) (any, error) {
	newActivity, err := groupChat.AddUsers(ctx, groupId, clientUsername, newUsers)
	if err != nil {
		return nil, err
	}

	done := newActivity.GroupInfo != nil
	if !done {
		return nil, nil
	}

	go eventStreamService.QueueGroupUsersAddedEvent(eventTypes.GroupUsersAddedEvent{
		GroupId:        groupId,
		AdminUser:      clientUsername,
		NewMembers:     newActivity.NewUsernames,
		MemberUsers:    newActivity.MemberUsernames,
		AdminUserCHE:   newActivity.ClientUserCHE,
		NewMembersCHE:  newActivity.NewUsersCHE,
		MemberUsersCHE: newActivity.MemberUsersCHE,
	})

	go func(newActivity groupChat.AddUsersActivity) {
		if len(newActivity.NewUsernames) == 0 {
			return
		}

		// history is the same for all init users, we only need separation for the cache
		// so we make use of one user's history
		che := newActivity.NewUsersCHE[newActivity.NewUsernames[0].(string)].(map[string]any)

		delete(che, "che_id")

		newUserData := map[string]any{
			"group":   newActivity.GroupInfo,
			"history": []map[string]any{che},
		}

		broadcastNewGroup(newActivity.NewUsernames, newUserData)
	}(newActivity)

	if memus := newActivity.MemberUsernames; len(memus) != 0 {
		activInfo := newActivity.MemberUsersCHE[memus[0].(string)].(map[string]any)["info"]

		go broadcastActivity(memus, activInfo, groupId)
	}

	return newActivity.ClientUserCHE["info"], nil
}

func RemoveUserFromGroup(ctx context.Context, groupId, clientUsername, targetUser string) (any, error) {
	newActivity, err := groupChat.RemoveUser(ctx, groupId, clientUsername, targetUser)
	if err != nil {
		return nil, err
	}

	if newActivity.MemberData != nil {
		go func() {
			broadcastActivity([]string{targetUser}, targetUserData, groupId)

			broadcastActivity(newActivity.MemberUsernames, newActivity.MemberData, groupId)
		}()
	}

	return newActivity.ClientData, nil
}

func JoinGroup(ctx context.Context, groupId, clientUsername string) (any, error) {
	newActivity, err := groupChat.Join(ctx, groupId, clientUsername)
	if err != nil {
		return nil, err
	}

	if newActivity.MemberData != nil {
		go broadcastActivity(newActivity.MemberUsernames, newActivity.MemberData, groupId)
	}

	return newActivity.ClientData, nil
}

func LeaveGroup(ctx context.Context, groupId, clientUsername string) (any, error) {
	newActivity, err := groupChat.Leave(ctx, groupId, clientUsername)
	if err != nil {
		return nil, err
	}

	if newActivity.MemberData != nil {
		go broadcastActivity(newActivity.MemberUsernames, newActivity.MemberData, groupId)
	}

	return newActivity.ClientData, nil
}

func MakeUserGroupAdmin(ctx context.Context, groupId, clientUsername, targetUser string) (any, error) {
	newActivity, err := groupChat.MakeUserAdmin(ctx, groupId, clientUsername, targetUser)
	if err != nil {
		return nil, err
	}

	if newActivity.MemberData != nil {
		go func() {
			broadcastActivity([]string{targetUser}, newAdminData, groupId)

			broadcastActivity(newActivity.MemberUsernames, newActivity.MemberData, groupId)
		}()
	}

	return newActivity.ClientData, nil
}

func RemoveUserFromGroupAdmins(ctx context.Context, groupId, clientUsername, targetUser string) (any, error) {
	newActivity, oldAdminData, err := groupChat.RemoveUserFromAdmins(ctx, groupId, clientUsername, targetUser)
	if err != nil {
		return nil, err
	}

	if newActivity.MemberData != nil {
		go func() {
			broadcastActivity([]string{targetUser}, oldAdminData, groupId)

			broadcastActivity(newActivity.MemberUsernames, newActivity.MemberData, groupId)
		}()
	}

	return newActivity.ClientData, nil
}

func GetGroupInfo(ctx context.Context, groupId string) (map[string]any, error) {
	return groupChat.GroupInfo(ctx, groupId)
}

func GetGroupMemInfo(ctx context.Context, clientUsername, groupId string) (map[string]any, error) {
	return groupChat.GroupMemInfo(ctx, clientUsername, groupId)
}

func SendMessage(ctx context.Context, clientUsername, groupId, replyTargetMsgId string, isReply bool, msgContent *appTypes.MsgContent, at int64) (map[string]any, error) {

	err := appServices.UploadMessageMedia(ctx, clientUsername, msgContent)
	if err != nil {
		return nil, err
	}

	msgContentJson := helpers.ToJson(*msgContent)

	var newMessage groupChat.NewMessage

	if !isReply {
		newMessage, err = groupChat.SendMessage(ctx, clientUsername, groupId, string(msgContentJson), at)
		if err != nil {
			return nil, err
		}
	} else {
		newMessage, err = groupChat.ReplyToMessage(ctx, clientUsername, groupId, replyTargetMsgId, string(msgContentJson), at)
		if err != nil {
			return nil, err
		}
	}

	if newMessage.MemberData != nil {
		go broadcastNewMessage(newMessage.MemberUsernames, newMessage.MemberData, groupId)
	}

	return newMessage.ClientData, nil
}

func AckMessageDelivered(ctx context.Context, clientUsername, groupId, msgId string, deliveredAt int64) (bool, error) {
	done, err := groupChat.AckMessageDelivered(ctx, clientUsername, groupId, msgId, deliveredAt)
	if err != nil {
		return done, err
	}

	/*
		** in bg worker,
		** in a group:id:message:id cache key, ZAdd the user acknowledging delivery, using deliveredAt as score,
		** after which you check if all groupMembers are present in this key (x in SScan: Zscore(x) != nil)
		** if no nil result is retured, then the message is now delivered to all users
		go broadcastMsgDelivered(msgAck.MemberUsernames, map[string]any{
			"group_id": groupId,
			"msg_id":   msgId,
		}) */

	return done, nil
}

func AckMessageRead(ctx context.Context, clientUsername, groupId, msgId string, readAt int64) (any, error) {
	done, err := groupChat.AckMessageRead(ctx, clientUsername, groupId, msgId, readAt)
	if err != nil {
		return nil, err
	}

	/*
		// broadcast in bg worker when read by all
		go broadcastMsgRead(msgAck.MemberUsernames, map[string]any{
			"group_id": groupId,
			"msg_id":   msgId,
		}) */

	return done, nil
}

func ReactToMessage(ctx context.Context, clientUsername, groupId, msgId, reaction string, at int64) (any, error) {
	rxnToMessage, err := groupChat.ReactToMessage(ctx, clientUsername, groupId, msgId, reaction, time.UnixMilli(at).UTC())
	if err != nil {
		return nil, err
	}

	if rxnToMessage.MemberData != nil {
		go broadcastMsgReaction(rxnToMessage.MemberUsernames, rxnToMessage.MemberData)
	}

	return rxnToMessage.ClientData, nil
}

func RemoveReactionToMessage(ctx context.Context, clientUsername, groupId, msgId string, at int64) (any, error) {
	memberUsernames, done, err := groupChat.RemoveReactionToMessage(ctx, clientUsername, groupId, msgId)
	if err != nil {
		return nil, err
	}

	if done {
		go broadcastMsgReactionRemoved(memberUsernames, map[string]any{
			"group_id":         groupId,
			"msg_id":           msgId,
			"reactor_username": clientUsername,
		})
	}

	return done, nil
}

func GetChatHistory(ctx context.Context, clientUsername, groupId string, limit int, offset int64) (any, error) {
	return groupChat.ChatHistory(ctx, clientUsername, groupId, limit, time.UnixMilli(offset).UTC())
}
