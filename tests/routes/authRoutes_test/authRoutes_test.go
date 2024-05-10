package authRoutes_test

import (
	"fmt"
	"i9chat/utils/appTypes"
	"testing"

	"github.com/fasthttp/websocket"
)

const pathPrefix string = "ws://localhost:8000/api/auth"

func TestRequestNewAccount(t *testing.T) {
	wsd := &websocket.Dialer{}

	connStream, _, err := wsd.Dial(pathPrefix+"/signup/request_new_account", nil)
	if err != nil {
		t.Error(err)
		return
	}

	defer connStream.Close()

	sendData := map[string]string{
		"email": "oluwarinolasam@gmail.com",
	}

	w_err := connStream.WriteJSON(sendData)
	if w_err != nil {
		t.Error(w_err)
		return
	}

	var recvData appTypes.WSResp

	r_err := connStream.ReadJSON(&recvData)
	if r_err != nil {
		t.Error(r_err)
		return
	}

	if recvData.Error != nil {
		t.Error(recvData.Error)
		return
	}

	fmt.Println(recvData.Body)
}
