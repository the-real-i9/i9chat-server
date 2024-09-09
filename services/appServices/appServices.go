package appServices

import (
	"fmt"
	"i9chat/helpers"
	"log"
	"os"
	"time"

	"gopkg.in/gomail.v2"
)

func SendMail(email string, subject string, body string) {

	user := os.Getenv("MAILING_EMAIL")
	pass := os.Getenv("MAILING_PASSWORD")

	m := gomail.NewMessage()
	m.SetHeader("From", user)
	m.SetHeader("To", email)
	m.SetHeader("Subject", fmt.Sprintf("i9chat - %s", subject))
	m.SetBody("text/html", body)

	d := gomail.NewDialer("smtp.gmail.com", 465, user, pass)
	// d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		log.Println(err)
		return
	}
}

func handleVoiceMsg(userId int, props map[string]any) map[string]any {
	var vd struct {
		Data []byte
	}

	helpers.MapToStruct(props, &vd)

	voiceUrl, _ := helpers.UploadFile(fmt.Sprintf("voice messages/user-%d/voice-%d.ogg", userId, time.Now().UnixNano()), vd.Data)

	delete(props, "data")

	props["url"] = voiceUrl

	return map[string]any{"type": "voice", "props": props}
}

func handleAudioMsg(userId int, props map[string]any) map[string]any {
	var ad struct {
		Data []byte
	}

	helpers.MapToStruct(props, &ad)

	audioUrl, _ := helpers.UploadFile(fmt.Sprintf("audio messages/user-%d/aud-%d.mp3", userId, time.Now().UnixNano()), ad.Data)

	delete(props, "data")

	props["url"] = audioUrl

	return map[string]any{"type": "audio", "props": props}
}

func handleVideoMsg(userId int, props map[string]any) map[string]any {
	var vd struct {
		Data []byte
	}

	helpers.MapToStruct(props, &vd)

	videoUrl, _ := helpers.UploadFile(fmt.Sprintf("video messages/user-%d/vid-%d.mp4", userId, time.Now().UnixNano()), vd.Data)

	delete(props, "data")

	props["url"] = videoUrl

	return map[string]any{"type": "video", "props": props}
}

func handleImageMsg(userId int, props map[string]any) map[string]any {
	var id struct {
		Data []byte
	}

	helpers.MapToStruct(props, &id)

	imageUrl, _ := helpers.UploadFile(fmt.Sprintf("image messages/user-%d/img-%d.jpg", userId, time.Now().UnixNano()), id.Data)

	delete(props, "data")

	props["url"] = imageUrl

	return map[string]any{"type": "image", "props": props}
}

func handleFileMsg(userId int, props map[string]any) map[string]any {
	var fd struct {
		Data []byte
	}

	helpers.MapToStruct(props, &fd)

	fileUrl, _ := helpers.UploadFile(fmt.Sprintf("file messages/user-%d/img-%d.jpg", userId, time.Now().UnixNano()), fd.Data)

	delete(props, "data")

	props["url"] = fileUrl

	return map[string]any{"type": "image", "props": props}
}

func MessageBinaryToUrl(userId int, msgBody map[string]any) map[string]any {
	var msg struct {
		Type  string
		Props map[string]any
	}

	helpers.MapToStruct(msgBody, &msg)

	switch msg.Type {
	case "voice":
		return handleVoiceMsg(userId, msg.Props)
	case "audio":
		return handleAudioMsg(userId, msg.Props)
	case "video":
		return handleVideoMsg(userId, msg.Props)
	case "image":
		return handleImageMsg(userId, msg.Props)
	case "file":
		return handleFileMsg(userId, msg.Props)
	default:
		return msgBody
	}
}
