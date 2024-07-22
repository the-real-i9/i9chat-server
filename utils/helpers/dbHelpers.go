package helpers

import (
	"context"
	"errors"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func initDBPool() error {
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

func QueryRowType[T any](sql string, params ...any) (*T, error) {
	rows, _ := dbPool.Query(context.Background(), sql, params...)

	res, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByNameLax[T])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return res, nil
}

func QueryRowsType[T any](sql string, params ...any) ([]*T, error) {
	rows, _ := dbPool.Query(context.Background(), sql, params...)

	res, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[T])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return res, nil
}

func BatchQuery[T any](sqls []string, params [][]any) ([]*T, error) {
	var res = make([]*T, len(sqls))

	batch := &pgx.Batch{}

	for i, sql := range sqls {
		batch.Queue(sql, params[i]...).QueryRow(func(row pgx.Row) error {
			return row.Scan(res[i])
		})
	}

	s_err := dbPool.SendBatch(context.Background(), batch).Close()

	return res, s_err
}
