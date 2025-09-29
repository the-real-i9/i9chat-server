package groupChatService

import (
	"context"
	"encoding/json"
	"fmt"
	"i9chat/src/appTypes"
	groupChat "i9chat/src/models/chatModel/groupChatModel"
	"i9chat/src/services/appServices"
	"i9chat/src/services/cloudStorageService"
	"log"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gofiber/fiber/v2"
)

func NewGroupChat(ctx context.Context, clientUsername, name, description string, pictureData []byte, initUsers []string, createdAt int64) (map[string]any, error) {
	picUrl, err := uploadGroupPicture(ctx, pictureData)
	if err != nil {
		return nil, err
	}

	newGroupChat, err := groupChat.New(ctx, clientUsername, name, description, picUrl, initUsers, time.UnixMilli(createdAt).UTC())
	if err != nil {
		return nil, err
	}

	if newGroupChat.InitMemberData != nil {
		go broadcastNewGroup(initUsers, newGroupChat.InitMemberData)
	}

	return newGroupChat.ClientData, nil
}

func GetChatHistory(ctx context.Context, clientUsername, groupId string, limit int, offset int64) (any, error) {
	return groupChat.ChatHistory(ctx, clientUsername, groupId, limit, time.UnixMilli(offset).UTC())
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

	msgContentJson, err := json.Marshal(*msgContent)
	if err != nil {
		log.Println("groupChatService.go: SendMessage: json.Marshal:", err)
		return nil, fiber.ErrInternalServerError
	}

	var newMessage groupChat.NewMessage

	if !isReply {
		newMessage, err = groupChat.SendMessage(ctx, clientUsername, groupId, string(msgContentJson), time.UnixMilli(at).UTC())
		if err != nil {
			return nil, err
		}
	} else {
		newMessage, err = groupChat.ReplyToMessage(ctx, clientUsername, groupId, replyTargetMsgId, string(msgContentJson), time.UnixMilli(at).UTC())
		if err != nil {
			return nil, err
		}
	}

	if newMessage.MemberData != nil {
		go broadcastNewMessage(newMessage.MemberUsernames, newMessage.MemberData, groupId)
	}

	return newMessage.ClientData, nil
}

func AckMessageDelivered(ctx context.Context, clientUsername, groupId, msgId string, deliveredAt int64) (any, error) {
	msgAck, err := groupChat.AckMessageDelivered(ctx, clientUsername, groupId, msgId, time.UnixMilli(deliveredAt).UTC())
	if err != nil {
		return nil, err
	}

	if msgAck.All {
		go broadcastMsgDelivered(msgAck.MemberUsernames, map[string]any{
			"group_id": groupId,
			"msg_id":   msgId,
		})
	}

	return map[string]any{"delivered_to_all": msgAck.All}, nil
}

func AckMessageRead(ctx context.Context, clientUsername, groupId, msgId string, readAt int64) (any, error) {
	msgAck, err := groupChat.AckMessageRead(ctx, clientUsername, groupId, msgId, time.UnixMilli(readAt).UTC())
	if err != nil {
		return nil, err
	}

	if msgAck.All {
		go broadcastMsgRead(msgAck.MemberUsernames, map[string]any{
			"group_id": groupId,
			"msg_id":   msgId,
		})
	}

	return map[string]any{"read_by_all": msgAck.All}, nil
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

func ChangeGroupName(ctx context.Context, groupId, clientUsername, newName string) (any, error) {
	newActivity, err := groupChat.ChangeName(ctx, groupId, clientUsername, newName)
	if err != nil {
		return nil, err
	}

	if newActivity.MemberData != nil {
		go broadcastActivity(newActivity.MemberUsernames, newActivity.MemberData, groupId)
	}

	return newActivity.ClientData, nil
}

func ChangeGroupDescription(ctx context.Context, groupId, clientUsername, newDescription string) (any, error) {
	newActivity, err := groupChat.ChangeDescription(ctx, groupId, clientUsername, newDescription)
	if err != nil {
		return nil, err
	}

	if newActivity.MemberData != nil {
		go broadcastActivity(newActivity.MemberUsernames, newActivity.MemberData, groupId)
	}

	return newActivity.ClientData, nil
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

	if newActivity.MemberData != nil {
		go broadcastActivity(newActivity.MemberUsernames, newActivity.MemberData, groupId)
	}

	return newActivity.ClientData, nil
}

func uploadGroupPicture(ctx context.Context, pictureData []byte) (string, error) {
	mediaMIME := mimetype.Detect(pictureData)
	mediaType, mediaExt := mediaMIME.String(), mediaMIME.Extension()

	if !strings.HasPrefix(mediaType, "image") {
		return "", fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("invalid picture type %s. expected image/*", mediaType))
	}

	picPath := fmt.Sprintf("group_chat_pictures/group_chat_pic_%d%s", time.Now().UnixNano(), mediaExt)

	picUrl, err := cloudStorageService.Upload(ctx, picPath, pictureData)

	if err != nil {
		return "", err
	}

	return picUrl, nil
}

func AddUsersToGroup(ctx context.Context, groupId, clientUsername string, newUsers []string) (any, error) {
	newActivity, newUserData, err := groupChat.AddUsers(ctx, groupId, clientUsername, newUsers)
	if err != nil {
		return nil, err
	}

	if newActivity.MemberData != nil {
		go func() {
			broadcastNewGroup(newUsers, newUserData)

			broadcastActivity(newActivity.MemberUsernames, newActivity.MemberData, groupId)
		}()
	}

	return newActivity.ClientData, nil
}

func RemoveUserFromGroup(ctx context.Context, groupId, clientUsername, targetUser string) (any, error) {
	newActivity, targetUserData, err := groupChat.RemoveUser(ctx, groupId, clientUsername, targetUser)
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
	newActivity, newAdminData, err := groupChat.MakeUserAdmin(ctx, groupId, clientUsername, targetUser)
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
