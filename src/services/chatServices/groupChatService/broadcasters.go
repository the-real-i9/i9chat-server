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

func broadcastNewMessage(memberUsernames []any, data any, groupId string) {
	for _, mu := range memberUsernames {

		realtimeService.SendEventMsg(mu.(string), appTypes.ServerEventMsg{
			Event: "group chat: new message",
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
			Event: "group chat: new activity",
			Data: map[string]any{
				"info":     data,
				"group_id": groupId,
			},
		})
	}
}

func broadcastMsgReaction(memberUsernames []any, data any) {
	for _, mu := range memberUsernames {
		realtimeService.SendEventMsg(mu.(string), appTypes.ServerEventMsg{
			Event: "group chat: message reaction",
			Data:  data,
		})
	}
}

func broadcastMsgReactionRemoved(memberUsernames []any, data any) {
	for _, mu := range memberUsernames {
		realtimeService.SendEventMsg(mu.(string), appTypes.ServerEventMsg{
			Event: "group chat: message reaction removed",
			Data:  data,
		})
	}
}
