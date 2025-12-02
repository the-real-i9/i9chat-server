package cache

import (
	"context"
	"fmt"
	"i9chat/src/helpers"

	"github.com/redis/go-redis/v9"
)

func GetGroup[T any](ctx context.Context, groupId string) (group T, err error) {
	groupJson, err := rdb().HGet(ctx, "groups", groupId).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return group, err
	}

	return helpers.FromJson[T](groupJson), nil
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
		return 0, err
	}

	return count, nil
}

func GetDirectChatHistoryEntry[T any](ctx context.Context, CHEId string) (CHE T, err error) {
	CHEJson, err := rdb().HGet(ctx, "direct_chat_history_entries", CHEId).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return CHE, err
	}

	return helpers.FromJson[T](CHEJson), nil
}

func GetGroupChatHistoryEntry[T any](ctx context.Context, CHEId string) (CHE T, err error) {
	CHEJson, err := rdb().HGet(ctx, "group_chat_history_entries", CHEId).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return CHE, err
	}

	return helpers.FromJson[T](CHEJson), nil
}

func GetMsgReactions(ctx context.Context, msgId string) (map[string]string, error) {
	msgReactions, err := rdb().HGetAll(ctx, fmt.Sprintf("message:%s:reactions", msgId)).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return nil, err
	}

	return msgReactions, nil
}
