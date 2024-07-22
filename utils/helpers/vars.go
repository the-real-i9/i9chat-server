package helpers

import (
	"errors"

	"cloud.google.com/go/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInternalServerError = errors.New("internal server error: check logger")

var GCSClient *storage.Client

var dbPool *pgxpool.Pool
