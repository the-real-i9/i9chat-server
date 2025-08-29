package userService

import (
	"context"
	"fmt"
	"i9chat/src/appTypes"
	user "i9chat/src/models/userModel"
	"i9chat/src/services/cloudStorageService"
	"i9chat/src/services/eventStreamService"
	"time"

	"github.com/gabriel-vasile/mimetype"
)

func GoOnline(ctx context.Context, clientUsername string) {
	dmPartners, err := user.ChangePresence(ctx, clientUsername, "online", time.Time{}.UTC())
	if err != nil {
		return
	}

	// go func(dmPartners []any) {
	for _, dmp := range dmPartners {
		eventStreamService.Send(dmp.(string), appTypes.ServerWSMsg{
			Event: "user online",
			Data: map[string]any{
				"user": clientUsername,
			},
		})
	}
	// }(dmPartners)

}

func GoOffline(ctx context.Context, clientUsername string) {
	lastSeen := time.Now().UTC()

	dmPartners, err := user.ChangePresence(ctx, clientUsername, "offline", lastSeen)
	if err != nil {
		return
	}

	// go func(dmPartners []any) {
	for _, dmp := range dmPartners {
		eventStreamService.Send(dmp.(string), appTypes.ServerWSMsg{
			Event: "user offline",
			Data: map[string]any{
				"user":      clientUsername,
				"last_seen": lastSeen.UnixMilli(),
			},
		})
	}
	// }(dmPartners)
}

func SessionUser(ctx context.Context, clientUsername string) (any, error) {
	user, err2 := user.SessionFind(ctx, clientUsername)
	if err2 != nil {
		return nil, err2
	}

	return user, nil
}

func ChangeProfilePicture(ctx context.Context, clientUsername string, pictureData []byte) (any, error) {
	// if any picture size validation error, do it here

	ext := mimetype.Detect(pictureData).Extension()
	picPath := fmt.Sprintf("profile_pictures/%s/profile_pic_%d.%s", clientUsername, time.Now().UnixNano(), ext)

	newPicUrl, err := cloudStorageService.Upload(ctx, picPath, pictureData)
	if err != nil {
		return nil, err
	}

	err2 := user.ChangeProfilePicture(ctx, clientUsername, newPicUrl)
	if err2 != nil {
		return nil, err2
	}

	return true, nil
}

func SetMyLocation(ctx context.Context, clientUsername string, newGeolocation appTypes.UserGeolocation) (any, error) {
	err := user.SetLocation(ctx, clientUsername, newGeolocation)
	if err != nil {
		return nil, err
	}

	return true, err
}

func FindUser(ctx context.Context, emailUsername string) (map[string]any, error) {
	return user.FindOne(ctx, emailUsername)
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
