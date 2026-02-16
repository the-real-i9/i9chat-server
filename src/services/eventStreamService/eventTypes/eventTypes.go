package eventTypes

import "i9chat/src/appTypes"

type NewUserEvent struct {
	Username string `redis:"username" msgpack:"username"`
	UserData string `redis:"userData" msgpack:"userData"`
}

type EditUserEvent struct {
	Username    string              `redis:"username" msgpack:"username"`
	UpdateKVMap appTypes.BinableMap `redis:"updateKVMap" msgpack:"updateKVMap"`
}

type UserPresenceChangeEvent struct {
	Username string `redis:"username" msgpack:"username"`
	Presence string `redis:"presence" msgpack:"presence"`
	LastSeen int64  `redis:"lastSeen" msgpack:"lastSeen"`
}

type NewGroupEvent struct {
	CreatorUser     string                `redis:"creatorUser" msgpack:"creatorUser"`
	GroupId         string                `redis:"groupId" msgpack:"groupId"`
	GroupData       string                `redis:"groupData" msgpack:"groupData"`
	ChatCursor      int64                 `redis:"chatCursor" msgpack:"chatCursor"`
	InitMembers     appTypes.BinableSlice `redis:"initMembers" msgpack:"initMembers"`
	CreatorUserCHEs appTypes.BinableSlice `redis:"creatorUserCHEs" msgpack:"creatorUserCHEs"`
	InitMembersCHEs appTypes.BinableMap   `redis:"initMembersCHEs" msgpack:"initMembersCHEs"`
}

type GroupEditEvent struct {
	GroupId       string              `redis:"groupId" msgpack:"groupId"`
	EditorUser    string              `redis:"editorUser" msgpack:"editorUser"`
	UpdateKVMap   appTypes.BinableMap `redis:"updateKVMap" msgpack:"updateKVMap"`
	EditorUserCHE appTypes.BinableMap `redis:"editorUserCHE" msgpack:"editorUserCHE"`
	MemInfo       string              `redis:"memInfo" msgpack:"memInfo"`
}

type GroupUsersAddedEvent struct {
	GroupId       string                `redis:"groupId" msgpack:"groupId"`
	Admin         string                `redis:"admin" msgpack:"admin"`
	NewMembers    appTypes.BinableSlice `redis:"newMembers" msgpack:"newMembers"`
	AdminCHE      appTypes.BinableMap   `redis:"adminCHE" msgpack:"adminCHE"`
	NewMembersCHE appTypes.BinableMap   `redis:"newMembersCHE" msgpack:"newMembersCHE"`
	MemInfo       string                `redis:"memInfo" msgpack:"memInfo"`
	ChatCursor    int64                 `redis:"chatCursor" msgpack:"chatCursor"`
}

type GroupUserRemovedEvent struct {
	GroupId      string              `redis:"groupId" msgpack:"groupId"`
	Admin        string              `redis:"admin" msgpack:"admin"`
	OldMember    string              `redis:"oldMember" msgpack:"oldMember"`
	AdminCHE     appTypes.BinableMap `redis:"adminCHE" msgpack:"adminCHE"`
	OldMemberCHE appTypes.BinableMap `redis:"oldMemberCHE" msgpack:"oldMemberCHE"`
	MemInfo      string              `redis:"memInfo" msgpack:"memInfo"`
}

type GroupUserJoinedEvent struct {
	GroupId      string              `redis:"groupId" msgpack:"groupId"`
	NewMember    string              `redis:"newMember" msgpack:"newMember"`
	NewMemberCHE appTypes.BinableMap `redis:"newMemberCHE" msgpack:"newMemberCHE"`
	MemInfo      string              `redis:"memInfo" msgpack:"memInfo"`
	ChatCursor   int64               `redis:"chatCursor" msgpack:"chatCursor"`
}

type GroupUserLeftEvent struct {
	GroupId      string              `redis:"groupId" msgpack:"groupId"`
	OldMember    string              `redis:"oldMember" msgpack:"oldMember"`
	OldMemberCHE appTypes.BinableMap `redis:"oldMemberCHE" msgpack:"oldMemberCHE"`
	MemInfo      string              `redis:"memInfo" msgpack:"memInfo"`
}

type GroupMakeUserAdminEvent struct {
	GroupId     string              `redis:"groupId" msgpack:"groupId"`
	Admin       string              `redis:"admin" msgpack:"admin"`
	NewAdmin    string              `redis:"newAdmin" msgpack:"newAdmin"`
	AdminCHE    appTypes.BinableMap `redis:"adminCHE" msgpack:"adminCHE"`
	NewAdminCHE appTypes.BinableMap `redis:"newAdminCHE" msgpack:"newAdminCHE"`
	MemInfo     string              `redis:"memInfo" msgpack:"memInfo"`
}

type GroupRemoveUserFromAdminsEvent struct {
	GroupId     string              `redis:"groupId" msgpack:"groupId"`
	Admin       string              `redis:"admin" msgpack:"admin"`
	OldAdmin    string              `redis:"oldAdmin" msgpack:"oldAdmin"`
	AdminCHE    appTypes.BinableMap `redis:"adminCHE" msgpack:"adminCHE"`
	OldAdminCHE appTypes.BinableMap `redis:"oldAdminCHE" msgpack:"oldAdminCHE"`
	MemInfo     string              `redis:"memInfo" msgpack:"memInfo"`
}

type NewDirectMessageEvent struct {
	FirstFromUser bool   `redis:"ffu" msgpack:"ffu"`
	FirstToUser   bool   `redis:"ftu" msgpack:"ftu"`
	FromUser      string `redis:"fromUser" msgpack:"fromUser"`
	ToUser        string `redis:"toUser" msgpack:"toUser"`
	CHEId         string `redis:"CHEId" msgpack:"CHEId"`
	MsgData       string `redis:"msgData" msgpack:"msgData"`
	CHECursor     int64  `redis:"cheCursor" msgpack:"cheCursor"`
}

type NewGroupMessageEvent struct {
	FromUser  string `redis:"fromUser" msgpack:"fromUser"`
	ToGroup   string `redis:"toGroup" msgpack:"toGroup"`
	CHEId     string `redis:"CHEId" msgpack:"CHEId"`
	MsgData   string `redis:"msgData" msgpack:"msgData"`
	CHECursor int64  `redis:"cheCursor" msgpack:"cheCursor"`
}

type NewDirectMsgReactionEvent struct {
	FromUser  string `redis:"fromUser" msgpack:"fromUser"`
	ToUser    string `redis:"toUser" msgpack:"toUser"`
	CHEId     string `redis:"CHEId" msgpack:"CHEId"`
	RxnData   string `redis:"rxnData" msgpack:"rxnData"`
	ToMsgId   string `redis:"toMsgId" msgpack:"toMsgId"`
	Emoji     string `redis:"emoji" msgpack:"emoji"`
	CHECursor int64  `redis:"cheCursor" msgpack:"cheCursor"`
}

type NewGroupMsgReactionEvent struct {
	FromUser  string `redis:"fromUser" msgpack:"fromUser"`
	ToGroup   string `redis:"toGroup" msgpack:"toGroup"`
	CHEId     string `redis:"CHEId" msgpack:"CHEId"`
	RxnData   string `redis:"rxnData" msgpack:"rxnData"`
	ToMsgId   string `redis:"toMsgId" msgpack:"toMsgId"`
	Emoji     string `redis:"emoji" msgpack:"emoji"`
	CHECursor int64  `redis:"cheCursor" msgpack:"cheCursor"`
}

type DirectMsgAckEvent struct {
	FromUser   string                `redis:"fromUser" msgpack:"fromUser"`
	ToUser     string                `redis:"toUser" msgpack:"toUser"`
	CHEIds     appTypes.BinableSlice `redis:"CHEIds" msgpack:"CHEIds"`
	Ack        string                `redis:"ack" msgpack:"ack"`
	At         int64                 `redis:"at" msgpack:"at"`
	ChatCursor int64                 `redis:"chatCursor" msgpack:"chatCursor"`
}

type GroupMsgAckEvent struct {
	FromUser   string                `redis:"fromUser" msgpack:"fromUser"`
	ToGroup    string                `redis:"toGroup" msgpack:"toGroup"`
	CHEIds     appTypes.BinableSlice `redis:"CHEIds" msgpack:"CHEIds"`
	Ack        string                `redis:"ack" msgpack:"ack"`
	At         int64                 `redis:"at" msgpack:"at"`
	ChatCursor int64                 `redis:"chatCursor" msgpack:"chatCursor"`
}

type DirectMsgReactionRemovedEvent struct {
	FromUser string `redis:"fromUser" msgpack:"fromUser"`
	ToUser   string `redis:"toUser" msgpack:"toUser"`
	ToMsgId  string `redis:"toMsgId" msgpack:"toMsgId"`
	CHEId    string `redis:"CHEId" msgpack:"CHEId"`
}

type GroupMsgReactionRemovedEvent struct {
	FromUser string `redis:"fromUser" msgpack:"fromUser"`
	ToGroup  string `redis:"toGroup" msgpack:"toGroup"`
	ToMsgId  string `redis:"toMsgId" msgpack:"toMsgId"`
	CHEId    string `redis:"CHEId" msgpack:"CHEId"`
}

type MsgDeletionEvent struct {
	CHEId string `redis:"CHEId" msgpack:"CHEId"`
	For   string `redis:"for" msgpack:"for"`
}
