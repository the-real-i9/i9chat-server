package chatTypes

import (
	"context"
	"fmt"
	"i9chat/src/services/cloudStorageService"
	"regexp"
	"slices"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type msgProps struct {
	TextContent    *string `json:"text_content,omitempty"`
	MediaCloudName *string `json:"media_cloud_name,omitempty"`
	Duration       *int64  `json:"duration,omitempty"`
	Caption        *string `json:"caption,omitempty"`
	Name           *string `json:"name,omitempty"`

	// fields to set when sending to client
	// Url            *string `json:"url,omitempty"`
	// MimeType       *string `json:"mime_type,omitempty"`
	// Size           *int64  `json:"size,omitempty"`
	// Extension      *string `json:"extension,omitempty"`
}

type MsgContent struct {
	Type     string `json:"type"`
	msgProps `json:"props"`
}

func (m MsgContent) Validate() error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.Type,
			validation.Required,
			validation.In("text", "voice", "audio", "video", "photo", "file").Error("invalid message type"),
		),
		validation.Field(&m.MediaCloudName,
			validation.When(m.Type == "text", validation.Nil.Error("invalid property for the specified type")).Else(
				validation.Required,
				validation.Match(regexp.MustCompile(
					`^blur_placeholder:uploads/chat/[\w-/]+\w actual:uploads/chat/[\w-/]+\w$`,
				)).Error("invalid media cloud name"),
			),
		),
		validation.Field(&m.msgProps, validation.Required),
		validation.Field(&m.TextContent, validation.When(m.Type != "text", validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
		validation.Field(&m.Duration, validation.When(m.Type != "voice", validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
		validation.Field(&m.Caption, validation.When(slices.Contains([]string{"text", "voice", "file", "audio"}, m.Type), validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
		validation.Field(&m.Name, validation.When(m.Type != "file", validation.Nil.Error("invalid property for the specified type")).Else(validation.Required)),
	)

	if err != nil {
		return err
	}

	if mediaCloudName := m.MediaCloudName; mediaCloudName != nil {
		go func(msgType, mediaCloudName string) {

			ctx := context.Background()

			switch msgType {
			case "photo", "video":
				var (
					mcnBlur   string
					mcnActual string
				)

				fmt.Sscanf(mediaCloudName, "blur_placeholder:%s actual:%s", &mcnBlur, &mcnActual)

				if mInfo := cloudStorageService.GetMediaInfo(ctx, mcnBlur); mInfo != nil {
					if mInfo.Size < 1*1024 || mInfo.Size > 100*1024 {
						cloudStorageService.DeleteCloudMedia(ctx, mcnBlur)
					}
				}

				if mInfo := cloudStorageService.GetMediaInfo(ctx, mcnActual); mInfo != nil {
					if msgType == "photo" && mInfo.Size < 1*1024 || mInfo.Size > 10*1024*1024 {
						cloudStorageService.DeleteCloudMedia(ctx, mcnActual)
					} else if mInfo.Size < 1*1024 || mInfo.Size > 40*1024*1024 {
						cloudStorageService.DeleteCloudMedia(ctx, mcnActual)
					}
				}
			case "voice":
				if mInfo := cloudStorageService.GetMediaInfo(ctx, mediaCloudName); mInfo != nil {
					if mInfo.Size < 500 || mInfo.Size > 10*1024*1024 {
						cloudStorageService.DeleteCloudMedia(ctx, mediaCloudName)
					}
				}
			case "audio":
				if mInfo := cloudStorageService.GetMediaInfo(ctx, mediaCloudName); mInfo != nil {
					if mInfo.Size < 500 || mInfo.Size > 20*1024*1024 {
						cloudStorageService.DeleteCloudMedia(ctx, mediaCloudName)
					}
				}
			default:
				if mInfo := cloudStorageService.GetMediaInfo(ctx, mediaCloudName); mInfo != nil {
					if mInfo.Size < 500 || mInfo.Size > 50*1024*1024 {
						cloudStorageService.DeleteCloudMedia(ctx, mediaCloudName)
					}
				}
			}
		}(m.Type, *mediaCloudName)
	}

	return nil
}
