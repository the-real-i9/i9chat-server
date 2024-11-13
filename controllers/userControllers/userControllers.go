package userControllers

import (
	"fmt"
	"i9chat/appTypes"
	"i9chat/helpers"
	"i9chat/services/userService"
	"i9chat/services/utils/authUtilServices"
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// This Controller essentially opens the stream for receiving messages
var GoOnline = authUtilServices.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("user").(*appTypes.ClientUser)

	// a channel for streaming data to client
	var myMailbox = make(chan any)

	userPOId := fmt.Sprintf("user-%d", clientUser.Id)

	app_err := userService.GoOnline(clientUser.Id, userPOId, myMailbox)
	if app_err != nil {
		w_err := c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, fmt.Errorf("operation failed: %s", app_err)))
		if w_err != nil {
			log.Println(w_err)
			return
		}
	}

	goOff := func() {
		userService.GoOffline(clientUser.Id, time.Now(), userPOId)
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

var ChangeProfilePicture = authUtilServices.WSHandlerProtected(func(c *websocket.Conn) {
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

		respData, app_err := userService.ChangeProfilePicture(clientUser.Id, clientUser.Username, body.PictureData)
		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, fmt.Errorf("operation failed: %s", app_err)))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: fiber.StatusOK,
			Body:       respData,
		})

	}
})

var UpdateMyLocation = authUtilServices.WSHandlerProtected(func(c *websocket.Conn) {
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

		respData, app_err := userService.UpdateMyLocation(clientUser.Id, body.NewGeolocation)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: 200,
			Body:       respData,
		})
	}
})

var GetAllUsers = authUtilServices.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("user").(*appTypes.ClientUser)

	respData, app_err := userService.GetAllUsers(clientUser.Id)

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
				Body:       respData,
			})
		}

		var body struct{}
		if r_err := c.ReadJSON(&body); r_err != nil {
			log.Println(r_err)
			break
		}
	}
})

var SearchUser = authUtilServices.WSHandlerProtected(func(c *websocket.Conn) {
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

		respData, app_err := userService.SearchUser(clientUser.Id, body.Query)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: 200,
			Body:       respData,
		})
	}
})

var FindNearbyUsers = authUtilServices.WSHandlerProtected(func(c *websocket.Conn) {
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

		respData, app_err := userService.FindNearbyUsers(clientUser.Id, body.LiveLocation)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: 200,
			Body:       respData,
		})
	}
})

// This handler merely get chats as is from the database, no updates accounted for yet.
// After closing this,  we must "GoOnline" to retrieve updates
var GetMyChats = authUtilServices.WSHandlerProtected(func(c *websocket.Conn) {
	clientUser := c.Locals("user").(*appTypes.ClientUser)

	respData, app_err := userService.GetMyChats(clientUser.Id)

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
				Body:       respData,
			})
		}

		var body struct{}
		if r_err := c.ReadJSON(&body); r_err != nil {
			log.Println(r_err)
			break
		}
	}
})
