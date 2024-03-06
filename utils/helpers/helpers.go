package helpers

import (
	"bufio"
	"context"
	"log"
	"os"
	"strings"

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

func GetDBPool() (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), os.Getenv("PGDATABASE_URL"))
	return pool, err
}

func QueryRow[T any](sql string, params ...any) (*T, error) {
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

func QueryRows[T any](sql string, params ...any) ([]*T, error) {
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
