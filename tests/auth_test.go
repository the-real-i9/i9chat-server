// User-story-based testing for server applications
package tests

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const signupPath = HOST_URL + "/api/auth/signup"
const signinPath = HOST_URL + "/api/auth/signin"
const signoutPath = userPath + "/signout"

func TestUserAuth(t *testing.T) {
	t.Parallel()

	t.Run("User-A Scenario", func(t *testing.T) {

		sessionCookie := ""

		t.Run("requests a new account", func(t *testing.T) {
			reqBody, err := reqBody(map[string]any{"email": "suberu@gmail.com"})
			require.NoError(t, err)

			res, err := http.Post(signupPath+"/request_new_account", "application/json", reqBody)
			require.NoError(t, err)

			require.Equal(t, http.StatusOK, res.StatusCode)
			bd, err := resBody(res.Body)
			require.NoError(t, err)
			require.NotEmpty(t, bd)

			sessionCookie = res.Header.Get("Set-Cookie")
		})

		t.Run("sends an incorrect email verf code", func(t *testing.T) {
			reqBody, err := reqBody(map[string]any{"code": "000111"})
			require.NoError(t, err)

			req, err := http.NewRequest("POST", signupPath+"/verify_email", reqBody)
			require.NoError(t, err)
			req.Header.Set("Cookie", sessionCookie)
			req.Header.Add("Content-Type", "application/json")

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			require.Equal(t, http.StatusBadRequest, res.StatusCode)
		})

		t.Run("sends the correct email verf code", func(t *testing.T) {
			reqBody, err := reqBody(map[string]any{"code": os.Getenv("DUMMY_VERF_TOKEN")})
			require.NoError(t, err)

			req, err := http.NewRequest("POST", signupPath+"/verify_email", reqBody)
			require.NoError(t, err)
			req.Header.Set("Cookie", sessionCookie)
			req.Header.Add("Content-Type", "application/json")

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			require.Equal(t, http.StatusOK, res.StatusCode)
			bd, err := resBody(res.Body)
			require.NoError(t, err)
			require.NotEmpty(t, bd)

			sessionCookie = res.Header.Get("Set-Cookie")
		})

		t.Run("submits her remaining credentials", func(t *testing.T) {
			reqBody, err := reqBody(map[string]any{
				"username": "suberu",
				"password": "sketeppy",
				"phone":    "08283443588",
				"geolocation": map[string]any{
					"x": 5.0,
					"y": 2.0,
				},
			})
			require.NoError(t, err)

			req, err := http.NewRequest("POST", signupPath+"/register_user", reqBody)
			require.NoError(t, err)
			req.Header.Add("Content-Type", "application/json")
			req.Header.Set("Cookie", sessionCookie)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			require.Equal(t, http.StatusOK, res.StatusCode)
			bd, err := resBody(res.Body)
			require.NoError(t, err)
			require.NotEmpty(t, bd)

			sessionCookie = res.Header.Get("Set-Cookie")
		})

		t.Run("user signs out", func(t *testing.T) {
			req, err := http.NewRequest("GET", signoutPath, nil)
			require.NoError(t, err)
			req.Header.Set("Cookie", sessionCookie)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			require.Equal(t, http.StatusOK, res.StatusCode)
			bd, err := resBody(res.Body)
			require.NoError(t, err)
			require.NotEmpty(t, bd)
		})

		t.Run("signs in with incorrect credentials", func(t *testing.T) {
			reqBody, err := reqBody(map[string]any{
				"emailOrUsername": "suberu@gmail.com",
				"password":        "millini",
			})
			require.NoError(t, err)

			res, err := http.Post(signinPath, "application/json", reqBody)
			require.NoError(t, err)
			require.Equal(t, http.StatusNotFound, res.StatusCode)
		})

		t.Run("signs in with correct credentials", func(t *testing.T) {
			reqBody, err := reqBody(map[string]any{
				"emailOrUsername": "suberu",
				"password":        "sketeppy",
			})
			require.NoError(t, err)

			res, err := http.Post(signinPath, "application/json", reqBody)
			require.NoError(t, err)

			require.Equal(t, http.StatusOK, res.StatusCode)
			bd, err := resBody(res.Body)
			require.NoError(t, err)
			require.NotEmpty(t, bd)

		})
	})

	t.Run("User-B Scenario", func(t *testing.T) {
		t.Run("requests an account with an already existing email", func(t *testing.T) {
			reqBody, err := reqBody(map[string]any{"email": "suberu@gmail.com"})
			require.NoError(t, err)

			res, err := http.Post(signupPath+"/request_new_account", "application/json", reqBody)
			require.NoError(t, err)
			require.Equal(t, http.StatusBadRequest, res.StatusCode)
		})
	})
}
