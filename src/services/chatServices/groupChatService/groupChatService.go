package groupChatService

import (
	"context"
	"fmt"
	"i9chat/src/appTypes/UITypes"
	"i9chat/src/cache"
	"i9chat/src/helpers"
	groupChat "i9chat/src/models/chatModel/groupChatModel"
	"i9chat/src/services/cloudStorageService"
	"i9chat/src/services/eventStreamService"
	"i9chat/src/services/eventStreamService/eventTypes"
	"maps"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AuthGroupPicDataT struct {
	UploadUrl         string `json:"uploadUrl"`
	GroupPicCloudName string `json:"groupPicCloudName"`
}

func AuthorizeGroupPicUpload(ctx context.Context, picMIME string) (AuthGroupPicDataT, error) {
	var res AuthGroupPicDataT

	for small0_medium1_large2 := range 3 {

		which := [3]string{"small", "medium", "large"}

		groupPicCloudName := fmt.Sprintf("uploads/group/group_pics/%d%d/%s-%s", time.Now().Year(), time.Now().Month(), uuid.NewString(), which[small0_medium1_large2])

		url, err := cloudStorageService.GetUploadUrl(groupPicCloudName, picMIME)
		if err != nil {
			return res, fiber.ErrInternalServerError
		}

		switch small0_medium1_large2 {
		case 0:
			res.UploadUrl += "small:"
			res.GroupPicCloudName += "small:"
		case 1:
			res.UploadUrl += "medium:"
			res.GroupPicCloudName += "medium:"
		default:
			res.UploadUrl += "large:"
			res.GroupPicCloudName += "large:"
		}

		res.UploadUrl += url
		res.GroupPicCloudName += groupPicCloudName

		if small0_medium1_large2 != 2 {
			res.UploadUrl += " "
			res.GroupPicCloudName += " "
		}
	}

	return res, nil
}

func NewGroup(ctx context.Context, clientUsername, name, description, pictureCloudName string, initUsers []string, createdAt int64) (map[string]any, error) {
	newGroup, err := groupChat.New(ctx, clientUsername, name, description, pictureCloudName, initUsers, createdAt)
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

		history := helpers.ToStruct[[]groupActivityCHE](cheMaps)

		initMemberData := map[string]any{
			"group":   newGroup,
			"history": history,
		}

		broadcastNewGroup(newGroup.InitUsers, initMemberData)
	}()

	history := helpers.ToStruct[[]groupActivityCHE](newGroup.ClientUserCHEs)

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
		GroupId:       groupId,
		EditorUser:    clientUsername,
		UpdateKVMap:   map[string]any{"name": newName},
		EditorUserCHE: newActivity.ClientUserCHE,
		MemInfo:       newActivity.MemInfo,
	})

	go broadcastActivityToAll(groupId, newActivity.MemInfo, []any{clientUsername})

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
		GroupId:       groupId,
		EditorUser:    clientUsername,
		UpdateKVMap:   map[string]any{"description": newDescription},
		EditorUserCHE: newActivity.ClientUserCHE,
		MemInfo:       newActivity.MemInfo,
	})

	go broadcastActivityToAll(groupId, newActivity.MemInfo, []any{clientUsername})

	return newActivity.ClientUserCHE["info"], nil
}

func ChangeGroupPicture(ctx context.Context, groupId, clientUsername, picCloudName string) (any, error) {
	newActivity, err := groupChat.ChangePicture(ctx, groupId, clientUsername, picCloudName)
	if err != nil {
		return nil, err
	}

	done := newActivity.ClientUserCHE != nil

	if !done {
		return nil, nil
	}

	go eventStreamService.QueueGroupEditEvent(eventTypes.GroupEditEvent{
		GroupId:       groupId,
		UpdateKVMap:   map[string]any{"picture_cloud_name": picCloudName},
		EditorUserCHE: newActivity.ClientUserCHE,
		MemInfo:       newActivity.MemInfo,
	})

	go broadcastActivityToAll(groupId, newActivity.MemInfo, []any{clientUsername})

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
		GroupId:       groupId,
		Admin:         clientUsername,
		NewMembers:    newActivity.NewUsernames,
		AdminCHE:      newActivity.ClientUserCHE,
		NewMembersCHE: newActivity.NewUsersCHE,
		MemInfo:       newActivity.MemInfo,
	})

	go func(newActivity groupChat.AddUsersActivity) {
		if len(newActivity.NewUsernames) == 0 {
			return
		}

		// history is the same for all init users, we only need separation for our cache
		// so we make use of one user's history
		che := maps.Clone(newActivity.NewUsersCHE[newActivity.NewUsernames[0].(string)].(map[string]any))

		delete(che, "che_id")

		newMemData := map[string]any{
			"group":   newActivity.GroupInfo,
			"history": []map[string]any{che},
		}

		broadcastNewGroup(newActivity.NewUsernames, newMemData)
	}(newActivity)

	go broadcastActivityToAll(groupId, newActivity.MemInfo, append(newActivity.NewUsernames, clientUsername))

	return newActivity.ClientUserCHE["info"], nil
}

func RemoveUserFromGroup(ctx context.Context, groupId, clientUsername, targetUser string) (any, error) {
	newActivity, err := groupChat.RemoveUser(ctx, groupId, clientUsername, targetUser)
	if err != nil {
		return nil, err
	}

	done := newActivity.ClientUserCHE != nil
	if !done {
		return nil, nil
	}

	go eventStreamService.QueueGroupUserRemovedEvent(eventTypes.GroupUserRemovedEvent{
		GroupId:      groupId,
		Admin:        clientUsername,
		OldMember:    targetUser,
		AdminCHE:     newActivity.ClientUserCHE,
		OldMemberCHE: newActivity.TargetUserCHE,
		MemInfo:      newActivity.MemInfo,
	})

	go broadcastActivityToOne(targetUser, newActivity.TargetUserCHE["info"], groupId)

	go broadcastActivityToAll(groupId, newActivity.MemInfo, []any{clientUsername, targetUser})

	return newActivity.ClientUserCHE["info"], nil
}

func JoinGroup(ctx context.Context, groupId, clientUsername string) (any, error) {
	newActivity, err := groupChat.Join(ctx, groupId, clientUsername)
	if err != nil {
		return nil, err
	}

	done := newActivity.GroupInfo != nil
	if !done {
		return nil, nil
	}

	go eventStreamService.QueueGroupUserJoinedEvent(eventTypes.GroupUserJoinedEvent{
		GroupId:      groupId,
		NewMember:    clientUsername,
		NewMemberCHE: newActivity.ClientUserCHE,
		MemInfo:      newActivity.MemInfo,
	})

	go broadcastActivityToAll(groupId, newActivity.MemInfo, []any{clientUsername})

	che := maps.Clone(newActivity.ClientUserCHE)

	delete(che, "che_id")

	newMemData := map[string]any{
		"group":   newActivity.GroupInfo,
		"history": []map[string]any{che},
	}

	return newMemData, nil
}

func LeaveGroup(ctx context.Context, groupId, clientUsername string) (any, error) {
	newActivity, err := groupChat.Leave(ctx, groupId, clientUsername)
	if err != nil {
		return nil, err
	}

	done := newActivity.ClientUserCHE != nil
	if !done {
		return nil, nil
	}

	go eventStreamService.QueueGroupUserLeftEvent(eventTypes.GroupUserLeftEvent{
		GroupId:      groupId,
		OldMember:    clientUsername,
		OldMemberCHE: newActivity.ClientUserCHE,
		MemInfo:      newActivity.MemInfo,
	})

	go broadcastActivityToAll(groupId, newActivity.MemInfo, []any{clientUsername})

	return newActivity.ClientUserCHE["info"], nil
}

func MakeUserGroupAdmin(ctx context.Context, groupId, clientUsername, targetUser string) (any, error) {
	newActivity, err := groupChat.MakeUserAdmin(ctx, groupId, clientUsername, targetUser)
	if err != nil {
		return nil, err
	}

	done := newActivity.ClientUserCHE != nil
	if !done {
		return nil, nil
	}

	go eventStreamService.QueueGroupMakeUserAdminEvent(eventTypes.GroupMakeUserAdminEvent{
		GroupId:     groupId,
		Admin:       clientUsername,
		NewAdmin:    targetUser,
		AdminCHE:    newActivity.ClientUserCHE,
		NewAdminCHE: newActivity.TargetUserCHE,
		MemInfo:     newActivity.MemInfo,
	})

	go broadcastActivityToOne(targetUser, newActivity.TargetUserCHE["info"], groupId)

	go broadcastActivityToAll(groupId, newActivity.MemInfo, []any{clientUsername, targetUser})

	return newActivity.ClientUserCHE["info"], nil
}

func RemoveUserFromGroupAdmins(ctx context.Context, groupId, clientUsername, targetUser string) (any, error) {
	newActivity, err := groupChat.RemoveUserFromAdmins(ctx, groupId, clientUsername, targetUser)
	if err != nil {
		return nil, err
	}

	done := newActivity.ClientUserCHE != nil
	if !done {
		return nil, nil
	}

	go eventStreamService.QueueGroupRemoveUserFromAdminsEvent(eventTypes.GroupRemoveUserFromAdminsEvent{
		GroupId:     groupId,
		Admin:       clientUsername,
		OldAdmin:    targetUser,
		AdminCHE:    newActivity.ClientUserCHE,
		OldAdminCHE: newActivity.TargetUserCHE,
		MemInfo:     newActivity.MemInfo,
	})

	go broadcastActivityToOne(targetUser, newActivity.TargetUserCHE["info"], groupId)

	go broadcastActivityToAll(groupId, newActivity.MemInfo, []any{clientUsername, targetUser})

	return newActivity.ClientUserCHE["info"], nil
}

func SendMessage(ctx context.Context, clientUsername, groupId, replyTargetMsgId string, isReply bool, msgContentJson string, at int64) (map[string]any, error) {
	var (
		newMessage groupChat.NewMessage
		err        error
	)

	if !isReply {
		newMessage, err = groupChat.SendMessage(ctx, clientUsername, groupId, msgContentJson, at)
		if err != nil {
			return nil, err
		}
	} else {
		newMessage, err = groupChat.ReplyToMessage(ctx, clientUsername, groupId, replyTargetMsgId, msgContentJson, at)
		if err != nil {
			return nil, err
		}
	}

	if newMessage.Id == "" {
		return nil, nil
	}

	// queue new message event
	go eventStreamService.QueueNewGroupMessageEvent(eventTypes.NewGroupMessageEvent{
		FromUser: clientUsername,
		ToGroup:  groupId,
		CHEId:    newMessage.Id,
		MsgData:  helpers.ToJson(newMessage),
	})

	go func(msgData groupChat.NewMessage) {
		var err error
		msgData.Sender, err = cache.GetUser[UITypes.ClientUser](context.Background(), clientUsername)
		if err != nil {
			return
		}

		msgData.Content = maps.Clone(msgData.Content)
		err = cloudStorageService.MessageMediaCloudNameToUrl(msgData.Content)
		if err != nil {
			return
		}

		broadcastNewMessage(groupId, msgData, clientUsername)
	}(newMessage)

	return map[string]any{"new_msg_id": newMessage.Id}, nil
}

func AckMessageDelivered(ctx context.Context, clientUsername, groupId, msgId string, deliveredAt int64) (bool, error) {
	done, err := groupChat.AckMessageDelivered(ctx, clientUsername, groupId, msgId, deliveredAt)
	if err != nil {
		return done, err
	}

	if done {
		// queue msg ack event
		go eventStreamService.QueueGroupMsgAckEvent(eventTypes.GroupMsgAckEvent{
			FromUser: clientUsername,
			ToGroup:  groupId,
			CHEId:    msgId,
			Ack:      "delivered",
			At:       deliveredAt,
		})
	}

	return done, nil
}

func AckMessageRead(ctx context.Context, clientUsername, groupId, msgId string, readAt int64) (any, error) {
	done, err := groupChat.AckMessageRead(ctx, clientUsername, groupId, msgId, readAt)
	if err != nil {
		return nil, err
	}

	if done {
		// queue msg ack event
		go eventStreamService.QueueGroupMsgAckEvent(eventTypes.GroupMsgAckEvent{
			FromUser: clientUsername,
			ToGroup:  groupId,
			CHEId:    msgId,
			Ack:      "read",
			At:       readAt,
		})
	}

	return done, nil
}

func ReactToMessage(ctx context.Context, clientUsername, groupId, msgId, emoji string, at int64) (any, error) {
	rxnToMessage, err := groupChat.ReactToMessage(ctx, clientUsername, groupId, msgId, emoji, at)
	if err != nil {
		return nil, err
	}

	done := rxnToMessage.CHEId != ""

	if !done {
		return done, nil
	}

	// queue msg reaction event
	go eventStreamService.QueueNewGroupMsgReactionEvent(eventTypes.NewGroupMsgReactionEvent{
		FromUser: clientUsername,
		ToGroup:  groupId,
		CHEId:    rxnToMessage.CHEId,
		RxnData:  helpers.ToJson(rxnToMessage),
		ToMsgId:  msgId,
		Emoji:    emoji,
	})

	go func() {
		reactor, err := cache.GetUser[UITypes.ClientUser](context.Background(), clientUsername)
		if err != nil {
			return
		}

		broadcastMsgReaction(groupId, clientUsername, map[string]any{
			"group_id":  groupId,
			"to_msg_id": msgId,
			"reaction": UITypes.MsgReaction{
				Emoji:   emoji,
				Reactor: reactor,
			},
		})
	}()

	return done, nil
}

func RemoveReactionToMessage(ctx context.Context, clientUsername, groupId, msgId string, at int64) (any, error) {
	rrtm, err := groupChat.RemoveReactionToMessage(ctx, clientUsername, groupId, msgId)
	if err != nil {
		return nil, err
	}

	done := rrtm.CHEId != ""

	if !done {
		return done, nil
	}

	// queue reaction removed event
	go eventStreamService.QueueGroupMsgReactionRemovedEvent(eventTypes.GroupMsgReactionRemovedEvent{
		FromUser: clientUsername,
		ToGroup:  groupId,
		ToMsgId:  msgId,
		CHEId:    rrtm.CHEId,
	})

	go broadcastMsgReactionRemoved(groupId, clientUsername, map[string]any{
		"group_id":     groupId,
		"msg_id":       msgId,
		"reactor_user": clientUsername,
	})

	return done, nil
}

func GetChatHistory(ctx context.Context, clientUsername, groupId string, limit int, cursor float64) ([]UITypes.ChatHistoryEntry, error) {
	return groupChat.ChatHistory(ctx, clientUsername, groupId, limit, cursor)
}

func GetGroupInfo(ctx context.Context, groupId string) (UITypes.GroupInfo, error) {
	return groupChat.GroupInfo(ctx, groupId)
}

func GetGroupMembers(ctx context.Context, clientUsername, groupId string, limit int, cursor float64) ([]UITypes.GroupMemberSnippet, error) {
	return groupChat.GroupMembers(ctx, clientUsername, groupId, limit, cursor)
}
