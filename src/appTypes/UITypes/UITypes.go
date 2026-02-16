package UITypes

type ClientUser struct {
	Username      string `msgpack:"username"`
	ProfilePicUrl string `msgpack:"profile_pic_url"`
	Presence      string `msgpack:"presence"`
}

type UserSnippet struct {
	Username      string `msgpack:"username" db:"username"`
	ProfilePicUrl string `msgpack:"profile_pic_url" db:"profile_pic_url"`
	Bio           string `msgpack:"bio" db:"bio"`
	Presence      string `msgpack:"presence" db:"presence"`
	LastSeen      int64  `msgpack:"last_seen" db:"last_seen"`
}

type UserProfile struct {
	Username      string         `msgpack:"username"`
	Name          string         `msgpack:"name"`
	ProfilePicUrl string         `msgpack:"profile_pic_url"`
	Bio           string         `msgpack:"bio"`
	Geolocation   map[string]any `msgpack:"geolocation"`
}

type GroupInfo struct {
	Id                 string `msgpack:"id"`
	Name               string `msgpack:"name"`
	PictureUrl         string `msgpack:"picture_url"`
	Description        string `msgpack:"description"`
	CreatedAt          int64  `msgpack:"created_at"`
	MembersCount       int64  `msgpack:"members_count"`
	OnlineMembersCount int    `msgpack:"online_members_count"`
}

type GroupMemberSnippet struct {
	Username      string  `msgpack:"username"`
	ProfilePicUrl string  `msgpack:"profile_pic_url"`
	Bio           string  `msgpack:"bio"`
	Cursor        float64 `msgpack:"cursor"`
}

type ChatPartnerUser struct {
	Username      string `msgpack:"username"`
	ProfilePicUrl string `msgpack:"profile_pic_url"`
}

type ChatGroup struct {
	Id         string `msgpack:"id"`
	Name       string `msgpack:"name"`
	PictureUrl string `msgpack:"picture_url"`
}

type ChatSnippet struct {
	Type string `msgpack:"type"`

	PartnerUser any `msgpack:"partner_user,omitempty"` /* stored as partnerUsername, then retrieved ChatPartnerUser */
	Group       any `msgpack:"group,omitempty"`        /* stored as groupId, then retrieved ChatGroup */

	UnreadMC int64 `msgpack:"unread_messages_count"`
	Cursor   int64 `msgpack:"cursor"`
}

type MsgReactor struct {
	Username      string `msgpack:"username"`
	ProfilePicUrl string `msgpack:"profile_pic_url"`
}

type MsgSender struct {
	Username      string `msgpack:"username"`
	ProfilePicUrl string `msgpack:"profile_pic_url"`
}

type MsgReaction struct {
	Emoji   string     `msgpack:"emoji"`
	Reactor MsgReactor `msgpack:"reactor"`
}

type ChatHistoryEntry struct {
	// appears always
	CHEType string `msgpack:"che_type"`

	// appears for "message" che_type
	Id             string         `msgpack:"id,omitempty"`
	Content        map[string]any `msgpack:"content,omitempty"`
	DeliveryStatus string         `msgpack:"delivery_status,omitempty"`
	CreatedAt      int64          `msgpack:"created_at,omitempty"`
	DeliveredAt    int64          `msgpack:"delivered_at,omitempty"`
	ReadAt         int64          `msgpack:"read_at,omitempty"`
	Sender         any            `msgpack:"sender,omitempty"`
	ReactionsCount map[string]int `msgpack:"reactions_count,omitempty"`
	Reactions      []MsgReaction  `msgpack:"reactions,omitempty"`

	// appears if che_type:message is a reply
	ReplyTargetMsg map[string]any `msgpack:"reply_target_msg,omitempty"`

	// appears for "reaction" che_type
	Reactor any    `msgpack:"reactor,omitempty"`
	Emoji   string `msgpack:"emoji,omitempty"`
	ToMsgId string `msgpack:"to_msg_id,omitempty"`

	// appears for "group activity" che_type
	Info string `msgpack:"info,omitempty"`

	// cursor for pagination
	Cursor int64 `msgpack:"cursor"`
}
