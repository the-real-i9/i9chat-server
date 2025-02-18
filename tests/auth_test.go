// User-story-based testing for server applications
package tests

import (
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

const signupPath = HOST_URL + "/api/auth/signup"
const signinPath = HOST_URL + "/api/auth/signin"

func TestUserAuth(t *testing.T) {
	t.Run("User-A Scenario", func(t *testing.T) {

		signupSessionCookie := ""

		t.Run("requests a new account", func(t *testing.T) {
			reqBody, err := reqBody(map[string]any{"email": "suberu@gmail.com"})
			assert.NoError(t, err)

			res, err := http.Post(signupPath+"/request_new_account", "application/json", reqBody)
			assert.NoError(t, err)

			if !assert.Equal(t, http.StatusOK, res.StatusCode) {
				bd, err := resBody(res.Body)
				assert.NoError(t, err)
				t.Log(bd)
			}

			signupSessionCookie = res.Header.Get("Set-Cookie")
		})

		t.Run("sends an incorrect email verf code", func(t *testing.T) {
			verfCode, err := strconv.Atoi(os.Getenv("DUMMY_VERF_TOKEN"))
			assert.NoError(t, err)

			reqBody, err := reqBody(map[string]any{"code": verfCode + 1})
			assert.NoError(t, err)

			req, err := http.NewRequest("POST", signupPath+"/verify_email", reqBody)
			assert.NoError(t, err)
			req.Header.Set("Cookie", signupSessionCookie)
			req.Header.Add("Content-Type", "application/json")

			res, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		})

		t.Run("sends the correct email verf code", func(t *testing.T) {
			verfCode, err := strconv.Atoi(os.Getenv("DUMMY_VERF_TOKEN"))
			assert.NoError(t, err)

			reqBody, err := reqBody(map[string]any{"code": verfCode})
			assert.NoError(t, err)

			req, err := http.NewRequest("POST", signupPath+"/verify_email", reqBody)
			assert.NoError(t, err)
			req.Header.Set("Cookie", signupSessionCookie)
			req.Header.Add("Content-Type", "application/json")

			res, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			if !assert.Equal(t, http.StatusOK, res.StatusCode) {
				bd, err := resBody(res.Body)
				assert.NoError(t, err)
				t.Log(bd)
			}
		})

		t.Run("submits her remaining credentials", func(t *testing.T) {
			reqBody, err := reqBody(map[string]any{
				"username": "suberu",
				"password": "sketeppy",
				"phone":    "08283443588",
				"geolocation": map[string]any{
					"longitude": 5.0,
					"latitude":  2.0,
				},
			})
			assert.NoError(t, err)

			req, err := http.NewRequest("POST", signupPath+"/register_user", reqBody)
			assert.NoError(t, err)
			req.Header.Add("Content-Type", "application/json")
			req.Header.Set("Cookie", signupSessionCookie)

			res, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			if !assert.Equal(t, http.StatusOK, res.StatusCode) {
				bd, err := resBody(res.Body)
				assert.NoError(t, err)
				t.Log(bd)
			}
		})

		t.Run("signs in with incorrect credentials", func(t *testing.T) {
			reqBody, err := reqBody(map[string]any{
				"emailOrUsername": "suberu@gmail.com",
				"password":        "millini",
			})
			assert.NoError(t, err)

			res, err := http.Post(signinPath, "application/json", reqBody)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusNotFound, res.StatusCode)
		})

		t.Run("signs in with correct credentials", func(t *testing.T) {
			reqBody, err := reqBody(map[string]any{
				"emailOrUsername": "suberu",
				"password":        "sketeppy",
			})
			assert.NoError(t, err)

			res, err := http.Post(signinPath, "application/json", reqBody)
			assert.NoError(t, err)

			if !assert.Equal(t, http.StatusOK, res.StatusCode) {
				bd, err := resBody(res.Body)
				assert.NoError(t, err)
				t.Log(bd)
			}

		})
	})

	t.Run("User-B Scenario", func(t *testing.T) {
		t.Run("requests an account with an already existing email", func(t *testing.T) {
			reqBody, err := reqBody(map[string]any{"email": "suberu@gmail.com"})
			assert.NoError(t, err)

			res, err := http.Post(signupPath+"/request_new_account", "application/json", reqBody)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		})
	})
}
