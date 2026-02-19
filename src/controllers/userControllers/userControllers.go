package userControllers

import (
	"i9chat/src/appTypes"
	"i9chat/src/appTypes/UITypes"
	"i9chat/src/helpers"
	"i9chat/src/services/userService"

	"github.com/gofiber/fiber/v3"
)

func AuthorizePPicUpload(c fiber.Ctx) error {
	ctx := c.Context()

	var body authorizePPicUploadBody

	err := c.Bind().MsgPack(&body)
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

	return c.MsgPack(respData)
}

func GetSessionUser(c fiber.Ctx) error {
	clientUser := c.Locals("user").(appTypes.ClientUser)

	user, err := userService.SigninUserFind(c.Context(), clientUser.Username)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	if user.Username == "" {
		return fiber.ErrNotFound
	}

	return c.MsgPack(UITypes.ClientUser{Username: user.Username, ProfilePicUrl: user.ProfilePicUrl, Presence: user.Presence})
}

func ChangeProfilePicture(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var body changeProfilePictureBody

	err := c.Bind().MsgPack(&body)
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

	return c.MsgPack(respData)
}

func ChangeBio(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var body changeBioBody

	err := c.Bind().MsgPack(&body)
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

	return c.MsgPack(respData)
}

func SetMyLocation(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var body updateMyGeolocationBody

	err := c.Bind().MsgPack(&body)
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

	return c.MsgPack(respData)
}

func FindUser(c fiber.Ctx) error {
	ctx := c.Context()

	respData, err := userService.FindUser(ctx, c.Query("username"))

	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func FindNearbyUsers(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var query struct {
		X      float64
		Y      float64
		Radius float64
	}

	if err := c.Bind().Query(&query); err != nil {
		return err
	}

	respData, err := userService.FindNearbyUsers(ctx, clientUser.Username, query.X, query.Y, query.Radius)
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func GetMyChats(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	var query struct {
		Limit  int64
		Cursor float64
	}

	if err := c.Bind().Query(&query); err != nil {
		return err
	}

	respData, err := userService.GetMyChats(ctx, clientUser.Username, helpers.CoalesceInt(query.Limit, 20), query.Cursor)
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func GetMyProfile(c fiber.Ctx) error {
	ctx := c.Context()

	clientUser := c.Locals("user").(appTypes.ClientUser)

	respData, err := userService.GetMyProfile(ctx, clientUser.Username)
	if err != nil {
		return err
	}

	return c.MsgPack(respData)
}

func SignOut(c fiber.Ctx) error {
	c.ClearCookie()

	return c.MsgPack("You've been logged out!")
}
