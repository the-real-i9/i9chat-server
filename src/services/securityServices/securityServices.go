package securityServices

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"i9chat/src/helpers"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/gofiber/fiber/v3"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		helpers.LogError(err)
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
			helpers.LogError(err)
			return false, fiber.ErrInternalServerError
		}
	}

	return true, nil
}

func GenerateTokenCodeExp() (string, time.Time) {
	var token string
	expires := time.Now().UTC().Add(1 * time.Hour)

	if os.Getenv("GO_ENV") != "production" {
		token = os.Getenv("DUMMY_TOKEN")
	} else {
		token = fmt.Sprint(rand.Intn(899999) + 100000)
	}

	return token, expires
}

func JwtSign(data any, secret string, expires time.Time) (string, error) {
	// create token -> (header.payload)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"data": helpers.ToJson(data),
		"exp":  expires.Unix(),
	})

	// sign token with secret -> (header.payload.signature)
	jwt, err := token.SignedString([]byte(secret))

	if err != nil {
		helpers.LogError(err)
		return "", fiber.ErrInternalServerError
	}

	return jwt, err
}

func JwtVerify[T any](tokenString, secret string) (T, error) {
	var data T

	parser := jwt.NewParser()
	token, err := parser.Parse(tokenString, func(t *jwt.Token) (any, error) {

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

		helpers.LogError(err)
		return data, fiber.ErrInternalServerError
	}

	data = helpers.FromJson[T](token.Claims.(jwt.MapClaims)["data"].(string))

	return data, nil
}
