package eventStreamService

import (
	"context"
	"i9chat/src/appGlobals"
	"i9chat/src/helpers"
	"i9chat/src/services/eventStreamService/eventTypes"

	"github.com/redis/go-redis/v9"
)

func rdb() *redis.Client {
	return appGlobals.RedisClient
}

func QueueNewUserEvent(nue eventTypes.NewUserEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "new_users",
		Values: nue,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueEditUserEvent(eue eventTypes.EditUserEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "user_edits",
		Values: eue,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueUserPresenceChangeEvent(upce eventTypes.UserPresenceChangeEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "user_presence_changes",
		Values: upce,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueNewDirectMessageEvent(ndme eventTypes.NewDirectMessageEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "new_direct_messages",
		Values: ndme,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueDirectMsgAckEvent(dmae eventTypes.DirectMsgAckEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "direct_msg_acks",
		Values: dmae,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueNewDirectMsgReactionEvent(nmre eventTypes.NewDirectMsgReactionEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "direct_msg_reactions",
		Values: nmre,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueDirectMsgReactionRemovedEvent(mrre eventTypes.DirectMsgReactionRemovedEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "direct_msg_reactions_removed",
		Values: mrre,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueNewGroupEvent(nue eventTypes.NewGroupEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "new_groups",
		Values: nue,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueGroupEditEvent(ege eventTypes.GroupEditEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "group_edits",
		Values: ege,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueGroupUsersAddedEvent(ege eventTypes.GroupUsersAddedEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "group_users_added",
		Values: ege,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueNewGroupMessageEvent(ndme eventTypes.NewGroupMessageEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "new_group_messages",
		Values: ndme,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueGroupMsgAckEvent(dmae eventTypes.GroupMsgAckEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "group_msg_acks",
		Values: dmae,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueNewGroupMsgReactionEvent(nmre eventTypes.NewGroupMsgReactionEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "group_msg_reactions",
		Values: nmre,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}

func QueueGroupMsgReactionRemovedEvent(mrre eventTypes.GroupMsgReactionRemovedEvent) {
	ctx := context.Background()

	err := rdb().XAdd(ctx, &redis.XAddArgs{
		Stream: "group_msg_reactions_removed",
		Values: mrre,
	}).Err()
	if err != nil {
		helpers.LogError(err)
	}
}
