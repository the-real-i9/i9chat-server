package groupChatService

import (
	"context"
	"fmt"
	"i9chat/src/appGlobals"
	"i9chat/src/appTypes"
	"i9chat/src/appTypes/UITypes"
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

func broadcastNewMessage(groupId string, data UITypes.ChatHistoryEntry, clientUsername string) {
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
				Event: "group chat: new che: message",
				Data: map[string]any{
					"group_id": groupId,
					"che":      data,
				},
			})
		}

		if nextCursor == 0 {
			break
		}

		cursor = nextCursor
	}
}

func broadcastActivityToOne(groupId string, gactCHE UITypes.ChatHistoryEntry, mu string) {
	go realtimeService.SendEventMsg(mu, appTypes.ServerEventMsg{
		Event: "group chat: new che: group activity",
		Data: map[string]any{
			"group_id": groupId,
			"che":      gactCHE,
		},
	})
}

func broadcastActivityToAll(groupId string, gactCHE UITypes.ChatHistoryEntry, except []any) {
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

		go func(musers []string) {
			for _, mu := range musers {
				if exceptUsers[mu] {
					continue
				}

				realtimeService.SendEventMsg(mu, appTypes.ServerEventMsg{
					Event: "group chat: new che: group activity",
					Data: map[string]any{
						"group_id": groupId,
						"che":      gactCHE,
					},
				})
			}
		}(musers)

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
