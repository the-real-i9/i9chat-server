package helpers

import (
	"fmt"
	"i9chat/appTypes"
	"log"
	"os"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func JwtSign(data any, secret string, expires time.Time) string {
	// create token -> (header.payload)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"data": data,
		"exp":  expires.Unix(),
	})

	// sign token with secret -> (header.payload.signature)
	jwt, err := token.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}

	return jwt
}

func JwtVerify[T any](tokenString, secret string) (*T, error) {
	parser := jwt.NewParser()
	token, err := parser.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {

		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	var data T

	MapToStruct(token.Claims.(jwt.MapClaims)["data"].(map[string]any), &data)

	return &data, nil
}

func WSHandlerProtected(handler func(*websocket.Conn), config ...websocket.Config) func(*fiber.Ctx) error {
	return websocket.New(func(c *websocket.Conn) {
		sessionToken := c.Headers("Authorization")

		clientUser, err := JwtVerify[appTypes.ClientUser](sessionToken, os.Getenv("AUTH_JWT_SECRET"))
		if err != nil {
			if w_err := c.WriteJSON(ErrResp(fiber.StatusUnauthorized, err)); w_err != nil {
				log.Println(w_err)
			}
			return
		}

		c.Locals("user", clientUser)

		handler(c)
	}, config...)
}
