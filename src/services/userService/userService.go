package userService

import (
	"context"
	"fmt"
	"i9chat/src/appTypes"
	"i9chat/src/appTypes/UITypes"
	"i9chat/src/helpers"
	user "i9chat/src/models/userModel"
	"i9chat/src/services/cloudStorageService"
	"i9chat/src/services/eventStreamService"
	"i9chat/src/services/eventStreamService/eventTypes"
	"i9chat/src/services/realtimeService"
	"time"

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

func UserExists(ctx context.Context, emailOrUsername string) (bool, error) {
	return user.Exists(ctx, emailOrUsername)
}

func NewUser(ctx context.Context, email, username, password, bio string) (user.NewUserT, error) {
	newUser, err := user.New(ctx, email, username, password, bio)
	if err != nil {
		return user.NewUserT{}, nil
	}

	if newUser.Email != "" {
		go eventStreamService.QueueNewUserEvent(eventTypes.NewUserEvent{
			Username: newUser.Username,
			UserData: helpers.ToJson(newUser),
		})
	}

	return newUser, nil
}

func SigninUserFind(ctx context.Context, uniqueIdent string) (user.SignedInUserT, error) {
	return user.SigninFind(ctx, uniqueIdent)
}

func ChangeUserPassword(ctx context.Context, email, newPassword string) (bool, error) {
	username, err := user.ChangePassword(ctx, email, newPassword)
	if err != nil {
		return false, err
	}

	done := username != ""

	if done {
		go eventStreamService.QueueEditUserEvent(eventTypes.EditUserEvent{
			Username:    username,
			UpdateKVMap: map[string]any{"password": newPassword},
		})
	}

	return done, nil
}

type AuthPPicDataT struct {
	UploadUrl     string `json:"uploadUrl"`
	PPicCloudName string `json:"profilePicCloudName"`
}

func AuthorizePPicUpload(ctx context.Context, picMIME string) (AuthPPicDataT, error) {
	var res AuthPPicDataT

	for small0_medium1_large2 := range 3 {

		which := [3]string{"small", "medium", "large"}

		pPicCloudName := fmt.Sprintf("uploads/user/profile_pics/%d%d/%s-%s", time.Now().Year(), time.Now().Month(), uuid.NewString(), which[small0_medium1_large2])

		url, err := cloudStorageService.GetUploadUrl(pPicCloudName, picMIME)
		if err != nil {
			return res, fiber.ErrInternalServerError
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

func ChangeProfilePicture(ctx context.Context, clientUsername, profilePicCloudName string) (bool, error) {
	done, err := user.ChangeProfilePicture(ctx, clientUsername, profilePicCloudName)
	if err != nil {
		return false, err
	}

	if done {
		go eventStreamService.QueueEditUserEvent(eventTypes.EditUserEvent{
			Username:    clientUsername,
			UpdateKVMap: map[string]any{"profile_pic_url": profilePicCloudName},
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
