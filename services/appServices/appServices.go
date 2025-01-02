package appServices

import (
	"context"
	"fmt"
	"i9chat/appTypes"
	"i9chat/services/cloudStorageService"
	"time"
)

func uploadVoice(ctx context.Context, userId int, vd *appTypes.MsgProps) error {
	voiceUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("voice_messages/user-%d/voice-%d.ogg", userId, time.Now().UnixNano()), vd.Data)

	if err != nil {
		return err
	}

	vd.Data = nil

	*vd.Url = voiceUrl

	return nil
}

func uploadAudio(ctx context.Context, userId int, ad *appTypes.MsgProps) error {
	audioUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("audio_messages/user-%d/aud-%d.mp3", userId, time.Now().UnixNano()), ad.Data)
	if err != nil {
		return err
	}

	ad.Data = nil

	*ad.Url = audioUrl

	return nil
}

func uploadVideo(ctx context.Context, userId int, vd *appTypes.MsgProps) error {
	videoUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("video_messages/user-%d/vid-%d.mp4", userId, time.Now().UnixNano()), vd.Data)
	if err != nil {
		return err
	}

	vd.Data = nil

	*vd.Url = videoUrl

	return nil
}

func uploadImage(ctx context.Context, userId int, id *appTypes.MsgProps) error {
	imageUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("image_messages/user-%d/img-%d.jpg", userId, time.Now().UnixNano()), id.Data)
	if err != nil {
		return err
	}

	id.Data = nil

	*id.Url = imageUrl

	return nil
}

func uploadFile(ctx context.Context, userId int, fd *appTypes.MsgProps) error {
	fileUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("file_messages/user-%d/img-%d.jpg", userId, time.Now().UnixNano()), fd.Data)
	if err != nil {
		return err
	}

	fd.Data = nil

	*fd.Url = fileUrl

	return nil
}

func UploadMessageMedia(ctx context.Context, userId int, msg *appTypes.MsgContent) error {

	switch msg.Type {
	case "voice":
		return uploadVoice(ctx, userId, msg.MsgProps)
	case "audio":
		return uploadAudio(ctx, userId, msg.MsgProps)
	case "video":
		return uploadVideo(ctx, userId, msg.MsgProps)
	case "image":
		return uploadImage(ctx, userId, msg.MsgProps)
	case "file":
		return uploadFile(ctx, userId, msg.MsgProps)
	default:
		return nil
	}
}
