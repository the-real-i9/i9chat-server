package helpers

import (
	"i9chat/utils/appTypes"
	"os"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func JwtSign(data any, secret string, expires time.Time) string {
	// create token -> (header.payload)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"data":  data,
		"admin": true,
		"exp":   expires,
	})

	// sign token with secret -> (header.payload.signature)
	jwt, err := token.SignedString([]byte(os.Getenv("SIGNUP_SESSION_JWT_SECRET")))
	if err != nil {
		panic(err)
	}

	return jwt
}

func WSHandlerProtected(handler func(*websocket.Conn), config ...websocket.Config) func(*fiber.Ctx) error {
	return websocket.New(func(c *websocket.Conn) {
		jwtData := c.Locals("auth").(*jwt.Token).Claims.(jwt.MapClaims)["data"].(map[string]any)

		var clientUser appTypes.ClientUser

		MapToStruct(jwtData, &clientUser)

		c.Locals("auth", clientUser)

		handler(c)
	}, config...)
}
