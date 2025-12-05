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

func TestDirectChat(t *testing.T) {
	// t.Parallel()

	user1 := UserT{
		Email:    "louislitt@gmail.com",
		Username: "louislitt",
		Password: "who's norma",
		Geolocation: UserGeolocation{
			X: 5.0,
			Y: 3.0,
		},
	}

	user2 := UserT{
		Email:    "jeffmalone@gmail.com",
		Username: "jeffyboy",
		Password: "jessica_",
		Geolocation: UserGeolocation{
			X: 4.0,
			Y: 3.0,
		},
	}

	{
		t.Log("Setup: create new accounts for users")

		for _, user := range []*UserT{&user1, &user2} {

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
				reqBody, err := makeReqBody(map[string]any{"code": os.Getenv("DUMMY_TOKEN")})
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
			wsConn, res, err := websocket.DefaultDialer.Dial(wsPath, header)
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
			user.ServerEventMsg = make(chan map[string]any)

			go func() {
				userCommChan := user.ServerEventMsg

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

	user1NewMsgId := ""

	{
		t.Log("Action: user1 sends message to user2")

		err := user1.WSConn.WriteJSON(map[string]any{
			"action": "direct chat: send message",
			"data": map[string]any{
				"partnerUsername": user2.Username,
				"msg": map[string]any{
					"type": "text",
					"props": map[string]any{
						"text_content": "Hi. How're you doing?",
					},
				},
				"at": time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(t, err)

		// user1's server reply (response) to action
		user1ServerReply := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1ServerReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "direct chat: send message",
			"data": td.Map(map[string]any{
				"new_msg_id": td.Ignore(),
			}, nil),
		}, nil))

		user1NewMsgId = user1ServerReply["data"].(map[string]any)["new_msg_id"].(string)
	}

	{
		t.Log("Action: user2 receives the message | acknowledges 'delivered'")

		user2NewMsgReceived := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2NewMsgReceived, td.Map(map[string]any{
			"event": "direct chat: new message",
			"data": td.SuperMapOf(map[string]any{
				"id": user1NewMsgId,
				"content": td.SuperMapOf(map[string]any{
					"type": "text",
					"props": td.SuperMapOf(map[string]any{
						"text_content": "Hi. How're you doing?",
					}, nil),
				}, nil),
				"delivery_status": "sent",
				"sender": td.SuperMapOf(map[string]any{
					"username": user1.Username,
				}, nil),
			}, nil),
		}, nil))

		err := user2.WSConn.WriteJSON(map[string]any{
			"action": "direct chat: ack message delivered",
			"data": map[string]any{
				"partnerUsername": user1.Username,
				"msgId":           user1NewMsgId,
				"at":              time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(t, err)

		user2ServerReply := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2ServerReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "direct chat: ack message delivered",
			"data":     true,
		}, nil))
	}

	{
		t.Log("Action: user1 receives the 'delivered' acknowledgement | marks message as 'delivered'")

		user1DelvAckReceipt := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1DelvAckReceipt, td.Map(map[string]any{
			"event": "direct chat: message delivered",
			"data": td.Map(map[string]any{
				"partner_username": user2.Username,
				"msg_id":           user1NewMsgId,
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user2 then acknowledges 'read'")

		err := user2.WSConn.WriteJSON(map[string]any{
			"action": "direct chat: ack message read",
			"data": map[string]any{
				"partnerUsername": user1.Username,
				"msgId":           user1NewMsgId,
				"at":              time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(t, err)

		user2ServerReply := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2ServerReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "direct chat: ack message read",
			"data":     true,
		}, nil))
	}

	{
		t.Log("Action: user1 receives the 'read' acknowledgement | marks message as 'read'")

		user1ReadAckReceipt := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1ReadAckReceipt, td.Map(map[string]any{
			"event": "direct chat: message read",
			"data": td.Map(map[string]any{
				"partner_username": user2.Username,
				"msg_id":           user1NewMsgId,
			}, nil),
		}, nil))
	}

	user2NewMsgId := ""

	{
		t.Log("Action: user2 sends message to user1")

		photo, err := os.ReadFile("./test_files/profile_pic.png")
		require.NoError(t, err)

		err = user2.WSConn.WriteJSON(map[string]any{
			"action": "direct chat: send message",
			"data": map[string]any{
				"partnerUsername": user1.Username,
				"msg": map[string]any{
					"type": "photo",
					"props": map[string]any{
						"data":    photo,
						"caption": "I'm guuud!",
					},
				},
				"at": time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(t, err)

		// user2's server reply (response) to action
		user2ServerReply := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2ServerReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "direct chat: send message",
			"data": td.Map(map[string]any{
				"new_msg_id": td.Ignore(),
			}, nil),
		}, nil))

		user2NewMsgId = user2ServerReply["data"].(map[string]any)["new_msg_id"].(string)
	}

	{
		t.Log("Action: user1 receives the message | acknowledges 'delivered'")

		user1NewMsgReceived := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1NewMsgReceived, td.Map(map[string]any{
			"event": "direct chat: new message",
			"data": td.SuperMapOf(map[string]any{
				"id": user2NewMsgId,
				"content": td.SuperMapOf(map[string]any{
					"type": "photo",
					"props": td.SuperMapOf(map[string]any{
						"caption": "I'm guuud!",
					}, nil),
				}, nil),
				"delivery_status": "sent",
				"sender": td.SuperMapOf(map[string]any{
					"username": user2.Username,
				}, nil),
			}, nil),
		}, nil))

		err := user1.WSConn.WriteJSON(map[string]any{
			"action": "direct chat: ack message delivered",
			"data": map[string]any{
				"partnerUsername": user2.Username,
				"msgId":           user2NewMsgId,
				"at":              time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(t, err)

		user1ServerReply := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1ServerReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "direct chat: ack message delivered",
			"data":     true,
		}, nil))
	}

	{
		t.Log("Action: user2 receives the 'delivered' acknowledgement | marks message as 'delivered'")

		user2DelvAckReceipt := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2DelvAckReceipt, td.Map(map[string]any{
			"event": "direct chat: message delivered",
			"data": td.Map(map[string]any{
				"partner_username": user1.Username,
				"msg_id":           user2NewMsgId,
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user1 then acknowledges 'read'")

		err := user1.WSConn.WriteJSON(map[string]any{
			"action": "direct chat: ack message read",
			"data": map[string]any{
				"partnerUsername": user2.Username,
				"msgId":           user2NewMsgId,
				"at":              time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(t, err)

		user1ServerReply := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1ServerReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "direct chat: ack message read",
			"data":     true,
		}, nil))
	}

	{
		t.Log("Action: user2 receives the 'read' acknowledgement | marks message as 'read'")

		user2ReadAckReceipt := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2ReadAckReceipt, td.Map(map[string]any{
			"event": "direct chat: message read",
			"data": td.Map(map[string]any{
				"partner_username": user1.Username,
				"msg_id":           user2NewMsgId,
			}, nil),
		}, nil))
	}

	{
		<-(time.NewTimer(100 * time.Millisecond).C)

		t.Log("Action: user1 opens his chat history with user2")

		req, err := http.NewRequest("GET", directChatPath+"/"+user2.Username+"/history", nil)
		require.NoError(t, err)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Set("Cookie", user1.SessionCookie)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[[]map[string]any](res.Body)
		require.NoError(t, err)

		td.Cmp(td.Require(t), rb,
			td.All(
				td.Contains(td.SuperMapOf(map[string]any{
					"id": user1NewMsgId,
					"content": td.SuperMapOf(map[string]any{
						"type": "text",
						"props": td.SuperMapOf(map[string]any{
							"text_content": "Hi. How're you doing?",
						}, nil),
					}, nil),
					"delivery_status": "read",
					"sender": td.SuperMapOf(map[string]any{
						"username": user1.Username,
					}, nil),
				}, nil)),
				td.Contains(td.SuperMapOf(map[string]any{
					"id": user2NewMsgId,
					"content": td.SuperMapOf(map[string]any{
						"type": "photo",
						"props": td.SuperMapOf(map[string]any{
							"caption": "I'm guuud!",
						}, nil),
					}, nil),
					"delivery_status": "read",
					"sender": td.SuperMapOf(map[string]any{
						"username": user2.Username,
					}, nil),
				}, nil)),
			),
		)
	}
}
