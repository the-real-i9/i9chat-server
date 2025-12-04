package cache

import (
	"context"
	"fmt"
	"i9chat/src/helpers"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

func stmsgIdToScore(val string) (res float64) {
	var err error
	// change the first "-" to a decimal point
	// delete subsequent "-"s, if any (this increments the fractional part),
	// for cases where even ordered stmsgIds clashes based on the logic in some background workers
	if res, err = strconv.ParseFloat(strings.Replace(strings.Replace(val, "-", ".", 1), "-", "", 1), 64); err != nil {
		helpers.LogError(err)
	}

	return
}

func StoreNewUsers(ctx context.Context, newUsers []string) error {
	if err := rdb().HSet(ctx, "users", newUsers).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreOfflineUsers(ctx context.Context, user_lastSeen_Pairs map[string]int64) error {
	members := []redis.Z{}
	for user, lastSeen := range user_lastSeen_Pairs {

		members = append(members, redis.Z{
			Score:  float64(lastSeen),
			Member: user,
		})
	}

	if err := rdb().ZAdd(ctx, "offline_users", members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreNewGroups(ctx context.Context, newGroups []string) error {
	if err := rdb().HSet(ctx, "groups", newGroups).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreGroupMembers(ctx context.Context, groupId string, members []any) error {
	if err := rdb().SAdd(ctx, fmt.Sprintf("group:%s:members", groupId), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreGroupAdmins(ctx context.Context, groupId string, admins []any) error {
	if err := rdb().SAdd(ctx, fmt.Sprintf("group:%s:admins", groupId), admins...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreNewUserChats(ctx context.Context, ownerUser string, chatIdentWithInfoPairs []string) error {
	if err := rdb().HSet(ctx, fmt.Sprintf("user:%s:chats", ownerUser), chatIdentWithInfoPairs).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserChatUnreadMsgs(ctx context.Context, ownerUser, chatIdent string, unreadMsgs []any) error {
	if err := rdb().SAdd(ctx, fmt.Sprintf("chat:owner:%s:ident:%s:unread_messages", ownerUser, chatIdent), unreadMsgs...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreUserChatIdents(ctx context.Context, ownerUser string, chatIdent_stmsgId_Pairs map[string]string) error {
	members := []redis.Z{}
	for partnerUser, stmsgId := range chatIdent_stmsgId_Pairs {

		members = append(members, redis.Z{
			Score:  stmsgIdToScore(stmsgId),
			Member: partnerUser,
		})
	}

	if err := rdb().ZAdd(ctx, fmt.Sprintf("user:%s:chats_sorted", ownerUser), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreDirectChatHistoryEntries(ctx context.Context, newCHEs []string) error {
	if err := rdb().HSet(ctx, "direct_chat_history_entries", newCHEs).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreGroupChatHistoryEntries(ctx context.Context, newCHEs []string) error {
	if err := rdb().HSet(ctx, "group_chat_history_entries", newCHEs).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreDirectChatHistory(ctx context.Context, ownerUser, partnerUser string, CHEId_stmsgId_Pairs [][2]string) error {
	members := []redis.Z{}
	for _, pair := range CHEId_stmsgId_Pairs {
		CHEId := pair[0]

		members = append(members, redis.Z{
			Score:  stmsgIdToScore(pair[1]),
			Member: CHEId,
		})
	}

	if err := rdb().ZAdd(ctx, fmt.Sprintf("direct_chat:owner:%s:partner:%s:history", ownerUser, partnerUser), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	if err := rdb().ZAdd(ctx, fmt.Sprintf("direct_chat:owner:%s:partner:%s:history", partnerUser, ownerUser), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreGroupChatHistory(ctx context.Context, ownerUser, groupId string, CHEId_stmsgId_Pairs [][2]string) error {
	members := []redis.Z{}
	for _, pair := range CHEId_stmsgId_Pairs {
		CHEId := pair[0]

		members = append(members, redis.Z{
			Score:  stmsgIdToScore(pair[1]),
			Member: CHEId,
		})
	}

	if err := rdb().ZAdd(ctx, fmt.Sprintf("group_chat:owner:%s:group_id:%s:history", ownerUser, groupId), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreGroupMsgDeliveredToUsers(ctx context.Context, groupId, msgId string, user_deliveredAt_Pairs [][2]any) error {
	members := []redis.Z{}
	for _, pair := range user_deliveredAt_Pairs {
		user := pair[0]

		members = append(members, redis.Z{
			Score:  float64(pair[1].(int64)),
			Member: user,
		})
	}

	if err := rdb().ZAdd(ctx, fmt.Sprintf("group:%s:msg:%s:delivered_to_users", groupId, msgId), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreGroupMsgReadByUsers(ctx context.Context, groupId, msgId string, user_deliveredAt_Pairs [][2]any) error {
	members := []redis.Z{}
	for _, pair := range user_deliveredAt_Pairs {
		user := pair[0]

		members = append(members, redis.Z{
			Score:  float64(pair[1].(int64)),
			Member: user,
		})
	}

	if err := rdb().ZAdd(ctx, fmt.Sprintf("group:%s:msg:%s:read_by_users", groupId, msgId), members...).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreMsgReactions(ctx context.Context, msgId string, userWithEmojiPairs []string) error {
	if err := rdb().HSet(ctx, fmt.Sprintf("message:%s:reactions", msgId), userWithEmojiPairs).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}
