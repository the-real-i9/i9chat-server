package chatModel

import (
	"fmt"
	"i9chat/utils/appTypes"
	"i9chat/utils/helpers"
	"log"
	"strconv"
	"time"
)

func NewDMChat(initiatorId int, partnerId int, initMsgContent map[string]any, createdAt time.Time) (map[string]any, error) {
	data, err := helpers.QueryRowFields("SELECT initiator_resp_data AS ird, partner_resp_data AS prd FROM new_dm_chat($1, $2, $3, $4)", initiatorId, partnerId, initMsgContent, createdAt)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: NewDMChat: %s", err))
		return nil, helpers.ErrInternalServerError
	}

	return data, nil
}

func BatchUpdateDMChatMessageDeliveryStatus(receiverId int, status string, delivDatas []*appTypes.DMChatMsgDeliveryData) error {
	var sqls = []string{}
	var params = [][]any{}

	for _, data := range delivDatas {
		sqls = append(sqls, "SELECT update_dm_chat_message_delivery_status($1, $2, $3, $4, $5)")
		params = append(params, []any{data.DmChatId, data.MsgId, receiverId, status, data.At})
	}

	_, err := helpers.BatchQuery[bool](sqls, params)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: BatchUpdateDMChatMessageDeliveryStatus: %s", err))
		return helpers.ErrInternalServerError
	}

	return nil
}

type DMChat struct {
	Id int
}

func (dmc DMChat) SendMessage(senderId int, msgContent map[string]any, createdAt time.Time) (map[string]any, error) {
	data, err := helpers.QueryRowFields("SELECT sender_resp_data AS srd, receiver_resp_data AS rrd, receiver_id AS receiverId FROM send_dm_chat_message($1, $2, $3, $4)", dmc.Id, senderId, msgContent, createdAt)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: DMChat_SendMessage: %s", err))
		return nil, helpers.ErrInternalServerError
	}

	return data, nil
}

func (dmc DMChat) GetChatHistory(offset int) ([]*map[string]any, error) {
	messages, err := helpers.QueryRowsField[map[string]any](`
	SELECT message FROM (
		SELECT message, created_at FROM get_dm_chat_history($1) 
		LIMIT 50 OFFSET $2
	) ORDER BY created_at ASC`, dmc.Id, offset)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: DMChat_GetChatHistory: %s", err))
		return nil, helpers.ErrInternalServerError
	}

	return messages, nil
}

type DMChatMessage struct {
	Id       int
	DmChatId int
}

func (dmcm DMChatMessage) UpdateDeliveryStatus(receiverId int, status string, updatedAt time.Time) error {
	_, err := helpers.QueryRowField[bool]("SELECT update_dm_chat_message_delivery_status($1, $2, $3, $4, $5)", dmcm.DmChatId, dmcm.Id, receiverId, status, updatedAt)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: DMChatMessage_UpdateDeliveryStatus: %s", err))
		return helpers.ErrInternalServerError
	}

	return nil
}

func (dmcm DMChatMessage) React(reactorId int, reaction rune) error {
	_, err := helpers.QueryRowField[bool]("SELECT react_to_dm_chat_message($1, $2, $3, $4)", dmcm.DmChatId, dmcm.Id, reactorId, strconv.QuoteRuneToASCII(reaction))
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: DMChatMessage_React: %s", err))
		return helpers.ErrInternalServerError
	}

	return nil
}
