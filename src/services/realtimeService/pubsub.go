package realtimeService

import (
	"context"
	"fmt"
	"i9chat/src/appTypes"
	"i9chat/src/helpers"

	"github.com/gofiber/contrib/websocket"
)

func PublishUserPresenceChange(ctx context.Context, targetUsername string, data map[string]any) {
	if err := rdb().Publish(ctx, fmt.Sprintf("user_%s_presence_change", targetUsername), helpers.ToMsgPack(appTypes.ServerEventMsg{
		Event: "user presence changed",
		Data:  data,
	})).Err(); err != nil {
		helpers.LogError(err)
	}
}

func SubscribeToUserPresence(ctx context.Context, clientUsername string, targetUsername string, ctxCancel context.CancelFunc) {
	pubsub := rdb().Subscribe(ctx, fmt.Sprintf("user_%s_presence_change", targetUsername))

	defer func() {
		if err := pubsub.Close(); err != nil {
			helpers.LogError(err)
		}
	}()

	go func(ctxCancel context.CancelFunc) {
		ch := pubsub.Channel()

		for msg := range ch {
			if userPipe, ok := AllClientSockets.Load(clientUsername); ok {
				pipe := userPipe.(*websocket.Conn)

				if err := pipe.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
					helpers.LogError(err)
					ctxCancel()
				}
			}
		}
	}(ctxCancel)
}
