package userService

import (
	"context"
	"fmt"
	"i9chat/src/appTypes"
	"i9chat/src/appTypes/UITypes"
	user "i9chat/src/models/userModel"
	"i9chat/src/services/cloudStorageService"
	"i9chat/src/services/eventStreamService"
	"i9chat/src/services/eventStreamService/eventTypes"
	"i9chat/src/services/realtimeService"
	"i9chat/src/services/securityServices"
	"os"
	"time"

	"github.com/gabriel-vasile/mimetype"
)

func GoOnline(ctx context.Context, clientUsername string) {
	done := user.ChangePresence(ctx, clientUsername, "online", 0)

	if done {
		realtimeService.PublishUserPresenceChange(ctx, clientUsername, map[string]any{
			"user":     clientUsername,
			"presence": "online",
		})

		go eventStreamService.QueueUserPresenceChangeEvent(eventTypes.UserPresenceChangeEvent{
			Username: clientUsername,
			Presence: "online",
			LastSeen: 0,
		})
	}

}

func GoOffline(ctx context.Context, clientUsername string) {
	lastSeen := time.Now().UTC().UnixMilli()

	done := user.ChangePresence(ctx, clientUsername, "offline", lastSeen)

	if done {
		realtimeService.PublishUserPresenceChange(ctx, clientUsername, map[string]any{
			"user":      clientUsername,
			"presence":  "offline",
			"last_seen": lastSeen,
		})

		go eventStreamService.QueueUserPresenceChangeEvent(eventTypes.UserPresenceChangeEvent{
			Username: clientUsername,
			Presence: "offline",
			LastSeen: lastSeen,
		})
	}
}

func ChangeProfilePicture(ctx context.Context, clientUsername string, pictureData []byte) (any, string, error) {
	// if any picture size validation error, do it here

	ext := mimetype.Detect(pictureData).Extension()
	picPath := fmt.Sprintf("profile_pictures/%s/profile_pic_%d.%s", clientUsername, time.Now().UnixNano(), ext)

	newPicUrl, err := cloudStorageService.Upload(ctx, picPath, pictureData)
	if err != nil {
		return nil, "", err
	}

	done, err := user.ChangeProfilePicture(ctx, clientUsername, newPicUrl)
	if err != nil {
		return nil, "", err
	}

	var authJwt string

	if done {
		authJwt, err = securityServices.JwtSign(appTypes.ClientUser{
			Username:      clientUsername,
			ProfilePicUrl: newPicUrl,
		}, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(10*24*time.Hour))
		if err != nil {
			return nil, "", err
		}

		go eventStreamService.QueueEditUserEvent(eventTypes.EditUserEvent{
			Username:    clientUsername,
			UpdateKVMap: map[string]any{"profile_pic_url": newPicUrl},
		})
	}

	return done, authJwt, nil
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

func GetMyChats(ctx context.Context, clientUsername string) (any, error) {
	return user.GetMyChats(ctx, clientUsername)
}

func GetMyProfile(ctx context.Context, clientUsername string) (UITypes.UserProfile, error) {
	return user.GetMyProfile(ctx, clientUsername)
}
