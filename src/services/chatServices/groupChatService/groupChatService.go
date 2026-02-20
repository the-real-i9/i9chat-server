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

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type AuthGroupPicDataT struct {
	UploadUrl         string `msgpack:"uploadUrl"`
	GroupPicCloudName string `msgpack:"groupPicCloudName"`
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

	go func(newGroup groupChat.NewGroup) {
		// history is the same for all init users, we only need separation for the cache
		// so we make use of one user's history
		history := newGroup.InitUsersCHEs[initUsers[0]].([]any)

		newGroup.PictureUrl = cloudStorageService.GroupPicCloudNameToUrl(newGroup.PictureUrl)

		var UIHistory []UITypes.ChatHistoryEntry

		for _, hist := range history {
			hist := hist.(map[string]any)

			UIHistory = append(UIHistory, UITypes.ChatHistoryEntry{CHEType: hist["che_type"].(string), Info: hist["info"].(string), Cursor: float64(hist["cursor"].(int64))})
		}

		initMemberData := map[string]any{
			"chat":    UITypes.ChatSnippet{Type: "group", Group: newGroup, UnreadMC: 2, Cursor: float64(newGroup.ChatCursor)},
			"history": UIHistory,
		}

		broadcastNewGroup(newGroup.InitUsers, initMemberData)
	}(newGroup)

	go func(newGroup groupChat.NewGroup, clientUsername string) {
		eventStreamService.QueueNewGroupEvent(eventTypes.NewGroupEvent{
			CreatorUser:     clientUsername,
			GroupId:         newGroup.Id,
			GroupData:       helpers.ToMsgPack(newGroup),
			InitMembers:     newGroup.InitUsers,
			CreatorUserCHEs: newGroup.ClientUserCHEs,
			InitMembersCHEs: newGroup.InitUsersCHEs,
			ChatCursor:      newGroup.ChatCursor,
		})
	}(newGroup, clientUsername)

	history := newGroup.ClientUserCHEs

	newGroup.PictureUrl = cloudStorageService.GroupPicCloudNameToUrl(newGroup.PictureUrl)

	var UIHistory []UITypes.ChatHistoryEntry

	for _, hist := range history {
		hist := hist.(map[string]any)

		UIHistory = append(UIHistory, UITypes.ChatHistoryEntry{CHEType: hist["che_type"].(string), Info: hist["info"].(string), Cursor: float64(hist["cursor"].(int64))})
	}

	return map[string]any{
		"chat":    UITypes.ChatSnippet{Type: "group", Group: newGroup, UnreadMC: 2, Cursor: float64(newGroup.ChatCursor)},
		"history": UIHistory,
	}, nil
}

func ChangeGroupName(ctx context.Context, groupId, clientUsername, newName string) (UITypes.ChatHistoryEntry, error) {
	newActivity, err := groupChat.ChangeName(ctx, groupId, clientUsername, newName)
	if err != nil {
		return UITypes.ChatHistoryEntry{}, err
	}

	done := newActivity.ClientUserCHE != nil

	if !done {
		return UITypes.ChatHistoryEntry{}, nil
	}

	go broadcastActivityToAll(groupId, UITypes.ChatHistoryEntry{
		CHEType: newActivity.MemberUserCHE["che_type"].(string),
		Info:    newActivity.MemberUserCHE["info"].(string),
		Cursor:  float64(newActivity.MemberUserCHE["cursor"].(int64)),
	}, []any{clientUsername})

	go eventStreamService.QueueGroupEditEvent(eventTypes.GroupEditEvent{
		GroupId:       groupId,
		EditorUser:    clientUsername,
		UpdateKVMap:   map[string]any{"name": newName},
		EditorUserCHE: newActivity.ClientUserCHE,
		MemInfo:       newActivity.MemInfo,
	})

	return UITypes.ChatHistoryEntry{
		CHEType: newActivity.ClientUserCHE["che_type"].(string),
		Info:    newActivity.ClientUserCHE["info"].(string),
		Cursor:  float64(newActivity.ClientUserCHE["cursor"].(int64)),
	}, nil
}

func ChangeGroupDescription(ctx context.Context, groupId, clientUsername, newDescription string) (UITypes.ChatHistoryEntry, error) {
	newActivity, err := groupChat.ChangeDescription(ctx, groupId, clientUsername, newDescription)
	if err != nil {
		return UITypes.ChatHistoryEntry{}, err
	}

	done := newActivity.ClientUserCHE != nil

	if !done {
		return UITypes.ChatHistoryEntry{}, nil
	}

	go broadcastActivityToAll(groupId, UITypes.ChatHistoryEntry{
		CHEType: newActivity.MemberUserCHE["che_type"].(string),
		Info:    newActivity.MemberUserCHE["info"].(string),
		Cursor:  float64(newActivity.MemberUserCHE["cursor"].(int64)),
	}, []any{clientUsername})

	go eventStreamService.QueueGroupEditEvent(eventTypes.GroupEditEvent{
		GroupId:       groupId,
		EditorUser:    clientUsername,
		UpdateKVMap:   map[string]any{"description": newDescription},
		EditorUserCHE: newActivity.ClientUserCHE,
		MemInfo:       newActivity.MemInfo,
	})

	return UITypes.ChatHistoryEntry{
		CHEType: newActivity.ClientUserCHE["che_type"].(string),
		Info:    newActivity.ClientUserCHE["info"].(string),
		Cursor:  float64(newActivity.ClientUserCHE["cursor"].(int64)),
	}, nil
}

func ChangeGroupPicture(ctx context.Context, groupId, clientUsername, picCloudName string) (UITypes.ChatHistoryEntry, error) {
	newActivity, err := groupChat.ChangePicture(ctx, groupId, clientUsername, picCloudName)
	if err != nil {
		return UITypes.ChatHistoryEntry{}, err
	}

	done := newActivity.ClientUserCHE != nil

	if !done {
		return UITypes.ChatHistoryEntry{}, nil
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
		Cursor:  float64(newActivity.MemberUserCHE["cursor"].(int64)),
	}, []any{clientUsername})

	return UITypes.ChatHistoryEntry{
		CHEType: newActivity.ClientUserCHE["che_type"].(string),
		Info:    newActivity.ClientUserCHE["info"].(string),
		Cursor:  float64(newActivity.ClientUserCHE["cursor"].(int64)),
	}, nil
}

func AddUsersToGroup(ctx context.Context, groupId, clientUsername string, newUsers []string) (UITypes.ChatHistoryEntry, error) {
	newActivity, err := groupChat.AddUsers(ctx, groupId, clientUsername, newUsers)
	if err != nil {
		return UITypes.ChatHistoryEntry{}, err
	}

	done := newActivity.GroupInfo != nil
	if !done {
		return UITypes.ChatHistoryEntry{}, nil
	}

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
			"chat":    UITypes.ChatSnippet{Type: "group", Group: groupInfo, UnreadMC: 1, Cursor: float64(newActivity.ChatCursor)},
			"history": []UITypes.ChatHistoryEntry{{CHEType: che["che_type"].(string), Info: che["info"].(string), Cursor: float64(che["cursor"].(int64))}},
		}

		broadcastNewGroup(newActivity.NewUsernames, newMemData)
	}(newActivity)

	go broadcastActivityToAll(groupId, UITypes.ChatHistoryEntry{
		CHEType: newActivity.MemberUserCHE["che_type"].(string),
		Info:    newActivity.MemberUserCHE["info"].(string),
		Cursor:  float64(newActivity.MemberUserCHE["cursor"].(int64)),
	}, append(newActivity.NewUsernames, clientUsername))

	go eventStreamService.QueueGroupUsersAddedEvent(eventTypes.GroupUsersAddedEvent{
		GroupId:       groupId,
		Admin:         clientUsername,
		NewMembers:    newActivity.NewUsernames,
		AdminCHE:      newActivity.ClientUserCHE,
		NewMembersCHE: newActivity.NewUsersCHE,
		MemInfo:       newActivity.MemInfo,
		ChatCursor:    newActivity.ChatCursor,
	})

	return UITypes.ChatHistoryEntry{
		CHEType: newActivity.ClientUserCHE["che_type"].(string),
		Info:    newActivity.ClientUserCHE["info"].(string),
		Cursor:  float64(newActivity.ClientUserCHE["cursor"].(int64)),
	}, nil
}

func RemoveUserFromGroup(ctx context.Context, groupId, clientUsername, targetUser string) (UITypes.ChatHistoryEntry, error) {
	newActivity, err := groupChat.RemoveUser(ctx, groupId, clientUsername, targetUser)
	if err != nil {
		return UITypes.ChatHistoryEntry{}, err
	}

	done := newActivity.ClientUserCHE != nil
	if !done {
		return UITypes.ChatHistoryEntry{}, nil
	}

	go broadcastActivityToOne(groupId, UITypes.ChatHistoryEntry{
		CHEType: newActivity.TargetUserCHE["che_type"].(string),
		Info:    newActivity.TargetUserCHE["info"].(string),
		Cursor:  float64(newActivity.TargetUserCHE["cursor"].(int64)),
	}, targetUser)

	go broadcastActivityToAll(groupId, UITypes.ChatHistoryEntry{
		CHEType: newActivity.MemberUserCHE["che_type"].(string),
		Info:    newActivity.MemberUserCHE["info"].(string),
		Cursor:  float64(newActivity.MemberUserCHE["cursor"].(int64)),
	}, []any{clientUsername, targetUser})

	go eventStreamService.QueueGroupUserRemovedEvent(eventTypes.GroupUserRemovedEvent{
		GroupId:      groupId,
		Admin:        clientUsername,
		OldMember:    targetUser,
		AdminCHE:     newActivity.ClientUserCHE,
		OldMemberCHE: newActivity.TargetUserCHE,
		MemInfo:      newActivity.MemInfo,
	})

	return UITypes.ChatHistoryEntry{
		CHEType: newActivity.ClientUserCHE["che_type"].(string),
		Info:    newActivity.ClientUserCHE["info"].(string),
		Cursor:  float64(newActivity.ClientUserCHE["cursor"].(int64)),
	}, nil
}

func JoinGroup(ctx context.Context, groupId, clientUsername string) (map[string]any, error) {
	newActivity, err := groupChat.Join(ctx, groupId, clientUsername)
	if err != nil {
		return nil, err
	}

	done := newActivity.GroupInfo != nil
	if !done {
		return nil, nil
	}

	go broadcastActivityToAll(groupId, UITypes.ChatHistoryEntry{
		CHEType: newActivity.MemberUserCHE["che_type"].(string),
		Info:    newActivity.MemberUserCHE["info"].(string),
		Cursor:  float64(newActivity.MemberUserCHE["cursor"].(int64)),
	}, []any{clientUsername})

	go eventStreamService.QueueGroupUserJoinedEvent(eventTypes.GroupUserJoinedEvent{
		GroupId:      groupId,
		NewMember:    clientUsername,
		NewMemberCHE: newActivity.ClientUserCHE,
		MemInfo:      newActivity.MemInfo,
		ChatCursor:   newActivity.ChatCursor,
	})

	che := newActivity.ClientUserCHE

	groupInfo := newActivity.GroupInfo

	groupInfo["picture_url"] = cloudStorageService.GroupPicCloudNameToUrl(groupInfo["picture_url"].(string))

	return map[string]any{
		"chat":    UITypes.ChatSnippet{Type: "group", Group: groupInfo, UnreadMC: 1, Cursor: float64(newActivity.ChatCursor)},
		"history": []UITypes.ChatHistoryEntry{{CHEType: che["che_type"].(string), Info: che["info"].(string), Cursor: float64(che["cursor"].(int64))}},
	}, nil
}

func LeaveGroup(ctx context.Context, groupId, clientUsername string) (UITypes.ChatHistoryEntry, error) {
	newActivity, err := groupChat.Leave(ctx, groupId, clientUsername)
	if err != nil {
		return UITypes.ChatHistoryEntry{}, err
	}

	done := newActivity.ClientUserCHE != nil
	if !done {
		return UITypes.ChatHistoryEntry{}, nil
	}

	go broadcastActivityToAll(groupId, UITypes.ChatHistoryEntry{
		CHEType: newActivity.MemberUserCHE["che_type"].(string),
		Info:    newActivity.MemberUserCHE["info"].(string),
		Cursor:  float64(newActivity.MemberUserCHE["cursor"].(int64)),
	}, []any{clientUsername})

	go eventStreamService.QueueGroupUserLeftEvent(eventTypes.GroupUserLeftEvent{
		GroupId:      groupId,
		OldMember:    clientUsername,
		OldMemberCHE: newActivity.ClientUserCHE,
		MemInfo:      newActivity.MemInfo,
	})

	return UITypes.ChatHistoryEntry{
		CHEType: newActivity.ClientUserCHE["che_type"].(string),
		Info:    newActivity.ClientUserCHE["info"].(string),
		Cursor:  float64(newActivity.ClientUserCHE["cursor"].(int64)),
	}, nil
}

func MakeUserGroupAdmin(ctx context.Context, groupId, clientUsername, targetUser string) (UITypes.ChatHistoryEntry, error) {
	newActivity, err := groupChat.MakeUserAdmin(ctx, groupId, clientUsername, targetUser)
	if err != nil {
		return UITypes.ChatHistoryEntry{}, err
	}

	done := newActivity.ClientUserCHE != nil
	if !done {
		return UITypes.ChatHistoryEntry{}, nil
	}

	go broadcastActivityToOne(groupId, UITypes.ChatHistoryEntry{
		CHEType: newActivity.TargetUserCHE["che_type"].(string),
		Info:    newActivity.TargetUserCHE["info"].(string),
		Cursor:  float64(newActivity.TargetUserCHE["cursor"].(int64)),
	}, targetUser)

	go broadcastActivityToAll(groupId, UITypes.ChatHistoryEntry{
		CHEType: newActivity.MemberUserCHE["che_type"].(string),
		Info:    newActivity.MemberUserCHE["info"].(string),
		Cursor:  float64(newActivity.MemberUserCHE["cursor"].(int64)),
	}, []any{clientUsername, targetUser})

	go eventStreamService.QueueGroupMakeUserAdminEvent(eventTypes.GroupMakeUserAdminEvent{
		GroupId:     groupId,
		Admin:       clientUsername,
		NewAdmin:    targetUser,
		AdminCHE:    newActivity.ClientUserCHE,
		NewAdminCHE: newActivity.TargetUserCHE,
		MemInfo:     newActivity.MemInfo,
	})

	return UITypes.ChatHistoryEntry{
		CHEType: newActivity.ClientUserCHE["che_type"].(string),
		Info:    newActivity.ClientUserCHE["info"].(string),
		Cursor:  float64(newActivity.ClientUserCHE["cursor"].(int64)),
	}, nil
}

func RemoveUserFromGroupAdmins(ctx context.Context, groupId, clientUsername, targetUser string) (UITypes.ChatHistoryEntry, error) {
	newActivity, err := groupChat.RemoveUserFromAdmins(ctx, groupId, clientUsername, targetUser)
	if err != nil {
		return UITypes.ChatHistoryEntry{}, err
	}

	done := newActivity.ClientUserCHE != nil
	if !done {
		return UITypes.ChatHistoryEntry{}, nil
	}

	go broadcastActivityToOne(groupId, UITypes.ChatHistoryEntry{
		CHEType: newActivity.TargetUserCHE["che_type"].(string),
		Info:    newActivity.TargetUserCHE["info"].(string),
		Cursor:  float64(newActivity.TargetUserCHE["cursor"].(int64)),
	}, targetUser)

	go broadcastActivityToAll(groupId, UITypes.ChatHistoryEntry{
		CHEType: newActivity.MemberUserCHE["che_type"].(string),
		Info:    newActivity.MemberUserCHE["info"].(string),
		Cursor:  float64(newActivity.MemberUserCHE["cursor"].(int64)),
	}, []any{clientUsername, targetUser})

	go eventStreamService.QueueGroupRemoveUserFromAdminsEvent(eventTypes.GroupRemoveUserFromAdminsEvent{
		GroupId:     groupId,
		Admin:       clientUsername,
		OldAdmin:    targetUser,
		AdminCHE:    newActivity.ClientUserCHE,
		OldAdminCHE: newActivity.TargetUserCHE,
		MemInfo:     newActivity.MemInfo,
	})

	return UITypes.ChatHistoryEntry{
		CHEType: newActivity.ClientUserCHE["che_type"].(string),
		Info:    newActivity.ClientUserCHE["info"].(string),
		Cursor:  float64(newActivity.ClientUserCHE["cursor"].(int64)),
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

	go func(msg groupChat.NewMessage, clientUsername string) {
		uisender, _ := cache.GetUser[UITypes.ClientUser](context.Background(), clientUsername)

		uisender.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(uisender.ProfilePicUrl)

		UImsg := UITypes.ChatHistoryEntry{
			CHEType: msg.CHEType, Id: msg.Id,
			Content:        cloudStorageService.MessageMediaCloudNameToUrl(msg.Content),
			DeliveryStatus: msg.DeliveryStatus, CreatedAt: msg.CreatedAt,
			Sender: uisender, ReplyTargetMsg: msg.ReplyTargetMsg, Cursor: float64(msg.Cursor),
		}

		broadcastNewMessage(groupId, UImsg, clientUsername)
	}(newMessage, clientUsername)

	// queue new message event
	go func(newMessage groupChat.NewMessage, clientUsername, groupId string) {
		eventStreamService.QueueNewGroupMessageEvent(eventTypes.NewGroupMessageEvent{
			FromUser:  clientUsername,
			ToGroup:   groupId,
			CHEId:     newMessage.Id,
			MsgData:   helpers.ToMsgPack(newMessage),
			CHECursor: newMessage.Cursor,
		})
	}(newMessage, clientUsername, groupId)

	return map[string]any{"new_msg_id": newMessage.Id, "che_cursor": newMessage.Cursor}, nil
}

func AckMessagesDelivered(ctx context.Context, clientUsername, groupId string, msgIds []any, deliveredAt int64) (map[string]any, error) {
	lastMsgCursor, err := groupChat.AckMessagesDelivered(ctx, clientUsername, groupId, msgIds, deliveredAt)
	if err != nil {
		return nil, err
	}

	if lastMsgCursor == 0 {
		return nil, nil
	}

	// queue msg ack event
	go eventStreamService.QueueGroupMsgAckEvent(eventTypes.GroupMsgAckEvent{
		FromUser:   clientUsername,
		ToGroup:    groupId,
		CHEIds:     msgIds,
		Ack:        "delivered",
		At:         deliveredAt,
		ChatCursor: lastMsgCursor,
	})

	return map[string]any{"updated_chat_cursor": lastMsgCursor}, nil
}

func AckMessagesRead(ctx context.Context, clientUsername, groupId string, msgIds []any, readAt int64) (bool, error) {
	done, err := groupChat.AckMessagesRead(ctx, clientUsername, groupId, msgIds, readAt)
	if err != nil {
		return false, err
	}

	if done {
		go eventStreamService.QueueGroupMsgAckEvent(eventTypes.GroupMsgAckEvent{
			FromUser: clientUsername,
			ToGroup:  groupId,
			CHEIds:   msgIds,
			Ack:      "read",
			At:       readAt,
		})
	}

	return done, nil
}

func ReactToMessage(ctx context.Context, clientUsername, groupId, msgId, emoji string, at int64) (map[string]any, error) {
	rxnToMessage, err := groupChat.ReactToMessage(ctx, clientUsername, groupId, msgId, emoji, at)
	if err != nil {
		return nil, err
	}

	if rxnToMessage.CHEId == "" {
		return nil, nil
	}

	go func(rxnData groupChat.RxnToMessage, clientUsername, groupId string) {
		uireactor, _ := cache.GetUser[UITypes.MsgReactor](context.Background(), clientUsername)

		uireactor.ProfilePicUrl = cloudStorageService.ProfilePicCloudNameToUrl(uireactor.ProfilePicUrl)

		broadcastMsgReaction(groupId, clientUsername, map[string]any{
			"group_id": groupId,
			"che":      UITypes.ChatHistoryEntry{CHEType: rxnData.CHEType, Reactor: clientUsername, Emoji: rxnData.Emoji, Cursor: float64(rxnData.Cursor)},
			"msg_reaction": map[string]any{
				"msg_id": rxnData.ToMsgId,
				"reaction": UITypes.MsgReaction{
					Emoji:   rxnData.Emoji,
					Reactor: uireactor,
				},
			},
		})
	}(rxnToMessage, clientUsername, groupId)

	go func(rxnData groupChat.RxnToMessage, clientUsername, groupId string) {
		eventStreamService.QueueNewGroupMsgReactionEvent(eventTypes.NewGroupMsgReactionEvent{
			FromUser:  clientUsername,
			ToGroup:   groupId,
			CHEId:     rxnData.CHEId,
			RxnData:   helpers.ToMsgPack(rxnData),
			ToMsgId:   rxnData.ToMsgId,
			Emoji:     rxnData.Emoji,
			CHECursor: rxnData.Cursor,
		})
	}(rxnToMessage, clientUsername, groupId)

	return map[string]any{"che_cursor": rxnToMessage.Cursor}, nil
}

func RemoveReactionToMessage(ctx context.Context, clientUsername, groupId, msgId string, at int64) (bool, error) {
	CHEId, err := groupChat.RemoveReactionToMessage(ctx, clientUsername, groupId, msgId)
	if err != nil {
		return false, err
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

func GetChatHistory(ctx context.Context, clientUsername, groupId string, limit int64, cursor float64) ([]UITypes.ChatHistoryEntry, error) {
	return groupChat.ChatHistory(ctx, clientUsername, groupId, limit, cursor)
}

func GetGroupInfo(ctx context.Context, groupId string) (UITypes.GroupInfo, error) {
	return groupChat.GroupInfo(ctx, groupId)
}

func GetGroupMembers(ctx context.Context, clientUsername, groupId string, limit int64, cursor float64) ([]UITypes.GroupMemberSnippet, error) {
	return groupChat.GroupMembers(ctx, clientUsername, groupId, limit, cursor)
}
