package chatmodel

import (
	"fmt"
	"log"
	"utils/appglobals"
	"utils/helpers"
)

func NewGroupChat(name string, description string, picture string, creator [2]string, initUsers [][2]string) (map[string]any, error) {
	data, err := helpers.QueryRowFields("SELECT creator_resp_data AS crd, new_members_resp_data AS nmrd FROM new_group_chat($1, $2, $3, $4, $5)", name, description, picture, creator, initUsers)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: NewGroupChat: %s", err))
		return nil, appglobals.ErrInternalServerError
	}

	return data, nil
}

type GroupChat struct {
	Id int
}

func (gpc GroupChat) ChangeName(admin [2]string, newName string) error {
	// go helpers.QueryRowField[bool]("SELECT change_group_name($1, $2, $3)", gpc.Id, admin, newName)
	_, err := helpers.QueryRowField[bool]("SELECT change_group_name($1, $2, $3)", gpc.Id, admin, newName)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: GroupChat_ChangeName: %s", err))
		return appglobals.ErrInternalServerError
	}

	return nil
}

func (gpc GroupChat) ChangeDescription(admin [2]string, newDescription string) error {
	// go helpers.QueryRowField[bool]("SELECT change_group_description($1, $2, $3)", gpc.Id, admin, newDescription)
	_, err := helpers.QueryRowField[bool]("SELECT change_group_description($1, $2, $3)", gpc.Id, admin, newDescription)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: GroupChat_ChangeDescription: %s", err))
		return appglobals.ErrInternalServerError
	}

	return nil
}

func (gpc GroupChat) ChangePicture(admin [2]string, newPicture string) error {
	// go helpers.QueryRowField[bool]("SELECT change_group_picture($1, $2, $3)", gpc.Id, admin, newPicture)
	_, err := helpers.QueryRowField[bool]("SELECT change_group_picture($1, $2, $3)", gpc.Id, admin, newPicture)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: GroupChat_ChangePicture: %s", err))
		return appglobals.ErrInternalServerError
	}

	return nil
}

func (gpc GroupChat) AddUsers(admin [2]string, users [][2]string) error {
	// go helpers.QueryRowField[bool]("SELECT add_users_to_group($1, $2, $3)", gpc.Id, admin, users)
	_, err := helpers.QueryRowField[bool]("SELECT add_users_to_group($1, $2, $3)", gpc.Id, admin, users)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: GroupChat_AddUsers: %s", err))
		return appglobals.ErrInternalServerError
	}

	return nil
}

func (gpc GroupChat) RemoveUser(admin [2]string, user [2]string) error {
	// go helpers.QueryRowField[bool]("SELECT remove_user_to_group($1, $2, $3)", gpc.Id, admin, user)
	_, err := helpers.QueryRowField[bool]("SELECT remove_user_to_group($1, $2, $3)", gpc.Id, admin, user)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: GroupChat_RemoveUser: %s", err))
		return appglobals.ErrInternalServerError
	}

	return nil
}

func (gpc GroupChat) Join(user [2]string) error {
	_, err := helpers.QueryRowField[bool]("SELECT join_group($1, $2)", gpc.Id, user)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: GroupChat_Join: %s", err))
		return appglobals.ErrInternalServerError
	}

	return nil
}

func (gpc GroupChat) Leave(user [2]string) error {
	_, err := helpers.QueryRowField[bool]("SELECT leave_group($1, $2)", gpc.Id, user)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: GroupChat_Leave: %s", err))
		return appglobals.ErrInternalServerError
	}

	return nil
}

func (gpc GroupChat) MakeUserAdmin(admin [2]string, user [2]string) error {
	// go helpers.QueryRowField[bool]("SELECT make_user_group_admin($1, $2, $3)", gpc.Id, admin, user)
	_, err := helpers.QueryRowField[bool]("SELECT make_user_group_admin($1, $2, $3)", gpc.Id, admin, user)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: GroupChat_MakeUserAdmin: %s", err))
		return appglobals.ErrInternalServerError
	}

	return nil
}

func (gpc GroupChat) RemoveUserFromAdmins(admin [2]string, user [2]string) error {
	// go helpers.QueryRowField[bool]("SELECT remove_user_from_group_admins($1, $2, $3)", gpc.Id, admin, user)
	_, err := helpers.QueryRowField[bool]("SELECT remove_user_from_group_admins($1, $2, $3)", gpc.Id, admin, user)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: GroupChat_RemoveUserFromAdmins: %s", err))
		return appglobals.ErrInternalServerError
	}

	return nil
}

func (gpc GroupChat) SendMessage(senderId int, msgContent map[string]any) (map[string]any, error) {
	data, err := helpers.QueryRowFields("SELECT sender_resp_data AS srd, members_resp_data AS mrd FROM send_dm_chat_message($1, $2, $3)", gpc.Id, senderId, msgContent)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: GroupChat_SendMessage: %s", err))
		return nil, appglobals.ErrInternalServerError
	}

	return data, nil
}

func (gpc GroupChat) GetChatHistory(offset int) ([]*map[string]any, error) {
	history, err := helpers.QueryRowsField[map[string]any](`
	SELECT history_item FROM (
		SELECT history_item, time_created FROM get_group_chat_history($1)
		LIMIT 50 OFFSET $2
	) ORDER BY time_created ASC`, gpc.Id, offset)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: GroupChat_GetChatHistory: %s", err))
		return nil, appglobals.ErrInternalServerError
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
		log.Println(fmt.Errorf("groupChatModel.go: GroupChatMessage_UpdateDeliveryStatus: %s", err))
		return "", appglobals.ErrInternalServerError
	}

	return *overallDeliveryStatus, nil
}

func (gpcm GroupChatMessage) React(reactorId int, reaction rune) error {
	// go helpers.QueryRowField[bool]("SELECT react_to_group_chat_message($1, $2, $3, $4)", gpcm.GroupChatId, gpcm.Id, reactorId, reaction)
	_, err := helpers.QueryRowField[bool]("SELECT react_to_group_chat_message($1, $2, $3, $4)", gpcm.GroupChatId, gpcm.Id, reactorId, reaction)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: GroupChatMessage_React: %s", err))
		return appglobals.ErrInternalServerError
	}

	return nil

}
