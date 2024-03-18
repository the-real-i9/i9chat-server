package chatmodel

import (
	"fmt"
	"log"
	"utils/appglobals"
	"utils/helpers"
)

func NewDMChat(initiatorId int, partnerId int, initMsgContent map[string]any) (map[string]any, error) {
	var respData struct {
		Ird map[string]any
		Prd map[string]any
	}

	data, err := helpers.QueryRowFields("SELECT initiator_resp_data AS ird, partner_resp_data AS prd FROM new_dm_chat($1, $2, $3)", initiatorId, partnerId, initMsgContent)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: NewDMChat: %s", err))
		return nil, appglobals.ErrInternalServerError
	}

	helpers.MapToStruct(data, &respData)

	go appglobals.SendNewChatUpdate(fmt.Sprintf("user-%d", partnerId), respData.Prd)

	return respData.Ird, nil
}

type DMChat struct {
	Id int
}

func (dmc DMChat) SendMessage(dmChatId int, senderId int, msgContent map[string]any) (int, error) {
	msgId, err := helpers.QueryRowField[int]("SELECT msg_id FROM send_dm_chat_message($1, $2, $3)", dmChatId, senderId, msgContent)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: DMChat_SendMessage: %s", err))
		return 0, appglobals.ErrInternalServerError
	}

	return *msgId, nil
}

func (dmc DMChat) GetChatHistory(offset int) ([]*map[string]any, error) {
	messages, err := helpers.QueryRowsField[map[string]any](`
	SELECT message FROM (
		SELECT message, created_at FROM get_dm_chat_history($1) 
		LIMIT 50 OFFSET $2
	) ORDER BY created_at ASC`, dmc.Id, offset)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: DMChat_GetChatHistory: %s", err))
		return nil, appglobals.ErrInternalServerError
	}

	return messages, nil
}

type DMChatMessage struct {
	Id       int
	DMChatId int
}

func (dmcm DMChatMessage) UpdateDeliveryStatus(receiverId int, status string) error {
	// go helpers.QueryRowField[bool]("SELECT update_dm_chat_message_delivery_status($1, $2, $3, $4)", dmcm.DMChatId, dmcm.Id, receiverId, status)
	_, err := helpers.QueryRowField[bool]("SELECT update_dm_chat_message_delivery_status($1, $2, $3, $4)", dmcm.DMChatId, dmcm.Id, receiverId, status)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: DMChatMessage_UpdateDeliveryStatus: %s", err))
		return appglobals.ErrInternalServerError
	}

	return nil
}

func (dmcm DMChatMessage) React(reactorId int, reaction rune) error {
	// go helpers.QueryRowField[bool]("SELECT react_to_dm_chat_message($1, $2, $3, $4)", dmcm.DMChatId, dmcm.Id, reactorId, reaction)
	_, err := helpers.QueryRowField[bool]("SELECT react_to_dm_chat_message($1, $2, $3, $4)", dmcm.DMChatId, dmcm.Id, reactorId, reaction)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: DMChatMessage_React: %s", err))
		return appglobals.ErrInternalServerError
	}

	return nil
}
