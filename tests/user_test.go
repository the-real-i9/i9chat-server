package tests

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

const userPath = HOST_URL + "/api/app/user"

func TestUserRoutes(t *testing.T) {
	// t.Parallel()

	accounts := map[string]map[string]any{
		"user1": {
			"email":    "mikeross@gmail.com",
			"username": "mikeross",
			"password": "recheal_zane",
			"phone":    "08183443588",
			"geolocation": map[string]any{
				"longitude": 5.0,
				"latitude":  2.0,
			},
		},
		"user2": {
			"email":    "harveyspecter@gmail.com",
			"username": "harvey",
			"password": "scottie_",
			"phone":    "08183443589",
			"geolocation": map[string]any{
				"longitude": 4.0,
				"latitude":  3.0,
			},
		},
		"user3": {
			"email":    "jessicapearson@gmail.com",
			"username": "jessica",
			"password": "jeff_malone",
			"phone":    "08183443590",
			"geolocation": map[string]any{
				"longitude": 3.0,
				"latitude":  4.0,
			},
		},
	}

	t.Run("create 3 accounts", func(t *testing.T) {

		for user, info := range accounts {
			t.Run("step one: request new account", func(t *testing.T) {
				reqBody, err := reqBody(map[string]any{"email": info["email"]})
				assert.NoError(t, err)

				res, err := http.Post(signupPath+"/request_new_account", "application/json", reqBody)
				assert.NoError(t, err)

				if !assert.Equal(t, http.StatusOK, res.StatusCode) {
					bd, err := resBody(res.Body)
					assert.NoError(t, err)
					t.Logf("%s", bd)
				}

				accounts[user]["signup_session_cookie"] = res.Header.Get("Set-Cookie")
			})

			t.Run("step two: verify email", func(t *testing.T) {
				verfCode, err := strconv.Atoi(os.Getenv("DUMMY_VERF_TOKEN"))
				assert.NoError(t, err)

				reqBody, err := reqBody(map[string]any{"code": verfCode})
				assert.NoError(t, err)

				req, err := http.NewRequest("POST", signupPath+"/verify_email", reqBody)
				assert.NoError(t, err)
				req.Header.Set("Cookie", accounts[user]["signup_session_cookie"].(string))
				req.Header.Add("Content-Type", "application/json")

				res, err := http.DefaultClient.Do(req)
				assert.NoError(t, err)

				if !assert.Equal(t, http.StatusOK, res.StatusCode) {
					bd, err := resBody(res.Body)
					assert.NoError(t, err)
					t.Logf("%s", bd)
				}
			})

			t.Run("step three: register user", func(t *testing.T) {
				reqBody, err := reqBody(map[string]any{
					"username":    info["username"],
					"password":    info["password"],
					"phone":       info["phone"],
					"geolocation": info["geolocation"],
				})
				assert.NoError(t, err)

				req, err := http.NewRequest("POST", signupPath+"/register_user", reqBody)
				assert.NoError(t, err)
				req.Header.Add("Content-Type", "application/json")
				req.Header.Set("Cookie", accounts[user]["signup_session_cookie"].(string))

				res, err := http.DefaultClient.Do(req)
				assert.NoError(t, err)

				if !assert.Equal(t, http.StatusOK, res.StatusCode) {
					bd, err := resBody(res.Body)
					assert.NoError(t, err)
					t.Logf("%s", bd)
				}

				accounts[user]["user_session_cookie"] = res.Header.Get("Set-Cookie")

				delete(accounts[user], "signup_session_cookie")
			})
		}
	})

	t.Run("change user1's phone number", func(t *testing.T) {
		reqBody, err := reqBody(map[string]any{"newPhoneNumber": "07083249523"})
		assert.NoError(t, err)

		req, err := http.NewRequest("POST", userPath+"/change_phone_number", reqBody)
		assert.NoError(t, err)
		req.Header.Set("Cookie", accounts["user1"]["user_session_cookie"].(string))
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			bd, err := resBody(res.Body)
			assert.NoError(t, err)
			t.Logf("%s", bd)
		}
	})

	t.Run("confirm user1's updated profile", func(t *testing.T) {
		req, err := http.NewRequest("GET", userPath+"/my_profile", nil)
		assert.NoError(t, err)
		req.Header.Set("Cookie", accounts["user1"]["user_session_cookie"].(string))

		res, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, res.StatusCode)

		bd, err := resBody(res.Body)
		assert.NoError(t, err)
		assert.NotEmpty(t, bd)

		var data map[string]any

		assert.NoError(t, json.Unmarshal(bd, &data))

		accounts["user1"] = data

		assert.Equal(t, accounts["user1"]["phone"], "07083249523")
	})

	t.Run("change user2's geolocation", func(t *testing.T) {
		reqBody, err := reqBody(map[string]any{"newGeolocation": map[string]any{"longitude": 3.0, "latitude": 3.0}})
		assert.NoError(t, err)

		req, err := http.NewRequest("POST", userPath+"/update_geolocation", reqBody)
		assert.NoError(t, err)
		req.Header.Set("Cookie", accounts["user2"]["user_session_cookie"].(string))
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			bd, err := resBody(res.Body)
			assert.NoError(t, err)
			t.Logf("%s", bd)
		}
	})

	t.Run("confirm user2's updated profile", func(t *testing.T) {
		username := accounts["user2"]["username"].(string)

		req, err := http.NewRequest("GET", userPath+"/find_user?emailUsernamePhone="+username, nil)
		assert.NoError(t, err)
		req.Header.Set("Cookie", accounts["user2"]["user_session_cookie"].(string))

		res, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, res.StatusCode)

		bd, err := resBody(res.Body)
		assert.NoError(t, err)
		assert.NotEmpty(t, bd)

		var data map[string]any

		assert.NoError(t, json.Unmarshal(bd, &data))

		accounts["user2"] = data

		assert.Equal(t, accounts["user2"]["geolocation"], map[string]any{"longitude": 3.0, "latitude": 3.0})
	})

	// cleanUpDB()
}
