package chatmodel

import (
	"fmt"
	"log"
	"strconv"
	"time"
	"utils/appglobals"
	"utils/apptypes"
	"utils/helpers"
)

func NewGroupChat(name string, description string, picture string, creator []string, initUsers [][]string) (map[string]any, error) {
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

func (gpc GroupChat) ChangeName(admin []string, newName string) (map[string]any, error) {
	data, err := helpers.QueryRowFields("SELECT member_ids AS memberIds, activity_data AS activityData change_group_name($1, $2, $3)", gpc.Id, admin, newName)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: GroupChat_ChangeName: %s", err))
		return nil, appglobals.ErrInternalServerError
	}

	return data, nil
}

func (gpc GroupChat) ChangeDescription(admin []string, newDescription string) (map[string]any, error) {
	data, err := helpers.QueryRowFields("SELECT member_ids AS memberIds, activity_data AS activityData change_group_description($1, $2, $3)", gpc.Id, admin, newDescription)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: GroupChat_ChangeDescription: %s", err))
		return nil, appglobals.ErrInternalServerError
	}

	return data, nil
}

func (gpc GroupChat) ChangePicture(admin []string, newPicture string) (map[string]any, error) {
	data, err := helpers.QueryRowFields("SELECT member_ids AS memberIds, activity_data AS activityData change_group_picture($1, $2, $3)", gpc.Id, admin, newPicture)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: GroupChat_ChangePicture: %s", err))
		return nil, appglobals.ErrInternalServerError
	}

	return data, nil
}

func (gpc GroupChat) AddUsers(admin []string, newUsers [][]string) (map[string]any, error) {
	data, err := helpers.QueryRowFields("SELECT member_ids AS memberIds, activity_data AS activityData add_users_to_group($1, $2, $3)", gpc.Id, admin, newUsers)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: GroupChat_AddUsers: %s", err))
		return nil, appglobals.ErrInternalServerError
	}

	return data, nil
}

func (gpc GroupChat) RemoveUser(admin []string, user []string) (map[string]any, error) {
	data, err := helpers.QueryRowFields("SELECT member_ids AS memberIds, activity_data AS activityData remove_user_to_group($1, $2, $3)", gpc.Id, admin, user)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: GroupChat_RemoveUser: %s", err))
		return nil, appglobals.ErrInternalServerError
	}

	return data, nil
}

func (gpc GroupChat) Join(newUser []string) (map[string]any, error) {
	data, err := helpers.QueryRowFields("SELECT member_ids AS memberIds, activity_data AS activityData join_group($1, $2)", gpc.Id, newUser)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: GroupChat_Join: %s", err))
		return nil, appglobals.ErrInternalServerError
	}

	return data, nil
}

func (gpc GroupChat) Leave(user []string) (map[string]any, error) {
	data, err := helpers.QueryRowFields("SELECT member_ids AS memberIds, activity_data AS activityData leave_group($1, $2)", gpc.Id, user)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: GroupChat_Leave: %s", err))
		return nil, appglobals.ErrInternalServerError
	}

	return data, nil
}

func (gpc GroupChat) MakeUserAdmin(admin []string, user []string) (map[string]any, error) {
	data, err := helpers.QueryRowFields("SELECT member_ids AS memberIds, activity_data AS activityData make_user_group_admin($1, $2, $3)", gpc.Id, admin, user)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: GroupChat_MakeUserAdmin: %s", err))
		return nil, appglobals.ErrInternalServerError
	}

	return data, nil
}

func (gpc GroupChat) RemoveUserFromAdmins(admin []string, user []string) (map[string]any, error) {
	data, err := helpers.QueryRowFields("SELECT member_ids AS memberIds, activity_data AS activityData remove_user_from_group_admins($1, $2, $3)", gpc.Id, admin, user)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: GroupChat_RemoveUserFromAdmins: %s", err))
		return nil, appglobals.ErrInternalServerError
	}

	return data, nil
}

func (gpc GroupChat) SendMessage(senderId int, msgContent map[string]any, createdAt time.Time) (map[string]any, error) {
	data, err := helpers.QueryRowFields("SELECT sender_resp_data AS srd, members_resp_data AS mrd, member_ids AS memberIds FROM send_group_chat_message($1, $2, $3, $4)", gpc.Id, senderId, msgContent, createdAt)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: GroupChat_SendMessage: %s", err))
		return nil, appglobals.ErrInternalServerError
	}

	return data, nil
}

func (gpc GroupChat) BatchUpdateGroupChatMessageDeliveryStatus(receiverId int, status string, delivDatas []*apptypes.GroupChatMsgDeliveryData) (string, error) {
	var sqls = []string{}
	var params = [][]any{}

	for _, data := range delivDatas {
		sqls = append(sqls, "SELECT overall_delivery_status FROM update_group_chat_message_delivery_status($1, $2, $3, $4, $5)")
		params = append(params, []any{gpc.Id, data.MsgId, receiverId, status, data.At})
	}

	overallDeliveryStatus, err := helpers.BatchQuery[string](sqls, params)
	if err != nil {
		log.Println(fmt.Errorf("DMChatModel.go: BatchUpdateGroupChatMessageDeliveryStatus: %s", err))
		return "", appglobals.ErrInternalServerError
	}

	return *overallDeliveryStatus, nil

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
	Id          int
	GroupChatId int
}

func (gpcm GroupChatMessage) UpdateDeliveryStatus(receiverId int, status string, updatedAt time.Time) (string, error) {
	overallDeliveryStatus, err := helpers.QueryRowField[string]("SELECT overall_delivery_status FROM update_group_chat_message_delivery_status($1, $2, $3, $4, $5)", gpcm.GroupChatId, gpcm.Id, receiverId, status, updatedAt)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: GroupChatMessage_UpdateDeliveryStatus: %s", err))
		return "", appglobals.ErrInternalServerError
	}

	return *overallDeliveryStatus, nil
}

func (gpcm GroupChatMessage) React(reactorId int, reaction rune) error {
	_, err := helpers.QueryRowField[bool]("SELECT react_to_group_chat_message($1, $2, $3, $4)", gpcm.GroupChatId, gpcm.Id, reactorId, strconv.QuoteRuneToASCII(reaction))
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: GroupChatMessage_React: %s", err))
		return appglobals.ErrInternalServerError
	}

	return nil

}
