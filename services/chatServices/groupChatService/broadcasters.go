package groupChatService

import (
	"fmt"
	groupChat "i9chat/models/chatModel/groupChatModel"
	"i9chat/services/messageBrokerService"
)

func broadcastNewGroup(targetUsers []string, targetUserData any) {
	for _, tu := range targetUsers {

		messageBrokerService.Send(fmt.Sprintf("user-%s-topic", tu), messageBrokerService.Message{
			Event: "new group chat",
			Data:  targetUserData,
		})
	}
}

func broadcastNewMessage(memberUsernames []string, memberData map[string]any) {
	for _, mu := range memberUsernames {

		messageBrokerService.Send(fmt.Sprintf("user-%s-topic", mu), messageBrokerService.Message{
			Event: "new group chat message",
			Data:  memberData,
		})
	}
}

func broadcastActivity(newActivity groupChat.NewActivity, groupId string) {

	for _, mu := range newActivity.MemberUsernames {
		messageBrokerService.Send(fmt.Sprintf("user-%s-topic", mu), messageBrokerService.Message{
			Event: "new group chat activity",
			Data: map[string]any{
				"info":     newActivity.MemberData,
				"group_id": groupId,
			},
		})
	}
}
