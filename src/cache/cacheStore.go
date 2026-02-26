package cache

import (
	"context"
	"fmt"
	"i9chat/src/helpers"

	"github.com/redis/go-redis/v9"
)

func StoreNewUsers(ctx context.Context, newUsers []string) error {
	if err := rdb().HSet(ctx, "users", newUsers).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreOfflineUsers(pipe redis.Pipeliner, ctx context.Context, user_lastSeen_Pairs map[string]int64) {
	members := []redis.Z{}
	membersUnsorted := []any{}

	for user, lastSeen := range user_lastSeen_Pairs {

		members = append(members, redis.Z{
			Score:  float64(lastSeen),
			Member: user,
		})

		membersUnsorted = append(membersUnsorted, user)
	}

	pipe.ZAdd(ctx, "offline_users", members...)
	pipe.SAdd(ctx, "offline_users_unsorted", membersUnsorted...)
}

func StoreNewGroups(ctx context.Context, newGroups []string) error {
	if err := rdb().HSet(ctx, "groups", newGroups).Err(); err != nil {
		helpers.LogError(err)

		return err
	}

	return nil
}

func StoreGroupMembers(pipe redis.Pipeliner, ctx context.Context, groupId string, members []any) {
	pipe.SAdd(ctx, fmt.Sprintf("group:%s:members", groupId), members...)
}

func StoreGroupAdmins(pipe redis.Pipeliner, ctx context.Context, groupId string, admins []any) {
	pipe.SAdd(ctx, fmt.Sprintf("group:%s:admins", groupId), admins...)
}

func StoreNewUserChats(pipe redis.Pipeliner, ctx context.Context, ownerUser string, chatIdentWithInfoPairs []string) {
	pipe.HSet(ctx, fmt.Sprintf("user:%s:chats", ownerUser), chatIdentWithInfoPairs)
}

func StoreUserChatUnreadMsgs(pipe redis.Pipeliner, ctx context.Context, ownerUser, chatIdent string, unreadMsgs []any) {
	pipe.SAdd(ctx, fmt.Sprintf("chat:owner:%s:ident:%s:unread_messages", ownerUser, chatIdent), unreadMsgs...)
}

func StoreUserChatIdents(pipe redis.Pipeliner, ctx context.Context, ownerUser string, chatIdent_score_Pairs map[string]float64) {
	members := []redis.Z{}
	for partnerUser, score := range chatIdent_score_Pairs {

		members = append(members, redis.Z{
			Score:  score,
			Member: partnerUser,
		})
	}

	pipe.ZAdd(ctx, fmt.Sprintf("user:%s:chats_sorted", ownerUser), members...)
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

func StoreDirectChatHistory(pipe redis.Pipeliner, ctx context.Context, ownerUser, partnerUser string, CHEId_score_Pairs [][2]any) {
	members := []redis.Z{}
	for _, pair := range CHEId_score_Pairs {
		CHEId := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: CHEId,
		})
	}

	pipe.ZAdd(ctx, fmt.Sprintf("direct_chat:owner:%s:partner:%s:history", ownerUser, partnerUser), members...)
	pipe.ZAdd(ctx, fmt.Sprintf("direct_chat:owner:%s:partner:%s:history", partnerUser, ownerUser), members...)
}

func StoreGroupChatHistory(pipe redis.Pipeliner, ctx context.Context, ownerUser, groupId string, CHEId_score_Pairs [][2]any) {
	members := []redis.Z{}
	for _, pair := range CHEId_score_Pairs {
		CHEId := pair[0]

		members = append(members, redis.Z{
			Score:  pair[1].(float64),
			Member: CHEId,
		})
	}

	pipe.ZAdd(ctx, fmt.Sprintf("group_chat:owner:%s:group_id:%s:history", ownerUser, groupId), members...)
}

func StoreGroupMsgDeliveredToUsers(pipe redis.Pipeliner, ctx context.Context, groupId, msgId string, user_deliveredAt_Pairs [][2]any) {
	members := []redis.Z{}
	for _, pair := range user_deliveredAt_Pairs {
		user := pair[0]

		members = append(members, redis.Z{
			Score:  float64(pair[1].(int64)),
			Member: user,
		})
	}

	pipe.ZAdd(ctx, fmt.Sprintf("group:%s:msg:%s:delivered_to_users", groupId, msgId), members...)
}

func StoreGroupMsgReadByUsers(pipe redis.Pipeliner, ctx context.Context, groupId, msgId string, user_deliveredAt_Pairs [][2]any) {
	members := []redis.Z{}
	for _, pair := range user_deliveredAt_Pairs {
		user := pair[0]

		members = append(members, redis.Z{
			Score:  float64(pair[1].(int64)),
			Member: user,
		})
	}

	pipe.ZAdd(ctx, fmt.Sprintf("group:%s:msg:%s:read_by_users", groupId, msgId), members...)
}

func StoreMsgReactions(pipe redis.Pipeliner, ctx context.Context, msgId string, userWithEmojiPairs []string) {
	pipe.HSet(ctx, fmt.Sprintf("message:%s:reactions", msgId), userWithEmojiPairs)
}
