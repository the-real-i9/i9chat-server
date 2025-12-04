package userControllers

import (
	"context"
	"i9chat/src/appTypes"
	"i9chat/src/helpers"
	"i9chat/src/services/userService"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GetSessionUser(c *fiber.Ctx) error {
	clientUser := c.Locals("user").(appTypes.ClientUser)

	return c.JSON(clientUser)
}

func ChangeProfilePicture(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var body changeProfilePictureBody

	err := c.BodyParser(&body)
	if err != nil {
		return err
	}

	if err := body.Validate(); err != nil {
		return err
	}

	respData, authJwt, err := userService.ChangeProfilePicture(ctx, clientUser.Username, body.PictureData)
	if err != nil {
		return err
	}

	if respData != nil {
		c.Cookie(helpers.Cookie("user", helpers.ToJson(map[string]any{"authJwt": authJwt}), int(10*24*time.Hour/time.Second)))
	}

	return c.JSON(respData)
}

func ChangeBio(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var body changeBioBody

	err := c.BodyParser(&body)
	if err != nil {
		return err
	}

	if err := body.Validate(); err != nil {
		return err
	}

	respData, err := userService.ChangeBio(ctx, clientUser.Username, body.NewBio)
	if err != nil {
		return err
	}

	return c.JSON(respData)
}

func SetMyLocation(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var body updateMyGeolocationBody

	err := c.BodyParser(&body)
	if err != nil {
		return err
	}

	if err := body.Validate(); err != nil {
		return err
	}

	respData, err := userService.SetMyLocation(ctx, clientUser.Username, body.NewGeolocation)

	if err != nil {
		return err
	}

	return c.JSON(respData)
}

func FindUser(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	respData, err := userService.FindUser(ctx, c.Query("username"))

	if err != nil {
		return err
	}

	return c.JSON(respData)
}

func FindNearbyUsers(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := userService.FindNearbyUsers(ctx, clientUser.Username, c.QueryFloat("x"), c.QueryFloat("y"), c.QueryFloat("radius"))
	if err != nil {
		return err
	}

	return c.JSON(respData)
}

func GetMyChats(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := userService.GetMyChats(ctx, clientUser.Username, c.QueryInt("limit", 20), c.QueryFloat("cursor"))
	if err != nil {
		return err
	}

	return c.JSON(respData)
}

func GetMyProfile(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := userService.GetMyProfile(ctx, clientUser.Username)
	if err != nil {
		return err
	}

	return c.JSON(respData)
}

func SignOut(c *fiber.Ctx) error {
	c.Cookie(helpers.Cookie("user", "", 0))

	return c.JSON("You've been logged out!")
}
