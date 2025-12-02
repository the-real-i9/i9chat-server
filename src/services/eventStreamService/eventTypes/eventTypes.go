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

type NewDirectMessageEvent struct {
	FirstFromUser bool   `redis:"ffu" json:"ffu"`
	FirstToUser   bool   `redis:"ftu" json:"ftu"`
	FromUser      string `redis:"fromUser" json:"fromUser"`
	ToUser        string `redis:"toUser" json:"toUser"`
	CHEId         string `redis:"CHEId" json:"CHEId"`
	MsgData       string `redis:"msgData" json:"msgData"`
}

type NewGroupEvent struct {
	CreatorUser     string                `redis:"creatorUser" json:"creatorUser"`
	GroupId         string                `redis:"groupId" json:"groupId"`
	GroupData       string                `redis:"groupData" json:"groupData"`
	InitMembers     appTypes.BinableSlice `redis:"initMembers" json:"initMembers"`
	CreatorUserCHEs appTypes.BinableSlice `redis:"creatorUserCHEs" json:"creatorUserCHEs"`
	InitMembersCHEs appTypes.BinableMap   `redis:"initMembersCHEs" json:"initMembersCHEs"`
}

type GroupEditEvent struct {
	GroupId        string                `redis:"groupId" json:"groupId"`
	EditorUser     string                `redis:"editorUser" json:"editorUser"`
	UpdateKVMap    appTypes.BinableMap   `redis:"updateKVMap" json:"updateKVMap"`
	MemberUsers    appTypes.BinableSlice `redis:"memberUsers" json:"memberUsers"`
	EditorUserCHE  appTypes.BinableMap   `redis:"editorUserCHE" json:"editorUserCHE"`
	MemberUsersCHE appTypes.BinableMap   `redis:"memberUsersCHE" json:"memberUsersCHE"`
}

type GroupUsersAddedEvent struct {
	GroupId        string                `redis:"groupId" json:"groupId"`
	AdminUser      string                `redis:"adminUser" json:"adminUser"`
	NewMembers     appTypes.BinableSlice `redis:"newMembers" json:"newMembers"`
	MemberUsers    appTypes.BinableSlice `redis:"memberUsers" json:"memberUsers"`
	AdminUserCHE   appTypes.BinableMap   `redis:"adminUserCHE" json:"adminUserCHE"`
	NewMembersCHE  appTypes.BinableMap   `redis:"newMembersCHE" json:"newMembersCHE"`
	MemberUsersCHE appTypes.BinableMap   `redis:"memberUsersCHE" json:"memberUsersCHE"`
}

type GroupUserJoinedEvent struct {
	GroupId string `redis:"groupId" json:"groupId"`
	User    string `redis:"user" json:"user"`
}

type GroupUserRemovedEvent struct {
	GroupId string `redis:"groupId" json:"groupId"`
	User    string `redis:"user" json:"user"`
}

type GroupUserLeftEvent struct {
	GroupId string `redis:"groupId" json:"groupId"`
	User    string `redis:"user" json:"user"`
}

type GroupMakeUserAdminEvent struct {
	GroupId string `redis:"groupId" json:"groupId"`
	User    string `redis:"user" json:"user"`
}

type GroupRemoveUserFromAdminsEvent struct {
	GroupId string `redis:"groupId" json:"groupId"`
	User    string `redis:"user" json:"user"`
}

type NewGroupMessageEvent struct {
	FromUser string `redis:"fromUser" json:"fromUser"`
	ToGroup  string `redis:"toGroup" json:"toGroup"`
	CHEId    string `redis:"CHEId" json:"CHEId"`
	MsgData  string `redis:"msgData" json:"msgData"`
}

type NewDirectMsgReactionEvent struct {
	FromUser string `redis:"fromUser" json:"fromUser"`
	ToUser   string `redis:"toUser" json:"toUser"`
	CHEId    string `redis:"CHEId" json:"CHEId"`
	RxnData  string `redis:"rxnData" json:"rxnData"`
	ToMsgId  string `redis:"toMsgId" json:"toMsgId"`
	Emoji    string `redis:"emoji" json:"emoji"`
}

type NewGroupMsgReactionEvent struct {
	FromUser string `redis:"fromUser" json:"fromUser"`
	ToGroup  string `redis:"toGroup" json:"toGroup"`
	CHEId    string `redis:"CHEId" json:"CHEId"`
	RxnData  string `redis:"rxnData" json:"rxnData"`
	ToMsgId  string `redis:"toMsgId" json:"toMsgId"`
	Emoji    string `redis:"emoji" json:"emoji"`
}

type DirectMsgAckEvent struct {
	FromUser string `redis:"fromUser" json:"fromUser"`
	ToUser   string `redis:"toUser" json:"toUser"`
	CHEId    string `redis:"CHEId" json:"CHEId"`
	Ack      string `redis:"ack" json:"ack"`
	At       int64  `redis:"at" json:"at"`
}

type GroupMsgAckEvent struct {
	FromUser string `redis:"fromUser" json:"fromUser"`
	ToGroup  string `redis:"toGroup" json:"toGroup"`
	CHEId    string `redis:"CHEId" json:"CHEId"`
	Ack      string `redis:"ack" json:"ack"`
	At       int64  `redis:"at" json:"at"`
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
