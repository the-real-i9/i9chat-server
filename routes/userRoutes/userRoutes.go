package userroutes

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	fmt.Println(router)
}
