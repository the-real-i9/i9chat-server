package cloudStorageService

import (
	"fmt"
	"i9chat/src/helpers"
)

func ProfilePicCloudNameToUrl(ppicCloudName string) string {
	if ppicCloudName != "{notset}" {
		var (
			smallPPicn  string
			mediumPPicn string
			largePPicn  string
		)

		_, err := fmt.Sscanf(ppicCloudName, "small:%s medium:%s large:%s", &smallPPicn, &mediumPPicn, &largePPicn)
		if err != nil {
			helpers.LogError(err)
		}

		smallPicUrl := GetMediaUrl(smallPPicn)
		mediumPicUrl := GetMediaUrl(mediumPPicn)
		largePicUrl := GetMediaUrl(largePPicn)

		return fmt.Sprintf("small:%s medium:%s large:%s", smallPicUrl, mediumPicUrl, largePicUrl)
	}

	return ppicCloudName
}

func GroupPicCloudNameToUrl(picCloudName string) string {
	var (
		smallPicn  string
		mediumPicn string
		largePicn  string
	)

	_, err := fmt.Sscanf(picCloudName, "small:%s medium:%s large:%s", &smallPicn, &mediumPicn, &largePicn)
	if err != nil {
		helpers.LogError(err)
	}

	smallPicUrl := GetMediaUrl(smallPicn)
	mediumPicUrl := GetMediaUrl(mediumPicn)
	largePicUrl := GetMediaUrl(largePicn)

	return fmt.Sprintf("small:%s medium:%s large:%s", smallPicUrl, mediumPicUrl, largePicUrl)
}

func MessageMediaCloudNameToUrl(msgContent map[string]any) {
	contentProps := msgContent["props"].(map[string]any)

	msgContentType := msgContent["type"].(string)

	if msgContentType != "text" {
		mediaCloudName := contentProps["media_cloud_name"].(string)

		if msgContentType == "photo" || msgContentType == "video" {
			var (
				blurPlchMcn string
				actualMcn   string
			)

			_, err := fmt.Sscanf(mediaCloudName, "blur_placeholder:%s actual:%s", &blurPlchMcn, &actualMcn)
			if err != nil {
				helpers.LogError(err)
			}

			blurPlchUrl := GetMediaUrl(blurPlchMcn)
			actualUrl := GetMediaUrl(actualMcn)

			contentProps["media_url"] = fmt.Sprintf("blur_placeholder:%s actual:%s", blurPlchUrl, actualUrl)
		} else {
			mediaUrl := GetMediaUrl(mediaCloudName)

			contentProps["media_url"] = mediaUrl
		}

		delete(contentProps, "media_cloud_name")
	}
}
