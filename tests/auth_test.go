// User-story-based testing for server applications
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

func TestUserAuth(t *testing.T) {
	t.Parallel()

	user1 := UserT{
		Email:    "suberu@gmail.com",
		Username: "suberu",
		Password: "sketeppy",
		Geolocation: UserGeolocation{
			X: 5.0,
			Y: 2.0,
		},
	}

	{
		t.Log("Action: user1 requests a new account")

		reqBody, err := makeReqBody(map[string]any{"email": user1.Email})
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
			"msg": "A 6-digit verification code has been sent to " + user1.Email,
		}, nil))

		user1.SessionCookie = res.Header.Get("Set-Cookie")
	}

	{
		t.Log("Action: user1 sends an incorrect email verf code")

		reqBody, err := makeReqBody(map[string]any{"code": "000111"})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", signupPath+"/verify_email", reqBody)
		require.NoError(t, err)
		req.Header.Set("Cookie", user1.SessionCookie)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusBadRequest, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := errResBody(res.Body)
		require.NoError(t, err)

		require.Equal(t, "Incorrect verification code! Check or Re-submit your email.", rb)
	}

	{
		t.Log("Action: user1 sends the correct email verf code")

		reqBody, err := makeReqBody(map[string]any{"code": os.Getenv("DUMMY_TOKEN")})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", signupPath+"/verify_email", reqBody)
		require.NoError(t, err)
		req.Header.Set("Cookie", user1.SessionCookie)
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
			"msg": fmt.Sprintf("Your email '%s' has been verified!", user1.Email),
		}, nil))

		user1.SessionCookie = res.Header.Get("Set-Cookie")
	}

	{
		t.Log("Action: user1 submits her remaining credentials")

		reqBody, err := makeReqBody(map[string]any{
			"username": user1.Username,
			"password": user1.Password,
			"geolocation": map[string]any{
				"x": user1.Geolocation.X,
				"y": user1.Geolocation.Y,
			},
		})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", signupPath+"/register_user", reqBody)
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

		rb, err := succResBody[map[string]any](res.Body)
		require.NoError(t, err)

		td.Cmp(td.Require(t), rb, td.SuperMapOf(map[string]any{
			"msg":  "Signup success!",
			"user": td.Ignore(),
		}, nil))

		user1.SessionCookie = res.Header.Get("Set-Cookie")
	}

	{
		t.Log("Action: user1 signs out")

		req, err := http.NewRequest("GET", signoutPath, nil)
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

		rb, err := succResBody[string](res.Body)
		require.NoError(t, err)

		require.Equal(t, "You've been logged out!", rb)
	}

	{
		t.Log("Action: user1 signs in with incorrect credentials")

		reqBody, err := makeReqBody(map[string]any{
			"emailOrUsername": user1.Email,
			"password":        "millini_x",
		})
		require.NoError(t, err)

		res, err := http.Post(signinPath, "application/json", reqBody)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusNotFound, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := errResBody(res.Body)
		require.NoError(t, err)

		require.Equal(t, "Incorrect email/username or password", rb)
	}

	{
		t.Log("Action: user1 signs in with correct credentials")

		reqBody, err := makeReqBody(map[string]any{
			"emailOrUsername": user1.Username,
			"password":        user1.Password,
		})
		require.NoError(t, err)

		res, err := http.Post(signinPath, "application/json", reqBody)
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
			"msg":  "Signin success!",
			"user": td.Ignore(),
		}, nil))
	}

	{
		t.Log("Action: userX requests an account with an already existing email")

		reqBody, err := makeReqBody(map[string]any{"email": user1.Email})
		require.NoError(t, err)

		res, err := http.Post(signupPath+"/request_new_account", "application/json", reqBody)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusBadRequest, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := errResBody(res.Body)
		require.NoError(t, err)

		require.Equal(t, fmt.Sprintf("An account with %s already exists", user1.Email), rb)
	}
}
