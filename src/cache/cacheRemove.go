package cache

import (
	"context"
	"fmt"
	"i9chat/src/helpers"
)

func RemoveOfflineUsers(ctx context.Context, users []any) error {
	if err := rdb().ZRem(ctx, "offline_users", users...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
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

func RemoveDirectChatHistory(ctx context.Context, ownerUser, partnerUser string, CHEIds []any) error {
	if err := rdb().ZRem(ctx, fmt.Sprintf("chat:owner:%s:partner:%s:history", ownerUser, partnerUser), CHEIds...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	if err := rdb().ZRem(ctx, fmt.Sprintf("chat:owner:%s:partner:%s:history", partnerUser, ownerUser), CHEIds...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func RemoveGroupChatHistory(ctx context.Context, groupId string, CHEIds []any) error {
	if err := rdb().ZRem(ctx, fmt.Sprintf("group_chat:%s:history", groupId), CHEIds...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func RemoveMsgReactions(ctx context.Context, msgId string, reactorUsers []string) error {
	if err := rdb().HDel(ctx, fmt.Sprintf("message:%s:reactions", msgId), reactorUsers...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func RemoveUserChatUnreadMsgs(ctx context.Context, ownerUser, chatIdent string, readMsgs []any) error {
	if err := rdb().SRem(ctx, fmt.Sprintf("chat:owner:%s:ident:%s:unread_messages", ownerUser, chatIdent), readMsgs...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}
