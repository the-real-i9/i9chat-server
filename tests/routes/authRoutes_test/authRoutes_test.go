package authRoutes_test

import (
	"encoding/json"
	"fmt"
	"i9chat/tests/testdata"
	"i9chat/tests/testhelpers"
	"i9chat/utils/appTypes"
	"net/http"
	"testing"

	"github.com/fasthttp/websocket"
)

const pathPrefix string = "ws://localhost:8000/api/auth"

func printJSON(data map[string]any) {
	res, _ := json.MarshalIndent(data, "", "  ")

	fmt.Println(string(res))
}

func XTestRequestNewAccount(t *testing.T) {
	wsd := &websocket.Dialer{}

	connStream, _, err := wsd.Dial(pathPrefix+"/signup/request_new_account", nil)
	if err != nil {
		t.Error(err)
		return
	}

	defer connStream.Close()

	sendData := map[string]any{
		"email": "ogunrinola.kehinde@yahoo.com",
	}

	var recvData appTypes.WSResp

	if wr_err := testhelpers.WSSendRecv(connStream, sendData, &recvData); wr_err != nil {
		t.Error(wr_err)
		return
	}

	if recvData.Error != "" {
		t.Error(recvData.Error)
		return
	}

	printJSON(recvData.Body)
}

func XTestVerifyEmail(t *testing.T) {
	wsd := &websocket.Dialer{}
	reqH := http.Header{}
	reqH.Set("Authorization", testdata.SignupSessionJwt)

	connStream, _, err := wsd.Dial(pathPrefix+"/signup/verify_email", reqH)
	if err != nil {
		t.Error(err)
		return
	}

	defer connStream.Close()

	sendData := map[string]any{
		"code": 133470,
	}

	var recvData appTypes.WSResp

	if wr_err := testhelpers.WSSendRecv(connStream, sendData, &recvData); wr_err != nil {
		t.Error(wr_err)
		return
	}

	if recvData.Error != "" {
		t.Error(recvData.Error)
		return
	}

	printJSON(recvData.Body)
}

func XTestRegisterUser(t *testing.T) {
	wsd := &websocket.Dialer{}
	reqH := http.Header{}
	reqH.Set("Authorization", testdata.SignupSessionJwt)

	connStream, _, err := wsd.Dial(pathPrefix+"/signup/register_user", reqH)
	if err != nil {
		t.Error(err)
		return
	}

	defer connStream.Close()

	sendData := map[string]any{
		"username":    "dollyp",
		"password":    "fhunmytor",
		"geolocation": "9, 5, 2",
	}

	w_err := connStream.WriteJSON(sendData)
	if w_err != nil {
		t.Error(w_err)
		return
	}

	var recvData appTypes.WSResp

	if wr_err := testhelpers.WSSendRecv(connStream, sendData, &recvData); wr_err != nil {
		t.Error(wr_err)
		return
	}

	if recvData.Error != "" {
		t.Error(recvData.Error)
		return
	}

	printJSON(recvData.Body)
}

func TestSignin(t *testing.T) {
	wsd := &websocket.Dialer{}

	connStream, _, err := wsd.Dial(pathPrefix+"/signin", nil)
	if err != nil {
		t.Error(err)
		return
	}

	defer connStream.Close()

	sendData := map[string]any{
		"emailOrUsername": "dollyp",
		"password":        "fhunmytor",
	}

	var recvData appTypes.WSResp

	if wr_err := testhelpers.WSSendRecv(connStream, sendData, &recvData); wr_err != nil {
		t.Error(wr_err)
		return
	}

	if recvData.Error != "" {
		t.Error(recvData.Error)
		return
	}

	printJSON(recvData.Body)
}
