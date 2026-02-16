package realtimeService

import (
	"context"
	"i9chat/src/appTypes"
	"i9chat/src/helpers"

	"github.com/gofiber/contrib/websocket"
)

func SendEventMsg(toUser string, msg appTypes.ServerEventMsg) {
	if userPipe, ok := AllClientSockets.Load(toUser); ok {
		pipe := userPipe.(*websocket.Conn)

		if err := pipe.WriteMessage(websocket.BinaryMessage, helpers.ToBtMsgPack(msg)); err != nil {
			helpers.LogError(err)
		}

		return
	}
}

func AddPipe(ctx context.Context, clientUsername string, pipe *websocket.Conn) {
	AllClientSockets.Store(clientUsername, pipe)
}

func RemovePipe(clientUsername string) {
	AllClientSockets.Delete(clientUsername)
}
