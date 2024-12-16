package appServices

import (
	"context"
	"fmt"
	"i9chat/helpers"
	"i9chat/services/cloudStorageService"
	"time"
)

func uploadVoice(ctx context.Context, userId int, props map[string]any) (map[string]any, error) {
	var vd struct {
		Data []byte
	}

	helpers.MapToStruct(props, &vd)

	voiceUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("voice_messages/user-%d/voice-%d.ogg", userId, time.Now().UnixNano()), vd.Data)

	if err != nil {
		return nil, err
	}

	delete(props, "data")

	props["url"] = voiceUrl

	return map[string]any{"type": "voice", "props": props}, nil
}

func uploadAudio(ctx context.Context, userId int, props map[string]any) (map[string]any, error) {
	var ad struct {
		Data []byte
	}

	helpers.MapToStruct(props, &ad)

	audioUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("audio_messages/user-%d/aud-%d.mp3", userId, time.Now().UnixNano()), ad.Data)
	if err != nil {
		return nil, err
	}

	delete(props, "data")

	props["url"] = audioUrl

	return map[string]any{"type": "audio", "props": props}, nil
}

func uploadVideo(ctx context.Context, userId int, props map[string]any) (map[string]any, error) {
	var vd struct {
		Data []byte
	}

	helpers.MapToStruct(props, &vd)

	videoUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("video_messages/user-%d/vid-%d.mp4", userId, time.Now().UnixNano()), vd.Data)
	if err != nil {
		return nil, err
	}

	delete(props, "data")

	props["url"] = videoUrl

	return map[string]any{"type": "video", "props": props}, nil
}

func uploadImage(ctx context.Context, userId int, props map[string]any) (map[string]any, error) {
	var id struct {
		Data []byte
	}

	helpers.MapToStruct(props, &id)

	imageUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("image_messages/user-%d/img-%d.jpg", userId, time.Now().UnixNano()), id.Data)
	if err != nil {
		return nil, err
	}

	delete(props, "data")

	props["url"] = imageUrl

	return map[string]any{"type": "image", "props": props}, nil
}

func uploadFile(ctx context.Context, userId int, props map[string]any) (map[string]any, error) {
	var fd struct {
		Data []byte
	}

	helpers.MapToStruct(props, &fd)

	fileUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("file_messages/user-%d/img-%d.jpg", userId, time.Now().UnixNano()), fd.Data)
	if err != nil {
		return nil, err
	}

	delete(props, "data")

	props["url"] = fileUrl

	return map[string]any{"type": "image", "props": props}, nil
}

func UploadMessageMedia(ctx context.Context, userId int, msgBody map[string]any) (map[string]any, error) {
	var msg struct {
		Type  string
		Props map[string]any
	}

	helpers.MapToStruct(msgBody, &msg)

	switch msg.Type {
	case "voice":
		return uploadVoice(ctx, userId, msg.Props)
	case "audio":
		return uploadAudio(ctx, userId, msg.Props)
	case "video":
		return uploadVideo(ctx, userId, msg.Props)
	case "image":
		return uploadImage(ctx, userId, msg.Props)
	case "file":
		return uploadFile(ctx, userId, msg.Props)
	default:
		return msgBody, nil
	}
}
