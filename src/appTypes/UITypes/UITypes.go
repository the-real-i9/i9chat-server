package UITypes

type ChatPartnerUser struct {
	Username      string `json:"username"`
	ProfilePicUrl string `json:"profile_pic_url"`
}

type ChatSnippet struct {
	PartnerUser any `json:"partner_user,omitempty"`
	GroupInfo   any `json:"group_info,omitempty"`

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

	// appears for message che_type
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

	// appears for reaction che_type
	Reactor any    `json:"reactor,omitempty"`
	Emoji   string `json:"emoji,omitempty"`

	// cursor for pagination
	Cursor float64 `json:"cursor"`
}
