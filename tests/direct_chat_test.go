package tests

import (
	"fmt"
	"i9chat/src/appGlobals"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"
)

func TestDirectChat(t *testing.T) {
	// t.Parallel()
	require := require.New(t)

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
				require.NoError(err)

				req := httptest.NewRequest("POST", signupPath+"/request_new_account", reqBody)
				req.Header.Add("Content-Type", "application/vnd.msgpack")

				res, err := app.Test(req)
				require.NoError(err)

				if !assert.Equal(t, http.StatusOK, res.StatusCode) {
					rb, err := errResBody(res.Body)
					require.NoError(err)
					t.Log("unexpected error:", rb)
					return
				}

				rb, err := succResBody[map[string]any](res.Body)
				require.NoError(err)

				td.Cmp(td.Require(t), rb, td.SuperMapOf(map[string]any{
					"msg": "A 6-digit verification code has been sent to " + user.Email,
				}, nil))

				user.SessionCookie = res.Header.Get("Set-Cookie")
			}

			{
				reqBody, err := makeReqBody(map[string]any{"code": os.Getenv("DUMMY_TOKEN")})
				require.NoError(err)

				req := httptest.NewRequest("POST", signupPath+"/verify_email", reqBody)
				req.Header.Set("Cookie", user.SessionCookie)
				req.Header.Add("Content-Type", "application/vnd.msgpack")

				res, err := app.Test(req)
				require.NoError(err)

				if !assert.Equal(t, http.StatusOK, res.StatusCode) {
					rb, err := errResBody(res.Body)
					require.NoError(err)
					t.Log("unexpected error:", rb)
					return
				}

				rb, err := succResBody[map[string]any](res.Body)
				require.NoError(err)

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
				require.NoError(err)

				req := httptest.NewRequest("POST", signupPath+"/register_user", reqBody)
				req.Header.Add("Content-Type", "application/vnd.msgpack")
				req.Header.Set("Cookie", user.SessionCookie)

				res, err := app.Test(req)
				require.NoError(err)

				if !assert.Equal(t, http.StatusCreated, res.StatusCode) {
					rb, err := errResBody(res.Body)
					require.NoError(err)
					t.Log("unexpected error:", rb)
					return
				}

				rb, err := succResBody[map[string]any](res.Body)
				require.NoError(err)

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
			header := http.Header{}
			header.Set("Cookie", user.SessionCookie)
			wsConn, res, err := websocket.DefaultDialer.Dial(wsPath, header)
			require.NoError(err)

			if !assert.Equal(t, http.StatusSwitchingProtocols, res.StatusCode) {
				rb, err := errResBody(res.Body)
				require.NoError(err)
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

					msgT, wsMsgBt, err := userWSConn.ReadMessage()
					if err != nil {
						break
					}
					require.Equal(websocket.BinaryMessage, msgT)

					err = msgpack.Unmarshal(wsMsgBt, &wsMsg)
					require.NoError(err)

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

		err := wsWriteMsgPack(user1.WSConn, map[string]any{
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
		require.NoError(err)

		// user1's server reply (response) to action
		user1ServerReply := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1ServerReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "direct chat: send message",
			"data": td.Map(map[string]any{
				"new_msg_id": td.Ignore(),
				"che_cursor": td.Ignore(),
			}, nil),
		}, nil))

		user1NewMsgId = user1ServerReply["data"].(map[string]any)["new_msg_id"].(string)
	}

	{
		t.Log("Action: user2 receives the message | acknowledges 'delivered'")

		user2NewMsgReceived := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2NewMsgReceived, td.SuperMapOf(map[string]any{
			"event": "new direct chat",
			"data": td.SuperMapOf(map[string]any{
				"chat": td.SuperMapOf(map[string]any{
					"type": "direct",
					"partner_user": td.SuperMapOf(map[string]any{
						"username": user1.Username,
					}, nil),
				}, nil),
				"history": td.Contains(td.SuperMapOf(map[string]any{
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
				}, nil)),
			}, nil),
		}, nil))

		err := wsWriteMsgPack(user2.WSConn, map[string]any{
			"action": "direct chat: ack messages delivered",
			"data": map[string]any{
				"partnerUsername": user1.Username,
				"msgIds":          []string{user1NewMsgId},
				"at":              time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(err)

		user2ServerReply := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2ServerReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "direct chat: ack messages delivered",
			"data": td.Map(map[string]any{
				"updated_chat_cursor": td.Ignore(),
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user1 receives the 'delivered' acknowledgement | marks message as 'delivered'")

		user1DelvAckReceipt := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1DelvAckReceipt, td.Map(map[string]any{
			"event": "direct chat: messages delivered",
			"data": td.SuperMapOf(map[string]any{
				"chat_partner": user2.Username,
				"msg_ids":      td.Contains(user1NewMsgId),
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user2 then acknowledges 'read'")

		err := wsWriteMsgPack(user2.WSConn, map[string]any{
			"action": "direct chat: ack messages read",
			"data": map[string]any{
				"partnerUsername": user1.Username,
				"msgIds":          []string{user1NewMsgId},
				"at":              time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(err)

		user2ServerReply := <-user2.ServerEventMsg

		// t.Log(user2ServerReply)

		td.Cmp(td.Require(t), user2ServerReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "direct chat: ack messages read",
			"data":     true,
		}, nil))
	}

	{
		t.Log("Action: user1 receives the 'read' acknowledgement | marks message as 'read'")

		user1ReadAckReceipt := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1ReadAckReceipt, td.Map(map[string]any{
			"event": "direct chat: messages read",
			"data": td.SuperMapOf(map[string]any{
				"chat_partner": user2.Username,
				"msg_ids":      td.Contains(user1NewMsgId),
			}, nil),
		}, nil))
	}

	user2NewMsgId := ""

	{
		t.Log("Action: user2 sends message to user1")

		var (
			uploadUrl       string
			mediaCloudName  string
			blurImagePath   = "./test_files/photo_blur.jpg"
			actualImagePath = "./test_files/photo.jpg"
			contentType     = "image/jpeg"
		)

		blurImageInfo, err := os.Stat(blurImagePath)
		require.NoError(err)
		actualImageInfo, err := os.Stat(actualImagePath)
		require.NoError(err)

		{

			t.Log("--- Authorize message media upload ---")

			reqBody, err := makeReqBody(map[string]any{
				"msg_type":   "photo",
				"media_mime": [2]string{contentType, contentType},
				"media_size": [2]int64{blurImageInfo.Size(), actualImageInfo.Size()},
			})
			require.NoError(err)

			req := httptest.NewRequest("POST", chatUploadPath+"/authorize/visual", reqBody)
			req.Header.Set("Cookie", user1.SessionCookie)
			req.Header.Add("Content-Type", "application/vnd.msgpack")

			res, err := app.Test(req)
			require.NoError(err)

			if !assert.Equal(t, http.StatusOK, res.StatusCode) {
				rb, err := errResBody(res.Body)
				require.NoError(err)
				t.Log("unexpected error:", rb)
				return
			}

			rb, err := succResBody[map[string]any](res.Body)
			require.NoError(err)

			td.Cmp(td.Require(t), rb, td.SuperMapOf(map[string]any{
				"uploadUrl":      td.Ignore(),
				"mediaCloudName": td.Ignore(),
			}, nil))

			uploadUrl = rb["uploadUrl"].(string)
			mediaCloudName = rb["mediaCloudName"].(string)
		}

		{
			t.Log("Upload session started:")

			varUploadUrl := make([]string, 2)
			_, err := fmt.Sscanf(uploadUrl, "blur_placeholder:%s actual:%s", &varUploadUrl[0], &varUploadUrl[1])
			require.NoError(err)

			for i, baUploadUrl := range varUploadUrl {
				varMedia := []string{"blur_placeholder", "actual"}
				varPath := []string{blurImagePath, actualImagePath}

				t.Logf("Uploading %s message media started", varMedia[i])

				sessionUrl := startResumableUpload(baUploadUrl, contentType, t)

				uploadFileInChunks(sessionUrl, varPath[i], contentType, logProgress, t)

				t.Logf("Uploading %s message media complete", varMedia[i])
			}

			defer func(mcn string) {
				varMediaCloudName := make([]string, 2)
				_, err = fmt.Sscanf(mcn, "blur_placeholder:%s actual:%s", &varMediaCloudName[0], &varMediaCloudName[1])
				require.NoError(err)

				for _, baMcn := range varMediaCloudName {
					err := appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).Object(baMcn).Delete(t.Context())
					require.NoError(err)
				}
			}(mediaCloudName)

			t.Log("Upload complete")
		}

		err = wsWriteMsgPack(user2.WSConn, map[string]any{
			"action": "direct chat: send message",
			"data": map[string]any{
				"partnerUsername": user1.Username,
				"msg": map[string]any{
					"type": "photo",
					"props": map[string]any{
						"media_cloud_name": mediaCloudName,
						"caption":          "I'm guuud!",
					},
				},
				"at": time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(err)

		// user2's server reply (response) to action
		user2ServerReply := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2ServerReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "direct chat: send message",
			"data": td.Map(map[string]any{
				"new_msg_id": td.Ignore(),
				"che_cursor": td.Ignore(),
			}, nil),
		}, nil))

		user2NewMsgId = user2ServerReply["data"].(map[string]any)["new_msg_id"].(string)
	}

	{
		t.Log("Action: user1 receives the message | acknowledges 'delivered'")

		user1NewMsgReceived := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1NewMsgReceived, td.Map(map[string]any{
			"event": "direct chat: new che: message",
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

		err := wsWriteMsgPack(user1.WSConn, map[string]any{
			"action": "direct chat: ack messages delivered",
			"data": map[string]any{
				"partnerUsername": user2.Username,
				"msgIds":          []string{user2NewMsgId},
				"at":              time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(err)

		user1ServerReply := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1ServerReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "direct chat: ack messages delivered",
			"data": td.Map(map[string]any{
				"updated_chat_cursor": td.Ignore(),
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user2 receives the 'delivered' acknowledgement | marks message as 'delivered'")

		user2DelvAckReceipt := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2DelvAckReceipt, td.Map(map[string]any{
			"event": "direct chat: messages delivered",
			"data": td.SuperMapOf(map[string]any{
				"chat_partner": user1.Username,
				"msg_ids":      td.Contains(user2NewMsgId),
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user1 then acknowledges 'read'")

		err := wsWriteMsgPack(user1.WSConn, map[string]any{
			"action": "direct chat: ack messages read",
			"data": map[string]any{
				"partnerUsername": user2.Username,
				"msgIds":          []string{user2NewMsgId},
				"at":              time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(err)

		user1ServerReply := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1ServerReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "direct chat: ack messages read",
			"data":     true,
		}, nil))
	}

	{
		t.Log("Action: user2 receives the 'read' acknowledgement | marks message as 'read'")

		user2ReadAckReceipt := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2ReadAckReceipt, td.Map(map[string]any{
			"event": "direct chat: messages read",
			"data": td.SuperMapOf(map[string]any{
				"chat_partner": user1.Username,
				"msg_ids":      td.Contains(user2NewMsgId),
			}, nil),
		}, nil))
	}

	{
		<-(time.NewTimer(100 * time.Millisecond).C)

		t.Log("Action: user1 opens his chat history with user2")

		req := httptest.NewRequest("GET", directChatPath+"/"+user2.Username+"/history", nil)
		req.Header.Add("Content-Type", "application/vnd.msgpack")
		req.Header.Set("Cookie", user1.SessionCookie)

		res, err := app.Test(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[[]map[string]any](res.Body)
		require.NoError(err)

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
