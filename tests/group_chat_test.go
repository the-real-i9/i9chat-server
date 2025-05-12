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

const groupChatPath string = HOST_URL + "/api/app/group_chat"

func TestGroupChat(t *testing.T) {
	t.Parallel()

	user1 := UserT{
		Email:    "harrydasouza@gmail.com",
		Username: "harry",
		Password: "harry_dasou",
		Phone:    "07049423518",
		Geolocation: UserGeolocation{
			X: 0.0,
			Y: 5.0,
		},
	}

	user2 := UserT{
		Email:    "conradharrigan@gmail.com",
		Username: "conrad",
		Password: "grandpa_harr",
		Phone:    "09113625189",
		Geolocation: UserGeolocation{
			X: 1.0,
			Y: 6.0,
		},
	}

	user3 := UserT{
		Email:    "kevinharrigan@gmail.com",
		Username: "kevin",
		Password: "daddy_harr",
		Phone:    "09113615682",
		Geolocation: UserGeolocation{
			X: 2.0,
			Y: 7.0,
		},
	}

	user4 := UserT{
		Email:    "eddieharrigan@gmail.com",
		Username: "eddie",
		Password: "badchild_harr",
		Phone:    "09125614672",
		Geolocation: UserGeolocation{
			X: 3.0,
			Y: 6.0,
		},
	}

	user5 := UserT{
		Email:    "meaveharrigan@gmail.com",
		Username: "meave",
		Password: "witchie_harr",
		Phone:    "07025514772",
		Geolocation: UserGeolocation{
			X: 4.0,
			Y: 5.0,
		},
	}

	{
		t.Log("Setup: create new accounts for users")

		for _, user := range []*UserT{&user1, &user2, &user3, &user4, &user5} {
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
		t.Log("Action: user1 creates group chat with user2 | user2 is notified")

		groupPic, err := os.ReadFile("./test_files/group_pic.png")
		require.NoError(t, err)

		reqBody, err := makeReqBody(map[string]any{
			"name":        "World Changers! 💪💪💪",
			"description": "We're world changers! Join the train!",
			"pictureData": groupPic,
			"initUsers":   []string{user2.Username},
			"createdAt":   time.Now().UTC().UnixMilli(),
		})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", groupChatPath+"/new", reqBody)
		require.NoError(t, err)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Set("Cookie", user1.SessionCookie)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusCreated, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[map[string]any](res.Body)
		require.NoError(t, err)

		td.Cmp(td.Require(t), rb, td.SuperMapOf(map[string]any{
			"id":          td.Ignore(),
			"name":        "World Changers! 💪💪💪",
			"description": "We're world changers! Join the train!",
			"last_activity": td.Map(map[string]any{
				"type": "group activity",
				"info": "You added " + user2.Username,
			}, nil),
		}, nil))
	}
}
