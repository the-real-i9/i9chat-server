package tests

import (
	"fmt"
	"i9chat/src/helpers"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func XTestGroupChat(t *testing.T) {
	t.Parallel()

	user1 := UserT{
		Email:    "harrydasouza@gmail.com",
		Username: "harry",
		Password: "harry_dasou",
		Geolocation: UserGeolocation{
			X: 0.0,
			Y: 5.0,
		},
	}

	user2 := UserT{
		Email:    "conradharrigan@gmail.com",
		Username: "conrad",
		Password: "grandpa_harr",
		Geolocation: UserGeolocation{
			X: 1.0,
			Y: 6.0,
		},
	}

	user3 := UserT{
		Email:    "kevinharrigan@gmail.com",
		Username: "kevin",
		Password: "daddy_harr",
		Geolocation: UserGeolocation{
			X: 2.0,
			Y: 7.0,
		},
	}

	user4 := UserT{
		Email:    "eddieharrigan@gmail.com",
		Username: "eddie",
		Password: "badchild_harr",
		Geolocation: UserGeolocation{
			X: 3.0,
			Y: 6.0,
		},
	}

	user5 := UserT{
		Email:    "meaveharrigan@gmail.com",
		Username: "meave",
		Password: "witchie_harr",
		Geolocation: UserGeolocation{
			X: 4.0,
			Y: 5.0,
		},
	}

	{
		t.Log("Setup: create new accounts for users")

		for _, user := range []*UserT{&user1, &user2, &user3, &user4, &user5} {

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

		for _, user := range []*UserT{&user1, &user2, &user3, &user4, &user5} {

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

	groupChatPic, err := os.ReadFile("./test_files/group_pic.png")
	require.NoError(t, err)

	newGroup := struct {
		Id          string
		Name        string
		Description string
		Picture     []byte
	}{Name: "World Changers! 💪💪💪", Description: "We're world changers! Join the train!", Picture: groupChatPic}

	{
		t.Log("Action: user1 creates group chat with user2 | user2 receives the new group")

		reqBody, err := makeReqBody(map[string]any{
			"name":        newGroup.Name,
			"description": newGroup.Description,
			"pictureData": newGroup.Picture,
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
			"group": td.SuperMapOf(map[string]any{
				"id":          td.Ignore(),
				"name":        newGroup.Name,
				"description": newGroup.Description,
			}, nil),
			"history": td.Contains(td.SuperMapOf(map[string]any{
				"che_type": "group activity",
				"info":     "You added " + user2.Username,
			}, nil)),
		}, nil))

		newGroup.Id = rb["group"].(map[string]any)["id"].(string)

		user2RecvNewGroup := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2RecvNewGroup, td.Map(map[string]any{
			"event": "new group chat",
			"data": td.SuperMapOf(map[string]any{
				"group": td.SuperMapOf(map[string]any{
					"id":          newGroup.Id,
					"name":        newGroup.Name,
					"description": newGroup.Description,
				}, nil),
				"history": td.Contains(td.SuperMapOf(map[string]any{
					"che_type": "group activity",
					"info":     "You were added",
				}, nil)),
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user3 joins group | user1 & user2 are notified")

		reqBody, err := makeReqBody(map[string]any{})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", groupChatPath+"/"+newGroup.Id+"/execute_action/join", reqBody)
		require.NoError(t, err)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Set("Cookie", user3.SessionCookie)

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
			"group": td.SuperMapOf(map[string]any{
				"id":          newGroup.Id,
				"name":        newGroup.Name,
				"description": newGroup.Description,
			}, nil),
			"history": td.Contains(td.SuperMapOf(map[string]any{
				"che_type": "group activity",
				"info":     "You joined",
			}, nil)),
		}, nil))

		user1GCJoinNotif := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1GCJoinNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     user3.Username + " joined",
				"group_id": newGroup.Id,
			}, nil),
		}, nil))

		user2GCJoinNotif := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2GCJoinNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     user3.Username + " joined",
				"group_id": newGroup.Id,
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user1 makes user2 group admin | user2 & other members are notified")

		reqBody, err := makeReqBody(map[string]any{
			"user": user2.Username,
		})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", groupChatPath+"/"+newGroup.Id+"/execute_action/make user admin", reqBody)
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

		rb, err := succResBody[string](res.Body)
		require.NoError(t, err)

		require.Equal(t, fmt.Sprintf("You made %s group admin", user2.Username), rb)

		user2GCNewAdminNotif := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2GCNewAdminNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     fmt.Sprintf("%s made you group admin", user1.Username),
				"group_id": newGroup.Id,
			}, nil),
		}, nil))

		user3GCNewAdminNotif := <-user3.ServerEventMsg

		td.Cmp(td.Require(t), user3GCNewAdminNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     fmt.Sprintf("%s made %s group admin", user1.Username, user2.Username),
				"group_id": newGroup.Id,
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user2 adds user4 & user5 | user4, user5 and other members are notified")

		reqBody, err := makeReqBody(map[string]any{
			"newUsers": []string{user4.Username, user5.Username},
		})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", groupChatPath+"/"+newGroup.Id+"/execute_action/add users", reqBody)
		require.NoError(t, err)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Set("Cookie", user2.SessionCookie)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[string](res.Body)
		require.NoError(t, err)

		require.Equal(t, fmt.Sprintf("You added %s", helpers.JoinWithCommaAnd(user4.Username, user5.Username)), rb)

		user4GCUserAddedNotif := <-user4.ServerEventMsg

		td.Cmp(td.Require(t), user4GCUserAddedNotif, td.Map(map[string]any{
			"event": "new group chat",
			"data": td.SuperMapOf(map[string]any{
				"group": td.SuperMapOf(map[string]any{
					"id":          newGroup.Id,
					"name":        newGroup.Name,
					"description": newGroup.Description,
				}, nil),
				"history": td.Contains(td.SuperMapOf(map[string]any{
					"che_type": "group activity",
					"info":     "You were added",
				}, nil)),
			}, nil),
		}, nil))

		user5GCUserAddedNotif := <-user5.ServerEventMsg

		td.Cmp(td.Require(t), user5GCUserAddedNotif, td.Map(map[string]any{
			"event": "new group chat",
			"data": td.SuperMapOf(map[string]any{
				"group": td.SuperMapOf(map[string]any{
					"id":          newGroup.Id,
					"name":        newGroup.Name,
					"description": newGroup.Description,
				}, nil),
				"history": td.Contains(td.SuperMapOf(map[string]any{
					"che_type": "group activity",
					"info":     "You were added",
				}, nil)),
			}, nil),
		}, nil))

		user1GCNewUsersAddedNotif := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1GCNewUsersAddedNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     fmt.Sprintf("%s added %s", user2.Username, helpers.JoinWithCommaAnd(user4.Username, user5.Username)),
				"group_id": newGroup.Id,
			}, nil),
		}, nil))

		user3GCNewUsersAddedNotif := <-user3.ServerEventMsg

		td.Cmp(td.Require(t), user3GCNewUsersAddedNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     fmt.Sprintf("%s added %s", user2.Username, helpers.JoinWithCommaAnd(user4.Username, user5.Username)),
				"group_id": newGroup.Id,
			}, nil),
		}, nil))
	}

	oldGroupName := newGroup.Name

	{
		t.Log("Action: user1 changes group name | other members are notified")

		newGroup.Name = "Programmer's Hub"

		reqBody, err := makeReqBody(map[string]any{
			"newName": newGroup.Name,
		})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", groupChatPath+"/"+newGroup.Id+"/execute_action/change name", reqBody)
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

		rb, err := succResBody[string](res.Body)
		require.NoError(t, err)

		require.Equal(t, fmt.Sprintf("You changed group name from %s to %s", oldGroupName, newGroup.Name), rb)

		user2GCNameChangeNotif := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2GCNameChangeNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     fmt.Sprintf("%s changed group name from %s to %s", user1.Username, oldGroupName, newGroup.Name),
				"group_id": newGroup.Id,
			}, nil),
		}, nil))

		user3GCNameChangeNotif := <-user3.ServerEventMsg

		td.Cmp(td.Require(t), user3GCNameChangeNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     fmt.Sprintf("%s changed group name from %s to %s", user1.Username, oldGroupName, newGroup.Name),
				"group_id": newGroup.Id,
			}, nil),
		}, nil))

		user4GCNameChangeNotif := <-user4.ServerEventMsg

		td.Cmp(td.Require(t), user4GCNameChangeNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     fmt.Sprintf("%s changed group name from %s to %s", user1.Username, oldGroupName, newGroup.Name),
				"group_id": newGroup.Id,
			}, nil),
		}, nil))

		user5GCNameChangeNotif := <-user5.ServerEventMsg

		td.Cmp(td.Require(t), user5GCNameChangeNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     fmt.Sprintf("%s changed group name from %s to %s", user1.Username, oldGroupName, newGroup.Name),
				"group_id": newGroup.Id,
			}, nil),
		}, nil))
	}

	oldGroupDescription := newGroup.Description

	{
		t.Log("Action: user1 changes group description | other members are notified")

		newGroup.Description = "We're all programmers here!"

		reqBody, err := makeReqBody(map[string]any{
			"newDescription": newGroup.Description,
		})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", groupChatPath+"/"+newGroup.Id+"/execute_action/change description", reqBody)
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

		rb, err := succResBody[string](res.Body)
		require.NoError(t, err)

		require.Equal(t, fmt.Sprintf("You changed group description from %s to %s", oldGroupDescription, newGroup.Description), rb)

		user2GCDescriptionChangeNotif := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2GCDescriptionChangeNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     fmt.Sprintf("%s changed group description from %s to %s", user1.Username, oldGroupDescription, newGroup.Description),
				"group_id": newGroup.Id,
			}, nil),
		}, nil))

		user3GCDescriptionChangeNotif := <-user3.ServerEventMsg

		td.Cmp(td.Require(t), user3GCDescriptionChangeNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     fmt.Sprintf("%s changed group description from %s to %s", user1.Username, oldGroupDescription, newGroup.Description),
				"group_id": newGroup.Id,
			}, nil),
		}, nil))

		user4GCDescriptionChangeNotif := <-user4.ServerEventMsg

		td.Cmp(td.Require(t), user4GCDescriptionChangeNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     fmt.Sprintf("%s changed group description from %s to %s", user1.Username, oldGroupDescription, newGroup.Description),
				"group_id": newGroup.Id,
			}, nil),
		}, nil))

		user5GCDescriptionChangeNotif := <-user5.ServerEventMsg

		td.Cmp(td.Require(t), user5GCDescriptionChangeNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     fmt.Sprintf("%s changed group description from %s to %s", user1.Username, oldGroupDescription, newGroup.Description),
				"group_id": newGroup.Id,
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user2 changes group picture | other members are notified")

		newGroup.Picture = groupChatPic

		reqBody, err := makeReqBody(map[string]any{
			"newPictureData": newGroup.Picture,
		})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", groupChatPath+"/"+newGroup.Id+"/execute_action/change picture", reqBody)
		require.NoError(t, err)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Set("Cookie", user2.SessionCookie)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[string](res.Body)
		require.NoError(t, err)

		require.Equal(t, "You changed group picture", rb)

		user1GCPictureChangeNotif := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1GCPictureChangeNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     fmt.Sprintf("%s changed group picture", user2.Username),
				"group_id": newGroup.Id,
			}, nil),
		}, nil))

		user3GCPictureChangeNotif := <-user3.ServerEventMsg

		td.Cmp(td.Require(t), user3GCPictureChangeNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     fmt.Sprintf("%s changed group picture", user2.Username),
				"group_id": newGroup.Id,
			}, nil),
		}, nil))

		user4GCPictureChangeNotif := <-user4.ServerEventMsg

		td.Cmp(td.Require(t), user4GCPictureChangeNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     fmt.Sprintf("%s changed group picture", user2.Username),
				"group_id": newGroup.Id,
			}, nil),
		}, nil))

		user5GCPictureChangeNotif := <-user5.ServerEventMsg

		td.Cmp(td.Require(t), user5GCPictureChangeNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     fmt.Sprintf("%s changed group picture", user2.Username),
				"group_id": newGroup.Id,
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user2 removes user1 from group admins | user1 & other members are notified")

		reqBody, err := makeReqBody(map[string]any{
			"user": user1.Username,
		})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", groupChatPath+"/"+newGroup.Id+"/execute_action/remove user from admins", reqBody)
		require.NoError(t, err)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Set("Cookie", user2.SessionCookie)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[string](res.Body)
		require.NoError(t, err)

		require.Equal(t, fmt.Sprintf("You removed %s from group admins", user1.Username), rb)

		user1GCAdminRemovalNotif := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1GCAdminRemovalNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     fmt.Sprintf("%s removed you from group admins", user2.Username),
				"group_id": newGroup.Id,
			}, nil),
		}, nil))

		user3GCAdminRemovalNotif := <-user3.ServerEventMsg

		td.Cmp(td.Require(t), user3GCAdminRemovalNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     fmt.Sprintf("%s removed %s from group admins", user2.Username, user1.Username),
				"group_id": newGroup.Id,
			}, nil),
		}, nil))

		user4GCAdminRemovalNotif := <-user4.ServerEventMsg

		td.Cmp(td.Require(t), user4GCAdminRemovalNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     fmt.Sprintf("%s removed %s from group admins", user2.Username, user1.Username),
				"group_id": newGroup.Id,
			}, nil),
		}, nil))

		user5GCAdminRemovalNotif := <-user5.ServerEventMsg

		td.Cmp(td.Require(t), user5GCAdminRemovalNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     fmt.Sprintf("%s removed %s from group admins", user2.Username, user1.Username),
				"group_id": newGroup.Id,
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user2 removes user1 from group | user1 & other members are notified")

		reqBody, err := makeReqBody(map[string]any{
			"user": user1.Username,
		})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", groupChatPath+"/"+newGroup.Id+"/execute_action/remove user", reqBody)
		require.NoError(t, err)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Set("Cookie", user2.SessionCookie)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[string](res.Body)
		require.NoError(t, err)

		require.Equal(t, fmt.Sprintf("You removed %s", user1.Username), rb)

		user1GCAdminRemovalNotif := <-user1.ServerEventMsg

		td.Cmp(td.Require(t), user1GCAdminRemovalNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     fmt.Sprintf("%s removed you", user2.Username),
				"group_id": newGroup.Id,
			}, nil),
		}, nil))

		user3GCAdminRemovalNotif := <-user3.ServerEventMsg

		td.Cmp(td.Require(t), user3GCAdminRemovalNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     fmt.Sprintf("%s removed %s", user2.Username, user1.Username),
				"group_id": newGroup.Id,
			}, nil),
		}, nil))

		user4GCAdminRemovalNotif := <-user4.ServerEventMsg

		td.Cmp(td.Require(t), user4GCAdminRemovalNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     fmt.Sprintf("%s removed %s", user2.Username, user1.Username),
				"group_id": newGroup.Id,
			}, nil),
		}, nil))

		user5GCAdminRemovalNotif := <-user5.ServerEventMsg

		td.Cmp(td.Require(t), user5GCAdminRemovalNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     fmt.Sprintf("%s removed %s", user2.Username, user1.Username),
				"group_id": newGroup.Id,
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user3 leaves group | other members are notified")

		reqBody, err := makeReqBody(map[string]any{})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", groupChatPath+"/"+newGroup.Id+"/execute_action/leave", reqBody)
		require.NoError(t, err)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Set("Cookie", user3.SessionCookie)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[string](res.Body)
		require.NoError(t, err)

		require.Equal(t, "You left", rb)

		user2GCLeaveNotif := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2GCLeaveNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     user3.Username + " left",
				"group_id": newGroup.Id,
			}, nil),
		}, nil))

		user4GCLeaveNotif := <-user4.ServerEventMsg

		td.Cmp(td.Require(t), user4GCLeaveNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     user3.Username + " left",
				"group_id": newGroup.Id,
			}, nil),
		}, nil))

		user5GCLeaveNotif := <-user5.ServerEventMsg

		td.Cmp(td.Require(t), user5GCLeaveNotif, td.Map(map[string]any{
			"event": "group chat: new activity",
			"data": td.Map(map[string]any{
				"info":     user3.Username + " left",
				"group_id": newGroup.Id,
			}, nil),
		}, nil))
	}

	{
		<-(time.NewTimer(100 * time.Millisecond).C)

		t.Log("user2 requests group info")

		err := user2.WSConn.WriteJSON(map[string]any{
			"action": "group: get info",
			"data": map[string]any{
				"groupId": newGroup.Id,
			},
		})
		require.NoError(t, err)

		// user2's server reply (response) to action
		user2ServerReply := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2ServerReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "group: get info",
			"data": td.SuperMapOf(map[string]any{
				"name":                 newGroup.Name,
				"description":          newGroup.Description,
				"members_count":        td.Lax(3),
				"online_members_count": td.Lax(3),
			}, nil),
		}, nil))
	}

	user4NewMsgId := ""

	{
		t.Log("Action: user4 sends message to group | other members receives the message")

		err := user4.WSConn.WriteJSON(map[string]any{
			"action": "group chat: send message",
			"data": map[string]any{
				"groupId": newGroup.Id,
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

		// user4's server reply (response) to action
		user4ServerReply := <-user4.ServerEventMsg

		td.Cmp(td.Require(t), user4ServerReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "group chat: send message",
			"data": td.Map(map[string]any{
				"new_msg_id": td.Ignore(),
			}, nil),
		}, nil))

		user4NewMsgId = user4ServerReply["data"].(map[string]any)["new_msg_id"].(string)
	}

	{
		t.Log("Action: user2 & user 5 receives the new message in group")

		user2NewMsgReceived := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2NewMsgReceived, td.Map(map[string]any{
			"event": "group chat: new message",
			"data": td.Map(map[string]any{
				"message": td.SuperMapOf(map[string]any{
					"id": user4NewMsgId,
					"content": td.SuperMapOf(map[string]any{
						"type": "text",
						"props": td.SuperMapOf(map[string]any{
							"text_content": "Hi. How're you doing?",
						}, nil),
					}, nil),
					"delivery_status": "sent",
					"sender": td.SuperMapOf(map[string]any{
						"username": user4.Username,
					}, nil),
				}, nil),
				"group_id": newGroup.Id,
			}, nil),
		}, nil))

		user5NewMsgReceived := <-user5.ServerEventMsg

		td.Cmp(td.Require(t), user5NewMsgReceived, td.Map(map[string]any{
			"event": "group chat: new message",
			"data": td.Map(map[string]any{
				"message": td.SuperMapOf(map[string]any{
					"id": user4NewMsgId,
					"content": td.SuperMapOf(map[string]any{
						"type": "text",
						"props": td.SuperMapOf(map[string]any{
							"text_content": "Hi. How're you doing?",
						}, nil),
					}, nil),
					"delivery_status": "sent",
					"sender": td.SuperMapOf(map[string]any{
						"username": user4.Username,
					}, nil),
				}, nil),
				"group_id": newGroup.Id,
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user2 & user5 acknowledges 'delivered'")

		// user2 acknowledges 'delivered'
		err := user2.WSConn.WriteJSON(map[string]any{
			"action": "group chat: ack message delivered",
			"data": map[string]any{
				"groupId": newGroup.Id,
				"msgId":   user4NewMsgId,
				"at":      time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(t, err)

		user2ServerReply := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2ServerReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "group chat: ack message delivered",
			"data":     true,
		}, nil))

		// user5 acknowledges 'delivered'
		err = user5.WSConn.WriteJSON(map[string]any{
			"action": "group chat: ack message delivered",
			"data": map[string]any{
				"groupId": newGroup.Id,
				"msgId":   user4NewMsgId,
				"at":      time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(t, err)

		user5ServerReply := <-user5.ServerEventMsg

		td.Cmp(td.Require(t), user5ServerReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "group chat: ack message delivered",
			"data":     true,
		}, nil))
	}

	{
		t.Log("Action: message is now 'delivered' to all other members | all receive the 'delivered' acknowledgement; marking message as 'delivered'")

		user2DelvAckReceipt := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2DelvAckReceipt, td.Map(map[string]any{
			"event": "group chat: message delivered",
			"data": td.Map(map[string]any{
				"group_id": newGroup.Id,
				"msg_id":   user4NewMsgId,
			}, nil),
		}, nil))

		user4DelvAckReceipt := <-user4.ServerEventMsg

		td.Cmp(td.Require(t), user4DelvAckReceipt, td.Map(map[string]any{
			"event": "group chat: message delivered",
			"data": td.Map(map[string]any{
				"group_id": newGroup.Id,
				"msg_id":   user4NewMsgId,
			}, nil),
		}, nil))

		user5DelvAckReceipt := <-user5.ServerEventMsg

		td.Cmp(td.Require(t), user5DelvAckReceipt, td.Map(map[string]any{
			"event": "group chat: message delivered",
			"data": td.Map(map[string]any{
				"group_id": newGroup.Id,
				"msg_id":   user4NewMsgId,
			}, nil),
		}, nil))

	}

	{
		t.Log("Action: user2 & user5 then acknowledges 'read'")

		// user5 acknowledges 'read'
		err := user5.WSConn.WriteJSON(map[string]any{
			"action": "group chat: ack message read",
			"data": map[string]any{
				"groupId": newGroup.Id,
				"msgId":   user4NewMsgId,
				"at":      time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(t, err)

		user5ServerReply := <-user5.ServerEventMsg

		td.Cmp(td.Require(t), user5ServerReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "group chat: ack message read",
			"data":     true,
		}, nil))

		// user2 acknowledges 'read'
		err = user2.WSConn.WriteJSON(map[string]any{
			"action": "group chat: ack message read",
			"data": map[string]any{
				"groupId": newGroup.Id,
				"msgId":   user4NewMsgId,
				"at":      time.Now().UTC().UnixMilli(),
			},
		})
		require.NoError(t, err)

		user2ServerReply := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2ServerReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "group chat: ack message read",
			"data":     true,
		}, nil))
	}

	{
		t.Log("Action: message is now 'read' by all other members | all members receive the 'read' acknowledgement; marking message as 'read'")

		user5ReadAckReceipt := <-user5.ServerEventMsg

		td.Cmp(td.Require(t), user5ReadAckReceipt, td.Map(map[string]any{
			"event": "group chat: message read",
			"data": td.Map(map[string]any{
				"group_id": newGroup.Id,
				"msg_id":   user4NewMsgId,
			}, nil),
		}, nil))

		user4ReadAckReceipt := <-user4.ServerEventMsg

		td.Cmp(td.Require(t), user4ReadAckReceipt, td.Map(map[string]any{
			"event": "group chat: message read",
			"data": td.Map(map[string]any{
				"group_id": newGroup.Id,
				"msg_id":   user4NewMsgId,
			}, nil),
		}, nil))

		user2ReadAckReceipt := <-user2.ServerEventMsg

		td.Cmp(td.Require(t), user2ReadAckReceipt, td.Map(map[string]any{
			"event": "group chat: message read",
			"data": td.Map(map[string]any{
				"group_id": newGroup.Id,
				"msg_id":   user4NewMsgId,
			}, nil),
		}, nil))
	}

	{
		<-(time.NewTimer(100 * time.Millisecond).C)

		t.Log("Action: user5 opens group chat history")

		req, err := http.NewRequest("GET", groupChatPath+"/"+newGroup.Id+"/history", nil)
		require.NoError(t, err)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Set("Cookie", user5.SessionCookie)

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
					"che_type": "group activity",
					"info":     "You were added",
				}, nil)),
				td.Contains(td.SuperMapOf(map[string]any{
					"che_type": "group activity",
					"info":     fmt.Sprintf("%s changed group name from %s to %s", user1.Username, oldGroupName, newGroup.Name),
				}, nil)),
				td.Contains(td.SuperMapOf(map[string]any{
					"che_type": "group activity",
					"info":     fmt.Sprintf("%s changed group description from %s to %s", user1.Username, oldGroupDescription, newGroup.Description),
				}, nil)),
				td.Contains(td.SuperMapOf(map[string]any{
					"che_type": "group activity",
					"info":     fmt.Sprintf("%s changed group picture", user2.Username),
				}, nil)),
				td.Contains(td.SuperMapOf(map[string]any{
					"che_type": "group activity",
					"info":     fmt.Sprintf("%s removed %s from group admins", user2.Username, user1.Username),
				}, nil)),
				td.Contains(td.SuperMapOf(map[string]any{
					"che_type": "group activity",
					"info":     fmt.Sprintf("%s removed %s", user2.Username, user1.Username),
				}, nil)),
				td.Contains(td.SuperMapOf(map[string]any{
					"che_type": "group activity",
					"info":     fmt.Sprintf("%s left", user3.Username),
				}, nil)),
				td.Contains(td.SuperMapOf(map[string]any{
					"che_type": "message",
					"id":       user4NewMsgId,
					"content": td.SuperMapOf(map[string]any{
						"type": "text",
						"props": td.SuperMapOf(map[string]any{
							"text_content": "Hi. How're you doing?",
						}, nil),
					}, nil),
					"delivery_status": "read",
					"sender": td.SuperMapOf(map[string]any{
						"username": user4.Username,
					}, nil),
				}, nil)),
			),
		)
	}
}
