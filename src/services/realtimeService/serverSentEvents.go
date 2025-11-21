package realtimeService

import (
	"context"
	"fmt"
	"i9chat/src/appTypes"
	"i9chat/src/helpers"

	"github.com/gofiber/contrib/websocket"
	"github.com/redis/go-redis/v9"
)

func SendEventMsg(toUser string, msg appTypes.ServerEventMsg) {
	ctx := context.Background()

	storeUndelvMsg := func() {
		// save for later in case of error
		_, err := rdb().RPush(ctx, fmt.Sprintf("user_event_msgs_queue:%s", toUser), helpers.ToJson(msg)).Result()
		if err != nil {
			helpers.LogError(err)
			return
		}
	}

	if userPipe, ok := AllClientSockets.Load(toUser); ok {
		pipe := userPipe.(*websocket.Conn)

		if err := pipe.WriteJSON(msg); err != nil {
			helpers.LogError(err)
			storeUndelvMsg()
		}

		return
	}

	storeUndelvMsg()
}

func AddPipe(ctx context.Context, clientUsername string, pipe *websocket.Conn) {
	for {
		msgJson, err := rdb().LPop(ctx, fmt.Sprintf("user_event_msgs_queue:%s", clientUsername)).Result()
		if err != nil {
			if err != redis.Nil {
				helpers.LogError(err)
			}
			break
		}

		if err := pipe.WriteMessage(websocket.TextMessage, []byte(msgJson)); err != nil {
			helpers.LogError(err)
		}
	}

	AllClientSockets.Store(clientUsername, pipe)
}

func RemovePipe(clientUsername string) {
	AllClientSockets.Delete(clientUsername)
}
