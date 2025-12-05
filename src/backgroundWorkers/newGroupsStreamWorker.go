package backgroundWorkers

import (
	"context"
	"fmt"
	"i9chat/src/appTypes"
	"i9chat/src/cache"
	"i9chat/src/helpers"
	"i9chat/src/services/eventStreamService/eventTypes"
	"log"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

func newGroupsStreamBgWorker(rdb *redis.Client) {
	var (
		streamName   = "new_groups"
		groupName    = "new_group_listeners"
		consumerName = "worker-1"
	)

	ctx := context.Background()

	err := rdb.XGroupCreateMkStream(ctx, streamName, groupName, "$").Err()
	if err != nil && (err.Error() != "BUSYGROUP Consumer Group name already exists") {
		helpers.LogError(err)
		log.Fatal()
	}

	go func() {
		for {
			streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    groupName,
				Consumer: consumerName,
				Streams:  []string{streamName, ">"},
				Count:    500,
				Block:    0,
			}).Result()

			if err != nil {
				helpers.LogError(err)
				continue
			}

			var stmsgIds []string
			var msgs []eventTypes.NewGroupEvent

			for _, stmsg := range streams[0].Messages {
				stmsgIds = append(stmsgIds, stmsg.ID)

				var msg eventTypes.NewGroupEvent
				msg.CreatorUser = stmsg.Values["creatorUser"].(string)
				msg.GroupId = stmsg.Values["groupId"].(string)
				msg.GroupData = stmsg.Values["groupData"].(string)
				msg.InitMembers = helpers.FromJson[appTypes.BinableSlice](stmsg.Values["initMembers"].(string))
				msg.CreatorUserCHEs = helpers.FromJson[appTypes.BinableSlice](stmsg.Values["creatorUserCHEs"].(string))
				msg.InitMembersCHEs = helpers.FromJson[appTypes.BinableMap](stmsg.Values["initMembersCHEs"].(string))

				msgs = append(msgs, msg)
			}

			msgsLen := len(msgs)

			newGroups := []string{}

			groupInitMembers := make(map[string][]any, msgsLen)

			groupCreators := make(map[string]any, msgsLen)

			newUserChats := make(map[string][]string)

			userChats := make(map[string]map[string]string)

			newGroupActivityEntries := []string{}

			chatGroupActivities := make(map[string][][2]string)

			// batch data for batch processing
			for i, msg := range msgs {
				newGroups = append(newGroups, msg.GroupId, msg.GroupData)

				groupInitMembers[msg.GroupId] = append(msg.InitMembers, msg.CreatorUser)

				groupCreators[msg.GroupId] = msg.CreatorUser

				newUserChats[msg.CreatorUser] = append(newUserChats[msg.CreatorUser], msg.GroupId, helpers.ToJson(map[string]any{"type": "group", "group": msg.GroupId}))

				if userChats[msg.CreatorUser] == nil {
					userChats[msg.CreatorUser] = make(map[string]string)
				}

				userChats[msg.CreatorUser][msg.GroupId] = stmsgIds[i]

				for gi, gactche := range msg.CreatorUserCHEs {
					gactData := gactche.(map[string]any)

					CHEId := gactData["che_id"].(string)

					newGroupActivityEntries = append(newGroupActivityEntries, CHEId, helpers.ToJson(gactData))

					chatGroupActivities[msg.CreatorUser+" "+msg.GroupId] = append(chatGroupActivities[msg.CreatorUser+" "+msg.GroupId], [2]string{CHEId, fmt.Sprintf("%s-%d", stmsgIds[i], gi)})
				}

				for imi, initMem := range msg.InitMembers {
					initMem := initMem.(string)
					newUserChats[initMem] = append(newUserChats[initMem], msg.GroupId, helpers.ToJson(map[string]any{"type": "group", "group": msg.GroupId}))

					if userChats[initMem] == nil {
						userChats[initMem] = make(map[string]string)
					}

					userChats[initMem][msg.GroupId] = fmt.Sprintf("%s-%d", stmsgIds[i], imi)

					for gi, gactche := range msg.InitMembersCHEs[initMem].([]any) {
						gactData := gactche.(map[string]any)

						CHEId := gactData["che_id"].(string)

						newGroupActivityEntries = append(newGroupActivityEntries, CHEId, helpers.ToJson(gactData))

						chatGroupActivities[initMem+" "+msg.GroupId] = append(chatGroupActivities[initMem+" "+msg.GroupId], [2]string{CHEId, fmt.Sprintf("%s-%d", stmsgIds[i], gi)})
					}
				}
			}

			// batch processing
			if err := cache.StoreNewGroups(ctx, newGroups); err != nil {
				return
			}

			if err := cache.StoreGroupChatHistoryEntries(ctx, newGroupActivityEntries); err != nil {
				return
			}

			eg, sharedCtx := errgroup.WithContext(ctx)

			for groupId, initMembers := range groupInitMembers {
				eg.Go(func() error {
					groupId, initMembers := groupId, initMembers

					return cache.StoreGroupMembers(sharedCtx, groupId, initMembers)
				})
			}

			for groupId, creatorUser := range groupCreators {
				eg.Go(func() error {
					groupId, creatorUser := groupId, creatorUser

					return cache.StoreGroupAdmins(sharedCtx, groupId, []any{creatorUser})
				})
			}

			for ownerUser, groupIdWithChatInfoPairs := range newUserChats {
				eg.Go(func() error {
					ownerUser, groupIdWithChatInfoPairs := ownerUser, groupIdWithChatInfoPairs

					return cache.StoreNewUserChats(sharedCtx, ownerUser, groupIdWithChatInfoPairs)
				})
			}

			for ownerUser, groupId_stmsgId_Pairs := range userChats {
				eg.Go(func() error {
					ownerUser, groupId_stmsgId_Pairs := ownerUser, groupId_stmsgId_Pairs

					return cache.StoreUserChatIdents(sharedCtx, ownerUser, groupId_stmsgId_Pairs)
				})
			}

			for ownerUserGroupId, CHEId_stmsgId_Pairs := range chatGroupActivities {
				eg.Go(func() error {
					ownerUserGroupId, CHEId_stmsgId_Pairs := ownerUserGroupId, CHEId_stmsgId_Pairs

					var ownerUser, groupId string

					fmt.Sscanf(ownerUserGroupId, "%s %s", &ownerUser, &groupId)

					return cache.StoreGroupChatHistory(sharedCtx, ownerUser, groupId, CHEId_stmsgId_Pairs)
				})
			}

			if eg.Wait() != nil {
				return
			}

			// acknowledge messages
			if err := rdb.XAck(ctx, streamName, groupName, stmsgIds...).Err(); err != nil {
				helpers.LogError(err)
			}
		}
	}()
}
