package chatUploadControllers

import (
	"errors"
	"i9chat/src/helpers"
	"slices"
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

					if val < 500 || val > 10*1024*1024 {
						return errors.New("voice media_size out of range; min: 500BiB; max: 10MeB")
					}

					return nil
				}),
			).Else(validation.When(
				b.MsgType == "audio", validation.By(func(value any) error {
					val := value.(int64)

					if val < 500 || val > 20*1024*1024 {
						return errors.New("audio media_size out of range; min: 500BiB; max: 20MeB")
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

	return helpers.ValidationError(err, "cucValidation.go", "authorizeUploadBody")
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
		validation.Field(&b.MediaMIME, validation.Required, validation.Length(2, 2).Error("expected array of 2 items."),
			validation.By(func(value any) error {
				val := value.([2]string)

				const (
					_             = iota
					BLUR_MIME int = iota - 1
					ACTUAL_MIME
				)

				if !slices.Contains([]string{"image/jpeg", "image/png", "image/webp", "image/avif"}, val[BLUR_MIME]) {
					return errors.New(`unsupported blur placeholder media_mime; use one of ["image/jpeg", "image/png", "image/webp", "image/avif"]`)
				}

				if b.MsgType == "photo" {
					if !slices.Contains([]string{"image/jpeg", "image/png", "image/webp", "image/avif"}, val[ACTUAL_MIME]) {
						return errors.New(`unsupported photo media_mime; use one of ["image/jpeg", "image/png", "image/webp", "image/avif"]`)
					}
				} else {
					if !slices.Contains([]string{"video/mp4", "video/webm"}, val[ACTUAL_MIME]) {
						return errors.New(`unsupported video/reel media_mime; use one of ["video/mp4", "video/webm"]`)
					}
				}

				return nil
			}),
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

				if media_size[BLUR_PLACEHOLDER] < 1*1024 || media_size[BLUR_PLACEHOLDER] > 100*1024 {
					return errors.New("blur placeholder media_size out of range; min: 1KiB; max: 100KiB")
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

	return helpers.ValidationError(err, "cucValidation.go", "authorizeVisualUploadBody")
}
