package testhelpers

import (
	"encoding/json"
	"fmt"
	"i9chat/utils/appTypes"

	"github.com/fasthttp/websocket"
)

func WSSendRecv(connStream *websocket.Conn, sendData map[string]any, recvData *appTypes.WSResp) (wr_err error) {
	w_err := connStream.WriteJSON(sendData)
	if w_err != nil {
		return w_err
	}

	r_err := connStream.ReadJSON(recvData)
	if r_err != nil {
		return r_err
	}

	return nil
}

func PrintJSON(data any) {
	res, _ := json.MarshalIndent(data, "", "  ")

	fmt.Println(string(res))
}
