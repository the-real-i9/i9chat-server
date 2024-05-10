package authRoutes_test

import (
	"encoding/json"
	"fmt"
	"i9chat/tests/testdata"
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

	sendData := map[string]int{
		"code": 910272,
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

	sendData := map[string]string{
		"username":    "i9x",
		"password":    "fhunmytor",
		"geolocation": "5, 2, 2",
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

	sendData := map[string]string{
		"emailOrUsername": "i9x",
		"password":        "fhunmytor",
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

	if recvData.Error != "" {
		t.Error(recvData.Error)
		return
	}

	printJSON(recvData.Body)
}
