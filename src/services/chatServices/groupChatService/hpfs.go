package groupChatService

import (
	"context"
	"fmt"
	"i9chat/src/services/cloudStorageService"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gofiber/fiber/v2"
)

func uploadGroupPicture(ctx context.Context, pictureData []byte) (string, error) {
	mediaMIME := mimetype.Detect(pictureData)
	mediaType, mediaExt := mediaMIME.String(), mediaMIME.Extension()

	if !strings.HasPrefix(mediaType, "image") {
		return "", fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("invalid picture type %s. expected image/*", mediaType))
	}

	picPath := fmt.Sprintf("group_chat_pictures/group_chat_pic_%d%s", time.Now().UnixNano(), mediaExt)

	picUrl, err := cloudStorageService.Upload(ctx, picPath, pictureData)

	if err != nil {
		return "", err
	}

	return picUrl, nil
}
