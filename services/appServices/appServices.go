package appservices

import (
	"crypto/tls"
	"fmt"
	"log"
	"model/appmodel"
	"os"
	"time"
	"utils/helpers"

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
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		log.Println(err)
	}
}

func EndSignupSession(sessionId string) {
	appmodel.EndSignupSession(sessionId)
}

func handleVoiceMsg(userId int, content map[string]any) map[string]any {
	var vo struct {
		Voice []byte
	}

	helpers.ParseToStruct(content, &vo)

	voiceUrl, _ := helpers.UploadFile(fmt.Sprintf("voice messages/user-%d/voice-%d.ogg", userId, time.Now().UnixMilli()), vo.Voice)

	content["voice"] = voiceUrl

	return map[string]any{"type": "voice", "content": content}
}

func handleAudioMsg(userId int, content map[string]any) map[string]any {
	var audd struct {
		Audios [][]byte
	}

	helpers.ParseToStruct(content, &audd)

	for i, aud := range audd.Audios {
		audioUrl, _ := helpers.UploadFile(fmt.Sprintf("audios/user-%d/aud-%d.mp3", userId, time.Now().UnixMilli()), aud)
		content["audios"].([]any)[i] = audioUrl
	}

	return map[string]any{"type": "audio", "content": content}
}

func handleVideoMsg(userId int, content map[string]any) map[string]any {
	var vd struct {
		Videos []struct {
			Video []byte
		}
	}

	helpers.ParseToStruct(content, &vd)

	for i, vid := range vd.Videos {
		videoUrl, _ := helpers.UploadFile(fmt.Sprintf("videos/user-%d/vid-%d.mp4", userId, time.Now().UnixMilli()), vid.Video)
		content["videos"].([]map[string]any)[i]["video"] = videoUrl
	}

	return map[string]any{"type": "video", "content": content}
}

func handleImageMsg(userId int, content map[string]any) map[string]any {
	var imgd struct {
		Images []struct {
			Image []byte
		}
	}

	helpers.ParseToStruct(content, &imgd)

	for i, img := range imgd.Images {
		imageUrl, _ := helpers.UploadFile(fmt.Sprintf("images/user-%d/img-%d.mp4", userId, time.Now().UnixMilli()), img.Image)
		content["images"].([]map[string]any)[i]["image"] = imageUrl
	}

	return map[string]any{"type": "image", "content": content}
}

func MessageBinaryToUrl(userId int, msgBody map[string]any) map[string]any {
	var msg struct {
		Type    string
		Content map[string]any
	}

	helpers.ParseToStruct(msgBody, &msg)

	switch msg.Type {
	case "voice":
		return handleVoiceMsg(userId, msg.Content)
	case "audio":
		return handleAudioMsg(userId, msg.Content)
	case "video":
		return handleVideoMsg(userId, msg.Content)
	case "image":
		return handleImageMsg(userId, msg.Content)
	default:
		return msgBody
	}
}
