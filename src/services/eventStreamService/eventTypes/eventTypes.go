package eventTypes

import "i9chat/src/appTypes"

type NewUserEvent struct {
	Username string `redis:"username"`
	UserData string `redis:"userData"`
}

type EditUserEvent struct {
	Username    string              `redis:"username"`
	UpdateKVMap appTypes.BinableMap `redis:"updateKVMap"`
}

type UserPresenceChangeEvent struct {
	Username string `redis:"username"`
	Presence string `redis:"presence"`
	LastSeen int64  `redis:"lastSeen"`
}

type NewGroupEvent struct {
	CreatorUser     string                `redis:"creatorUser"`
	GroupId         string                `redis:"groupId"`
	GroupData       string                `redis:"groupData"`
	ChatCursor      int64                 `redis:"chatCursor"`
	InitMembers     appTypes.BinableSlice `redis:"initMembers"`
	CreatorUserCHEs appTypes.BinableSlice `redis:"creatorUserCHEs"`
	InitMembersCHEs appTypes.BinableMap   `redis:"initMembersCHEs"`
}

type GroupEditEvent struct {
	GroupId       string              `redis:"groupId"`
	EditorUser    string              `redis:"editorUser"`
	UpdateKVMap   appTypes.BinableMap `redis:"updateKVMap"`
	EditorUserCHE appTypes.BinableMap `redis:"editorUserCHE"`
	MemInfo       string              `redis:"memInfo"`
}

type GroupUsersAddedEvent struct {
	GroupId       string                `redis:"groupId"`
	Admin         string                `redis:"admin"`
	NewMembers    appTypes.BinableSlice `redis:"newMembers"`
	AdminCHE      appTypes.BinableMap   `redis:"adminCHE"`
	NewMembersCHE appTypes.BinableMap   `redis:"newMembersCHE"`
	MemInfo       string                `redis:"memInfo"`
	ChatCursor    int64                 `redis:"chatCursor"`
}

type GroupUserRemovedEvent struct {
	GroupId      string              `redis:"groupId"`
	Admin        string              `redis:"admin"`
	OldMember    string              `redis:"oldMember"`
	AdminCHE     appTypes.BinableMap `redis:"adminCHE"`
	OldMemberCHE appTypes.BinableMap `redis:"oldMemberCHE"`
	MemInfo      string              `redis:"memInfo"`
}

type GroupUserJoinedEvent struct {
	GroupId      string              `redis:"groupId"`
	NewMember    string              `redis:"newMember"`
	NewMemberCHE appTypes.BinableMap `redis:"newMemberCHE"`
	MemInfo      string              `redis:"memInfo"`
	ChatCursor   int64               `redis:"chatCursor"`
}

type GroupUserLeftEvent struct {
	GroupId      string              `redis:"groupId"`
	OldMember    string              `redis:"oldMember"`
	OldMemberCHE appTypes.BinableMap `redis:"oldMemberCHE"`
	MemInfo      string              `redis:"memInfo"`
}

type GroupMakeUserAdminEvent struct {
	GroupId     string              `redis:"groupId"`
	Admin       string              `redis:"admin"`
	NewAdmin    string              `redis:"newAdmin"`
	AdminCHE    appTypes.BinableMap `redis:"adminCHE"`
	NewAdminCHE appTypes.BinableMap `redis:"newAdminCHE"`
	MemInfo     string              `redis:"memInfo"`
}

type GroupRemoveUserFromAdminsEvent struct {
	GroupId     string              `redis:"groupId"`
	Admin       string              `redis:"admin"`
	OldAdmin    string              `redis:"oldAdmin"`
	AdminCHE    appTypes.BinableMap `redis:"adminCHE"`
	OldAdminCHE appTypes.BinableMap `redis:"oldAdminCHE"`
	MemInfo     string              `redis:"memInfo"`
}

type NewDirectMessageEvent struct {
	FirstFromUser bool   `redis:"ffu"`
	FirstToUser   bool   `redis:"ftu"`
	FromUser      string `redis:"fromUser"`
	ToUser        string `redis:"toUser"`
	CHEId         string `redis:"CHEId"`
	MsgData       string `redis:"msgData"`
	CHECursor     int64  `redis:"cheCursor"`
}

type NewGroupMessageEvent struct {
	FromUser  string `redis:"fromUser"`
	ToGroup   string `redis:"toGroup"`
	CHEId     string `redis:"CHEId"`
	MsgData   string `redis:"msgData"`
	CHECursor int64  `redis:"cheCursor"`
}

type NewDirectMsgReactionEvent struct {
	FromUser  string `redis:"fromUser"`
	ToUser    string `redis:"toUser"`
	CHEId     string `redis:"CHEId"`
	RxnData   string `redis:"rxnData"`
	ToMsgId   string `redis:"toMsgId"`
	Emoji     string `redis:"emoji"`
	CHECursor int64  `redis:"cheCursor"`
}

type NewGroupMsgReactionEvent struct {
	FromUser  string `redis:"fromUser"`
	ToGroup   string `redis:"toGroup"`
	CHEId     string `redis:"CHEId"`
	RxnData   string `redis:"rxnData"`
	ToMsgId   string `redis:"toMsgId"`
	Emoji     string `redis:"emoji"`
	CHECursor int64  `redis:"cheCursor"`
}

type DirectMsgAckEvent struct {
	FromUser   string                `redis:"fromUser"`
	ToUser     string                `redis:"toUser"`
	CHEIds     appTypes.BinableSlice `redis:"CHEIds"`
	Ack        string                `redis:"ack"`
	At         int64                 `redis:"at"`
	ChatCursor int64                 `redis:"chatCursor"`
}

type GroupMsgAckEvent struct {
	FromUser      string                `redis:"fromUser"`
	ToGroup       string                `redis:"toGroup"`
	CHEIds        appTypes.BinableSlice `redis:"CHEIds"`
	Ack           string                `redis:"ack"`
	At            int64                 `redis:"at"`
	ChatCursor    int64                 `redis:"chatCursor"`
	MsgIdtoSender appTypes.BinableSlice `redis:"msgIdtoSender"`
}

type DirectMsgReactionRemovedEvent struct {
	FromUser string `redis:"fromUser"`
	ToUser   string `redis:"toUser"`
	ToMsgId  string `redis:"toMsgId"`
	CHEId    string `redis:"CHEId"`
}

type GroupMsgReactionRemovedEvent struct {
	FromUser string `redis:"fromUser"`
	ToGroup  string `redis:"toGroup"`
	ToMsgId  string `redis:"toMsgId"`
	CHEId    string `redis:"CHEId"`
}

type MsgDeletionEvent struct {
	CHEId string `redis:"CHEId"`
	For   string `redis:"for"`
}
