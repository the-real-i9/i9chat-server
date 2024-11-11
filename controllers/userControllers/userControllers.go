package userControllers

import (
	"fmt"
	"i9chat/appTypes"
	"i9chat/helpers"
	user "i9chat/models/userModel"
	"i9chat/services/appServices"
	"i9chat/services/authServices"
	"i9chat/services/messageBrokerService"
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// This Controller essentially opens the stream for receiving messages
var GoOnline = authServices.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("user").(*appTypes.ClientUser)

	// a channel for streaming data to client
	var myMailbox = make(chan any)

	userPOId := fmt.Sprintf("user-%d", clientUser.Id)

	messageBrokerService.AddMailbox(userPOId, myMailbox)
	goOnline(clientUser.Id)

	goOff := func() {
		goOffline(clientUser.Id, time.Now())
		messageBrokerService.RemoveMailbox(userPOId)
	}

	go goOnlineSocketControl(c, goOff)

	for data := range myMailbox {
		w_err := c.WriteJSON(data)
		if w_err != nil {
			log.Println(w_err)
			goOff()
			break
		}
	}
})

func goOnlineSocketControl(c *websocket.Conn, goOff func()) {
	for {
		var body struct{}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}
	}

	goOff()
}

var ChangeProfilePicture = authServices.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("user").(*appTypes.ClientUser)

	var w_err error

	for {
		var body changeProfilePictureBody

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

		if app_err := changeMyProfilePicture(clientUser.Id, clientUser.Username, body.PictureData); app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, fmt.Errorf("operation failed: %s", app_err)))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: fiber.StatusOK,
			Body: map[string]any{
				"msg": "Operation Successful",
			},
		})

	}
})

var UpdateMyGeolocation = authServices.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("user").(*appTypes.ClientUser)

	var body updateMyGeolocationBody

	var w_err error

	for {
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

		app_err := user.UpdateLocation(clientUser.Id, body.NewGeolocation)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: 200,
			Body: map[string]any{
				"msg": "Operation Successful",
			},
		})
	}
})

var GetAllUsers = authServices.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("user").(*appTypes.ClientUser)

	allUsers, app_err := user.GetAll(clientUser.Id)

	var w_err error

	for {
		if w_err != nil {
			log.Println(w_err)
			break
		}

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
		} else {
			w_err = c.WriteJSON(appTypes.WSResp{
				StatusCode: fiber.StatusOK,
				Body:       allUsers,
			})
		}

		var body struct{}
		if r_err := c.ReadJSON(&body); r_err != nil {
			log.Println(r_err)
			break
		}
	}
})

var SearchUser = authServices.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("user").(*appTypes.ClientUser)

	var w_err error

	for {
		var body struct {
			Query string
		}

		if w_err != nil {
			log.Println(w_err)
			break
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		searchResult, app_err := user.Search(clientUser.Id, body.Query)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: 200,
			Body:       searchResult,
		})
	}
})

var FindNearbyUsers = authServices.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("user").(*appTypes.ClientUser)

	var w_err error

	for {
		var body findNearbyUsersBody

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

		nearbyUsers, app_err := user.FindNearby(clientUser.Id, body.LiveLocation)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: 200,
			Body:       nearbyUsers,
		})
	}
})

// This handler merely get chats as is from the database, no updates accounted for yet.
// After closing this,  we must "GoOnline" to retrieve updates
var GetMyChats = authServices.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("user").(*appTypes.ClientUser)

	myChats, app_err := user.GetChats(clientUser.Id)

	var w_err error

	for {
		if w_err != nil {
			log.Println(w_err)
			break
		}

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusInternalServerError, app_err))
		} else {

			w_err = c.WriteJSON(appTypes.WSResp{
				StatusCode: fiber.StatusOK,
				Body:       myChats,
			})
		}

		var body struct{}
		if r_err := c.ReadJSON(&body); r_err != nil {
			log.Println(r_err)
			break
		}
	}
})

var CreateNewDMChatAndAckMessages = authServices.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("user").(*appTypes.ClientUser)

	var w_err error

	for {
		var body openDMChatStreamBody

		var newChatData newDMChatDataT

		// For DM Chat, we allowed options for both single and batch acknowledgements
		var ackMsgData ackMsgDataT

		var batchAckMsgData batchAckMsgDataT

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

		createNewChat := func() error {

			helpers.MapToStruct(body.Data, &newChatData)

			if val_err := newChatData.Validate(); val_err != nil {
				return c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			}

			initiatorData, app_err := newDMChat(
				clientUser.Id,
				newChatData.PartnerId,
				appServices.MessageBinaryToUrl(clientUser.Id, newChatData.InitMsg),
				newChatData.CreatedAt,
			)

			if app_err != nil {
				return c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			}

			return c.WriteJSON(initiatorData)

		}

		// acknowledge messages singly
		acknowledgeMessage := func() error {

			helpers.MapToStruct(body.Data, &ackMsgData)

			if val_err := ackMsgData.Validate(); val_err != nil {
				return c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			}

			go updateDMChatMessageDeliveryStatus(ackMsgData.DMChatId, ackMsgData.MsgId, ackMsgData.SenderId, clientUser.Id, ackMsgData.Status, ackMsgData.At)

			return nil
		}

		// acknowledge messages in batch
		batchAcknowledgeMessages := func() error {

			helpers.MapToStruct(body.Data, &batchAckMsgData)

			if val_err := batchAckMsgData.Validate(); val_err != nil {
				return c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			}

			go batchUpdateDMChatMessageDeliveryStatus(clientUser.Id, batchAckMsgData.Status, batchAckMsgData.MsgAckDatas)

			return nil
		}

		if body.Action == "create new chat" {

			w_err = createNewChat()

		} else if body.Action == "acknowledge message" {

			w_err = acknowledgeMessage()

		} else if body.Action == "batch acknowledge messages" {

			w_err = batchAcknowledgeMessages()

		} else {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, fmt.Errorf("invalid 'action' value")))
		}
	}
})

var CreateNewGroupChatAndAckMessages = authServices.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("user").(*appTypes.ClientUser)

	var w_err error

	for {
		var body openGroupChatStreamBody

		var newChatData newGroupChatDataT

		// For Group chat, messages should be acknowledged in batches,
		// and it's only for a single group chat at a time
		var ackMsgsData ackMsgsDataT

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

		createNewChat := func() error {

			helpers.MapToStruct(body.Data, &newChatData)

			if val_err := newChatData.Validate(); val_err != nil {
				return c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			}

			data, app_err := newGroupChat(
				newChatData.Name,
				newChatData.Description,
				newChatData.PictureData,
				[]string{fmt.Sprint(clientUser.Id), clientUser.Username},
				newChatData.InitUsers,
			)
			if app_err != nil {
				return c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			}

			return c.WriteJSON(data)
		}

		// For Group chat, messages can only be acknowledged in batches,
		acknowledgeMessages := func() error {

			helpers.MapToStruct(body.Data, &ackMsgsData)

			if val_err := ackMsgsData.Validate(); val_err != nil {
				return c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			}

			go batchUpdateGroupChatMessageDeliveryStatus(ackMsgsData.GroupChatId, clientUser.Id, ackMsgsData.Status, ackMsgsData.MsgAckDatas)

			return nil
		}

		if body.Action == "create new chat" {

			w_err = createNewChat()

		} else if body.Action == "acknowledge messages" {

			w_err = acknowledgeMessages()

		} else {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, fmt.Errorf("invalid 'action' value")))
		}
	}
})
