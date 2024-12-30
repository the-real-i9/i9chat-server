package dmChat

import (
	"context"
	"fmt"
	"i9chat/appTypes"
	"i9chat/helpers"
	user "i9chat/models/userModel"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgtype"
)

type InitiatorData struct {
	NewDMChatId int `db:"new_dm_chat_id" json:"new_dm_chat_id"`
	InitMsgId   int `db:"init_msg_id" json:"init_msg_id"`
}

type PartnerData struct {
	Type     string    `json:"type"`
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

func New(ctx context.Context, initiatorId int, partnerId int, initMsgContent map[string]any, createdAt time.Time) (*NewDMChat, error) {
	newDMChat, err := helpers.QueryRowType[NewDMChat](ctx, "SELECT initiator_resp_data AS ird, partner_resp_data AS prd FROM new_dm_chat($1, $2, $3, $4)", initiatorId, partnerId, initMsgContent, createdAt)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: New: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	return newDMChat, nil
}

type SenderData struct {
	NewMsgId int `db:"new_msg_id" json:"new_msg_id"`
}

type ReceiverData struct {
	In       string         `db:"in" json:"in"`
	MsgId    int            `db:"msg_id" json:"msg_id"`
	DMChatId int            `db:"dm_chat_id" json:"dm_chat_id"`
	Sender   user.User      `json:"sender"`
	Content  map[string]any `json:"content"`
}

type NewMessage struct {
	*SenderData   `db:"srd"`
	*ReceiverData `db:"rrd"`
	ReceiverId    int `db:"receiver_id"`
}

func SendMessage(ctx context.Context, dmChatId, senderId int, msgContent map[string]any, createdAt time.Time) (*NewMessage, error) {
	newMessage, err := helpers.QueryRowType[NewMessage](ctx, "SELECT sender_resp_data AS srd, receiver_resp_data AS rrd, receiver_id FROM send_dm_chat_message($1, $2, $3, $4)", dmChatId, senderId, msgContent, createdAt)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: SendMessage: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	return newMessage, nil
}

func ReactToMessage(ctx context.Context, dmChatId, msgId, reactorId int, reaction rune) error {
	_, err := helpers.QueryRowField[bool](ctx, "SELECT react_to_dm_chat_message($1, $2, $3, $4)", dmChatId, msgId, reactorId, strconv.QuoteRuneToASCII(reaction))
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: ReactToMessage: %s", err))
		return fiber.ErrInternalServerError
	}

	return nil
}

type messageReaction struct {
	Reaction rune       `json:"reaction,omitempty"`
	Reactor  *user.User `json:"reactor,omitempty"`
}

type Message struct {
	Id             int               `db:"msg_id" json:"msg_id"`
	Sender         user.User         `json:"sender"`
	MsgContent     map[string]any    `db:"msg_content" json:"msg_content"`
	DeliveryStatus string            `db:"delivery_status" json:"delivery_status"`
	CreatedAt      pgtype.Timestamp  `db:"created_at" json:"created_at"`
	Edited         bool              `json:"edited"`
	EditedAt       *pgtype.Timestamp `db:"edited_at" json:"edited_at,omitempty"`
	Reactions      []messageReaction `json:"reactions"`
}

func GetChatHistory(ctx context.Context, dmChatId, offset int) ([]*Message, error) {
	messages, err := helpers.QueryRowsType[Message](ctx, `
	SELECT * FROM (
		SELECT * FROM get_dm_chat_history($1) 
		LIMIT 50 OFFSET $2
	) ORDER BY created_at ASC`, dmChatId, offset)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: GetChatHistory: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	return messages, nil
}

func BatchUpdateMessageDeliveryStatus(ctx context.Context, receiverId int, status string, ackDatas []*appTypes.DMChatMsgAckData) error {
	var sqls = []string{}
	var params = [][]any{}

	for _, data := range ackDatas {
		sqls = append(sqls, "SELECT update_dm_chat_message_delivery_status($1, $2, $3, $4, $5)")
		params = append(params, []any{data.DMChatId, data.MsgId, receiverId, status, data.At})
	}

	_, err := helpers.BatchQuery[bool](ctx, sqls, params)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: BatchUpdateMessageDeliveryStatus: %s", err))
		return fiber.ErrInternalServerError
	}

	return nil
}

func UpdateMessageDeliveryStatus(ctx context.Context, dmChatId, msgId, receiverId int, status string, updatedAt time.Time) error {
	_, err := helpers.QueryRowField[bool](ctx, "SELECT update_dm_chat_message_delivery_status($1, $2, $3, $4, $5)", dmChatId, msgId, receiverId, status, updatedAt)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: UpdateMessageDeliveryStatus: %s", err))
		return fiber.ErrInternalServerError
	}

	return nil
}
