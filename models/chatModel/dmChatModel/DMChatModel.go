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

type ClientNewDMChatData struct {
	NewDMChatId string `db:"new_dm_chat_id" json:"new_dm_chat_id"` // client's dm chat id
	InitMsgId   int    `db:"init_msg_id" json:"init_msg_id"`
}

type PartnerNewDMChatData struct {
	Type        string    `json:"type"`
	DMChatId    string    `db:"dm_chat_id" json:"dm_chat_id"` // partner's dm chat id
	PartnerUser user.User `json:"partner_user"`
	InitMsg     struct {
		Id      int            `json:"id"`
		Content map[string]any `json:"content"`
	} `db:"init_msg" json:"init_msg"`
}

type NewDMChat struct {
	*ClientNewDMChatData  `db:"crd"`
	*PartnerNewDMChatData `db:"prd"`
}

func New(ctx context.Context, clientUserId int, partnerUserId int, initMsgContent appTypes.MsgContent, createdAt time.Time) (*NewDMChat, error) {
	newDMChat, err := helpers.QueryRowType[NewDMChat](ctx, "SELECT client_resp_data AS crd, partner_resp_data AS prd FROM new_dm_chat($1, $2, $3, $4)", clientUserId, partnerUserId, initMsgContent, createdAt)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: New: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	return newDMChat, nil
}

type ClientNewMsgData struct {
	NewMsgId int `db:"new_msg_id" json:"new_msg_id"`
}

type PartnerNewMsgData struct {
	In       string         `db:"in" json:"in"`
	MsgId    int            `db:"msg_id" json:"msg_id"`
	DMChatId string         `db:"dm_chat_id" json:"dm_chat_id"` // partner's dm chat id
	Sender   user.User      `json:"sender"`
	Content  map[string]any `json:"content"`
}

type NewMessage struct {
	*ClientNewMsgData  `db:"crd"`
	*PartnerNewMsgData `db:"prd"`
	PartnerUserId      int `db:"partner_user_id"`
}

func SendMessage(ctx context.Context, clientDMChatId string, clientUserId int, msgContent appTypes.MsgContent, createdAt time.Time) (*NewMessage, error) {
	newMessage, err := helpers.QueryRowType[NewMessage](ctx, "SELECT client_resp_data AS crd, partner_resp_data AS prd, partner_user_id FROM send_dm_chat_message($1, $2, $3, $4)", clientDMChatId, clientUserId, msgContent, createdAt)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: SendMessage: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	return newMessage, nil
}

func ReactToMessage(ctx context.Context, clientDMChatId string, msgId, clientUserId int, reaction rune) error {
	_, err := helpers.QueryRowField[bool](ctx, "SELECT react_to_dm_chat_message($1, $2, $3, $4)", clientDMChatId, msgId, clientUserId, strconv.QuoteRuneToASCII(reaction))
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

func GetChatHistory(ctx context.Context, dmChatId string, offset int) ([]*Message, error) {
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

type PartnerMsgDelivStatusUpdateData struct {
	PartnerUserId   int    `db:"partner_user_id"`
	PartnerDMChatId string `db:"partner_dm_chat_id"`
	MsgId           int    `db:"msg_id"`
}

func UpdateMessageDeliveryStatus(ctx context.Context, clientDMChatId string, msgId, clientUserId int, status string, updatedAt time.Time) (*PartnerMsgDelivStatusUpdateData, error) {
	res, err := helpers.QueryRowType[PartnerMsgDelivStatusUpdateData](ctx, "SELECT partner_user_id, partner_dm_chat_id, $2 AS msg_id FROM update_dm_chat_message_delivery_status($1, $2, $3, $4, $5)", clientDMChatId, msgId, clientUserId, status, updatedAt)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: UpdateMessageDeliveryStatus: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	return res, nil
}

func BatchUpdateMessageDeliveryStatus(ctx context.Context, clientUserId int, status string, ackDatas []*appTypes.DMChatMsgAckData) ([]*PartnerMsgDelivStatusUpdateData, error) {
	var sqls = []string{}
	var params = [][]any{}

	for _, data := range ackDatas {
		sqls = append(sqls, "SELECT partner_user_id, partner_dm_chat_id, $2 AS msg_id FROM update_dm_chat_message_delivery_status($1, $2, $3, $4, $5)")
		params = append(params, []any{data.ClientDMChatId, data.MsgId, clientUserId, status, data.At})
	}

	res, err := helpers.BatchQuery[PartnerMsgDelivStatusUpdateData](ctx, sqls, params)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: BatchUpdateMessageDeliveryStatus: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	return res, nil
}
