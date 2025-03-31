package appServices

import (
	"context"
	"fmt"
	"i9chat/src/appTypes"
	"i9chat/src/services/cloudStorageService"
	"time"
)

func uploadVoice(ctx context.Context, username string, vd *appTypes.MsgProps) error {
	voiceUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("voice_messages/user-%s/voice-%d.ogg", username, time.Now().UnixNano()), vd.Data)

	if err != nil {
		return err
	}

	vd.Data = nil

	*vd.Url = voiceUrl

	return nil
}

func uploadAudio(ctx context.Context, username string, ad *appTypes.MsgProps) error {
	audioUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("audio_messages/user-%s/aud-%d.mp3", username, time.Now().UnixNano()), ad.Data)
	if err != nil {
		return err
	}

	ad.Data = nil

	*ad.Url = audioUrl

	return nil
}

func uploadVideo(ctx context.Context, username string, vd *appTypes.MsgProps) error {
	videoUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("video_messages/user-%s/vid-%d.mp4", username, time.Now().UnixNano()), vd.Data)
	if err != nil {
		return err
	}

	vd.Data = nil

	*vd.Url = videoUrl

	return nil
}

func uploadImage(ctx context.Context, username string, id *appTypes.MsgProps) error {
	imageUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("image_messages/user-%s/img-%d.jpg", username, time.Now().UnixNano()), id.Data)
	if err != nil {
		return err
	}

	id.Data = nil

	*id.Url = imageUrl

	return nil
}

func uploadFile(ctx context.Context, username string, fd *appTypes.MsgProps) error {
	fileUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("file_messages/user-%s/img-%d.jpg", username, time.Now().UnixNano()), fd.Data)
	if err != nil {
		return err
	}

	fd.Data = nil

	*fd.Url = fileUrl

	return nil
}

func UploadMessageMedia(ctx context.Context, username string, msg *appTypes.MsgContent) error {

	switch msg.Type {
	case "voice":
		return uploadVoice(ctx, username, msg.MsgProps)
	case "audio":
		return uploadAudio(ctx, username, msg.MsgProps)
	case "video":
		return uploadVideo(ctx, username, msg.MsgProps)
	case "image":
		return uploadImage(ctx, username, msg.MsgProps)
	case "file":
		return uploadFile(ctx, username, msg.MsgProps)
	default:
		return nil
	}
}
