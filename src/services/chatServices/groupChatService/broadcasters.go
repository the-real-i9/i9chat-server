package groupChatService

import (
	"i9chat/src/appTypes"
	"i9chat/src/services/realtimeService"
)

func broadcastNewGroup(targetUsers []any, data any) {
	for _, tu := range targetUsers {

		realtimeService.SendEventMsg(tu.(string), appTypes.ServerEventMsg{
			Event: "new group chat",
			Data:  data,
		})
	}
}

func broadcastNewMessage(memberUsernames []string, data any, groupId string) {
	for _, mu := range memberUsernames {

		realtimeService.SendEventMsg(mu, appTypes.ServerEventMsg{
			Event: "new group chat message",
			Data: map[string]any{
				"message":  data,
				"group_id": groupId,
			},
		})
	}
}

func broadcastActivity(memberUsernames []any, data any, groupId string) {
	for _, mu := range memberUsernames {

		realtimeService.SendEventMsg(mu.(string), appTypes.ServerEventMsg{
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
		realtimeService.SendEventMsg(mu, appTypes.ServerEventMsg{
			Event: "group chat message delivered",
			Data:  data,
		})
	}
}

func broadcastMsgRead(memberUsernames []string, data any) {
	for _, mu := range memberUsernames {
		realtimeService.SendEventMsg(mu, appTypes.ServerEventMsg{
			Event: "group chat message read",
			Data:  data,
		})
	}
}

func broadcastMsgReaction(memberUsernames []string, data any) {
	for _, mu := range memberUsernames {
		realtimeService.SendEventMsg(mu, appTypes.ServerEventMsg{
			Event: "group chat message reaction",
			Data:  data,
		})
	}
}

func broadcastMsgReactionRemoved(memberUsernames []any, data any) {
	for _, mu := range memberUsernames {
		mu := mu.(string)
		realtimeService.SendEventMsg(mu, appTypes.ServerEventMsg{
			Event: "group chat message reaction removed",
			Data:  data,
		})
	}
}
