package securityServices

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"i9chat/appGlobals"
	"i9chat/appTypes"
	"i9chat/helpers"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("securityServices.go: HashPassword:", err)
		return "", appGlobals.ErrInternalServerError
	}

	return string(hash), nil
}

func PasswordMatchesHash(hash string, password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		} else {
			log.Println("securityServices.go: HashPassword:", err)
			return false, appGlobals.ErrInternalServerError
		}
	}

	return true, nil
}

func GenerateVerifCodeExp() (int, time.Time) {
	return rand.Intn(899999) + 100000, time.Now().UTC().Add(1 * time.Hour)
}

func JwtSign(data any, secret string, expires time.Time) (string, error) {
	// create token -> (header.payload)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"data": data,
		"exp":  expires.Unix(),
	})

	// sign token with secret -> (header.payload.signature)
	jwt, err := token.SignedString([]byte(secret))

	if err != nil {
		log.Println("securityServices.go: JwtSign:", err)
		return "", appGlobals.ErrInternalServerError
	}

	return jwt, err
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
		if errors.Is(err, jwt.ErrTokenMalformed) ||
			errors.Is(err, jwt.ErrTokenSignatureInvalid) ||
			errors.Is(err, jwt.ErrTokenUnverifiable) ||
			errors.Is(err, jwt.ErrTokenInvalidClaims) ||
			errors.Is(err, jwt.ErrTokenExpired) {

			return nil, fiber.NewError(fiber.StatusUnauthorized, "jwt error:", err.Error())
		}

		log.Println("securityServices.go: JwtVerify:", err)
		return nil, appGlobals.ErrInternalServerError
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
			if w_err := c.WriteJSON(helpers.ErrResp(err)); w_err != nil {
				log.Println(w_err)
			}
			return
		}

		c.Locals("user", clientUser)

		handler(c)
	}, config...)
}
