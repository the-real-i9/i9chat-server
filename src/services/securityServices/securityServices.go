package securityServices

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"i9chat/src/helpers"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/gofiber/fiber/v2"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("securityServices.go: HashPassword:", err)
		return "", fiber.ErrInternalServerError
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
			return false, fiber.ErrInternalServerError
		}
	}

	return true, nil
}

func GenerateVerifCodeExp() (string, time.Time) {
	var token string
	expires := time.Now().UTC().Add(1 * time.Hour)

	if os.Getenv("GO_ENV") != "production" {
		token = os.Getenv("DUMMY_VERF_TOKEN")
	} else {
		token = fmt.Sprint(rand.Intn(899999) + 100000)
	}

	return token, expires
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
		return "", fiber.ErrInternalServerError
	}

	return jwt, err
}

func JwtVerify[T any](tokenString, secret string) (T, error) {
	var data T

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

			return data, fiber.NewError(fiber.StatusUnauthorized, "jwt error:", err.Error())
		}

		log.Println("securityServices.go: JwtVerify:", err)
		return data, fiber.ErrInternalServerError
	}

	helpers.ToStruct(token.Claims.(jwt.MapClaims)["data"], &data)

	return data, nil
}
