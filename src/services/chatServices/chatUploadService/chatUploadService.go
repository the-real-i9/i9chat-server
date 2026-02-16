package chatUploadService

import (
	"context"
	"fmt"
	"i9chat/src/services/cloudStorageService"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type AuthDataT struct {
	UploadUrl      string `msgpack:"uploadUrl"`
	MediaCloudName string `msgpack:"mediaCloudName"`
}

func Authorize(ctx context.Context, msgType, mediaMIME string) (AuthDataT, error) {
	var res AuthDataT

	mediaCloudName := fmt.Sprintf("uploads/chat/%s/%d%d/%s", msgType, time.Now().Year(), time.Now().Month(), uuid.NewString())

	url, err := cloudStorageService.GetUploadUrl(mediaCloudName, mediaMIME)
	if err != nil {
		return res, fiber.ErrInternalServerError
	}

	res.UploadUrl = url
	res.MediaCloudName = mediaCloudName

	return res, nil
}

func AuthorizeVisual(ctx context.Context, msgType string, mediaMIME [2]string) (AuthDataT, error) {
	var res AuthDataT

	for blurPlch0_actual1, mime := range mediaMIME {

		which := [2]string{"blur_placeholder", "actual"}

		mediaCloudName := fmt.Sprintf("uploads/chat/%s/%d%d/%s-%s", msgType, time.Now().Year(), time.Now().Month(), uuid.NewString(), which[blurPlch0_actual1])

		url, err := cloudStorageService.GetUploadUrl(mediaCloudName, mime)
		if err != nil {
			return res, fiber.ErrInternalServerError
		}

		if blurPlch0_actual1 == 0 {
			res.UploadUrl += "blur_placeholder:"
			res.MediaCloudName += "blur_placeholder:"
		} else {
			res.UploadUrl += "actual:"
			res.MediaCloudName += "actual:"
		}

		res.UploadUrl += url
		res.MediaCloudName += mediaCloudName

		if blurPlch0_actual1 == 0 {
			res.UploadUrl += " "
			res.MediaCloudName += " "
		}
	}

	return res, nil
}
