package chatUploadControllers

import (
	"errors"
	"i9chat/src/helpers"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type authorizeUploadBody struct {
	MsgType   string `json:"msg_type"`
	MediaMIME string `json:"media_mime"`
	MediaSize int64  `json:"media_size"`
}

func (b authorizeUploadBody) Validate() error {

	err := validation.ValidateStruct(&b,
		validation.Field(&b.MsgType,
			validation.Required,
			validation.In("voice", "audio", "file").Error("invalid message type"),
		),
		validation.Field(&b.MediaMIME,
			validation.Required,
			validation.When(b.MsgType == "voice",
				validation.In("audio/ogg", "audio/aac").Error(`unsupported media_mime for voice; use one of ["audio/ogg", "audio/aac"]`),
			).Else(validation.When(b.MsgType == "audio",
				validation.In("audio/ogg", "audio/aac", "audio/mpeg", "audio/mp4", "audio/webm").Error(`unsupported media_mime for audio; use one of ["audio/ogg", "audio/aac", "audio/mpeg", "audio/mp4", "audio/webm"]`),
			).Else(validation.By(func(value any) error {
				val := value.(string)

				if strings.HasPrefix(val, "/") || strings.HasSuffix(val, "/") || !strings.Contains(val, "/") {
					return errors.New("invalid mime type")
				}

				return nil
			})),
			),
		),
		validation.Field(&b.MediaSize,
			validation.Required,
			validation.When(
				b.MsgType == "voice", validation.By(func(value any) error {
					val := value.(int64)

					if val < 500 || val > 4*1024*1024 {
						return errors.New("voice media_size out of range; min: 500BiB; max: 4MeB")
					}

					return nil
				}),
			).Else(validation.When(
				b.MsgType == "audio", validation.By(func(value any) error {
					val := value.(int64)

					if val < 1024 || val > 20*1024*1024 {
						return errors.New("audio media_size out of range; min: 1KiB; max: 20MeB")
					}

					return nil
				}),
			).Else(validation.By(func(value any) error {
				val := value.(int64)

				if val < 500 || val > 50*1024*1024 {
					return errors.New("file media_size out of range; min: 500BiB; max: 50MeB")
				}

				return nil
			})),
			),
		),
	)

	return helpers.ValidationError(err, "chatUploadControllers_validation.go", "authorizeUploadBody")
}

type authorizeVisualUploadBody struct {
	MsgType   string    `json:"msg_type"`
	MediaMIME [2]string `json:"media_mime"`
	MediaSize [2]int64  `json:"media_size"`
}

func (b authorizeVisualUploadBody) Validate() error {

	err := validation.ValidateStruct(&b,
		validation.Field(&b.MsgType,
			validation.Required,
			validation.In("photo", "video").Error("invalid message type"),
		),
		validation.Field(&b.MediaMIME, validation.Required, validation.Length(2, 2).Error("expected array of 2 items.")),
		validation.Field(&b.MediaMIME[0], validation.Required,
			validation.In("image/jpeg", "image/png", "image/webp", "image/avif").Error(`unsupported blur placeholder media_mime; use one of ["image/jpeg", "image/png", "image/webp", "image/avif"]`),
		),
		validation.Field(&b.MediaMIME[1], validation.Required,
			validation.When(b.MsgType == "photo",
				validation.In("image/jpeg", "image/png", "image/webp", "image/avif").Error(`unsupported photo media_mime; use one of ["image/jpeg", "image/png", "image/webp", "image/avif"]`),
			).Else(validation.In("video/mp4", "video/webm").Error(`unsupported video media_mime; use one of ["video/mp4", "video/webm"]`)),
		),
		validation.Field(&b.MediaSize,
			validation.Required,
			validation.Length(2, 2).Error("expected a media_size array of 2 items"),
			validation.By(func(value any) error {
				media_size := value.([2]int64)

				const (
					_                    = iota
					BLUR_PLACEHOLDER int = iota - 1
					ACTUAL_MEDIA
				)

				if media_size[BLUR_PLACEHOLDER] < 1*1024 || media_size[BLUR_PLACEHOLDER] > 10*1024 {
					return errors.New("blur placeholder media_size out of range; min: 1KiB; max: 10KiB")
				}

				switch b.MsgType {
				case "photo":
					if media_size[ACTUAL_MEDIA] < 1*1024 || media_size[ACTUAL_MEDIA] > 10*1024*1024 {
						return errors.New("photo media_size out of range; min: 1KiB; max: 10MeB")
					}
				case "video":
					if media_size[ACTUAL_MEDIA] < 1*1024 || media_size[ACTUAL_MEDIA] > 40*1024*1024 {
						return errors.New("video media_size out of range; min: 1KiB; max: 40MeB")
					}
				default:
				}

				return nil
			}),
		),
	)

	return helpers.ValidationError(err, "chatUploadControllers_validation.go", "authorizeVisualUploadBody")
}
