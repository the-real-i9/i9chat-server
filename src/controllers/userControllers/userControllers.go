package userControllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"i9chat/src/appTypes"
	"i9chat/src/helpers"
	"i9chat/src/services/messageBrokerService"
	"i9chat/src/services/userService"
	"io"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/segmentio/kafka-go"
)

var OpenWSStream = websocket.New(func(c *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	go userService.GoOnline(ctx, clientUser.Username)

	r := messageBrokerService.ConsumeTopic(fmt.Sprintf("user-%s-alerts", clientUser.Username))

	go serverToClientStream(c, r)

	// clientToServerStream
	var w_err error

	for {
		var body clientEventBody

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
			w_err = c.WriteJSON(helpers.WSErrResp(val_err, body.Event))
			continue
		}

		switch body.Event {
		case "new dm chat message":
			respData, err := newDMChatMsgHndl(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrResp(err, body.Event))
				continue
			}
			c.WriteJSON(respData)
		case "ack dm chat message delivered":

			err := ackDMChatMsgDeliveredHndl(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrResp(err, body.Event))
				continue
			}
		case "ack dm chat message read":

			err := ackDMChatMsgReadHndl(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrResp(err, body.Event))
				continue
			}

		case "get dm chat history":

			respData, err := getDMChatHistoryHndl(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrResp(err, body.Event))
				continue
			}
			c.WriteJSON(respData)
		case "new group chat message":

			respData, err := newGroupChatMsgHndl(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrResp(err, body.Event))
				continue
			}
			c.WriteJSON(respData)
		case "ack group chat message delivered":

			err := ackGroupChatMsgDeliveredHndl(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrResp(err, body.Event))
				continue
			}
		case "ack group chat message read":

			err := ackGroupChatMsgReadHndl(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrResp(err, body.Event))
				continue
			}
		case "get group chat history":

			respData, err := getGroupChatHistoryHndl(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrResp(err, body.Event))
				continue
			}
			c.WriteJSON(respData)
		case "get group info":

			respData, err := getGroupInfoHndl(ctx, body.Data)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrResp(err, body.Event))
				continue
			}
			c.WriteJSON(respData)
		case "get group membership info":

			respData, err := getGroupMemInfoHndl(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrResp(err, body.Event))
				continue
			}
			c.WriteJSON(respData)
		default:
			w_err = c.WriteJSON(helpers.WSErrResp(fmt.Errorf("invalid event"), body.Event))
			continue
		}
	}

	go func() {
		userService.GoOffline(context.Background(), clientUser.Username)

		if err := r.Close(); err != nil {
			log.Println("failed to close reader:", err)
		}
	}()
})

func serverToClientStream(c *websocket.Conn, r *kafka.Reader) {

	ctx := context.Background()

	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				log.Println("userControllers.go: serverToClientStream: r.ReadMessage:", err)
			}
			break
		}

		var msg any
		json.Unmarshal(m.Value, &msg)

		c.WriteJSON(msg)
	}
}

func ChangeProfilePicture(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var body changeProfilePictureBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return val_err
	}

	respData, app_err := userService.ChangeProfilePicture(ctx, clientUser.Username, body.PictureData)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func ChangePhone(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var body changePhoneBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return val_err
	}

	respData, app_err := userService.ChangePhone(ctx, clientUser.Username, body.Phone)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func UpdateMyLocation(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var body updateMyGeolocationBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return val_err
	}

	respData, app_err := userService.UpdateMyLocation(ctx, clientUser.Username, body.NewGeolocation)

	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func FindUser(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var query searchUserQuery

	q_err := c.QueryParser(&query)
	if q_err != nil {
		return q_err
	}

	if val_err := query.Validate(); val_err != nil {
		return val_err
	}

	respData, app_err := userService.FindUser(ctx, query.EmailUsernamePhone)

	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func FindNearbyUsers(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var query findNearbyUsersQuery

	query_err := c.QueryParser(&query)
	if query_err != nil {
		return query_err
	}

	if val_err := query.Validate(); val_err != nil {
		return val_err
	}

	respData, app_err := userService.FindNearbyUsers(ctx, clientUser.Username, query.X, query.Y, query.Radius)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func GetMyChats(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, app_err := userService.GetMyChats(ctx, clientUser.Username)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func GetMyProfile(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, app_err := userService.GetMyProfile(ctx, clientUser.Username)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func SignOut(c *fiber.Ctx) error {
	c.ClearCookie("user")

	return c.SendString("You've been logged out!")
}
