package backgroundWorkers

import "github.com/redis/go-redis/v9"

func Start(rdb *redis.Client) {
	newUsersStreamBgWorker(rdb)
	userEditsStreamBgWorker(rdb)
	userPresenceChangesStreamBgWorker(rdb)

	newDirectMessagesStreamBgWorker(rdb)
	directMsgAcksStreamBgWorker(rdb)

	newGroupsStreamBgWorker(rdb)
	groupEditsStreamBgWorker(rdb)
	groupUsersAddedStreamBgWorker(rdb)
	groupUsersRemovedStreamBgWorker(rdb)
	groupUsersJoinedStreamBgWorker(rdb)
	groupUsersLeftStreamBgWorker(rdb)
	groupNewAdminsStreamBgWorker(rdb)
	groupRemovedAdminsStreamBgWorker(rdb)

	newGroupMessagesStreamBgWorker(rdb)
	groupMsgAcksStreamBgWorker(rdb)
}
