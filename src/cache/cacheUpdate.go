package cache

import (
	"context"
	"i9chat/src/helpers"
	"maps"
)

func UpdateDirectMessageDelivery(ctx context.Context, CHEId string, updateKVMap map[string]any) error {
	msgDataMsgPack, err := rdb().HGet(ctx, "direct_chat_history_entries", CHEId).Result()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	msgData := helpers.FromMsgPack[map[string]any](msgDataMsgPack)

	// if a client skips the "delivered" ack, and acks "read"
	// it means the message is delivered and read at the same time
	if updateKVMap["read_at"] != nil && msgData["delivered_at"] == nil {
		msgData["delivered_at"] = updateKVMap["read_at"]
	}

	maps.Copy(msgData, updateKVMap)

	err = rdb().HSet(ctx, "direct_chat_history_entries", CHEId, helpers.ToMsgPack(msgData)).Err()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}

func UpdateGroupMessageDelivery(ctx context.Context, CHEId string, updateKVMap map[string]any) error {
	msgDataMsgPack, err := rdb().HGet(ctx, "group_chat_history_entries", CHEId).Result()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	msgData := helpers.FromMsgPack[map[string]any](msgDataMsgPack)

	// if a client skips the "delivered" ack, and acks "read"
	// it means the message is delivered and read at the same time
	if updateKVMap["read_at"] != nil && msgData["delivered_at"] == nil {
		msgData["delivered_at"] = updateKVMap["read_at"]
	}

	maps.Copy(msgData, updateKVMap)

	err = rdb().HSet(ctx, "group_chat_history_entries", CHEId, helpers.ToMsgPack(msgData)).Err()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}
