package appUtilServices

import (
	"fmt"
	"i9chat/helpers"
	"i9chat/services/cloudStorageService"
	"time"
)

func uploadVoice(userId int, props map[string]any) map[string]any {
	var vd struct {
		Data []byte
	}

	helpers.MapToStruct(props, &vd)

	voiceUrl, _ := cloudStorageService.UploadFile(fmt.Sprintf("voice_messages/user-%d/voice-%d.ogg", userId, time.Now().UnixNano()), vd.Data)

	delete(props, "data")

	props["url"] = voiceUrl

	return map[string]any{"type": "voice", "props": props}
}

func uploadAudio(userId int, props map[string]any) map[string]any {
	var ad struct {
		Data []byte
	}

	helpers.MapToStruct(props, &ad)

	audioUrl, _ := cloudStorageService.UploadFile(fmt.Sprintf("audio_messages/user-%d/aud-%d.mp3", userId, time.Now().UnixNano()), ad.Data)

	delete(props, "data")

	props["url"] = audioUrl

	return map[string]any{"type": "audio", "props": props}
}

func uploadVideo(userId int, props map[string]any) map[string]any {
	var vd struct {
		Data []byte
	}

	helpers.MapToStruct(props, &vd)

	videoUrl, _ := cloudStorageService.UploadFile(fmt.Sprintf("video_messages/user-%d/vid-%d.mp4", userId, time.Now().UnixNano()), vd.Data)

	delete(props, "data")

	props["url"] = videoUrl

	return map[string]any{"type": "video", "props": props}
}

func uploadImage(userId int, props map[string]any) map[string]any {
	var id struct {
		Data []byte
	}

	helpers.MapToStruct(props, &id)

	imageUrl, _ := cloudStorageService.UploadFile(fmt.Sprintf("image_messages/user-%d/img-%d.jpg", userId, time.Now().UnixNano()), id.Data)

	delete(props, "data")

	props["url"] = imageUrl

	return map[string]any{"type": "image", "props": props}
}

func uploadFile(userId int, props map[string]any) map[string]any {
	var fd struct {
		Data []byte
	}

	helpers.MapToStruct(props, &fd)

	fileUrl, _ := cloudStorageService.UploadFile(fmt.Sprintf("file_messages/user-%d/img-%d.jpg", userId, time.Now().UnixNano()), fd.Data)

	delete(props, "data")

	props["url"] = fileUrl

	return map[string]any{"type": "image", "props": props}
}

func UploadMessageMedia(userId int, msgBody map[string]any) map[string]any {
	var msg struct {
		Type  string
		Props map[string]any
	}

	helpers.MapToStruct(msgBody, &msg)

	switch msg.Type {
	case "voice":
		return uploadVoice(userId, msg.Props)
	case "audio":
		return uploadAudio(userId, msg.Props)
	case "video":
		return uploadVideo(userId, msg.Props)
	case "image":
		return uploadImage(userId, msg.Props)
	case "file":
		return uploadFile(userId, msg.Props)
	default:
		return msgBody
	}
}
