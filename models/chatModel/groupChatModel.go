package chatmodel

import "utils/helpers"

func NewGroupChat(name string, description string, creator [2]string, initUsers [][2]string) (int, error) {
	newGroupChatId, err := helpers.QueryRowField[int]("SELECT new_group_chat_id FROM new_group_chat($1, $2, $3, $4)", name, description, creator, initUsers)
	if err != nil {
		return 0, err
	}

	return *newGroupChatId, nil
}

type GroupChat struct {
	Id int
}

func (gpc GroupChat) ChangeName(admin [2]string, newName string) error {
	// go helpers.QueryRowField[bool]("SELECT change_group_name($1, $2, $3)", gpc.Id, admin, newName)
	_, err := helpers.QueryRowField[bool]("SELECT change_group_name($1, $2, $3)", gpc.Id, admin, newName)
	if err != nil {
		return err
	}

	return nil
}

func (gpc GroupChat) ChangeDescription(admin [2]string, newDescription string) error {
	// go helpers.QueryRowField[bool]("SELECT change_group_description($1, $2, $3)", gpc.Id, admin, newDescription)
	_, err := helpers.QueryRowField[bool]("SELECT change_group_description($1, $2, $3)", gpc.Id, admin, newDescription)
	if err != nil {
		return err
	}

	return nil
}

func (gpc GroupChat) ChangePicture(admin [2]string, newPicture string) error {
	// go helpers.QueryRowField[bool]("SELECT change_group_picture($1, $2, $3)", gpc.Id, admin, newPicture)
	_, err := helpers.QueryRowField[bool]("SELECT change_group_picture($1, $2, $3)", gpc.Id, admin, newPicture)
	if err != nil {
		return err
	}

	return nil
}

func (gpc GroupChat) AddUsers(admin [2]string, users [][2]string) error {
	// go helpers.QueryRowField[bool]("SELECT add_users_to_group($1, $2, $3)", gpc.Id, admin, users)
	_, err := helpers.QueryRowField[bool]("SELECT add_users_to_group($1, $2, $3)", gpc.Id, admin, users)
	if err != nil {
		return err
	}

	return nil
}

func (gpc GroupChat) RemoveUser(admin [2]string, user [2]string) error {
	// go helpers.QueryRowField[bool]("SELECT remove_user_to_group($1, $2, $3)", gpc.Id, admin, user)
	_, err := helpers.QueryRowField[bool]("SELECT remove_user_to_group($1, $2, $3)", gpc.Id, admin, user)
	if err != nil {
		return err
	}

	return nil
}

func (gpc GroupChat) Join(user [2]string) error {
	_, err := helpers.QueryRowField[bool]("SELECT join_group($1, $2)", gpc.Id, user)
	if err != nil {
		return err
	}

	return nil
}

func (gpc GroupChat) Leave(user [2]string) error {
	_, err := helpers.QueryRowField[bool]("SELECT leave_group($1, $2)", gpc.Id, user)
	if err != nil {
		return err
	}

	return nil
}

func (gpc GroupChat) MakeUserAdmin(admin [2]string, user [2]string) error {
	// go helpers.QueryRowField[bool]("SELECT make_user_group_admin($1, $2, $3)", gpc.Id, admin, user)
	_, err := helpers.QueryRowField[bool]("SELECT make_user_group_admin($1, $2, $3)", gpc.Id, admin, user)
	if err != nil {
		return err
	}

	return nil
}

func (gpc GroupChat) RemoveUserFromAdmins(admin [2]string, user [2]string) error {
	// go helpers.QueryRowField[bool]("SELECT remove_user_from_group_admins($1, $2, $3)", gpc.Id, admin, user)
	_, err := helpers.QueryRowField[bool]("SELECT remove_user_from_group_admins($1, $2, $3)", gpc.Id, admin, user)
	if err != nil {
		return err
	}

	return nil
}

func (gpc GroupChat) SendMessage(senderId int, msgContent map[string]any) (int, error) {
	msgId, err := helpers.QueryRowField[int]("SELECT msg_id FROM send_group_chat_message($1, $2, $3)", gpc.Id, senderId, msgContent)
	if err != nil {
		return 0, err
	}

	return *msgId, nil
}

func (gpc GroupChat) GetChatHistory() ([]*map[string]any, error) {
	history, err := helpers.QueryRowsField[map[string]any]("SELECT history_item FROM get_group_chat_history($1)", gpc.Id)
	if err != nil {
		return nil, err
	}

	return history, nil
}

type GroupChatMessage struct {
	GroupChatId int
	Id          int
}

func (gpcm GroupChatMessage) UpdateDeliveryStatus(receiverId int, status string) (string, error) {
	overallDeliveryStatus, err := helpers.QueryRowField[string]("SELECT overall_delivery_status FROM update_group_chat_message_delivery_status($1, $2, $3, $4)", gpcm.GroupChatId, gpcm.Id, receiverId, status)
	if err != nil {
		return "", nil
	}

	return *overallDeliveryStatus, nil
}

func (gpcm GroupChatMessage) React(reactorId int, reaction rune) error {
	// go helpers.QueryRowField[bool]("SELECT react_to_group_chat_message($1, $2, $3, $4)", gpcm.GroupChatId, gpcm.Id, reactorId, reaction)
	_, err := helpers.QueryRowField[bool]("SELECT react_to_group_chat_message($1, $2, $3, $4)", gpcm.GroupChatId, gpcm.Id, reactorId, reaction)
	if err != nil {
		return err
	}

	return nil

}
