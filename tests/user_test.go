package tests

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserOps(t *testing.T) {
	t.Parallel()

	user1 := UserT{
		Email:    "mikeross@gmail.com",
		Username: "mikeross",
		Password: "recheal_zane",
		Geolocation: UserGeolocation{
			X: 5.0,
			Y: 2.0,
		},
	}

	user2 := UserT{
		Email:    "harveyspecter@gmail.com",
		Username: "harvey",
		Password: "scottie_",
		Geolocation: UserGeolocation{
			X: 4.0,
			Y: 3.0,
		},
	}

	user3 := UserT{
		Email:    "jessicapearson@gmail.com",
		Username: "jessica",
		Password: "jeff_malone",
		Geolocation: UserGeolocation{
			X: 3.0,
			Y: 4.0,
		},
	}

	{
		t.Log("Setup: create new accounts for users")

		for _, user := range []*UserT{&user1, &user2, &user3} {
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
		t.Log("Action: user2 changes her geolocation")

		user2.Geolocation.X = 3.0
		user2.Geolocation.Y = 3.0

		reqBody, err := makeReqBody(map[string]any{"newGeolocation": map[string]any{"x": user2.Geolocation.X, "y": user2.Geolocation.Y}})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", userPath+"/update_geolocation", reqBody)
		require.NoError(t, err)
		req.Header.Set("Cookie", user2.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[bool](res.Body)
		require.NoError(t, err)
		require.True(t, rb)
	}

	{
		t.Log("Action: user2 confirms her updated profile | geolocation changed")

		req, err := http.NewRequest("GET", userPath+"/find_user?eup="+user2.Username, nil)
		require.NoError(t, err)
		req.Header.Set("Cookie", user2.SessionCookie)

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
			"username": user2.Username,
			"geolocation": td.SuperMapOf(map[string]any{
				"x": user2.Geolocation.X,
				"y": user2.Geolocation.Y,
			}, nil),
		}, nil))
	}

	{
		t.Log("Action: user1 finds users nearby")

		x, y, radius := 2.0, 2.0, 10.0

		req, err := http.NewRequest("GET", fmt.Sprintf("%s%s?x=%f&y=%f&radius=%f", userPath, "/find_nearby_users", x, y, radius), nil)
		require.NoError(t, err)
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

		td.Cmp(td.Require(t), rb, td.Contains(td.Any(
			td.SuperMapOf(map[string]any{
				"username": user2.Username,
			}, nil),
			td.SuperMapOf(map[string]any{
				"username": user3.Username,
			}, nil),
		)))
	}
}
