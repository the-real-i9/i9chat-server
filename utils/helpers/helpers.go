package helpers

import (
	"bufio"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func LoadEnv() error {
	dotenv, err := os.Open(".env")
	if err != nil {
		return err
	}

	env := bufio.NewScanner(dotenv)

	for env.Scan() {
		key, value, found := strings.Cut(env.Text(), "=")
		if !found || strings.HasPrefix(key, "#") {
			continue
		}

		err := os.Setenv(key, value)
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}

func JwtSign(userData map[string]any) string {
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	byteHeader, _ := json.Marshal(header)
	encodedHeader := base64.RawURLEncoding.EncodeToString(byteHeader)

	payload := map[string]any{
		"data": userData,
		"jwtClaims": map[string]any{
			"issuer": "i9chat",
			"iat":    time.Now(),
			"exp":    time.Now().Add(24 * time.Hour),
		},
	}

	bytePayload, _ := json.Marshal(payload)
	encodedPayload := base64.RawURLEncoding.EncodeToString(bytePayload)

	h := hmac.New(sha256.New, []byte(os.Getenv("JWT_SECRET")))

	h.Write([]byte(encodedHeader + "." + encodedPayload))

	sig, _ := json.Marshal(h.Sum(nil))

	var signature string

	json.Unmarshal(sig, &signature)

	return fmt.Sprintf("%s.%s.%s", encodedHeader, encodedPayload, signature)
}

func JwtParse(jwtToken string) (map[string]any, error) {
	jwtParts := strings.Split(jwtToken, ".")

	var (
		encodedHeader  = jwtParts[0]
		encodedPayload = jwtParts[1]
		signature      = jwtParts[2]
	)

	h := hmac.New(sha256.New, []byte(os.Getenv("JWT_SECRET")))

	h.Write([]byte(encodedHeader + "." + encodedPayload))

	expSig, _ := json.Marshal(h.Sum(nil))

	var expectedSignature string

	json.Unmarshal(expSig, &expectedSignature)

	tokenValid := hmac.Equal([]byte(signature), []byte(expectedSignature))
	if !tokenValid {
		return nil, fmt.Errorf("%s", "invalid jwt")
	}

	decPay, _ := base64.RawURLEncoding.DecodeString(encodedPayload)

	var decodedPayload struct {
		Data      map[string]any
		JwtClaims struct {
			Exp time.Time
		}
	}

	json.Unmarshal(decPay, &decodedPayload)

	if decodedPayload.JwtClaims.Exp.Before(time.Now()) {
		return nil, fmt.Errorf("%s", "jwt expired")
	}

	return decodedPayload.Data, nil
}

func GetDBPool() (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), os.Getenv("PGDATABASE_URL"))
	return pool, err
}

func QueryInstance[T any](sql string, params ...any) (*T, error) {
	pool, err := GetDBPool()
	if err != nil {
		return nil, err
	}

	rows, _ := pool.Query(context.Background(), sql, params...)

	data, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[T])
	if err != nil {
		return nil, err
	}

	return data, nil
}

func QueryInstances[T any](sql string, params ...any) ([]*T, error) {
	pool, err := GetDBPool()
	if err != nil {
		return nil, err
	}

	rows, _ := pool.Query(context.Background(), sql, params...)

	data, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[T])
	if err != nil {
		return nil, err
	}

	return data, nil
}

func QueryRowField[T any](sql string, params ...any) (*T, error) {
	pool, err := GetDBPool()
	if err != nil {
		return nil, err
	}

	rows, _ := pool.Query(context.Background(), sql, params...)

	res, err := pgx.CollectOneRow(rows, pgx.RowToAddrOf[T])
	if err != nil {
		return nil, err
	}

	return res, err
}

func QueryRowsField[T any](sql string, params ...any) ([]*T, error) {
	pool, err := GetDBPool()
	if err != nil {
		return nil, err
	}

	rows, _ := pool.Query(context.Background(), sql, params...)

	res, err := pgx.CollectRows(rows, pgx.RowToAddrOf[T])
	if err != nil {
		return nil, err
	}

	return res, nil
}

func QueryRowFields(sql string, params ...any) (map[string]any, error) {
	pool, err := GetDBPool()
	if err != nil {
		return nil, err
	}

	rows, _ := pool.Query(context.Background(), sql, params...)

	res, err := pgx.CollectOneRow(rows, pgx.RowToMap)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func QueryRowsFields(sql string, params ...any) ([]map[string]any, error) {
	pool, err := GetDBPool()
	if err != nil {
		return nil, err
	}

	rows, _ := pool.Query(context.Background(), sql, params...)

	res, err := pgx.CollectRows(rows, pgx.RowToMap)
	if err != nil {
		return nil, err
	}

	return res, nil
}
