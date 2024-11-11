package groupChatControllers

import (
	"fmt"
	"i9chat/appTypes"
	"i9chat/helpers"
	groupChat "i9chat/models/chatModel/groupChatModel"
	"i9chat/services/appServices"
	"i9chat/services/authServices"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var GetChatHistory = authServices.WSHandlerProtected(func(c *websocket.Conn) {

	var w_err error

	for {
		var body getChatHistoryBody

		if w_err != nil {
			log.Println(w_err)
			return
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			return
		}

		if val_err := body.Validate(); val_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			continue
		}

		groupChatHistory, app_err := groupChat.GetChatHistory(body.GroupChatId, body.Offset)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: 200,
			Body:       groupChatHistory,
		})
	}
})

var SendMessage = authServices.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("user").(*appTypes.ClientUser)

	var groupChatId int

	_, err := fmt.Sscanf(c.Params("group_chat_id"), "%d", &groupChatId)
	if err != nil {
		panic(err)
	}

	var w_err error

	for {
		var body sendMessageBody

		if w_err != nil {
			log.Println(w_err)
			break
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		if val_err := body.Validate(); val_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			continue
		}

		senderData, app_err := sendMessage(
			groupChatId,
			clientUser.Id,
			appServices.UploadMessageMedia(clientUser.Id, body.Msg),
			body.At,
		)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(senderData)

	}
})

var ExecuteAction = authServices.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("user").(*appTypes.ClientUser)

	type handler func(clientUser []string, data map[string]any) error

	actionToHandlerMap := map[action]handler{
		"change name":             changeGroupName,
		"change description":      changeGroupDescription,
		"change picture":          changeGroupPicture,
		"add users":               addUsersToGroup,
		"remove user":             removeUserFromGroup,
		"join":                    joinGroup,
		"leave":                   leaveGroup,
		"make user admin":         makeUserGroupAdmin,
		"remove user from admins": removeUserFromGroupAdmins,
	}

	var w_err error

	for {
		var body executeActionBody

		if w_err != nil {
			log.Println(w_err)
			break
		}

		if r_err := c.ReadJSON(&body); r_err != nil {
			log.Println(r_err)
			break
		}

		app_err := actionToHandlerMap[body.Action]([]string{fmt.Sprint(clientUser.Id), clientUser.Username}, body.Data)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, fmt.Errorf("action failed: %s", app_err)))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: 200,
			Body: map[string]any{
				"msg": "Action executed",
			},
		})
	}
})
