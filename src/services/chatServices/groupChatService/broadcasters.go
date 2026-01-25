package groupChatService

import (
	"context"
	"fmt"
	"i9chat/src/appGlobals"
	"i9chat/src/appTypes"
	"i9chat/src/helpers"
	"i9chat/src/services/realtimeService"

	"github.com/redis/go-redis/v9"
)

func broadcastNewGroup(targetUsers []any, data any) {
	for _, tu := range targetUsers {

		realtimeService.SendEventMsg(tu.(string), appTypes.ServerEventMsg{
			Event: "new group chat",
			Data:  data,
		})
	}
}

func broadcastNewMessage(groupId string, data any, clientUsername string) {
	ctx := context.Background()

	var cursor uint64 = 0

	for {
		musers, nextCursor, err := appGlobals.RedisClient.SScan(ctx, fmt.Sprintf("group:%s:members", groupId), cursor, "*", 100).Result()
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return
		}

		for _, mu := range musers {
			if mu == clientUsername {
				continue
			}

			go realtimeService.SendEventMsg(mu, appTypes.ServerEventMsg{
				Event: "group chat: new message",
				Data: map[string]any{
					"message":  data,
					"group_id": groupId,
				},
			})
		}

		if nextCursor == 0 {
			break
		}

		cursor = nextCursor
	}
}

func broadcastActivityToOne(mu string, info any, groupId string) {
	go realtimeService.SendEventMsg(mu, appTypes.ServerEventMsg{
		Event: "group chat: new activity",
		Data: map[string]any{
			"info":     info,
			"group_id": groupId,
		},
	})
}

func broadcastActivityToAll(groupId string, info string, except []any) {
	ctx := context.Background()

	exceptUsers := make(map[string]bool, len(except))
	for _, u := range except {
		exceptUsers[u.(string)] = true
	}

	var cursor uint64 = 0

	for {
		musers, nextCursor, err := appGlobals.RedisClient.SScan(ctx, fmt.Sprintf("group:%s:members", groupId), cursor, "*", 100).Result()
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return
		}

		for _, mu := range musers {
			if exceptUsers[mu] {
				continue
			}

			go realtimeService.SendEventMsg(mu, appTypes.ServerEventMsg{
				Event: "group chat: new activity",
				Data: map[string]any{
					"info":     info,
					"group_id": groupId,
				},
			})
		}

		if nextCursor == 0 {
			break
		}

		cursor = nextCursor
	}
}

func broadcastMsgReaction(groupId, clientUsername string, data any) {
	ctx := context.Background()

	var cursor uint64 = 0

	for {
		musers, nextCursor, err := appGlobals.RedisClient.SScan(ctx, fmt.Sprintf("group:%s:members", groupId), cursor, "*", 100).Result()
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return
		}

		for _, mu := range musers {
			if mu == clientUsername {
				continue
			}

			go realtimeService.SendEventMsg(mu, appTypes.ServerEventMsg{
				Event: "group chat: message reaction",
				Data:  data,
			})
		}

		if nextCursor == 0 {
			break
		}

		cursor = nextCursor
	}
}

func broadcastMsgReactionRemoved(groupId, clientUsername string, data any) {
	ctx := context.Background()

	var cursor uint64 = 0

	for {
		musers, nextCursor, err := appGlobals.RedisClient.SScan(ctx, fmt.Sprintf("group:%s:members", groupId), cursor, "*", 100).Result()
		if err != nil && err != redis.Nil {
			helpers.LogError(err)
			return
		}

		for _, mu := range musers {
			if mu == clientUsername {
				continue
			}

			go realtimeService.SendEventMsg(mu, appTypes.ServerEventMsg{
				Event: "group chat: message reaction removed",
				Data:  data,
			})
		}

		if nextCursor == 0 {
			break
		}

		cursor = nextCursor
	}
}
