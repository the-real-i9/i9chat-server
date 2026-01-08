package UITypes

type ClientUser struct {
	Username      string `json:"username"`
	ProfilePicUrl string `json:"profile_pic_url"`
	Presence      string `json:"presence"`
}

type UserSnippet struct {
	Username      string `json:"username" db:"username"`
	ProfilePicUrl string `json:"profile_pic_url" db:"profile_pic_url"`
	Bio           string `json:"bio" db:"bio"`
	Presence      string `json:"presence" db:"presence"`
	LastSeen      int64  `json:"last_seen" db:"last_seen"`
}

type UserProfile struct {
	Username      string         `json:"username"`
	Name          string         `json:"name"`
	ProfilePicUrl string         `json:"profile_pic_url"`
	Bio           string         `json:"bio"`
	Geolocation   map[string]any `json:"geolocation"`
}

type GroupInfo struct {
	Id                 string `json:"id"`
	Name               string `json:"name"`
	PictureUrl         string `json:"picture_url"`
	Description        string `json:"description"`
	CreatedAt          int64  `json:"created_at"`
	MembersCount       int64  `json:"members_count"`
	OnlineMembersCount int    `json:"online_members_count"`
}

type GroupMemberSnippet struct {
	Username      string  `json:"username"`
	ProfilePicUrl string  `json:"profile_pic_url"`
	Bio           string  `json:"bio"`
	Cursor        float64 `json:"cursor"`
}

type ChatPartnerUser struct {
	Username      string `json:"username"`
	ProfilePicUrl string `json:"profile_pic_url"`
}

type ChatGroup struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	PictureUrl string `json:"picture_url"`
}

type ChatSnippet struct {
	Type string `json:"type"`

	PartnerUser any `json:"partner_user,omitempty"` /* stored as partnerUsername, then retrieved ChatPartnerUser */
	Group       any `json:"group,omitempty"`        /* stored as groupId, then retrieved ChatGroup */

	UnreadMC int64   `json:"unread_messages_count"`
	Cursor   float64 `json:"cursor"`
}

type msgReactorKind interface {
	MsgReactor | any
}

type MsgReactor struct {
	Username      string `json:"username"`
	ProfilePicUrl string `json:"profile_pic_url"`
}

type MsgSender struct {
	Username      string `json:"username"`
	ProfilePicUrl string `json:"profile_pic_url"`
}

type MsgReaction struct {
	Emoji   string         `json:"emoji"`
	Reactor msgReactorKind `json:"reactor"`
}

type ChatHistoryEntry struct {
	// appears always
	CHEType string `json:"che_type"`

	// appears for "message" che_type
	Id             string         `json:"id,omitempty"`
	Content        map[string]any `json:"content,omitempty"`
	DeliveryStatus string         `json:"delivery_status,omitempty"`
	CreatedAt      int64          `json:"created_at,omitempty"`
	DeliveredAt    int64          `json:"delivered_at,omitempty"`
	ReadAt         int64          `json:"read_at,omitempty"`
	Sender         any            `json:"sender,omitempty"`
	ReactionsCount map[string]int `json:"reactions_count,omitempty"`
	Reactions      []MsgReaction  `json:"reactions,omitempty"`

	// appears if che_type:message is a reply
	ReplyTargetMsg map[string]any `json:"reply_target_msg,omitempty"`

	// appears for "reaction" che_type
	Reactor any    `json:"reactor,omitempty"`
	Emoji   string `json:"emoji,omitempty"`
	ToMsgId string `json:"to_msg_id,omitempty"`

	// appears for "group activity" che_type
	Info string `json:"info,omitempty"`

	// cursor for pagination
	Cursor float64 `json:"cursor"`
}
