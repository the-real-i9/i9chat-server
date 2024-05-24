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
	// base64-encode header
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	byteHeader, _ := json.Marshal(header)
	encodedHeader := base64.RawURLEncoding.EncodeToString(byteHeader)

	// base64-encode header
	payload := map[string]any{
		"data": userData,
		"exp":  expires,
	}

	bytePayload, _ := json.Marshal(payload)
	encodedPayload := base64.RawURLEncoding.EncodeToString(bytePayload)

	// generate HMAC signature | sign the `header.payload` portion
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(encodedHeader + "." + encodedPayload))
	hashRes := h.Sum(nil)

	signature := base64.RawURLEncoding.EncodeToString(hashRes[:])

	// construct jwt
	jwt := fmt.Sprintf("%s.%s.%s", encodedHeader, encodedPayload, signature)

	return jwt
}

func parseJwtString(jwt string) (encodedHeader, encodedPayload, signature string, err error) {
	token, found := strings.CutPrefix(jwt, "Bearer ")
	if !found {
		return "", "", "", fmt.Errorf("%s", "authorization error: jwt missing bearer prefix")
	}

	jwtParts := strings.Split(token, ".")

	return jwtParts[0], jwtParts[1], jwtParts[2], nil
}

func JwtVerify(jwt string, secret string) (map[string]any, error) {

	encodedHeader, encodedPayload, signature, p_err := parseJwtString(jwt)
	if p_err != nil {
		return nil, p_err
	}

	// generate HMAC expected signature | re-sign the `header.payload` portion
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(encodedHeader + "." + encodedPayload))
	hashRes := h.Sum(nil)

	expectedSignature := base64.RawURLEncoding.EncodeToString(hashRes[:])

	// check jwt validity
	jwtValid := hmac.Equal([]byte(signature), []byte(expectedSignature))
	if !jwtValid {
		return nil, fmt.Errorf("%s", "authorization error: invalid jwt")
	}

	// decode the payload
	decPay, _ := base64.RawURLEncoding.DecodeString(encodedPayload)

	var decodedPayload struct {
		Data map[string]any
		Exp  time.Time
	}

	json.Unmarshal(decPay, &decodedPayload)

	// check jwt expiration
	if decodedPayload.Exp.Before(time.Now()) {
		return nil, fmt.Errorf("%s", "authorization error: jwt expired")
	}

	return decodedPayload.Data, nil
}

func WSHandlerProtected(handler func(*websocket.Conn), config ...websocket.Config) func(*fiber.Ctx) error {
	return websocket.New(func(c *websocket.Conn) {
		authJwt := c.Headers("Authorization")

		if authJwt == "" {
			w_err := c.WriteJSON(ErrResp(fiber.StatusUnauthorized, fmt.Errorf("authorization error: authorization token required")))
			if w_err != nil {
				return
			}
			return
		}

		userData, err := JwtVerify(authJwt, os.Getenv("AUTH_JWT_SECRET"))
		if err != nil {
			w_err := c.WriteJSON(ErrResp(fiber.StatusUnauthorized, err))
			if w_err != nil {
				return
			}
			return
		}

		c.Locals("auth", userData)

		handler(c)
	}, config...)
}
