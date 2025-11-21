package userService

import (
	"context"
	"fmt"
	"i9chat/src/appTypes"
	user "i9chat/src/models/userModel"
	"i9chat/src/services/cloudStorageService"
	"i9chat/src/services/realtimeService"
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
	}
}

func ChangeProfilePicture(ctx context.Context, clientUsername string, pictureData []byte) (any, error) {
	// if any picture size validation error, do it here

	ext := mimetype.Detect(pictureData).Extension()
	picPath := fmt.Sprintf("profile_pictures/%s/profile_pic_%d.%s", clientUsername, time.Now().UnixNano(), ext)

	newPicUrl, err := cloudStorageService.Upload(ctx, picPath, pictureData)
	if err != nil {
		return nil, err
	}

	done, err := user.ChangeProfilePicture(ctx, clientUsername, newPicUrl)
	if err != nil {
		return nil, err
	}

	return done, nil
}

func ChangeBio(ctx context.Context, clientUsername, newBio string) (any, error) {
	done, err := user.ChangeBio(ctx, clientUsername, newBio)
	if err != nil {
		return nil, err
	}

	return done, nil
}

func SetMyLocation(ctx context.Context, clientUsername string, newGeolocation appTypes.UserGeolocation) (any, error) {
	done, err := user.SetLocation(ctx, clientUsername, newGeolocation)
	if err != nil {
		return nil, err
	}

	return done, err
}

func FindNearbyUsers(ctx context.Context, clientUsername string, x, y, radius float64) ([]any, error) {
	return user.FindNearby(ctx, clientUsername, x, y, radius)
}

func GetMyChats(ctx context.Context, clientUsername string) (any, error) {
	return user.GetMyChats(ctx, clientUsername)
}

func GetMyProfile(ctx context.Context, clientUsername string) (map[string]any, error) {
	return user.GetMyProfile(ctx, clientUsername)
}
