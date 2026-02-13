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

	if newGroup.Id == "" {
		return nil, nil
	}

	go eventStreamService.QueueNewGroupEvent(eventTypes.NewGroupEvent{
		CreatorUser:     clientUsername,
		GroupId:         newGroup.Id,
		GroupData:       helpers.ToJson(newGroup),
		InitMembers:     newGroup.InitUsers,
		CreatorUserCHEs: newGroup.ClientUserCHEs,
		InitMembersCHEs: newGroup.InitUsersCHEs,
		ChatCursor:      newGroup.ChatCursor,
	})

	go func(newGroup groupChat.NewGroup) {
		// history is the same for all init users, we only need separation for the cache
		// so we make use of one user's history
		history := newGroup.InitUsersCHEs[initUsers[0]].([]any)

		newGroup.PictureUrl = cloudStorageService.GroupPicCloudNameToUrl(newGroup.PictureUrl)

		var UIHistory []UITypes.ChatHistoryEntry

		for _, hist := range history {
			hist := hist.(map[string]any)

			UIHistory = append(UIHistory, UITypes.ChatHistoryEntry{CHEType: hist["che_type"].(string), Info: hist["info"].(string), Cursor: hist["cursor"].(int64)})
		}

		initMemberData := map[string]any{
			"chat":    UITypes.ChatSnippet{Type: "group", Group: newGroup, UnreadMC: 2, Cursor: newGroup.ChatCursor},
			"history": UIHistory,
		}

		broadcastNewGroup(newGroup.InitUsers, initMemberData)
	}(newGroup)

	history := newGroup.ClientUserCHEs

	newGroup.PictureUrl = cloudStorageService.GroupPicCloudNameToUrl(newGroup.PictureUrl)

	var UIHistory []UITypes.ChatHistoryEntry

	for _, hist := range history {
		hist := hist.(map[string]any)

		UIHistory = append(UIHistory, UITypes.ChatHistoryEntry{CHEType: hist["che_type"].(string), Info: hist["info"].(string), Cursor: hist["cursor"].(int64)})
	}

	clientData := map[string]any{
		"chat":    UITypes.ChatSnippet{Type: "group", Group: newGroup, UnreadMC: 2, Cursor: newGroup.ChatCursor},
		"history": UIHistory,
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

	go broadcastActivityToAll(groupId, UITypes.ChatHistoryEntry{
		CHEType: newActivity.MemberUserCHE["che_type"].(string),
		Info:    newActivity.MemberUserCHE["info"].(string),
		Cursor:  newActivity.MemberUserCHE["cursor"].(int64),
	}, []any{clientUsername})

	return UITypes.ChatHistoryEntry{
		CHEType: newActivity.ClientUserCHE["che_type"].(string),
		Info:    newActivity.ClientUserCHE["info"].(string),
		Cursor:  newActivity.ClientUserCHE["cursor"].(int64),
	}, nil
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

	go broadcastActivityToAll(groupId, UITypes.ChatHistoryEntry{
		CHEType: newActivity.MemberUserCHE["che_type"].(string),
		Info:    newActivity.MemberUserCHE["info"].(string),
		Cursor:  newActivity.MemberUserCHE["cursor"].(int64),
	}, []any{clientUsername})

	return UITypes.ChatHistoryEntry{
		CHEType: newActivity.ClientUserCHE["che_type"].(string),
		Info:    newActivity.ClientUserCHE["info"].(string),
		Cursor:  newActivity.ClientUserCHE["cursor"].(int64),
	}, nil
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
		UpdateKVMap:   map[string]any{"picture_url": picCloudName},
		EditorUserCHE: newActivity.ClientUserCHE,
		MemInfo:       newActivity.MemInfo,
	})

	go broadcastActivityToAll(groupId, UITypes.ChatHistoryEntry{
		CHEType: newActivity.MemberUserCHE["che_type"].(string),
		Info:    newActivity.MemberUserCHE["info"].(string),
		Cursor:  newActivity.MemberUserCHE["cursor"].(int64),
	}, []any{clientUsername})

	return UITypes.ChatHistoryEntry{
		CHEType: newActivity.ClientUserCHE["che_type"].(string),
		Info:    newActivity.ClientUserCHE["info"].(string),
		Cursor:  newActivity.ClientUserCHE["cursor"].(int64),
	}, nil
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
		che := newActivity.NewUsersCHE[newActivity.NewUsernames[0].(string)].(map[string]any)

		groupInfo := newActivity.GroupInfo

		groupInfo["picture_url"] = cloudStorageService.GroupPicCloudNameToUrl(groupInfo["picture_url"].(string))

		newMemData := map[string]any{
			"chat":    UITypes.ChatSnippet{Type: "group", Group: groupInfo, UnreadMC: 1, Cursor: newActivity.ChatCursor},
			"history": []UITypes.ChatHistoryEntry{UITypes.ChatHistoryEntry{CHEType: che["che_type"].(string), Info: che["info"].(string), Cursor: che["cursor"].(int64)}},
		}

		broadcastNewGroup(newActivity.NewUsernames, newMemData)
	}(newActivity)

	go broadcastActivityToAll(groupId, UITypes.ChatHistoryEntry{
		CHEType: newActivity.MemberUserCHE["che_type"].(string),
		Info:    newActivity.MemberUserCHE["info"].(string),
		Cursor:  newActivity.MemberUserCHE["cursor"].(int64),
	}, append(newActivity.NewUsernames, clientUsername))

	return UITypes.ChatHistoryEntry{
		CHEType: newActivity.ClientUserCHE["che_type"].(string),
		Info:    newActivity.ClientUserCHE["info"].(string),
		Cursor:  newActivity.ClientUserCHE["cursor"].(int64),
	}, nil
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

	go broadcastActivityToOne(groupId, UITypes.ChatHistoryEntry{
		CHEType: newActivity.TargetUserCHE["che_type"].(string),
		Info:    newActivity.TargetUserCHE["info"].(string),
		Cursor:  newActivity.TargetUserCHE["cursor"].(int64),
	}, targetUser)

	go broadcastActivityToAll(groupId, UITypes.ChatHistoryEntry{
		CHEType: newActivity.MemberUserCHE["che_type"].(string),
		Info:    newActivity.MemberUserCHE["info"].(string),
		Cursor:  newActivity.MemberUserCHE["cursor"].(int64),
	}, []any{clientUsername, targetUser})

	return UITypes.ChatHistoryEntry{
		CHEType: newActivity.ClientUserCHE["che_type"].(string),
		Info:    newActivity.ClientUserCHE["info"].(string),
		Cursor:  newActivity.ClientUserCHE["cursor"].(int64),
	}, nil
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

	go broadcastActivityToAll(groupId, UITypes.ChatHistoryEntry{
		CHEType: newActivity.MemberUserCHE["che_type"].(string),
		Info:    newActivity.MemberUserCHE["info"].(string),
		Cursor:  newActivity.MemberUserCHE["cursor"].(int64),
	}, []any{clientUsername})

	che := newActivity.ClientUserCHE

	groupInfo := newActivity.GroupInfo

	groupInfo["picture_url"] = cloudStorageService.GroupPicCloudNameToUrl(groupInfo["picture_url"].(string))

	clientData := map[string]any{
		"chat":    UITypes.ChatSnippet{Type: "group", Group: groupInfo, UnreadMC: 1, Cursor: newActivity.ChatCursor},
		"history": []UITypes.ChatHistoryEntry{UITypes.ChatHistoryEntry{CHEType: che["che_type"].(string), Info: che["info"].(string), Cursor: che["cursor"].(int64)}},
	}

	return clientData, nil
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

	go broadcastActivityToAll(groupId, UITypes.ChatHistoryEntry{
		CHEType: newActivity.MemberUserCHE["che_type"].(string),
		Info:    newActivity.MemberUserCHE["info"].(string),
		Cursor:  newActivity.MemberUserCHE["cursor"].(int64),
	}, []any{clientUsername})

	return UITypes.ChatHistoryEntry{
		CHEType: newActivity.ClientUserCHE["che_type"].(string),
		Info:    newActivity.ClientUserCHE["info"].(string),
		Cursor:  newActivity.ClientUserCHE["cursor"].(int64),
	}, nil
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

	go broadcastActivityToOne(groupId, UITypes.ChatHistoryEntry{
		CHEType: newActivity.TargetUserCHE["che_type"].(string),
		Info:    newActivity.TargetUserCHE["info"].(string),
		Cursor:  newActivity.TargetUserCHE["cursor"].(int64),
	}, targetUser)

	go broadcastActivityToAll(groupId, UITypes.ChatHistoryEntry{
		CHEType: newActivity.MemberUserCHE["che_type"].(string),
		Info:    newActivity.MemberUserCHE["info"].(string),
		Cursor:  newActivity.MemberUserCHE["cursor"].(int64),
	}, []any{clientUsername, targetUser})

	return UITypes.ChatHistoryEntry{
		CHEType: newActivity.ClientUserCHE["che_type"].(string),
		Info:    newActivity.ClientUserCHE["info"].(string),
		Cursor:  newActivity.ClientUserCHE["cursor"].(int64),
	}, nil
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

	go broadcastActivityToOne(groupId, UITypes.ChatHistoryEntry{
		CHEType: newActivity.TargetUserCHE["che_type"].(string),
		Info:    newActivity.TargetUserCHE["info"].(string),
		Cursor:  newActivity.TargetUserCHE["cursor"].(int64),
	}, targetUser)

	go broadcastActivityToAll(groupId, UITypes.ChatHistoryEntry{
		CHEType: newActivity.MemberUserCHE["che_type"].(string),
		Info:    newActivity.MemberUserCHE["info"].(string),
		Cursor:  newActivity.MemberUserCHE["cursor"].(int64),
	}, []any{clientUsername, targetUser})

	return UITypes.ChatHistoryEntry{
		CHEType: newActivity.ClientUserCHE["che_type"].(string),
		Info:    newActivity.ClientUserCHE["info"].(string),
		Cursor:  newActivity.ClientUserCHE["cursor"].(int64),
	}, nil
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
		FromUser:  clientUsername,
		ToGroup:   groupId,
		CHEId:     newMessage.Id,
		MsgData:   helpers.ToJson(newMessage),
		CHECursor: newMessage.Cursor,
	})

	go func(msg groupChat.NewMessage) {
		uisender, _ := cache.GetUser[UITypes.ClientUser](context.Background(), clientUsername)

		uisender.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(uisender.ProfilePicUrl)

		cloudStorageService.MessageMediaCloudNameToUrl(msg.Content)

		UImsg := UITypes.ChatHistoryEntry{CHEType: msg.CHEType, Id: msg.Id, Content: msg.Content, DeliveryStatus: msg.DeliveryStatus, CreatedAt: msg.CreatedAt, Sender: uisender, ReplyTargetMsg: msg.ReplyTargetMsg, Cursor: msg.Cursor}

		broadcastNewMessage(groupId, UImsg, clientUsername)
	}(newMessage)

	return map[string]any{"new_msg_id": newMessage.Id, "che_cursor": newMessage.Cursor}, nil
}

func AckMessageDelivered(ctx context.Context, clientUsername, groupId, msgId string, deliveredAt int64) (map[string]any, error) {
	msgCursor, err := groupChat.AckMessageDelivered(ctx, clientUsername, groupId, msgId, deliveredAt)
	if err != nil {
		return nil, err
	}

	if msgCursor == nil {
		return nil, nil
	}

	// queue msg ack event
	go eventStreamService.QueueGroupMsgAckEvent(eventTypes.GroupMsgAckEvent{
		FromUser:   clientUsername,
		ToGroup:    groupId,
		CHEId:      msgId,
		Ack:        "delivered",
		At:         deliveredAt,
		ChatCursor: *msgCursor,
	})

	return map[string]any{"updated_chat_cursor": *msgCursor}, nil
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

	if rxnToMessage.CHEId == "" {
		return nil, nil
	}

	go func(rxnData groupChat.RxnToMessage) {
		uireactor, _ := cache.GetUser[UITypes.MsgReactor](context.Background(), clientUsername)

		uireactor.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(uireactor.ProfilePicUrl)

		broadcastMsgReaction(groupId, clientUsername, map[string]any{
			"group_id": groupId,
			"che":      UITypes.ChatHistoryEntry{CHEType: rxnData.CHEType, Reactor: clientUsername, Emoji: rxnData.Emoji, Cursor: rxnData.Cursor},
			"msg_reaction": map[string]any{
				"msg_id": msgId,
				"reaction": UITypes.MsgReaction{
					Emoji:   emoji,
					Reactor: uireactor,
				},
			},
		})
	}(rxnToMessage)

	go func(rxnData groupChat.RxnToMessage) {
		eventStreamService.QueueNewGroupMsgReactionEvent(eventTypes.NewGroupMsgReactionEvent{
			FromUser:  clientUsername,
			ToGroup:   groupId,
			CHEId:     rxnData.CHEId,
			RxnData:   helpers.ToJson(rxnData),
			ToMsgId:   msgId,
			Emoji:     emoji,
			CHECursor: rxnData.Cursor,
		})
	}(rxnToMessage)

	return map[string]any{"che_cursor": rxnToMessage.Cursor}, nil
}

func RemoveReactionToMessage(ctx context.Context, clientUsername, groupId, msgId string, at int64) (any, error) {
	CHEId, err := groupChat.RemoveReactionToMessage(ctx, clientUsername, groupId, msgId)
	if err != nil {
		return nil, err
	}

	done := CHEId != ""

	// queue reaction removed event
	if done {
		go broadcastMsgReactionRemoved(groupId, clientUsername, map[string]any{
			"group_id":     groupId,
			"msg_id":       msgId,
			"reactor_user": clientUsername,
		})

		go eventStreamService.QueueGroupMsgReactionRemovedEvent(eventTypes.GroupMsgReactionRemovedEvent{
			FromUser: clientUsername,
			ToGroup:  groupId,
			ToMsgId:  msgId,
			CHEId:    CHEId,
		})
	}

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
