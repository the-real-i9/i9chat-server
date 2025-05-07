package appServices

import (
	"context"
	"fmt"
	"i9chat/src/appTypes"
	"i9chat/src/services/cloudStorageService"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gofiber/fiber/v2"
)

func uploadVoice(ctx context.Context, username string, msg *appTypes.MsgContent) error {
	voiceUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("voice_messages/user-%s/voice-%d.ogg", username, time.Now().UnixNano()), msg.Data)

	if err != nil {
		return err
	}

	msg.SetMediaUrl(voiceUrl)

	return nil
}

func uploadAudio(ctx context.Context, username string, msg *appTypes.MsgContent) error {
	audioUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("audio_messages/user-%s/aud-%d.mp3", username, time.Now().UnixNano()), msg.Data)
	if err != nil {
		return err
	}

	msg.SetMediaUrl(audioUrl)

	return nil
}

func uploadVideo(ctx context.Context, username string, msg *appTypes.MsgContent) error {
	videoUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("video_messages/user-%s/vid-%d.mp4", username, time.Now().UnixNano()), msg.Data)
	if err != nil {
		return err
	}

	msg.SetMediaUrl(videoUrl)

	return nil
}

func uploadImage(ctx context.Context, username string, msg *appTypes.MsgContent) error {
	imageUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("image_messages/user-%s/img-%d.jpg", username, time.Now().UnixNano()), msg.Data)
	if err != nil {
		return err
	}

	msg.SetMediaUrl(imageUrl)

	return nil
}

func uploadFile(ctx context.Context, username string, msg *appTypes.MsgContent) error {
	fileUrl, err := cloudStorageService.Upload(ctx, fmt.Sprintf("file_messages/user-%s/img-%d.jpg", username, time.Now().UnixNano()), msg.Data)
	if err != nil {
		return err
	}

	msg.SetMediaUrl(fileUrl)

	return nil
}

func UploadMessageMedia(ctx context.Context, username string, msg *appTypes.MsgContent) error {
	if msg.Data == nil {
		return nil
	}

	mediaMIME := mimetype.Detect(msg.Data)
	mediaType, mediaExt := mediaMIME.String(), mediaMIME.Extension()

	if ((msg.Type == "voice" || msg.Type == "audio") && !strings.HasPrefix(mediaType, "audio")) || (msg.Type == "video" && !strings.HasPrefix(mediaType, "video")) || (msg.Type == "image" && !strings.HasPrefix(mediaType, "image")) {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("invalid media type %s, for the message type %s", mediaType, msg.Type))
	}

	msg.SetMediaMIME(mediaType, mediaExt)
	msg.SetMediaSize(int64(len(msg.Data)))

	switch msg.Type {
	case "voice":
		return uploadVoice(ctx, username, msg)
	case "audio":
		return uploadAudio(ctx, username, msg)
	case "video":
		return uploadVideo(ctx, username, msg)
	case "image":
		return uploadImage(ctx, username, msg)
	case "file":
		return uploadFile(ctx, username, msg)
	default:
		return nil
	}
}
