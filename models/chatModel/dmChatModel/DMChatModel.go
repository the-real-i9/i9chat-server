package dmChat

import (
	"fmt"
	user "i9chat/models/userModel"
	"i9chat/utils/appTypes"
	"i9chat/utils/helpers"
	"log"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type InitiatorData struct {
	NewDMChatId int `db:"new_dm_chat_id" json:"new_dm_chat_id"`
	InitMsgId   int `db:"init_msg_id" json:"init_msg_id"`
}

type PartnerData struct {
	Type     string    `json:"typm"`
	DMChatId int       `db:"dm_chat_id" json:"dm_chat_id"`
	Partner  user.User `json:"partner"`
	InitMsg  struct {
		Id      int            `json:"id"`
		Content map[string]any `json:"content"`
	} `db:"init_msg" json:"init_msg"`
}

type NewDMChat struct {
	*InitiatorData `db:"ird"`
	*PartnerData   `db:"prd"`
}

func New(initiatorId int, partnerId int, initMsgContent map[string]any, createdAt time.Time) (*NewDMChat, error) {
	newDMChat, err := helpers.QueryRowType[NewDMChat]("SELECT initiator_resp_data AS ird, partner_resp_data AS prd FROM new_dm_chat($1, $2, $3, $4)", initiatorId, partnerId, initMsgContent, createdAt)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: NewDMChat: %s", err))
		return nil, helpers.ErrInternalServerError
	}

	return newDMChat, nil
}

type SenderData struct {
	NewMsgId int `db:"new_msg_id" json:"new_msg_id"`
}

type ReceiverData struct {
	MsgId    int            `db:"msg_id" json:"msg_im"`
	DMChatId int            `db:"dm_chat_id" json:"dm_chat_id"`
	Sender   user.User      `json:"sender"`
	Content  map[string]any `json:"content"`
}

type NewMessage struct {
	*SenderData   `db:"srd"`
	*ReceiverData `db:"rrd"`
	ReceiverId    int `db:"receiver_id"`
}

func SendMessage(dmChatId, senderId int, msgContent map[string]any, createdAt time.Time) (*NewMessage, error) {
	newMessage, err := helpers.QueryRowType[NewMessage]("SELECT sender_resp_data AS srd, receiver_resp_data AS rrd, receiver_id FROM send_dm_chat_message($1, $2, $3, $4)", dmChatId, senderId, msgContent, createdAt)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: DMChat_SendMessage: %s", err))
		return nil, helpers.ErrInternalServerError
	}

	return newMessage, nil
}

func ReactToMessage(dmChatId, msgId, reactorId int, reaction rune) error {
	_, err := helpers.QueryRowField[bool]("SELECT react_to_dm_chat_message($1, $2, $3, $4)", dmChatId, msgId, reactorId, strconv.QuoteRuneToASCII(reaction))
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: ReactToMessage: %s", err))
		return helpers.ErrInternalServerError
	}

	return nil
}

type messageReaction struct {
	Reaction rune      `json:"reaction,omitempty"`
	Reactor  user.User `json:"reactor,omitempty"`
}

type Message struct {
	Id             int               `json:"id"`
	Sender         user.User         `json:"sender"`
	Content        map[string]any    `json:"content"`
	DeliveryStatus string            `db:"delivery_status" json:"delivery_status"`
	CreatedAt      pgtype.Timestamp  `db:"created_at" json:"created_at"`
	Edited         bool              `json:"edited"`
	EditedAt       pgtype.Timestamp  `db:"edited_at" json:"edited_at"`
	Reactions      []messageReaction `json:"reactions"`
}

func GetChatHistory(dmChatId, offset int) ([]*Message, error) {
	messages, err := helpers.QueryRowsType[Message](`
	SELECT message FROM (
		SELECT message, created_at FROM get_dm_chat_history($1) 
		LIMIT 50 OFFSET $2
	) ORDER BY created_at ASC`, dmChatId, offset)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: GetChatHistory: %s", err))
		return nil, helpers.ErrInternalServerError
	}

	return messages, nil
}

func BatchUpdateMessageDeliveryStatus(receiverId int, status string, ackDatas []*appTypes.DMChatMsgAckData) error {
	var sqls = []string{}
	var params = [][]any{}

	for _, data := range ackDatas {
		sqls = append(sqls, "SELECT update_dm_chat_message_delivery_status($1, $2, $3, $4, $5)")
		params = append(params, []any{data.DMChatId, data.MsgId, receiverId, status, data.At})
	}

	_, err := helpers.BatchQuery[bool](sqls, params)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: BatchUpdateMessageDeliveryStatus: %s", err))
		return helpers.ErrInternalServerError
	}

	return nil
}

func UpdateMessageDeliveryStatus(dmChatId, msgId, receiverId int, status string, updatedAt time.Time) error {
	_, err := helpers.QueryRowField[bool]("SELECT update_dm_chat_message_delivery_status($1, $2, $3, $4, $5)", dmChatId, msgId, receiverId, status, updatedAt)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: UpdateMessageDeliveryStatus: %s", err))
		return helpers.ErrInternalServerError
	}

	return nil
}
