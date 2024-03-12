package chatmodel

import "utils/helpers"

func NewDMChat(initiatorId int, partnerId int, initMsgContent map[string]any) (map[string]any, error) {
	data, err := helpers.QueryRowFields("SELECT new_dm_chat_id, init_msg_id FROM new_dm_chat($1, $2, $3)", initiatorId, partnerId, initMsgContent)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type DMChat struct {
	Id int
}

func (dmc DMChat) SendMessage(senderId int, msgContent map[string]any) (int, error) {
	msgId, err := helpers.QueryRowField[int]("SELECT msg_id FROM send_dm_chat_message($1, $2, $3)", dmc.Id, senderId, msgContent)
	if err != nil {
		return 0, err
	}

	return *msgId, nil
}

func (dmc DMChat) GetChatHistory() ([]*map[string]any, error) {
	messages, err := helpers.QueryRowsField[map[string]any]("SELECT message FROM get_dm_chat_history($1)", dmc.Id)
	if err != nil {
		return nil, err
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
		return err
	}

	return nil
}

func (dmcm DMChatMessage) React(reactorId int, reaction rune) error {
	// go helpers.QueryRowField[bool]("SELECT react_to_dm_chat_message($1, $2, $3, $4)", dmcm.DMChatId, dmcm.Id, reactorId, reaction)
	_, err := helpers.QueryRowField[bool]("SELECT react_to_dm_chat_message($1, $2, $3, $4)", dmcm.DMChatId, dmcm.Id, reactorId, reaction)
	if err != nil {
		return err
	}

	return nil
}
