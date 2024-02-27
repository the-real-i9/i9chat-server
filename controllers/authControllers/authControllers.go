package authcontrollers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func Signin(c *fiber.Ctx) error {
	var p struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&p); err != nil {
		log.Println(err)
		return err
	}

	fmt.Printf("{ email: %s, password: %s }\n", p.Email, p.Password)

	return c.SendStatus(http.StatusOK)
}
