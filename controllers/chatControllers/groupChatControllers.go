package chatcontrollers

import (
	"fmt"
	"log"
	"services/chatservice"
	"services/userservice"
	"time"
	"utils/appglobals"
	"utils/apptypes"
	"utils/helpers"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var GetGroupChatHistory = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var user apptypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	var body struct {
		GroupChatId int
		Offset      int
	}

	r_err := c.ReadJSON(&body)
	if r_err != nil {
		log.Println(r_err)
		return
	}

	groupChatHistory, app_err := chatservice.GroupChat{Id: body.GroupChatId}.GetChatHistory(body.Offset)

	var w_err error
	if app_err != nil {
		w_err = c.WriteJSON(helpers.AppError(fiber.StatusUnprocessableEntity, app_err))
	} else {
		w_err = c.WriteJSON(map[string]any{"chat_history": groupChatHistory})
	}

	if w_err != nil {
		log.Println(w_err)
		return
	}
})

var ActivateGroupChatSession = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	// this goroutine receives message acknowlegement for sent messages
	// and in turn changes the delivery status of messages sent by the child goroutine
	var user apptypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	var groupChatId int

	fmt.Sscanf(c.Query("chat_id"), "%d", &groupChatId)

	var myMailbox = make(chan map[string]any, 2)

	mailboxKey := fmt.Sprintf("user-%d--groupchat-%d", user.UserId, groupChatId)

	gcso := appglobals.GroupChatSessionObserver{}

	gcso.Subscribe(mailboxKey, myMailbox)

	endSession := func() {
		gcso.Unsubscribe(mailboxKey)
	}

	go sendGroupChatMessages(c, user, groupChatId, endSession)

	/* ---- stream group chat message events pending dispatch to the channel ---- */
	// observe that this happens once every new connection
	// A "What did I miss?" sort of query
	if event_data_kvps, err := (userservice.User{Id: user.UserId}).GetGroupChatMessageEventsPendingDispatch(groupChatId); err == nil {
		for _, evk := range event_data_kvps {
			evk := *evk
			myMailbox <- evk
		}
	}

	for data := range myMailbox {
		w_err := c.WriteJSON(data)
		if w_err != nil {
			log.Println(w_err)
			endSession()
			return
		}
	}
})

func sendGroupChatMessages(c *websocket.Conn, user apptypes.JWTUserData, groupChatId int, endSession func()) {
	// this goroutine sends messages
	var body struct {
		Msg map[string]any
		At  time.Time
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			return
		}

		data, app_err := chatservice.GroupChat{Id: groupChatId}.SendMessage(
			user.UserId, body.Msg, body.At,
		)

		var w_err error
		if app_err != nil {
			w_err = c.WriteJSON(helpers.AppError(fiber.StatusUnprocessableEntity, app_err))
		} else {
			w_err = c.WriteJSON(data)
		}

		if w_err != nil {
			log.Println(w_err)
			endSession()
			return
		}
	}
}

var PerformGroupOperation = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var user apptypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	operationHandlerMap := map[string]func(client []string, data map[string]any){
		"change_name":             changeGroupName,
		"change_description":      changeGroupDescription,
		"change_picture":          changeGroupPicture,
		"add_users":               addUsersToGroup,
		"remove_user":             removeUserFromGroup,
		"join":                    joinGroup,
		"leave":                   leaveGroup,
		"make_user_admin":         makeUserGroupAdmin,
		"remove_user_from_admins": removeUserFromGroupAdmins,
	}

	for {
		var body struct {
			Operation string
			Data      map[string]any
		}

		if r_err := c.ReadJSON(&body); r_err != nil {
			log.Println(r_err)
			break
		}

		client := []string{fmt.Sprint(user.UserId), user.Username}

		go operationHandlerMap[body.Operation](client, body.Data)

		if w_err := c.WriteJSON(map[string]any{"code": 200, "msg": "Operation Successful"}); w_err != nil {
			log.Println(w_err)
			break
		}
	}
})

func changeGroupName(client []string, data map[string]any) {
	var d struct {
		GroupChatId int
		NewName     string
	}

	helpers.MapToStruct(data, &d)

	chatservice.GroupChat{Id: d.GroupChatId}.ChangeName(client, d.NewName)

}

func changeGroupDescription(client []string, data map[string]any) {
	var d struct {
		GroupChatId    int
		NewDescription string
	}

	helpers.MapToStruct(data, &d)

	chatservice.GroupChat{Id: d.GroupChatId}.ChangeDescription(client, d.NewDescription)
}

func changeGroupPicture(client []string, data map[string]any) {
	var d struct {
		GroupChatId int
		NewPicture  []byte
	}

	helpers.MapToStruct(data, &d)

	chatservice.GroupChat{Id: d.GroupChatId}.ChangePicture(client, d.NewPicture)
}

func addUsersToGroup(client []string, data map[string]any) {
	var d struct {
		GroupChatId int
		NewUsers    [][]string
	}

	helpers.MapToStruct(data, &d)

	chatservice.GroupChat{Id: d.GroupChatId}.AddUsers(client, d.NewUsers)
}

func removeUserFromGroup(client []string, data map[string]any) {
	var d struct {
		GroupChatId int
		User        []string
	}

	helpers.MapToStruct(data, &d)

	chatservice.GroupChat{Id: d.GroupChatId}.RemoveUser(client, d.User)
}

func joinGroup(client []string, data map[string]any) {
	var d struct {
		GroupChatId int
	}

	helpers.MapToStruct(data, &d)

	chatservice.GroupChat{Id: d.GroupChatId}.Join(client)
}

func leaveGroup(client []string, data map[string]any) {
	var d struct {
		GroupChatId int
	}

	helpers.MapToStruct(data, &d)

	chatservice.GroupChat{Id: d.GroupChatId}.Leave(client)
}

func makeUserGroupAdmin(client []string, data map[string]any) {
	var d struct {
		GroupChatId int
		User        []string
	}

	helpers.MapToStruct(data, &d)

	chatservice.GroupChat{Id: d.GroupChatId}.MakeUserAdmin(client, d.User)
}

func removeUserFromGroupAdmins(client []string, data map[string]any) {
	var d struct {
		GroupChatId int
		User        []string
	}

	helpers.MapToStruct(data, &d)

	chatservice.GroupChat{Id: d.GroupChatId}.RemoveUserFromAdmins(client, d.User)
}
