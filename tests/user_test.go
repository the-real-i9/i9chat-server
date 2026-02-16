package tests

import (
	"fmt"
	"i9chat/src/appGlobals"
	"net/http"
	"os"
	"testing"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserOps(t *testing.T) {
	// t.Parallel()
	require := require.New(t)

	user1 := UserT{
		Email:    "mikeross@gmail.com",
		Username: "mikeross",
		Password: "recheal_zane",
		Geolocation: UserGeolocation{
			X: -0.1276,
			Y: 51.5074,
		},
	}

	user2 := UserT{
		Email:    "harveyspecter@gmail.com",
		Username: "harvey",
		Password: "scottie_",
		Geolocation: UserGeolocation{
			X: 2.3522,
			Y: 48.8566,
		},
	}

	user3 := UserT{
		Email:    "jessicapearson@gmail.com",
		Username: "jessica",
		Password: "jeff_malone",
		Geolocation: UserGeolocation{
			X: 2.3522,
			Y: 48.8566,
		},
	}

	{
		t.Log("Setup: create new accounts for users")

		for _, user := range []*UserT{&user1, &user2, &user3} {

			{
				reqBody, err := makeReqBody(map[string]any{"email": user.Email})
				require.NoError(err)

				res, err := http.Post(signupPath+"/request_new_account", "application/vnd.msgpack", reqBody)
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

				req, err := http.NewRequest("POST", signupPath+"/verify_email", reqBody)
				require.NoError(err)
				req.Header.Set("Cookie", user.SessionCookie)
				req.Header.Add("Content-Type", "application/vnd.msgpack")

				res, err := http.DefaultClient.Do(req)
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

				req, err := http.NewRequest("POST", signupPath+"/register_user", reqBody)
				require.NoError(err)
				req.Header.Add("Content-Type", "application/vnd.msgpack")
				req.Header.Set("Cookie", user.SessionCookie)

				res, err := http.DefaultClient.Do(req)
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
		t.Log("Action: user1 sets her geolocation")

		reqBody, err := makeReqBody(map[string]any{"newGeolocation": map[string]any{"x": user1.Geolocation.X, "y": user1.Geolocation.Y}})
		require.NoError(err)

		req, err := http.NewRequest("POST", userPath+"/set_geolocation", reqBody)
		require.NoError(err)
		req.Header.Set("Cookie", user1.SessionCookie)
		req.Header.Add("Content-Type", "application/vnd.msgpack")

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[bool](res.Body)
		require.NoError(err)
		require.True(rb)
	}

	{
		t.Log("Action: user2 sets her geolocation")

		reqBody, err := makeReqBody(map[string]any{"newGeolocation": map[string]any{"x": user2.Geolocation.X, "y": user2.Geolocation.Y}})
		require.NoError(err)

		req, err := http.NewRequest("POST", userPath+"/set_geolocation", reqBody)
		require.NoError(err)
		req.Header.Set("Cookie", user2.SessionCookie)
		req.Header.Add("Content-Type", "application/vnd.msgpack")

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[bool](res.Body)
		require.NoError(err)
		require.True(rb)
	}

	{
		t.Log("Action: user1 finds users nearby")

		x, y, radius := user1.Geolocation.X, user1.Geolocation.Y, 350000.0

		req, err := http.NewRequest("GET", fmt.Sprintf("%s%s?x=%f&y=%f&radius=%f", userPath, "/find_nearby_users", x, y, radius), nil)
		require.NoError(err)
		req.Header.Set("Cookie", user1.SessionCookie)

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[[]map[string]any](res.Body)
		require.NoError(err)

		td.Cmp(td.Require(t), rb, td.Contains(td.Any(
			td.SuperMapOf(map[string]any{
				"username": user2.Username,
			}, nil),
			td.SuperMapOf(map[string]any{
				"username": user3.Username,
			}, nil),
		)))
	}

	{
		t.Log("Action: user1 changes her bio")

		reqBody, err := makeReqBody(map[string]any{"newBio": "This is my new bio!"})
		require.NoError(err)

		req, err := http.NewRequest("POST", userPath+"/change_bio", reqBody)
		require.NoError(err)
		req.Header.Set("Cookie", user1.SessionCookie)
		req.Header.Add("Content-Type", "application/vnd.msgpack")

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[bool](res.Body)
		require.NoError(err)
		require.True(rb)
	}

	{
		t.Log("Action: user2 changes her profile picture")

		var (
			uploadUrl           string
			profilePicCloudName string
			filePath            = "./test_files/profile_pic.png"
			contentType         = "image/png"
		)

		{
			fileInfo, err := os.Stat(filePath)
			require.NoError(err)

			t.Log("--- Authorize profile picture upload ---")

			reqBody, err := makeReqBody(map[string]any{"pic_mime": contentType, "pic_size": [3]int64{fileInfo.Size(), fileInfo.Size(), fileInfo.Size()}})
			require.NoError(err)

			req, err := http.NewRequest("POST", userPath+"/profile_pic_upload/authorize", reqBody)
			require.NoError(err)
			req.Header.Set("Cookie", user2.SessionCookie)
			req.Header.Add("Content-Type", "application/vnd.msgpack")

			res, err := http.DefaultClient.Do(req)
			require.NoError(err)

			if !assert.Equal(t, http.StatusOK, res.StatusCode) {
				rb, err := errResBody(res.Body)
				require.NoError(err)
				t.Log("unexpected error:", rb)
				return
			}

			rb, err := succResBody[map[string]any](res.Body)
			require.NoError(err)

			td.Cmp(td.Require(t), rb, td.Map(map[string]any{
				"uploadUrl":           td.Ignore(),
				"profilePicCloudName": td.Ignore(),
			}, nil))

			uploadUrl = rb["uploadUrl"].(string)
			profilePicCloudName = rb["profilePicCloudName"].(string)
		}

		{
			t.Log("Upload session started:")

			varUploadUrl := make([]string, 3)
			_, err := fmt.Sscanf(uploadUrl, "small:%s medium:%s large:%s", &varUploadUrl[0], &varUploadUrl[1], &varUploadUrl[2])
			require.NoError(err)

			for i, smlUploadUrl := range varUploadUrl {
				varSize := []string{"small", "medium", "large"}

				t.Logf("Uploading %s profile pic started", varSize[i])

				sessionUrl := startResumableUpload(smlUploadUrl, contentType, t)

				uploadFileInChunks(sessionUrl, filePath, contentType, logProgress, t)

				t.Logf("Uploading %s profile pic complete", varSize[i])
			}

			defer func(ppcn string) {
				varPPicCloudName := make([]string, 3)
				_, err = fmt.Sscanf(ppcn, "small:%s medium:%s large:%s", &varPPicCloudName[0], &varPPicCloudName[1], &varPPicCloudName[2])
				require.NoError(err)

				for _, smlPPicCn := range varPPicCloudName {
					err := appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).Object(smlPPicCn).Delete(t.Context())
					require.NoError(err)
				}
			}(profilePicCloudName)

			t.Log("Upload complete")
		}

		reqBody, err := makeReqBody(map[string]any{"profile_pic_cloud_name": profilePicCloudName})
		require.NoError(err)

		req, err := http.NewRequest("POST", userPath+"/change_profile_picture", reqBody)
		require.NoError(err)
		req.Header.Set("Cookie", user2.SessionCookie)
		req.Header.Add("Content-Type", "application/vnd.msgpack")

		res, err := http.DefaultClient.Do(req)
		require.NoError(err)

		if !assert.Equal(t, http.StatusOK, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(err)
			t.Log("unexpected error:", rb)
			return
		}

		rb, err := succResBody[bool](res.Body)
		require.NoError(err)
		require.True(rb)
	}
}
