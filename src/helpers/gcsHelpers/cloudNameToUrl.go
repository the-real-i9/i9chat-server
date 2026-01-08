package gcsHelpers

import (
	"fmt"
	"i9chat/src/helpers"
	"slices"
)

func ProfilePicCloudNameToUrl(userMap map[string]any) error {
	ppicCloudName := userMap["profile_pic_cloud_name"].(string)

	if ppicCloudName != "{notset}" {
		var (
			smallPPicn  string
			mediumPPicn string
			largePPicn  string
		)

		_, err := fmt.Sscanf(ppicCloudName, "small:%s medium:%s large:%s", &smallPPicn, &mediumPPicn, &largePPicn)
		if err != nil {
			helpers.LogError(err)
			return err
		}

		smallPicUrl, err := GetMediaurl(smallPPicn)
		if err != nil {
			return err
		}

		mediumPicUrl, err := GetMediaurl(mediumPPicn)
		if err != nil {
			return err
		}

		largePicUrl, err := GetMediaurl(largePPicn)
		if err != nil {
			return err
		}

		userMap["profile_pic_url"] = fmt.Sprintf("small:%s medium:%s large:%s", smallPicUrl, mediumPicUrl, largePicUrl)
	} else {
		userMap["profile_pic_url"] = "{notset}"
	}

	delete(userMap, "profile_pic_cloud_name")

	return nil
}

func GroupPicCloudNameToUrl(groupMap map[string]any) error {
	picCloudName := groupMap["picture_cloud_name"].(string)

	var (
		smallPicn  string
		mediumPicn string
		largePicn  string
	)

	_, err := fmt.Sscanf(picCloudName, "small:%s medium:%s large:%s", &smallPicn, &mediumPicn, &largePicn)
	if err != nil {
		helpers.LogError(err)
		return err
	}

	smallPicUrl, err := GetMediaurl(smallPicn)
	if err != nil {
		return err
	}

	mediumPicUrl, err := GetMediaurl(mediumPicn)
	if err != nil {
		return err
	}

	largePicUrl, err := GetMediaurl(largePicn)
	if err != nil {
		return err
	}

	groupMap["picture_url"] = fmt.Sprintf("small:%s medium:%s large:%s", smallPicUrl, mediumPicUrl, largePicUrl)

	delete(groupMap, "picture_cloud_name")

	return nil
}

func MessageMediaCloudNameToUrl(msgContent map[string]any) error {
	contentProps := msgContent["props"].(map[string]any)

	if msgContent["type"].(string) != "text" {
		mediaCloudName := contentProps["media_cloud_name"].(string)

		if slices.Contains([]string{"photo", "video"}, msgContent["type"].(string)) {
			var (
				blurPlchMcn string
				actualMcn   string
			)

			_, err := fmt.Sscanf(mediaCloudName, "blur_placeholder:%s actual:%s", &blurPlchMcn, &actualMcn)
			if err != nil {
				helpers.LogError(err)
				return err
			}

			blurPlchUrl, err := GetMediaurl(blurPlchMcn)
			if err != nil {
				return err
			}

			actualUrl, err := GetMediaurl(actualMcn)
			if err != nil {
				return err
			}

			contentProps["media_url"] = fmt.Sprintf("blur_placeholder:%s actual:%s", blurPlchUrl, actualUrl)
		} else {
			mediaUrl, err := GetMediaurl(mediaCloudName)
			if err != nil {
				return err
			}

			contentProps["media_url"] = mediaUrl
		}

		delete(contentProps, "media_cloud_name")
	}

	return nil
}
