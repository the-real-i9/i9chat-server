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
	Admin          string                `redis:"admin" json:"admin"`
	NewMembers     appTypes.BinableSlice `redis:"newMembers" json:"newMembers"`
	MemberUsers    appTypes.BinableSlice `redis:"memberUsers" json:"memberUsers"`
	AdminCHE       appTypes.BinableMap   `redis:"adminCHE" json:"adminCHE"`
	NewMembersCHE  appTypes.BinableMap   `redis:"newMembersCHE" json:"newMembersCHE"`
	MemberUsersCHE appTypes.BinableMap   `redis:"memberUsersCHE" json:"memberUsersCHE"`
}

type GroupUserRemovedEvent struct {
	GroupId        string                `redis:"groupId" json:"groupId"`
	Admin          string                `redis:"admin" json:"admin"`
	OldMember      string                `redis:"oldMember" json:"oldMember"`
	MemberUsers    appTypes.BinableSlice `redis:"memberUsers" json:"memberUsers"`
	AdminCHE       appTypes.BinableMap   `redis:"adminCHE" json:"adminCHE"`
	OldMemberCHE   appTypes.BinableMap   `redis:"oldMemberCHE" json:"oldMemberCHE"`
	MemberUsersCHE appTypes.BinableMap   `redis:"memberUsersCHE" json:"memberUsersCHE"`
}

type GroupUserJoinedEvent struct {
	GroupId        string                `redis:"groupId" json:"groupId"`
	NewMember      string                `redis:"newMember" json:"newMember"`
	MemberUsers    appTypes.BinableSlice `redis:"memberUsers" json:"memberUsers"`
	NewMemberCHE   appTypes.BinableMap   `redis:"newMemberCHE" json:"newMemberCHE"`
	MemberUsersCHE appTypes.BinableMap   `redis:"memberUsersCHE" json:"memberUsersCHE"`
}

type GroupUserLeftEvent struct {
	GroupId        string                `redis:"groupId" json:"groupId"`
	OldMember      string                `redis:"oldMember" json:"oldMember"`
	MemberUsers    appTypes.BinableSlice `redis:"memberUsers" json:"memberUsers"`
	OldMemberCHE   appTypes.BinableMap   `redis:"oldMemberCHE" json:"oldMemberCHE"`
	MemberUsersCHE appTypes.BinableMap   `redis:"memberUsersCHE" json:"memberUsersCHE"`
}

type GroupMakeUserAdminEvent struct {
	GroupId        string                `redis:"groupId" json:"groupId"`
	Admin          string                `redis:"admin" json:"admin"`
	NewAdmin       string                `redis:"newAdmin" json:"newAdmin"`
	MemberUsers    appTypes.BinableSlice `redis:"memberUsers" json:"memberUsers"`
	AdminCHE       appTypes.BinableMap   `redis:"adminCHE" json:"adminCHE"`
	NewAdminCHE    appTypes.BinableMap   `redis:"newAdminCHE" json:"newAdminCHE"`
	MemberUsersCHE appTypes.BinableMap   `redis:"memberUsersCHE" json:"memberUsersCHE"`
}

type GroupRemoveUserFromAdminsEvent struct {
	GroupId        string                `redis:"groupId" json:"groupId"`
	Admin          string                `redis:"admin" json:"admin"`
	OldAdmin       string                `redis:"oldAdmin" json:"oldAdmin"`
	MemberUsers    appTypes.BinableSlice `redis:"memberUsers" json:"memberUsers"`
	AdminCHE       appTypes.BinableMap   `redis:"adminCHE" json:"adminCHE"`
	OldAdminCHE    appTypes.BinableMap   `redis:"oldAdminCHE" json:"oldAdminCHE"`
	MemberUsersCHE appTypes.BinableMap   `redis:"memberUsersCHE" json:"memberUsersCHE"`
}

type NewDirectMessageEvent struct {
	FirstFromUser bool   `redis:"ffu" json:"ffu"`
	FirstToUser   bool   `redis:"ftu" json:"ftu"`
	FromUser      string `redis:"fromUser" json:"fromUser"`
	ToUser        string `redis:"toUser" json:"toUser"`
	CHEId         string `redis:"CHEId" json:"CHEId"`
	MsgData       string `redis:"msgData" json:"msgData"`
}

type NewGroupMessageEvent struct {
	FromUser    string                `redis:"fromUser" json:"fromUser"`
	ToGroup     string                `redis:"toGroup" json:"toGroup"`
	CHEId       string                `redis:"CHEId" json:"CHEId"`
	MsgData     string                `redis:"msgData" json:"msgData"`
	MemberUsers appTypes.BinableSlice `redis:"memberUsers" json:"memberUsers"`
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
