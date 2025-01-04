package userControllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"i9chat/appGlobals"
	"i9chat/appTypes"
	"i9chat/helpers"
	"i9chat/services/chatServices/dmChatService"
	"i9chat/services/chatServices/groupChatService"
	"i9chat/services/messageBrokerService"
	"i9chat/services/userService"
	"io"
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var OpenWSStream = websocket.New(func(c *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(*appTypes.ClientUser)

	app_err := userService.GoOnline(ctx, clientUser.Id)
	if app_err != nil {
		w_err := c.WriteJSON(helpers.ErrResp(app_err))
		if w_err != nil {
			log.Println(w_err)
			return
		}
	}

	r := messageBrokerService.ConsumeTopic(fmt.Sprintf("i9chat-user-%d-topic", clientUser.Id))

	goOff := func() {
		if err := r.Close(); err != nil {
			log.Println("failed to close reader:", err)
		}
		userService.GoOffline(context.TODO(), clientUser.Id, time.Now())
	}

	go clientEventStream(c, clientUser, goOff)

	for {
		m, err := r.FetchMessage(ctx)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				log.Println(err)
				goOff()
			}
			break
		}

		var msg any
		json.Unmarshal(m.Value, &msg)

		c.WriteJSON(msg)

		if err := r.CommitMessages(ctx, m); err != nil {
			log.Println("failed to commit messages:", err)
		}
	}
})

func clientEventStream(c *websocket.Conn, clientUser *appTypes.ClientUser, goOff func()) {

	var w_err error

	for {
		var body clientEventBody

		var dmChatMsgAckData dmChatMsgAckDataT

		var batchDMChatMsgAckData batchDMChatMsgAckDataT

		var groupChatMsgsAckData groupChatMsgsAckDataT

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
			w_err = c.WriteJSON(helpers.ErrResp(val_err))
			continue
		}

		// acknowledge messages singly
		acknowledgeDMChatMessage := func() error {

			helpers.MapToStruct(body.Data, &dmChatMsgAckData)

			if val_err := dmChatMsgAckData.Validate(); val_err != nil {
				return c.WriteJSON(helpers.ErrResp(val_err))
			}

			go dmChatService.UpdateMessageDeliveryStatus(context.TODO(), clientUser.Id, dmChatMsgAckData.PartnerUserId, dmChatMsgAckData.MsgId, dmChatMsgAckData.Status, dmChatMsgAckData.At)

			return nil
		}

		// acknowledge messages in batch
		batchAcknowledgeDMChatMessages := func() error {

			helpers.MapToStruct(body.Data, &batchDMChatMsgAckData)

			if val_err := batchDMChatMsgAckData.Validate(); val_err != nil {
				return c.WriteJSON(helpers.ErrResp(val_err))
			}

			go dmChatService.BatchUpdateMessageDeliveryStatus(context.TODO(), clientUser.Id, batchDMChatMsgAckData.Status, batchDMChatMsgAckData.MsgAckDatas)

			return nil
		}

		batchAcknowledgeGroupChatMessages := func() error {

			helpers.MapToStruct(body.Data, &groupChatMsgsAckData)

			if val_err := groupChatMsgsAckData.Validate(); val_err != nil {
				return c.WriteJSON(helpers.ErrResp(val_err))
			}

			go groupChatService.BatchUpdateMessageDeliveryStatus(context.TODO(), groupChatMsgsAckData.GroupChatId, clientUser.Id, groupChatMsgsAckData.Status, groupChatMsgsAckData.MsgAckDatas)

			return nil
		}

		if body.Action == "ACK dm chat message" {

			w_err = acknowledgeDMChatMessage()

		} else if body.Action == "batch ACK dm chat messages" {

			w_err = batchAcknowledgeDMChatMessages()

		} else if body.Action == "batch ACK group chat messages" {

			w_err = batchAcknowledgeGroupChatMessages()

		} else {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.NewError(fiber.StatusBadRequest, "invalid 'action' value")))
		}
	}

	goOff()
}

func ChangeProfilePicture(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(*appTypes.ClientUser)

	var body changeProfilePictureBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(val_err.Error())
	}

	respData, app_err := userService.ChangeProfilePicture(ctx, clientUser.Id, clientUser.Username, body.PictureData)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func UpdateMyLocation(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(*appTypes.ClientUser)

	var body updateMyGeolocationBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(val_err.Error())
	}

	respData, app_err := userService.UpdateMyLocation(ctx, clientUser.Id, body.NewGeolocation)

	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func GetAllUsers(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(*appTypes.ClientUser)

	respData, app_err := userService.GetAllUsers(ctx, clientUser.Id)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func SearchUser(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(*appTypes.ClientUser)

	query := c.Query("q")

	respData, app_err := userService.SearchUser(ctx, clientUser.Id, query)

	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func FindNearbyUsers(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(*appTypes.ClientUser)

	var body findNearbyUsersBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(val_err.Error())
	}

	respData, app_err := userService.FindNearbyUsers(ctx, clientUser.Id, body.LiveLocation)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func GetMyChats(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(*appTypes.ClientUser)

	respData, app_err := userService.GetMyChats(ctx, clientUser.Id)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func Logout(c *fiber.Ctx) error {
	sess, err := appGlobals.UserSessionStore.Get(c)
	if err != nil {
		log.Println("userControllers.go: Logout: UserSignupSession.Get:", err)
		return fiber.ErrInternalServerError
	}

	if err := sess.Destroy(); err != nil {
		log.Println("userControllers.go: Logout: sess.Destroy:", err)
		return fiber.ErrInternalServerError
	}

	return c.Status(fiber.StatusOK).SendString("You've been logged out!")
}
