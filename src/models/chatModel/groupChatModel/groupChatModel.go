package groupChat

import (
	"fmt"
	"i9chat/src/appGlobals"
	"i9chat/src/appTypes"
	"i9chat/src/helpers"
	user "i9chat/src/models/userModel"
	"log"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type CreatorData struct {
	NewGroupChatId int `db:"new_group_chat_id" json:"new_group_chat_id"`
}

type InitMemberData struct {
	Type        string `json:"type"`
	GroupChatId int    `db:"group_chat_id" json:"group_chat_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PictureUrl  string `db:"picture_url" json:"picture_url"`
}

type NewGroupChat struct {
	*CreatorData    `db:"crd"`
	*InitMemberData `db:"imrd"`
}

func New(name string, description string, pictureUrl string, creator []string, initUsers [][]appTypes.String) (*NewGroupChat, error) {
	newGroupChat, err := helpers.QueryRowType[NewGroupChat]("SELECT creator_resp_data AS crd, init_member_resp_data AS imrd FROM new_group_chat($1, $2, $3, $4, $5)", name, description, pictureUrl, creator, initUsers)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: New: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return newGroupChat, nil
}

type NewActivity struct {
	MembersIds   []int          `db:"members_ids"`
	ActivityInfo map[string]any `db:"activity_data"`
}

func ChangeName(groupChatId int, admin []string, newName string) (*NewActivity, error) {
	newActivity, err := helpers.QueryRowType[NewActivity]("SELECT * change_group_name($1, $2, $3)", groupChatId, admin, newName)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: ChangeName: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return newActivity, nil
}

func ChangeDescription(groupChatId int, admin []string, newDescription string) (*NewActivity, error) {
	newActivity, err := helpers.QueryRowType[NewActivity]("SELECT * change_group_description($1, $2, $3)", groupChatId, admin, newDescription)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: ChangeDescription: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return newActivity, nil
}

func ChangePicture(groupChatId int, admin []string, newPictureUrl string) (*NewActivity, error) {
	newActivity, err := helpers.QueryRowType[NewActivity]("SELECT * change_group_picture($1, $2, $3)", groupChatId, admin, newPictureUrl)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: ChangePicture: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return newActivity, nil
}

func AddUsers(groupChatId int, admin []string, newUsers [][]appTypes.String) (*NewActivity, error) {
	newActivity, err := helpers.QueryRowType[NewActivity]("SELECT * add_users_to_group($1, $2, $3)", groupChatId, admin, newUsers)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: AddUsers: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return newActivity, nil
}

func RemoveUser(groupChatId int, admin []string, user []appTypes.String) (*NewActivity, error) {
	newActivity, err := helpers.QueryRowType[NewActivity]("SELECT * remove_user_to_group($1, $2, $3)", groupChatId, admin, user)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: RemoveUser: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return newActivity, nil
}

func Join(groupChatId int, newUser []string) (*NewActivity, error) {
	newActivity, err := helpers.QueryRowType[NewActivity]("SELECT * join_group($1, $2)", groupChatId, newUser)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: Join: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return newActivity, nil
}

func Leave(groupChatId int, user []string) (*NewActivity, error) {
	newActivity, err := helpers.QueryRowType[NewActivity]("SELECT * leave_group($1, $2)", groupChatId, user)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: Leave: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return newActivity, nil
}

func MakeUserAdmin(groupChatId int, admin []string, user []appTypes.String) (*NewActivity, error) {
	newActivity, err := helpers.QueryRowType[NewActivity]("SELECT * make_user_group_admin($1, $2, $3)", groupChatId, admin, user)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: MakeUserAdmin: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return newActivity, nil
}

func RemoveUserFromAdmins(groupChatId int, admin []string, user []appTypes.String) (*NewActivity, error) {
	newActivity, err := helpers.QueryRowType[NewActivity]("SELECT * remove_user_from_group_admins($1, $2, $3)", groupChatId, admin, user)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: RemoveUserFromAdmins: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return newActivity, nil
}

type SenderData struct {
	NewMsgId int `db:"new_msg_id" json:"new_msg_id"`
}

type MemberData struct {
	In          string         `db:"in" json:"in"`
	MsgId       int            `db:"msg_id" json:"msg_id"`
	GroupChatId int            `db:"group_chat_id" json:"group_chat_id"`
	Sender      user.User      `json:"sender"`
	Content     map[string]any `json:"content"`
}

type NewMessage struct {
	*SenderData `db:"srd"`
	*MemberData `db:"mrd"`
	MembersIds  []int `db:"members_ids"`
}

func SendMessage(groupChatId int, senderId int, msgContent map[string]any, createdAt time.Time) (*NewMessage, error) {
	newMessage, err := helpers.QueryRowType[NewMessage]("SELECT sender_resp_data AS srd, member_resp_data AS mrd, members_ids FROM send_group_chat_message($1, $2, $3, $4)", groupChatId, senderId, msgContent, createdAt)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: SendMessage: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return newMessage, nil
}

func ReactToMessage(groupChatId int, msgId, reactorId int, reaction rune) error {
	_, err := helpers.QueryRowField[bool]("SELECT react_to_group_chat_message($1, $2, $3, $4)", groupChatId, msgId, reactorId, strconv.QuoteRuneToASCII(reaction))
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: ReactToMessage: %s", err))
		return appGlobals.ErrInternalServerError
	}

	return nil
}

type BatchStatusUpdateResult struct {
	OverallDeliveryStatus string `db:"overall_delivery_status"`
	ShouldBroadcast       bool   `db:"should_broadcast"`
}

func BatchUpdateMessageDeliveryStatus(groupChatId int, receiverId int, status string, ackDatas []*appTypes.GroupChatMsgAckData) (*BatchStatusUpdateResult, error) {
	var sqls = []string{}
	var params = [][]any{}

	for _, data := range ackDatas {
		sqls = append(sqls, "SELECT * FROM update_group_chat_message_delivery_status($1, $2, $3, $4, $5)")
		params = append(params, []any{groupChatId, data.MsgId, receiverId, status, data.At})
	}

	resultList, err := helpers.BatchQuery[BatchStatusUpdateResult](sqls, params)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: BatchUpdateMessageDeliveryStatus: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	lastResult := resultList[len(resultList)]

	return lastResult, nil
}

type messageReaction struct {
	Reaction rune       `json:"reaction,omitempty"`
	Reactor  *user.User `json:"reactor,omitempty"`
}

type HistoryItem struct {
	Type string `json:"type"`

	// if Type = "message"
	Id             int               `json:"id,omitempty"`
	Sender         *user.User        `json:"sender,omitempty"`
	Content        map[string]any    `json:"content,omitempty"`
	DeliveryStatus string            `db:"delivery_status" json:"delivery_status,omitempty"`
	CreatedAt      *pgtype.Timestamp `db:"created_at" json:"created_at,omitempty"`
	Edited         bool              `json:"edited,omitempty"`
	EditedAt       *pgtype.Timestamp `db:"edited_at" json:"edited_at,omitempty"`
	Reactions      []messageReaction `json:"reactions,omitempty"`

	// if Type = "activity"
	ActivityType string         `db:"activity_type" json:"activity_type,omitempty"`
	ActivityInfo map[string]any `db:"activity_info" json:"activity_info,omitempty"`
}

func GetChatHistory(groupChatId, offset int) ([]*HistoryItem, error) {
	history, err := helpers.QueryRowsType[HistoryItem](`
	SELECT * FROM (
		SELECT * FROM get_group_chat_history($1)
		LIMIT 50 OFFSET $2
	) ORDER BY created_at ASC`, groupChatId, offset)
	if err != nil {
		log.Println(fmt.Errorf("groupChatModel.go: GetChatHistory: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return history, nil
}
