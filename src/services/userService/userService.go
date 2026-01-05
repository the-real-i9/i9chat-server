package userService

import (
	"context"
	"fmt"
	"i9chat/src/appGlobals"
	"i9chat/src/appTypes"
	"i9chat/src/appTypes/UITypes"
	"i9chat/src/helpers"
	user "i9chat/src/models/userModel"
	"i9chat/src/services/eventStreamService"
	"i9chat/src/services/eventStreamService/eventTypes"
	"i9chat/src/services/realtimeService"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func GoOnline(ctx context.Context, clientUsername string) {
	done := user.ChangePresence(ctx, clientUsername, "online", 0)

	if done {
		go eventStreamService.QueueUserPresenceChangeEvent(eventTypes.UserPresenceChangeEvent{
			Username: clientUsername,
			Presence: "online",
			LastSeen: 0,
		})

		realtimeService.PublishUserPresenceChange(ctx, clientUsername, map[string]any{
			"user":     clientUsername,
			"presence": "online",
		})
	}

}

type AuthPPicDataT struct {
	UploadUrl     string `json:"uploadUrl"`
	PPicCloudName string `json:"profilePicCloudName"`
}

func AuthorizePPicUpload(ctx context.Context, picMIME string, picSize [3]int64) (AuthPPicDataT, error) {
	var res AuthPPicDataT

	for small0_medium1_large2, size := range picSize {

		which := [3]string{"small", "medium", "large"}

		pPicCloudName := fmt.Sprintf("uploads/user/profile_pics/%d%d/%s-%s", time.Now().Year(), time.Now().Month(), uuid.NewString(), which[small0_medium1_large2])

		url, err := appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).SignedURL(
			pPicCloudName,
			&storage.SignedURLOptions{
				Scheme:      storage.SigningSchemeV4,
				Method:      "PUT",
				ContentType: picMIME,
				Expires:     time.Now().Add(15 * time.Minute),
				Headers:     []string{fmt.Sprintf("x-goog-content-length-range: %d,%[1]d", size)},
			},
		)
		if err != nil {
			helpers.LogError(err)
			return AuthPPicDataT{}, fiber.ErrInternalServerError
		}

		switch small0_medium1_large2 {
		case 0:
			res.UploadUrl += "small:"
			res.PPicCloudName += "small:"
		case 1:
			res.UploadUrl += "medium:"
			res.PPicCloudName += "medium:"
		default:
			res.UploadUrl += "large:"
			res.PPicCloudName += "large:"
		}

		res.UploadUrl += url
		res.PPicCloudName += pPicCloudName

		if small0_medium1_large2 != 2 {
			res.UploadUrl += " "
			res.PPicCloudName += " "
		}
	}

	return res, nil
}

func GoOffline(ctx context.Context, clientUsername string) {
	lastSeen := time.Now().UTC().UnixMilli()

	done := user.ChangePresence(ctx, clientUsername, "offline", lastSeen)

	if done {
		go eventStreamService.QueueUserPresenceChangeEvent(eventTypes.UserPresenceChangeEvent{
			Username: clientUsername,
			Presence: "offline",
			LastSeen: lastSeen,
		})

		realtimeService.PublishUserPresenceChange(ctx, clientUsername, map[string]any{
			"user":      clientUsername,
			"presence":  "offline",
			"last_seen": lastSeen,
		})
	}
}

func ChangeProfilePicture(ctx context.Context, clientUsername, profilePicCloudName string) (any, error) {
	done, err := user.ChangeProfilePicture(ctx, clientUsername, profilePicCloudName)
	if err != nil {
		return nil, err
	}

	if done {
		go eventStreamService.QueueEditUserEvent(eventTypes.EditUserEvent{
			Username:    clientUsername,
			UpdateKVMap: map[string]any{"profile_pic_cloud_name": profilePicCloudName},
		})
	}

	return done, nil
}

func ChangeBio(ctx context.Context, clientUsername, newBio string) (any, error) {
	done, err := user.ChangeBio(ctx, clientUsername, newBio)
	if err != nil {
		return nil, err
	}

	if done {
		go eventStreamService.QueueEditUserEvent(eventTypes.EditUserEvent{
			Username:    clientUsername,
			UpdateKVMap: map[string]any{"bio": newBio},
		})
	}

	return done, nil
}

func SetMyLocation(ctx context.Context, clientUsername string, newGeolocation appTypes.UserGeolocation) (any, error) {
	location, err := user.SetLocation(ctx, clientUsername, newGeolocation)
	if err != nil {
		return nil, err
	}

	done := location != nil

	if done {
		go eventStreamService.QueueEditUserEvent(eventTypes.EditUserEvent{
			Username:    clientUsername,
			UpdateKVMap: map[string]any{"geolocation": location},
		})
	}

	return done, err
}

func FindUser(ctx context.Context, username string) (UITypes.UserSnippet, error) {
	return user.Find(ctx, username)
}

func FindNearbyUsers(ctx context.Context, clientUsername string, x, y, radius float64) ([]UITypes.UserSnippet, error) {
	return user.FindNearby(ctx, clientUsername, x, y, radius)
}

func GetMyChats(ctx context.Context, clientUsername string, limit int, cursor float64) ([]UITypes.ChatSnippet, error) {
	return user.GetMyChats(ctx, clientUsername, limit, cursor)
}

func GetMyProfile(ctx context.Context, clientUsername string) (UITypes.UserProfile, error) {
	return user.GetMyProfile(ctx, clientUsername)
}
