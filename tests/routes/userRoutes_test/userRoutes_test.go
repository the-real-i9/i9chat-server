package userRoutes_test

import (
	"i9chat/tests/testdata"
	"i9chat/tests/testhelpers"
	"i9chat/utils/appTypes"
	"net/http"
	"testing"

	"github.com/fasthttp/websocket"
)

const pathPrefix string = "ws://localhost:8000/api/app/user"

func XTestGetAllUsers(t *testing.T) {
	wsd := websocket.Dialer{}

	reqH := http.Header{"Authorization": {testdata.I9xAuthJwt}}

	connStream, _, err := wsd.Dial(pathPrefix+"/all_users", reqH)
	if err != nil {
		t.Error(err)
		return
	}

	defer connStream.Close()

	var recvData appTypes.WSResp

	wr_err := testhelpers.WSSendRecv(connStream, nil, &recvData)
	if wr_err != nil {
		t.Error(wr_err)
		return
	}

	if recvData.Error != "" {
		t.Error(recvData.Error)
		return
	}

	testhelpers.PrintJSON(recvData.Body)
}

func TestGetMyChats(t *testing.T) {
	wsd := websocket.Dialer{}

	reqH := http.Header{"Authorization": {testdata.I9xAuthJwt}}

	connStream, _, err := wsd.Dial(pathPrefix+"/my_chats", reqH)
	if err != nil {
		t.Error(err)
		return
	}

	defer connStream.Close()

	var recvData appTypes.WSResp

	wr_err := testhelpers.WSSendRecv(connStream, nil, &recvData)
	if wr_err != nil {
		t.Error(wr_err)
		return
	}

	if recvData.Error != "" {
		t.Error(recvData.Error)
		return
	}

	testhelpers.PrintJSON(recvData.Body)
}
