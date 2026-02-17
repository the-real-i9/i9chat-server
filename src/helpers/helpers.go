package helpers

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/vmihailenco/msgpack/v5"
)

func JoinWithCommaAnd(items ...string) string {
	n := len(items)
	if n == 0 {
		return ""
	}
	if n == 1 {
		return items[0]
	}
	if n == 2 {
		return items[0] + " and " + items[1]
	}

	// Join all except the last with commas, then append "and last"
	return strings.Join(items[:n-1], ", ") + ", and " + items[n-1]
}

func WSErrReply(err error, toAction string) map[string]any {

	errCode := fiber.StatusInternalServerError

	if ferr, ok := err.(*fiber.Error); ok {
		errCode = ferr.Code
	}

	errResp := map[string]any{
		"event":    "server error",
		"toAction": toAction,
		"data": map[string]any{
			"statusCode": errCode,
			"errorMsg":   fmt.Sprint(err),
		},
	}

	return errResp
}

func WSReply(data any, toAction string) map[string]any {

	reply := map[string]any{
		"event":    "server reply",
		"toAction": toAction,
		"data":     data,
	}

	return reply
}

func Session(kvPairs map[string]any, path string, maxAge int) *fiber.Cookie {
	c := &fiber.Cookie{
		HTTPOnly: true,
		Secure:   false,
		Domain:   os.Getenv("SERVER_HOST"),
	}

	c.Name = "session"
	c.Value = base64.RawURLEncoding.EncodeToString(ToBtMsgPack(kvPairs))
	c.Path = path
	c.MaxAge = maxAge

	return c
}

func LogError(err error) {
	if err == nil {
		return
	}

	pc, file, line, ok := runtime.Caller(1)
	fn := "unknown"
	if !ok {
		file = "???"
		line = 0
	} else {
		fn = runtime.FuncForPC(pc).Name()
	}

	log.Printf("[ERROR] %s:%d %s(): %v\n", file, line, fn, err)
}

func ToMsgPack(data any) string {
	d, err := msgpack.Marshal(data)
	if err != nil {
		LogError(err)
	}
	return string(d)
}

func ToBtMsgPack(data any) []byte {
	d, err := msgpack.Marshal(data)
	if err != nil {
		LogError(err)
	}
	return d
}

func ToJson(data any) string {
	d, err := json.Marshal(data)
	if err != nil {
		LogError(err)
	}
	return string(d)
}

func FromMsgPack[T any](msgPackStr string) (res T) {
	err := msgpack.Unmarshal([]byte(msgPackStr), &res)
	if err != nil {
		LogError(err)
	}

	return
}

func FromJson[T any](jsonStr string) (res T) {
	err := json.Unmarshal([]byte(jsonStr), &res)
	if err != nil {
		LogError(err)
	}

	return
}

func FromBtMsgPack[T any](msgPackBt []byte) (res T) {
	err := msgpack.Unmarshal(msgPackBt, &res)
	if err != nil {
		LogError(err)
	}

	return
}

func MaxCursor(cursor float64) string {
	if cursor == 0 {
		return "+inf"
	}

	return fmt.Sprintf("(%f", cursor)
}

func CoalesceInt(input int64, def int64) int64 {
	if input == 0 {
		return def
	}

	return input
}
