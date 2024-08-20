package globals

import (
	"errors"

	"cloud.google.com/go/storage"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInternalServerError = errors.New("internal server error: check logger")

var GCSClient *storage.Client

var DBPool *pgxpool.Pool

var SignupSessionStore *session.Store
