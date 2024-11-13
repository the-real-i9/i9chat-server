package authUtilServices

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"i9chat/appTypes"
	"i9chat/helpers"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func HashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	return string(hash)
}

func PasswordMatchesHash(hash string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false
		} else {
			panic(err)
		}
	}

	return true
}

func GenerateVerifCodeExp() (int, time.Time) {
	return rand.Intn(899999) + 100000, time.Now().UTC().Add(1 * time.Hour)
}

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

	helpers.MapToStruct(token.Claims.(jwt.MapClaims)["data"].(map[string]any), &data)

	return &data, nil
}

func WSHandlerProtected(handler func(*websocket.Conn), config ...websocket.Config) func(*fiber.Ctx) error {
	return websocket.New(func(c *websocket.Conn) {
		authHeaderValue := c.Headers("Authorization")

		sessionToken := strings.Fields(authHeaderValue)[1]

		clientUser, err := JwtVerify[appTypes.ClientUser](sessionToken, os.Getenv("AUTH_JWT_SECRET"))
		if err != nil {
			if w_err := c.WriteJSON(helpers.ErrResp(fiber.StatusUnauthorized, err)); w_err != nil {
				log.Println(w_err)
			}
			return
		}

		c.Locals("user", clientUser)

		handler(c)
	}, config...)
}
