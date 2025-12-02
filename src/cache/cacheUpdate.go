package cache

import (
	"context"
	"i9chat/src/helpers"
	"maps"
)

func UpdateUser(ctx context.Context, username string, updateKVMap map[string]any) error {
	userDataJson, err := rdb().HGet(ctx, "users", username).Result()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	userData := helpers.FromJson[map[string]any](userDataJson)

	maps.Copy(userData, updateKVMap)

	err = rdb().HSet(ctx, "users", username, helpers.ToJson(userData)).Err()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}

func UpdateGroup(ctx context.Context, groupId string, updateKVMap map[string]any) error {
	groupDataJson, err := rdb().HGet(ctx, "groups", groupId).Result()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	userData := helpers.FromJson[map[string]any](groupDataJson)

	maps.Copy(userData, updateKVMap)

	err = rdb().HSet(ctx, "groups", groupId, helpers.ToJson(userData)).Err()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}

func UpdateDirectMessage(ctx context.Context, CHEId string, updateKVMap map[string]any) error {
	msgDataJson, err := rdb().HGet(ctx, "direct_chat_history_entries", CHEId).Result()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	msgData := helpers.FromJson[map[string]any](msgDataJson)

	maps.Copy(msgData, updateKVMap)

	err = rdb().HSet(ctx, "direct_chat_history_entries", CHEId, helpers.ToJson(msgData)).Err()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}

func UpdateGroupMessage(ctx context.Context, CHEId string, updateKVMap map[string]any) error {
	msgDataJson, err := rdb().HGet(ctx, "group_chat_history_entries", CHEId).Result()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	msgData := helpers.FromJson[map[string]any](msgDataJson)

	maps.Copy(msgData, updateKVMap)

	err = rdb().HSet(ctx, "group_chat_history_entries", CHEId, helpers.ToJson(msgData)).Err()
	if err != nil {
		helpers.LogError(err)
		return err
	}

	return nil
}
