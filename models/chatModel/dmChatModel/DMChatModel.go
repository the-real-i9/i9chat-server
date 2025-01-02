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

type ClientNewMsgData struct {
	MsgId int `db:"msg_id" json:"msg_id"`
}

type PartnerNewMsgData struct {
	In     string    `db:"in" json:"in"`
	Sender user.User `json:"sender"`
	Msg    struct {
		Id      int            `db:"msg_id" json:"msg_id"`
		Content map[string]any `json:"content"`
	}
}

type NewMessage struct {
	*ClientNewMsgData  `db:"crd"`
	*PartnerNewMsgData `db:"prd"`
}

func SendMessage(ctx context.Context, clientUserId, partnerUserId int, msgContent appTypes.MsgContent, createdAt time.Time) (*NewMessage, error) {
	res, err := helpers.QueryRowType[NewMessage](ctx, "SELECT client_resp_data AS crd, partner_resp_data AS prd FROM send_dm_chat_message($1, $2, $3, $4)", clientUserId, partnerUserId, msgContent, createdAt)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: SendMessage: %s", err))
		return nil, fiber.ErrInternalServerError
	}

	return res, nil
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

func UpdateMessageDeliveryStatus(ctx context.Context, clientUserId, partnerUserId, msgId int, status string, updatedAt time.Time) error {
	_, err := helpers.QueryRowField[bool](ctx, "SELECT update_dm_chat_message_delivery_status($1, $2, $3, $4, $5)", clientUserId, partnerUserId, msgId, status, updatedAt)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: UpdateMessageDeliveryStatus: %s", err))
		return fiber.ErrInternalServerError
	}

	return nil
}

func BatchUpdateMessageDeliveryStatus(ctx context.Context, clientUserId int, status string, ackDatas []*appTypes.DMChatMsgAckData) error {
	var sqls = []string{}
	var params = [][]any{}

	for _, data := range ackDatas {
		sqls = append(sqls, "SELECT update_dm_chat_message_delivery_status($1, $2, $3, $4, $5)")
		params = append(params, []any{clientUserId, data.PartnerUserId, data.MsgId, status, data.At})
	}

	_, err := helpers.BatchQuery[bool](ctx, sqls, params)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: BatchUpdateMessageDeliveryStatus: %s", err))
		return fiber.ErrInternalServerError
	}

	return nil
}
