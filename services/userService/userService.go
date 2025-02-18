package userService

import (
	"context"
	"fmt"
	"i9chat/appTypes"
	user "i9chat/models/userModel"
	"i9chat/services/cloudStorageService"
	"i9chat/services/messageBrokerService"
	"time"

	"github.com/gabriel-vasile/mimetype"
)

func GoOnline(ctx context.Context, clientUsername string) {
	go func() {
		dmPartners := user.ChangePresence(ctx, clientUsername, "online", time.Time{})

		for _, dmp := range dmPartners {
			messageBrokerService.Send(fmt.Sprintf("user-%s-topic", dmp), messageBrokerService.Message{
				Event: "user online",
				Data: map[string]any{
					"user": clientUsername,
				},
			})
		}
	}()

}

func GoOffline(ctx context.Context, clientUsername string, lastSeen time.Time) {
	go func() {
		dmPartners := user.ChangePresence(ctx, clientUsername, "offline", lastSeen)

		for _, dmp := range dmPartners {
			messageBrokerService.Send(fmt.Sprintf("user-%s-topic", dmp), messageBrokerService.Message{
				Event: "user offline",
				Data: map[string]any{
					"user":      clientUsername,
					"last_seen": lastSeen,
				},
			})
		}
	}()
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

	respData := map[string]any{
		"msg": "Operation Successful",
	}

	return respData, nil
}

func ChangePhone(ctx context.Context, clientUsername string, newPhone string) (any, error) {
	err := user.ChangePhone(ctx, clientUsername, newPhone)
	if err != nil {
		return nil, err
	}

	respData := map[string]any{
		"msg": "Operation Successful",
	}

	return respData, nil
}

func UpdateMyLocation(ctx context.Context, clientUsername string, newGeolocation appTypes.UserGeolocation) (any, error) {
	err := user.UpdateLocation(ctx, clientUsername, newGeolocation)
	if err != nil {
		return nil, err
	}

	respData := map[string]any{
		"msg": "Operation Successful",
	}

	return respData, err
}

func SearchUser(ctx context.Context, clientUsername, emailUsernamePhone string) ([]any, error) {
	return user.Search(ctx, clientUsername, emailUsernamePhone)
}

func FindNearbyUsers(ctx context.Context, clientUsername string, long, lat, radius float64) ([]any, error) {
	return user.FindNearby(ctx, clientUsername, long, lat, radius)
}

func GetMyChats(ctx context.Context, clientUsername string) ([]user.ChatItem, error) {
	return user.GetChats(ctx, clientUsername)
}
