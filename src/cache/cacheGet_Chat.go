package cache

import (
	"context"
	"fmt"
	"i9chat/src/helpers"
	"slices"

	"github.com/redis/go-redis/v9"
)

func GetGroup[T any](ctx context.Context, groupId string) (group T, err error) {
	groupJson, err := rdb().HGet(ctx, "groups", groupId).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return group, err
	}

	groupMap := helpers.FromJson[map[string]any](groupJson)

	picCloudName := groupMap["picture_cloud_name"].(string)

	var (
		smallPicn  string
		mediumPicn string
		largePicn  string
	)

	_, err = fmt.Sscanf(picCloudName, "small:%s medium:%s large:%s", &smallPicn, &mediumPicn, &largePicn)
	if err != nil {
		return group, err
	}

	smallPicUrl, err := getMediaurl(smallPicn)
	if err != nil {
		return group, err
	}

	mediumPicUrl, err := getMediaurl(mediumPicn)
	if err != nil {
		return group, err
	}

	largePicUrl, err := getMediaurl(largePicn)
	if err != nil {
		return group, err
	}

	groupMap["picture_url"] = fmt.Sprintf("small:%s medium:%s large:%s", smallPicUrl, mediumPicUrl, largePicUrl)

	delete(groupMap, "picture_cloud_name")

	return helpers.ToStruct[T](groupMap), nil
}

func GetGroupMembersList(ctx context.Context, groupId string) ([]string, error) {
	list, err := rdb().SMembers(ctx, fmt.Sprintf("group:%s:members", groupId)).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return list, err
	}

	return list, nil
}

func GetGroupMembersCount(ctx context.Context, groupId string) (int64, error) {
	count, err := rdb().SCard(ctx, fmt.Sprintf("group:%s:members", groupId)).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return count, err
	}

	return count, nil
}

func GetGroupOnlineMembersCount(ctx context.Context, groupId string) (int, error) {
	onMems, err := rdb().SDiff(ctx, fmt.Sprintf("group:%s:members", groupId), "offline_users_unsorted").Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return 0, err
	}

	return len(onMems), nil
}

func GetChat[T any](ctx context.Context, ownerUser, chatIdent string) (chat T, err error) {
	chatJson, err := rdb().HGet(ctx, fmt.Sprintf("user:%s:chats", ownerUser), chatIdent).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return chat, err
	}

	return helpers.FromJson[T](chatJson), nil
}

func GetChatUnreadMsgsCount(ctx context.Context, ownerUser, chatIdent string) (int64, error) {
	count, err := rdb().SCard(ctx, fmt.Sprintf("chat:owner:%s:ident:%s:unread_messages", ownerUser, chatIdent)).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return count, err
	}

	return count, nil
}

func GetDirectChatHistoryEntry[T any](ctx context.Context, CHEId string) (CHE T, err error) {
	CHEJson, err := rdb().HGet(ctx, "direct_chat_history_entries", CHEId).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return CHE, err
	}

	CHEMap := helpers.FromJson[map[string]any](CHEJson)

	cheType := CHEMap["che_type"].(string)

	if cheType == "message" {
		content := CHEMap["content"].(map[string]any)
		contentProps := content["props"].(map[string]any)

		if content["type"].(string) != "text" {
			mediaCloudName := contentProps["media_cloud_name"].(string)

			if slices.Contains([]string{"photo", "video"}, content["type"].(string)) {
				var (
					blurPlchMcn string
					actualMcn   string
				)

				_, err = fmt.Sscanf(mediaCloudName, "blur_placeholder:%s actual:%s", &blurPlchMcn, &actualMcn)
				if err != nil {
					return CHE, err
				}

				blurPlchUrl, err := getMediaurl(blurPlchMcn)
				if err != nil {
					return CHE, err
				}

				actualUrl, err := getMediaurl(actualMcn)
				if err != nil {
					return CHE, err
				}

				contentProps["media_url"] = fmt.Sprintf("blur_placeholder:%s actual:%s", blurPlchUrl, actualUrl)
			} else {
				var mcn string

				_, err = fmt.Sscanf(mediaCloudName, "%s", &mcn)
				if err != nil {
					return CHE, err
				}

				mediaUrl, err := getMediaurl(mcn)
				if err != nil {
					return CHE, err
				}

				contentProps["media_url"] = mediaUrl
			}

			delete(contentProps, "media_cloud_name")
		}
	}

	return helpers.ToStruct[T](CHEMap), nil
}

func GetGroupChatHistoryEntry[T any](ctx context.Context, CHEId string) (CHE T, err error) {
	CHEJson, err := rdb().HGet(ctx, "group_chat_history_entries", CHEId).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return CHE, err
	}

	CHEMap := helpers.FromJson[map[string]any](CHEJson)

	cheType := CHEMap["che_type"].(string)

	if cheType == "message" {
		content := CHEMap["content"].(map[string]any)
		contentProps := content["props"].(map[string]any)

		if content["type"].(string) != "text" {
			mediaCloudName := contentProps["media_cloud_name"].(string)

			if slices.Contains([]string{"photo", "video"}, content["type"].(string)) {
				var (
					blurPlchMcn string
					actualMcn   string
				)

				_, err = fmt.Sscanf(mediaCloudName, "blur_placeholder:%s actual:%s", &blurPlchMcn, &actualMcn)
				if err != nil {
					return CHE, err
				}

				blurPlchUrl, err := getMediaurl(blurPlchMcn)
				if err != nil {
					return CHE, err
				}

				actualUrl, err := getMediaurl(actualMcn)
				if err != nil {
					return CHE, err
				}

				contentProps["media_url"] = fmt.Sprintf("blur_placeholder:%s actual:%s", blurPlchUrl, actualUrl)
			} else {
				var mcn string

				_, err = fmt.Sscanf(mediaCloudName, "%s", &mcn)
				if err != nil {
					return CHE, err
				}

				mediaUrl, err := getMediaurl(mcn)
				if err != nil {
					return CHE, err
				}

				contentProps["media_url"] = mediaUrl
			}

			delete(contentProps, "media_cloud_name")
		}
	}

	return helpers.ToStruct[T](CHEMap), nil
}

func GetMsgReactions(ctx context.Context, msgId string) (map[string]string, error) {
	msgReactions, err := rdb().HGetAll(ctx, fmt.Sprintf("message:%s:reactions", msgId)).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return nil, err
	}

	return msgReactions, nil
}

func GetGroupMsgDeliveredToUsersCount(ctx context.Context, groupId, msgId string) (int64, error) {
	count, err := rdb().ZCard(ctx, fmt.Sprintf("group:%s:msg:%s:delivered_to_users", groupId, msgId)).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return count, err
	}

	return count, nil
}

func GetGroupMsgReadByUsersCount(ctx context.Context, groupId, msgId string) (int64, error) {
	count, err := rdb().ZCard(ctx, fmt.Sprintf("group:%s:msg:%s:read_by_users", groupId, msgId)).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return count, err
	}

	return count, nil
}
