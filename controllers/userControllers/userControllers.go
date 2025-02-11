package userControllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"i9chat/appGlobals"
	"i9chat/appTypes"
	"i9chat/helpers"
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

	userService.GoOnline(ctx, clientUser.Username)

	r := messageBrokerService.ConsumeTopic(fmt.Sprintf("i9chat-user-%s-topic", clientUser.Username))

	goOff := func() {
		if err := r.Close(); err != nil {
			log.Println("failed to close reader:", err)
		}
		userService.GoOffline(ctx, clientUser.Username, time.Now())
	}

	go clientEventStream(c, ctx, clientUser, goOff)

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

// Events
//   - new dm chat message
//   - dm chat message delivered ack
//   - dm chat message read ack
//   - new group chat message
//   - group chat message delivered ack
//   - group chat message read ack
func clientEventStream(c *websocket.Conn, ctx context.Context, clientUser *appTypes.ClientUser, goOff func()) {

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
			// do
			respData, err := newDMChatMsgEventHandler(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrResp(err, body.Event))
				continue
			}
			c.WriteJSON(respData)
		case "dm chat message delivered ack":

			err := dmChatMsgDeliveredAckEventHandler(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrResp(err, body.Event))
				continue
			}

		case "dm chat message read ack":

			err := dmChatMsgReadAckEventHandler(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrResp(err, body.Event))
				continue
			}

		case "new group chat message":
			// do
			respData, err := newGroupChatMsgEventHandler(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrResp(err, body.Event))
				continue
			}
			c.WriteJSON(respData)
		case "group chat message delivered ack":
			err := groupChatMsgDeliveredAckEventHandler(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrResp(err, body.Event))
				continue
			}
		case "group chat message read ack":
			err := groupChatMsgReadAckEventHandler(ctx, clientUser.Username, body.Data)
			if err != nil {
				w_err = c.WriteJSON(helpers.WSErrResp(err, body.Event))
				continue
			}
		default:
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
		return val_err
	}

	respData, app_err := userService.ChangeProfilePicture(ctx, clientUser.Username, body.PictureData)
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
		return val_err
	}

	respData, app_err := userService.UpdateMyLocation(ctx, clientUser.Username, body.NewGeolocation)

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

	respData, app_err := userService.SearchUser(ctx, clientUser.Username, query)

	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func FindNearbyUsers(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(*appTypes.ClientUser)

	var query findNearbyUsersQuery

	query_err := c.QueryParser(&query)
	if query_err != nil {
		return query_err
	}

	if val_err := query.Validate(); val_err != nil {
		return val_err
	}

	respData, app_err := userService.FindNearbyUsers(ctx, clientUser.Username, query.Long, query.Lat, query.Radius)
	if app_err != nil {
		return app_err
	}

	return c.JSON(respData)
}

func GetMyChats(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientUser := c.Locals("user").(*appTypes.ClientUser)

	respData, app_err := userService.GetMyChats(ctx, clientUser.Username)
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

	return fiber.NewError(fiber.StatusOK, "You've been logged out!")
}
