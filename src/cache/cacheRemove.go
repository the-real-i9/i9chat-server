package cache

import (
	"context"
	"fmt"
	"i9chat/src/helpers"

	"github.com/redis/go-redis/v9"
)

func RemoveOfflineUsers(pipe redis.Pipeliner, ctx context.Context, users []any) {
	pipe.ZRem(ctx, "offline_users", users...)
}

func RemoveGroupMembers(pipe redis.Pipeliner, ctx context.Context, groupId string, members []any) {
	pipe.SRem(ctx, fmt.Sprintf("group:%s:members", groupId), members...)
}

func RemoveGroupAdmins(pipe redis.Pipeliner, ctx context.Context, groupId string, admins []any) {
	pipe.SRem(ctx, fmt.Sprintf("group:%s:admins", groupId), admins...)
}

func RemoveDirectChatHistoryEntries(ctx context.Context, CHEIds []string) error {
	if err := rdb().HDel(ctx, "direct_chat_history_entries", CHEIds...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func RemoveGroupChatHistoryEntries(ctx context.Context, CHEIds []string) error {
	if err := rdb().HDel(ctx, "group_chat_history_entries", CHEIds...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func RemoveDirectChatHistory(pipe redis.Pipeliner, ctx context.Context, ownerUser, partnerUser string, CHEIds []any) {
	pipe.ZRem(ctx, fmt.Sprintf("chat:owner:%s:partner:%s:history", ownerUser, partnerUser), CHEIds...)
	pipe.ZRem(ctx, fmt.Sprintf("chat:owner:%s:partner:%s:history", partnerUser, ownerUser), CHEIds...)
}

func RemoveGroupChatHistory(pipe redis.Pipeliner, ctx context.Context, ownerUser, groupId string, CHEIds []any) {
	pipe.ZRem(ctx, fmt.Sprintf("group_chat:owner:%s:group_id:%s:history", ownerUser, groupId), CHEIds...)
}

func RemoveMsgReactions(pipe redis.Pipeliner, ctx context.Context, msgId string, reactorUsers []string) {
	pipe.HDel(ctx, fmt.Sprintf("message:%s:reactions", msgId), reactorUsers...)
}

func RemoveUserChatUnreadMsgs(pipe redis.Pipeliner, ctx context.Context, ownerUser, chatIdent string, readMsgs []any) {
	pipe.SRem(ctx, fmt.Sprintf("chat:owner:%s:ident:%s:unread_messages", ownerUser, chatIdent), readMsgs...)
}
