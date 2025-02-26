package tests

import (
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/stretchr/testify/require"
)

const userWSPath = WSHOST_URL + "/api/app/user/go_online"

func Test(t *testing.T) {
	t.Parallel()

	accounts := map[string]map[string]any{
		"user1": {
			"email":    "louislitt@gmail.com",
			"username": "louislitt",
			"password": "who's norma",
			"phone":    "08145423518",
			"geolocation": map[string]any{
				"x": 5.0,
				"y": 3.0,
			},
		},
		"user2": {
			"email":    "jeffmalone@gmail.com",
			"username": "jeffyboy",
			"password": "jessica_",
			"phone":    "08113425589",
			"geolocation": map[string]any{
				"x": 4.0,
				"y": 3.0,
			},
		},
	}

	t.Run("signup 2 users", func(t *testing.T) {

		for user, info := range accounts {
			t.Run("step one: request new account", func(t *testing.T) {
				reqBody, err := reqBody(map[string]any{"email": info["email"]})
				require.NoError(t, err)

				res, err := http.Post(signupPath+"/request_new_account", "application/json", reqBody)
				require.NoError(t, err)

				require.Equal(t, http.StatusOK, res.StatusCode)

				bd, err := resBody(res.Body)
				require.NoError(t, err)
				require.NotEmpty(t, bd)

				accounts[user]["session_cookie"] = res.Header.Get("Set-Cookie")
			})

			t.Run("step two: verify email", func(t *testing.T) {
				verfCode, err := strconv.Atoi(os.Getenv("DUMMY_VERF_TOKEN"))
				require.NoError(t, err)

				reqBody, err := reqBody(map[string]any{"code": verfCode})
				require.NoError(t, err)

				req, err := http.NewRequest("POST", signupPath+"/verify_email", reqBody)
				require.NoError(t, err)
				req.Header.Set("Cookie", accounts[user]["session_cookie"].(string))
				req.Header.Add("Content-Type", "application/json")

				res, err := http.DefaultClient.Do(req)
				require.NoError(t, err)

				require.Equal(t, http.StatusOK, res.StatusCode)

				bd, err := resBody(res.Body)
				require.NoError(t, err)
				require.NotEmpty(t, bd)

				accounts[user]["session_cookie"] = res.Header.Get("Set-Cookie")
			})

			t.Run("step three: register user", func(t *testing.T) {
				reqBody, err := reqBody(map[string]any{
					"username":    info["username"],
					"password":    info["password"],
					"phone":       info["phone"],
					"geolocation": info["geolocation"],
				})
				require.NoError(t, err)

				req, err := http.NewRequest("POST", signupPath+"/register_user", reqBody)
				require.NoError(t, err)
				req.Header.Add("Content-Type", "application/json")
				req.Header.Set("Cookie", accounts[user]["session_cookie"].(string))

				res, err := http.DefaultClient.Do(req)
				require.NoError(t, err)

				require.Equal(t, http.StatusOK, res.StatusCode)

				bd, err := resBody(res.Body)
				require.NoError(t, err)
				require.NotEmpty(t, bd)

				accounts[user]["session_cookie"] = res.Header.Get("Set-Cookie")
			})
		}
	})

	var (
		user1wsConn *websocket.Conn
		user1res    *http.Response
		user1err    error

		user2wsConn *websocket.Conn
		user2res    *http.Response
		user2err    error
	)

	t.Run("bring user1 online", func(t *testing.T) {
		header := http.Header{}
		header.Set("Cookie", accounts["user1"]["session_cookie"].(string))
		user1wsConn, user1res, user1err = websocket.DefaultDialer.Dial(userWSPath, header)

		require.NoError(t, user1err)
		require.Equal(t, http.StatusSwitchingProtocols, user1res.StatusCode)
	})

	t.Run("bring user2 online", func(t *testing.T) {
		header := http.Header{}
		header.Set("Cookie", accounts["user2"]["session_cookie"].(string))
		user2wsConn, user2res, user2err = websocket.DefaultDialer.Dial(userWSPath, header)

		require.NoError(t, user2err)
		require.Equal(t, http.StatusSwitchingProtocols, user2res.StatusCode)
	})

	/* t.Run("confirm user1 is online", func(t *testing.T) {
		req, err := http.NewRequest("GET", userPath+"/my_profile", nil)
		require.NoError(t, err)
		req.Header.Set("Cookie", accounts["user1"]["session_cookie"].(string))

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, res.StatusCode)

		bd, err := resBody(res.Body)
		require.NoError(t, err)
		require.NotEmpty(t, bd)

		var data map[string]any

		require.NoError(t, json.Unmarshal(bd, &data))

		maps.Copy(accounts["user1"], data)

		require.Equal(t, accounts["user1"]["presence"], "online")
	})

	t.Run("confirm user2 is online", func(t *testing.T) {
		req, err := http.NewRequest("GET", userPath+"/my_profile", nil)
		require.NoError(t, err)
		req.Header.Set("Cookie", accounts["user2"]["session_cookie"].(string))

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, res.StatusCode)

		bd, err := resBody(res.Body)
		require.NoError(t, err)
		require.NotEmpty(t, bd)

		var data map[string]any

		require.NoError(t, json.Unmarshal(bd, &data))

		maps.Copy(accounts["user2"], data)

		require.Equal(t, accounts["user2"]["presence"], "online")
	}) */

	t.Run("user1 messages user2, who then acks 'delivered'", func(t *testing.T) {
		var (
			ndcmEvent  = "new dm chat message"
			adcmdEvent = "ack dm chat message delivered"
			// adcmrEvent = "ack dm chat message read"
		)

		var (
			user1Username = accounts["user1"]["username"].(string)
			user2Username = accounts["user2"]["username"].(string)
		)

		var (
			err error
		)

		// user1 sends message to user2
		err = user1wsConn.WriteJSON(map[string]any{
			"event": ndcmEvent,
			"data": map[string]any{
				"partnerUsername": user2Username,
				"msg": map[string]any{
					"type": "text",
					"props": map[string]any{
						"textContent": "Hi. How're you doing?",
					},
				},
				"createdAt": time.Now().UTC(),
			},
		})
		require.NoError(t, err)

		// user1's server reply to event send
		var user1Reply map[string]any

		require.NoError(t, user1wsConn.ReadJSON(&user1Reply))
		require.Contains(t, user1Reply, "event")
		require.Equal(t, user1Reply["event"], "server reply")

		// user2 receives a new message
		var user2NewMsgReceipt map[string]any

		require.NoError(t, user2wsConn.ReadJSON(&user2NewMsgReceipt))
		require.Contains(t, user2NewMsgReceipt, "event")
		require.Equal(t, user2NewMsgReceipt["event"], "new dm chat message")
		require.Contains(t, user2NewMsgReceipt["data"], "id")
		require.Contains(t, user2NewMsgReceipt["data"], "content")

		recvdMsgId := user2NewMsgReceipt["data"].(map[string]any)["id"].(string)

		// user2 acknowledges message as 'delivered'
		err = user2wsConn.WriteJSON(map[string]any{
			"event": adcmdEvent,
			"data": map[string]any{
				"partnerUsername": user1Username,
				"msgId":           recvdMsgId,
				"at":              time.Now().UTC(),
			},
		})
		require.NoError(t, err)

		// user1 receives the 'delivered' acknowledgement
		var user1DelvReceipt map[string]any

		require.NoError(t, user1wsConn.ReadJSON(&user1DelvReceipt))

		require.Contains(t, user1DelvReceipt, "event")
		require.Equal(t, user1DelvReceipt["event"], "dm chat message delivered")
		require.Equal(t, user1DelvReceipt["data"], map[string]any{"partner_username": user2Username, "msg_id": recvdMsgId})
	})

	require.NoError(t, user1wsConn.CloseHandler()(websocket.CloseNormalClosure, "user1 done"))
	require.NoError(t, user2wsConn.CloseHandler()(websocket.CloseNormalClosure, "user2 done"))
}
