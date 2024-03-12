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

func GenerateJwtToken(userData map[string]any) string {
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	byteHeader, _ := json.Marshal(header)
	encodedHeader := base64.RawURLEncoding.EncodeToString(byteHeader)

	payload := map[string]any{
		"data": userData,
		"jwtClaims": map[string]any{
			"issuer": "i9chat",
			"iat":    time.Now().UnixMilli(),
			"exp":    time.Now().Add(1 * time.Hour).UnixMilli(),
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
