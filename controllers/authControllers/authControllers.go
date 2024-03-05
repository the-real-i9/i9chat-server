package authcontrollers

import (
	"fmt"
	"log"
	"net/http"
	authservices "services/auth"

	"github.com/gofiber/fiber/v2"
)

func RequestNewAccount(c *fiber.Ctx) error {
	var p struct {
		Email string `json:"email"`
	}

	if err := c.BodyParser(&p); err != nil {
		log.Println(err)
		return err
	}

	_, _, err := authservices.RequestNewAccount(p.Email)
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	return c.SendStatus(http.StatusOK)
}
