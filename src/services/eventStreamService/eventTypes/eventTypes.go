package eventTypes

import "i9chat/src/appTypes"

type NewUserEvent struct {
	Username string `redis:"username" json:"username"`
	UserData string `redis:"userData" json:"userData"`
}

type EditUserEvent struct {
	Username    string              `redis:"username" json:"username"`
	UpdateKVMap appTypes.BinableMap `redis:"updateKVMap" json:"updateKVMap"`
}

type UserPresenceChangeEvent struct {
	Username string `redis:"username" json:"username"`
	Presence string `redis:"presence" json:"presence"`
	LastSeen int64  `redis:"lastSeen" json:"lastSeen"`
}

type NewGroupEvent struct {
	CreatorUser     string                `redis:"creatorUser" json:"creatorUser"`
	GroupId         string                `redis:"groupId" json:"groupId"`
	GroupData       string                `redis:"groupData" json:"groupData"`
	ChatCursor      int64                 `redis:"chatCursor" json:"chatCursor"`
	InitMembers     appTypes.BinableSlice `redis:"initMembers" json:"initMembers"`
	CreatorUserCHEs appTypes.BinableSlice `redis:"creatorUserCHEs" json:"creatorUserCHEs"`
	InitMembersCHEs appTypes.BinableMap   `redis:"initMembersCHEs" json:"initMembersCHEs"`
}

type GroupEditEvent struct {
	GroupId       string              `redis:"groupId" json:"groupId"`
	EditorUser    string              `redis:"editorUser" json:"editorUser"`
	UpdateKVMap   appTypes.BinableMap `redis:"updateKVMap" json:"updateKVMap"`
	EditorUserCHE appTypes.BinableMap `redis:"editorUserCHE" json:"editorUserCHE"`
	MemInfo       string              `redis:"memInfo" json:"memInfo"`
}

type GroupUsersAddedEvent struct {
	GroupId       string                `redis:"groupId" json:"groupId"`
	Admin         string                `redis:"admin" json:"admin"`
	NewMembers    appTypes.BinableSlice `redis:"newMembers" json:"newMembers"`
	AdminCHE      appTypes.BinableMap   `redis:"adminCHE" json:"adminCHE"`
	NewMembersCHE appTypes.BinableMap   `redis:"newMembersCHE" json:"newMembersCHE"`
	MemInfo       string                `redis:"memInfo" json:"memInfo"`
	ChatCursor    int64                 `redis:"chatCursor" json:"chatCursor"`
}

type GroupUserRemovedEvent struct {
	GroupId      string              `redis:"groupId" json:"groupId"`
	Admin        string              `redis:"admin" json:"admin"`
	OldMember    string              `redis:"oldMember" json:"oldMember"`
	AdminCHE     appTypes.BinableMap `redis:"adminCHE" json:"adminCHE"`
	OldMemberCHE appTypes.BinableMap `redis:"oldMemberCHE" json:"oldMemberCHE"`
	MemInfo      string              `redis:"memInfo" json:"memInfo"`
}

type GroupUserJoinedEvent struct {
	GroupId      string              `redis:"groupId" json:"groupId"`
	NewMember    string              `redis:"newMember" json:"newMember"`
	NewMemberCHE appTypes.BinableMap `redis:"newMemberCHE" json:"newMemberCHE"`
	MemInfo      string              `redis:"memInfo" json:"memInfo"`
	ChatCursor   int64               `redis:"chatCursor" json:"chatCursor"`
}

type GroupUserLeftEvent struct {
	GroupId      string              `redis:"groupId" json:"groupId"`
	OldMember    string              `redis:"oldMember" json:"oldMember"`
	OldMemberCHE appTypes.BinableMap `redis:"oldMemberCHE" json:"oldMemberCHE"`
	MemInfo      string              `redis:"memInfo" json:"memInfo"`
}

type GroupMakeUserAdminEvent struct {
	GroupId     string              `redis:"groupId" json:"groupId"`
	Admin       string              `redis:"admin" json:"admin"`
	NewAdmin    string              `redis:"newAdmin" json:"newAdmin"`
	AdminCHE    appTypes.BinableMap `redis:"adminCHE" json:"adminCHE"`
	NewAdminCHE appTypes.BinableMap `redis:"newAdminCHE" json:"newAdminCHE"`
	MemInfo     string              `redis:"memInfo" json:"memInfo"`
}

type GroupRemoveUserFromAdminsEvent struct {
	GroupId     string              `redis:"groupId" json:"groupId"`
	Admin       string              `redis:"admin" json:"admin"`
	OldAdmin    string              `redis:"oldAdmin" json:"oldAdmin"`
	AdminCHE    appTypes.BinableMap `redis:"adminCHE" json:"adminCHE"`
	OldAdminCHE appTypes.BinableMap `redis:"oldAdminCHE" json:"oldAdminCHE"`
	MemInfo     string              `redis:"memInfo" json:"memInfo"`
}

type NewDirectMessageEvent struct {
	FirstFromUser bool   `redis:"ffu" json:"ffu"`
	FirstToUser   bool   `redis:"ftu" json:"ftu"`
	FromUser      string `redis:"fromUser" json:"fromUser"`
	ToUser        string `redis:"toUser" json:"toUser"`
	CHEId         string `redis:"CHEId" json:"CHEId"`
	MsgData       string `redis:"msgData" json:"msgData"`
	CHECursor     int64  `redis:"cheCursor" json:"cheCursor"`
}

type NewGroupMessageEvent struct {
	FromUser  string `redis:"fromUser" json:"fromUser"`
	ToGroup   string `redis:"toGroup" json:"toGroup"`
	CHEId     string `redis:"CHEId" json:"CHEId"`
	MsgData   string `redis:"msgData" json:"msgData"`
	CHECursor int64  `redis:"cheCursor" json:"cheCursor"`
}

type NewDirectMsgReactionEvent struct {
	FromUser  string `redis:"fromUser" json:"fromUser"`
	ToUser    string `redis:"toUser" json:"toUser"`
	CHEId     string `redis:"CHEId" json:"CHEId"`
	RxnData   string `redis:"rxnData" json:"rxnData"`
	ToMsgId   string `redis:"toMsgId" json:"toMsgId"`
	Emoji     string `redis:"emoji" json:"emoji"`
	CHECursor int64  `redis:"cheCursor" json:"cheCursor"`
}

type NewGroupMsgReactionEvent struct {
	FromUser  string `redis:"fromUser" json:"fromUser"`
	ToGroup   string `redis:"toGroup" json:"toGroup"`
	CHEId     string `redis:"CHEId" json:"CHEId"`
	RxnData   string `redis:"rxnData" json:"rxnData"`
	ToMsgId   string `redis:"toMsgId" json:"toMsgId"`
	Emoji     string `redis:"emoji" json:"emoji"`
	CHECursor int64  `redis:"cheCursor" json:"cheCursor"`
}

type DirectMsgAckEvent struct {
	FromUser   string `redis:"fromUser" json:"fromUser"`
	ToUser     string `redis:"toUser" json:"toUser"`
	CHEId      string `redis:"CHEId" json:"CHEId"`
	Ack        string `redis:"ack" json:"ack"`
	At         int64  `redis:"at" json:"at"`
	ChatCursor int64  `redis:"chatCursor" json:"chatCursor"`
}

type GroupMsgAckEvent struct {
	FromUser   string `redis:"fromUser" json:"fromUser"`
	ToGroup    string `redis:"toGroup" json:"toGroup"`
	CHEId      string `redis:"CHEId" json:"CHEId"`
	Ack        string `redis:"ack" json:"ack"`
	At         int64  `redis:"at" json:"at"`
	ChatCursor int64  `redis:"chatCursor" json:"chatCursor"`
}

type DirectMsgReactionRemovedEvent struct {
	FromUser string `redis:"fromUser" json:"fromUser"`
	ToUser   string `redis:"toUser" json:"toUser"`
	ToMsgId  string `redis:"toMsgId" json:"toMsgId"`
	CHEId    string `redis:"CHEId" json:"CHEId"`
}

type GroupMsgReactionRemovedEvent struct {
	FromUser string `redis:"fromUser" json:"fromUser"`
	ToGroup  string `redis:"toGroup" json:"toGroup"`
	ToMsgId  string `redis:"toMsgId" json:"toMsgId"`
	CHEId    string `redis:"CHEId" json:"CHEId"`
}

type MsgDeletionEvent struct {
	CHEId string `redis:"CHEId" json:"CHEId"`
	For   string `redis:"for" json:"for"`
}
