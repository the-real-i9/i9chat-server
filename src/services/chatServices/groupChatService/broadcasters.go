package groupChatService

import (
	"fmt"
	"i9chat/src/appTypes"
	"i9chat/src/services/eventStreamService"
)

func broadcastNewGroup(targetUsers []string, data any) {
	for _, tu := range targetUsers {

		eventStreamService.Send(fmt.Sprintf("user-%s-alerts", tu), appTypes.ServerWSMsg{
			Event: "new group chat",
			Data:  data,
		})
	}
}

func broadcastNewMessage(memberUsernames []string, data any, groupId string) {
	for _, mu := range memberUsernames {

		eventStreamService.Send(mu, appTypes.ServerWSMsg{
			Event: "new group chat message",
			Data: map[string]any{
				"message":  data,
				"group_id": groupId,
			},
		})
	}
}

func broadcastActivity(memberUsernames []string, data any, groupId string) {

	for _, mu := range memberUsernames {
		eventStreamService.Send(mu, appTypes.ServerWSMsg{
			Event: "new group chat activity",
			Data: map[string]any{
				"info":     data,
				"group_id": groupId,
			},
		})
	}
}

func broadcastMsgDelivered(memberUsernames []string, data any) {
	for _, mu := range memberUsernames {
		eventStreamService.Send(mu, appTypes.ServerWSMsg{
			Event: "group chat message delivered",
			Data:  data,
		})
	}
}

func broadcastMsgRead(memberUsernames []string, data any) {
	for _, mu := range memberUsernames {
		eventStreamService.Send(mu, appTypes.ServerWSMsg{
			Event: "group chat message read",
			Data:  data,
		})
	}
}
