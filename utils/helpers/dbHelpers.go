package helpers

import (
	"context"
	"errors"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var dbPool *pgxpool.Pool

func InitDBPool() error {
	pool, err := pgxpool.New(context.Background(), os.Getenv("PGDATABASE_URL"))
	if err != nil {
		return err
	}
	dbPool = pool

	return nil
}

func QueryRowField[T any](sql string, params ...any) (*T, error) {
	rows, _ := dbPool.Query(context.Background(), sql, params...)

	res, err := pgx.CollectOneRow(rows, pgx.RowToAddrOf[T])
	if err != nil {
		return nil, err
	}

	return res, err
}

func QueryRowsField[T any](sql string, params ...any) ([]*T, error) {
	rows, _ := dbPool.Query(context.Background(), sql, params...)

	res, err := pgx.CollectRows(rows, pgx.RowToAddrOf[T])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return make([]*T, 0), nil
		}
		return nil, err
	}

	return res, nil
}

func QueryRowFields(sql string, params ...any) (map[string]any, error) {
	rows, _ := dbPool.Query(context.Background(), sql, params...)

	res, err := pgx.CollectOneRow(rows, pgx.RowToMap)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return res, nil
}

func QueryRowsFields(sql string, params ...any) ([]map[string]any, error) {
	rows, _ := dbPool.Query(context.Background(), sql, params...)

	res, err := pgx.CollectRows(rows, pgx.RowToMap)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return make([]map[string]any, 0), nil
		}
		return nil, err
	}

	return res, nil
}

func BatchQuery[T any](sqls []string, params [][]any) (*T, error) {
	var res *T

	batch := &pgx.Batch{}

	for i, sql := range sqls {
		batch.Queue(sql, params[i]...).QueryRow(func(row pgx.Row) error {
			return row.Scan(res)
		})
	}

	s_err := dbPool.SendBatch(context.Background(), batch).Close()

	return res, s_err
}
