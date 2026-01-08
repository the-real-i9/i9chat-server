package userControllers

import (
	"i9chat/src/appTypes"
	"i9chat/src/appTypes/UITypes"
	"i9chat/src/cache"
	"i9chat/src/services/userService"

	"github.com/gofiber/fiber/v2"
)

func AuthorizePPicUpload(c *fiber.Ctx) error {
	ctx := c.Context()

	var body authorizePPicUploadBody

	err := c.BodyParser(&body)
	if err != nil {
		return err
	}

	if err = body.Validate(); err != nil {
		return err
	}

	respData, err := userService.AuthorizePPicUpload(ctx, body.PicMIME)
	if err != nil {
		return err
	}

	return c.JSON(respData)
}

func GetSessionUser(c *fiber.Ctx) error {
	clientUser := c.Locals("user").(appTypes.ClientUser)

	user, err := cache.GetUser[UITypes.ClientUser](c.Context(), clientUser.Username)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	return c.JSON(user)
}

func ChangeProfilePicture(c *fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var body changeProfilePictureBody

	err := c.BodyParser(&body)
	if err != nil {
		return err
	}

	if err := body.Validate(ctx); err != nil {
		return err
	}

	respData, err := userService.ChangeProfilePicture(ctx, clientUser.Username, body.ProfilePicCloudName)
	if err != nil {
		return err
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
	ctx := c.Context()

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
	c.ClearCookie()

	return c.JSON("You've been logged out!")
}
