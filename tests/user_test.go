package tests

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

const userPath = HOST_URL + "/api/app/user"

func TestUserOps(t *testing.T) {
	t.Parallel()

	accounts := map[string]map[string]any{
		"user1": {
			"email":    "mikeross@gmail.com",
			"username": "mikeross",
			"password": "recheal_zane",
			"phone":    "08183443588",
			"geolocation": map[string]any{
				"x": 5.0,
				"y": 2.0,
			},
		},
		"user2": {
			"email":    "harveyspecter@gmail.com",
			"username": "harvey",
			"password": "scottie_",
			"phone":    "08183443589",
			"geolocation": map[string]any{
				"x": 4.0,
				"y": 3.0,
			},
		},
		"user3": {
			"email":    "jessicapearson@gmail.com",
			"username": "jessica",
			"password": "jeff_malone",
			"phone":    "08183443590",
			"geolocation": map[string]any{
				"x": 3.0,
				"y": 4.0,
			},
		},
	}

	t.Run("create 3 accounts", func(t *testing.T) {

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

	t.Run("bring users online", func(t *testing.T) {

	})

	t.Run("change user1's phone number", func(t *testing.T) {
		reqBody, err := reqBody(map[string]any{"newPhoneNumber": "07083249523"})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", userPath+"/change_phone_number", reqBody)
		require.NoError(t, err)
		req.Header.Set("Cookie", accounts["user1"]["session_cookie"].(string))
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, res.StatusCode)
		bd, err := resBody(res.Body)
		require.NoError(t, err)
		require.NotEmpty(t, bd)
	})

	t.Run("confirm user1's updated profile", func(t *testing.T) {
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

		require.Equal(t, accounts["user1"]["phone"], "07083249523")
	})

	t.Run("change user2's geolocation", func(t *testing.T) {
		reqBody, err := reqBody(map[string]any{"newGeolocation": map[string]any{"x": 3.0, "y": 3.0}})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", userPath+"/update_geolocation", reqBody)
		require.NoError(t, err)
		req.Header.Set("Cookie", accounts["user2"]["session_cookie"].(string))
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, res.StatusCode)
		bd, err := resBody(res.Body)
		require.NoError(t, err)
		require.NotEmpty(t, bd)
	})

	t.Run("confirm user2's updated profile", func(t *testing.T) {
		username := accounts["user2"]["username"].(string)

		req, err := http.NewRequest("GET", userPath+"/find_user?emailUsernamePhone="+username, nil)
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

		require.Equal(t, accounts["user2"]["geolocation"], map[string]any{"x": 3.0, "y": 3.0})
	})

	t.Run("find nearby users", func(t *testing.T) {
		x, y, radius := 2.0, 2.0, 10.0

		req, err := http.NewRequest("GET", fmt.Sprintf("%s%s?x=%f&y=%f&radius=%f", userPath, "/find_nearby_users", x, y, radius), nil)
		require.NoError(t, err)
		req.Header.Set("Cookie", accounts["user1"]["session_cookie"].(string))

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, res.StatusCode)

		bd, err := resBody(res.Body)
		require.NoError(t, err)
		require.NotEmpty(t, bd)

		var data []any

		require.NoError(t, json.Unmarshal(bd, &data))

		require.NotEmpty(t, data)
	})
}
