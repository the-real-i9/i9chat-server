package tests

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const userWSPath = WSHOST_URL + "/api/app/user/go_online"

func TestDMChat(t *testing.T) {
	t.Parallel()

	user1 := UserT{
		Email:    "louislitt@gmail.com",
		Username: "louislitt",
		Password: "who's norma",
		Phone:    "08145423518",
		Geolocation: UserGeolocation{
			X: 5.0,
			Y: 3.0,
		},
	}

	user2 := UserT{
		Email:    "jeffmalone@gmail.com",
		Username: "jeffyboy",
		Password: "jessica_",
		Phone:    "08113425589",
		Geolocation: UserGeolocation{
			X: 4.0,
			Y: 3.0,
		},
	}

	{
		t.Log("Setup: create new accounts for users")

		for _, user := range []*UserT{&user1, &user2} {
			user := user

			{
				reqBody, err := makeReqBody(map[string]any{"email": user.Email})
				require.NoError(t, err)

				res, err := http.Post(signupPath+"/request_new_account", "application/json", reqBody)
				require.NoError(t, err)

				if !assert.Equal(t, http.StatusOK, res.StatusCode) {
					rb, err := errResBody(res.Body)
					require.NoError(t, err)
					t.Log("unexpected error:", rb)
					return
				}

				rb, err := succResBody[map[string]any](res.Body)
				require.NoError(t, err)

				td.Cmp(td.Require(t), rb, td.SuperMapOf(map[string]any{
					"msg": "A 6-digit verification code has been sent to " + user.Email,
				}, nil))

				user.SessionCookie = res.Header.Get("Set-Cookie")
			}

			{
				reqBody, err := makeReqBody(map[string]any{"code": os.Getenv("DUMMY_VERF_TOKEN")})
				require.NoError(t, err)

				req, err := http.NewRequest("POST", signupPath+"/verify_email", reqBody)
				require.NoError(t, err)
				req.Header.Set("Cookie", user.SessionCookie)
				req.Header.Add("Content-Type", "application/json")

				res, err := http.DefaultClient.Do(req)
				require.NoError(t, err)

				if !assert.Equal(t, http.StatusOK, res.StatusCode) {
					rb, err := errResBody(res.Body)
					require.NoError(t, err)
					t.Log("unexpected error:", rb)
					return
				}

				rb, err := succResBody[map[string]any](res.Body)
				require.NoError(t, err)

				td.Cmp(td.Require(t), rb, td.SuperMapOf(map[string]any{
					"msg": fmt.Sprintf("Your email '%s' has been verified!", user.Email),
				}, nil))

				user.SessionCookie = res.Header.Get("Set-Cookie")
			}

			{
				reqBody, err := makeReqBody(map[string]any{
					"username": user.Username,
					"password": user.Password,
					"phone":    user.Phone,
					"geolocation": map[string]any{
						"x": user.Geolocation.X,
						"y": user.Geolocation.Y,
					},
				})
				require.NoError(t, err)

				req, err := http.NewRequest("POST", signupPath+"/register_user", reqBody)
				require.NoError(t, err)
				req.Header.Add("Content-Type", "application/json")
				req.Header.Set("Cookie", user.SessionCookie)

				res, err := http.DefaultClient.Do(req)
				require.NoError(t, err)

				if !assert.Equal(t, http.StatusOK, res.StatusCode) {
					rb, err := errResBody(res.Body)
					require.NoError(t, err)
					t.Log("unexpected error:", rb)
					return
				}

				rb, err := succResBody[map[string]any](res.Body)
				require.NoError(t, err)

				td.Cmp(td.Require(t), rb, td.SuperMapOf(map[string]any{
					"msg":  "Signup success!",
					"user": td.Ignore(),
				}, nil))

				user.SessionCookie = res.Header.Get("Set-Cookie")
			}
		}
	}

	{
		t.Log("Setup: Init user sockets")

		for _, user := range []*UserT{&user1, &user2} {
			user := user

			header := http.Header{}
			header.Set("Cookie", user.SessionCookie)
			wsConn, res, err := websocket.DefaultDialer.Dial(userWSPath, header)
			require.NoError(t, err)

			if !assert.Equal(t, http.StatusSwitchingProtocols, res.StatusCode) {
				rb, err := errResBody(res.Body)
				require.NoError(t, err)
				t.Log("unexpected error:", rb)
				return
			}

			require.NotNil(t, wsConn)

			defer wsConn.CloseHandler()(websocket.CloseNormalClosure, user.Username+": GoodBye!")

			user.WSConn = wsConn
			user.ServerWSMsg = make(chan map[string]any)

			go func() {
				userCommChan := user.ServerWSMsg

				for {
					userCommChan := userCommChan
					userWSConn := user.WSConn

					var wsMsg map[string]any

					if err := userWSConn.ReadJSON(&wsMsg); err != nil {
						break
					}

					if wsMsg == nil {
						continue
					}

					userCommChan <- wsMsg
				}

				close(userCommChan)
			}()
		}
	}

	{
		t.Log("Action: user1 sends message to user2")

		err := user1.WSConn.WriteJSON(map[string]any{
			"event": "new dm chat message",
			"data": map[string]any{
				"partnerUsername": user2.Username,
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

		// user1's server reply (response) to event sent
		user1ServerReply := <-user1.ServerWSMsg

		td.Cmp(td.Require(t), user1ServerReply, td.SuperMapOf(map[string]any{
			"event":   "server reply",
			"toEvent": "new dm chat message",
			"data": td.Map(map[string]any{
				"new_msg_id": td.Ignore(),
			}, nil),
		}, nil))
	}

	user2RecvdMsgId := ""

	{
		t.Log("Action: user2 receives the message | acknowledges 'delivered'")

		user2NewMsgReceived := <-user2.ServerWSMsg

		td.Cmp(td.Require(t), user2NewMsgReceived, td.SuperMapOf(map[string]any{
			"event": "new dm chat message",
			"data": td.SuperMapOf(map[string]any{
				"id":              td.Ignore(),
				"content":         td.Ignore(),
				"delivery_status": "sent",
				"sender": td.SuperMapOf(map[string]any{
					"username": user1.Username,
				}, nil),
			}, nil),
		}, nil))

		user2RecvdMsgId = user2NewMsgReceived["data"].(map[string]any)["id"].(string)

		err := user2.WSConn.WriteJSON(map[string]any{
			"event": "ack dm chat message delivered",
			"data": map[string]any{
				"partnerUsername": user1.Username,
				"msgId":           user2RecvdMsgId,
				"at":              time.Now().UTC(),
			},
		})
		require.NoError(t, err)
	}

	{
		t.Log("user1 receives the 'delivered' acknowledgement | marks message as 'delivered'")

		user1DelvAckReceipt := <-user1.ServerWSMsg

		td.Cmp(td.Require(t), user1DelvAckReceipt, td.SuperMapOf(map[string]any{
			"event": "dm chat message delivered",
			"data": td.Map(map[string]any{
				"partner_username": user2.Username,
				"msg_id":           user2RecvdMsgId,
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user2 then acknowledges 'read'")

		err := user2.WSConn.WriteJSON(map[string]any{
			"event": "ack dm chat message read",
			"data": map[string]any{
				"partnerUsername": user1.Username,
				"msgId":           user2RecvdMsgId,
				"at":              time.Now().UTC(),
			},
		})
		require.NoError(t, err)
	}

	{
		t.Log("user1 receives the 'read' acknowledgement | marks message as 'read'")

		user1ReadAckReceipt := <-user1.ServerWSMsg

		td.Cmp(td.Require(t), user1ReadAckReceipt, td.SuperMapOf(map[string]any{
			"event": "dm chat message read",
			"data": td.Map(map[string]any{
				"partner_username": user2.Username,
				"msg_id":           user2RecvdMsgId,
			}, nil),
		}, nil))
	}
}
