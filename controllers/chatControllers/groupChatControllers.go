package chatcontrollers

import (
	"fmt"
	"log"
	"services/chatservice"
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

var WatchGroupChatMessage = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var user apptypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	c.WriteJSON(map[string]string{"msg": "Enter the {groupChatId: {id}} to start watching. Any message after this closes the connection."})

	var body struct {
		GroupChatId int
	}

	r_err := c.ReadJSON(&body)
	if r_err != nil {
		log.Println(r_err)
		return
	}

	var myMailbox = make(chan map[string]any, 2)
	var closeMailBox = make(chan int)

	mailboxKey := fmt.Sprintf("user-%d--groupchat-%d", user.UserId, body.GroupChatId)

	gcmo := appglobals.GroupChatMessageObserver{}

	gcmo.Subscribe(mailboxKey, myMailbox)

	go func() {
		c.ReadMessage()
		closeMailBox <- 1
		close(closeMailBox)
	}()

	for {
		select {
		case data := <-myMailbox:
			w_err := c.WriteJSON(data)
			if w_err != nil {
				log.Println(w_err)
				gcmo.Unsubscribe(mailboxKey)
				return
			}
		case <-closeMailBox:
			gcmo.Unsubscribe(mailboxKey)
			return
		}
	}
})

var SendGroupChatMessage = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var user apptypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	var body struct {
		GroupChatId int
		MsgContent  map[string]any
		CreatedAt   time.Time
	}

	for {
		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			return
		}

		data, app_err := chatservice.GroupChat{Id: body.GroupChatId}.SendMessage(
			user.UserId, body.MsgContent, body.CreatedAt,
		)

		var w_err error
		if app_err != nil {
			w_err = c.WriteJSON(helpers.AppError(fiber.StatusUnprocessableEntity, app_err))
		} else {
			w_err = c.WriteJSON(data)
		}

		if w_err != nil {
			log.Println(w_err)
			return
		}
	}
})

var WatchGroupActivity = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var user apptypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	c.WriteJSON(map[string]string{"msg": "Enter the {groupChatId: {id}} to start watching. Any message after this closes the connection."})

	var body struct {
		GroupChatId int
	}

	r_err := c.ReadJSON(&body)
	if r_err != nil {
		log.Println(r_err)
		return
	}

	var myMailbox = make(chan map[string]any, 2)
	var closeMailBox = make(chan int)

	mailboxKey := fmt.Sprintf("user-%d--groupchat-%d", user.UserId, body.GroupChatId)

	gcao := appglobals.GroupChatActivityObserver{}

	gcao.Subscribe(mailboxKey, myMailbox)

	go func() {
		c.ReadMessage()
		closeMailBox <- 1
		close(closeMailBox)
	}()

	for {
		select {
		case data := <-myMailbox:
			w_err := c.WriteJSON(data)
			if w_err != nil {
				log.Println(w_err)
				gcao.Unsubscribe(mailboxKey)
				return
			}
		case <-closeMailBox:
			gcao.Unsubscribe(mailboxKey)
			return
		}
	}
})

var BatchUpdateGroupChatMessageDeliveryStatus = helpers.WSHandlerProtected(func(c *websocket.Conn) {
	var user apptypes.JWTUserData

	helpers.MapToStruct(c.Locals("auth").(map[string]any), &user)

	for {
		var body struct {
			GroupChatId int
			Status      string
			MsgDatas    []*apptypes.GroupChatMsgDeliveryData
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		go chatservice.GroupChat{Id: body.GroupChatId}.BatchUpdateGroupChatMessageDeliveryStatus(user.UserId, body.Status, body.MsgDatas)

	}
})

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
