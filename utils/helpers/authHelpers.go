package helpers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func JwtSign(userData map[string]any, secret string, expires time.Time) string {
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	byteHeader, _ := json.Marshal(header)
	encodedHeader := base64.RawURLEncoding.EncodeToString(byteHeader)

	payload := map[string]any{
		"data": userData,
		"exp":  expires,
	}

	bytePayload, _ := json.Marshal(payload)
	encodedPayload := base64.RawURLEncoding.EncodeToString(bytePayload)

	h := hmac.New(sha256.New, []byte(secret))

	h.Write([]byte(encodedHeader + "." + encodedPayload))

	sig, _ := json.Marshal(h.Sum(nil))

	var signature string

	json.Unmarshal(sig, &signature)

	return fmt.Sprintf("%s.%s.%s", encodedHeader, encodedPayload, signature)
}

func JwtVerify(jwtToken string, secret string) (map[string]any, error) {
	jwtParts := strings.Split(jwtToken, ".")

	var (
		encodedHeader  = jwtParts[0]
		encodedPayload = jwtParts[1]
		signature      = jwtParts[2]
	)

	h := hmac.New(sha256.New, []byte(secret))

	h.Write([]byte(encodedHeader + "." + encodedPayload))

	expSig, _ := json.Marshal(h.Sum(nil))

	var expectedSignature string

	json.Unmarshal(expSig, &expectedSignature)

	tokenValid := hmac.Equal([]byte(signature), []byte(expectedSignature))
	if !tokenValid {
		return nil, fmt.Errorf("%s", "authorization error: invalid jwt")
	}

	decPay, _ := base64.RawURLEncoding.DecodeString(encodedPayload)

	var decodedPayload struct {
		Data map[string]any
		Exp  time.Time
	}

	json.Unmarshal(decPay, &decodedPayload)

	if decodedPayload.Exp.Before(time.Now()) {
		return nil, fmt.Errorf("%s", "authorization error: jwt expired")
	}

	return decodedPayload.Data, nil
}

func WSHandlerProtected(handler func(*websocket.Conn), config ...websocket.Config) func(*fiber.Ctx) error {
	return websocket.New(func(c *websocket.Conn) {
		jwtToken := c.Headers("Authorization")

		if jwtToken == "" {
			w_err := c.WriteJSON(AppError(fiber.StatusUnauthorized, fmt.Errorf("authorization error: authorization token required")))
			if w_err != nil {
				return
			}
			return
		}

		userData, err := JwtVerify(jwtToken, os.Getenv("AUTH_JWT_SECRET"))
		if err != nil {
			w_err := c.WriteJSON(AppError(fiber.StatusUnprocessableEntity, err))
			if w_err != nil {
				return
			}
			return
		}

		c.Locals("auth", userData)

		handler(c)
	}, config...)
}
