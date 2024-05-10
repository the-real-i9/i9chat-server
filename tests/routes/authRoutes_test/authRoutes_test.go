package authRoutes_test

import (
	"fmt"
	"i9chat/utils/appTypes"
	"net/http"
	"testing"

	"github.com/fasthttp/websocket"
)

const pathPrefix string = "ws://localhost:8000/api/auth"

var signupSessesionJwt = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImVtYWlsIjoib2x1d2FyaW5vbGFzYW1AZ21haWwuY29tIiwic2Vzc2lvbklkIjoiMWQ5YzA0MjktN2M0ZS00NDIwLTgzZDEtNDg4Njc5OGFkMzExIn0sImp3dENsYWltcyI6eyJleHAiOiIyMDI0LTA1LTEwVDE1OjAxOjE4LjE1NjE0ODM4N1oiLCJpYXQiOiIyMDI0LTA1LTEwVDE0OjAxOjE4LjE4NjIzOTk4N1oiLCJpc3N1ZXIiOiJpOWNoYXQifX0.+Z1WSQmTlqKV+khtVwTbILho89eYOymxPHjEGXI6BJE="

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

	fmt.Println(recvData.Body)
}

func TestVerifyEmail(t *testing.T) {
	wsd := &websocket.Dialer{}
	reqH := http.Header{}
	reqH.Set("Authorization", signupSessesionJwt)

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

	fmt.Println(recvData.Body)
}
