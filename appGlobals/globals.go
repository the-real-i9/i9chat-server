package appGlobals

import (
	"cloud.google.com/go/storage"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

var GCSClient *storage.Client

var DBPool *pgxpool.Pool

var SignupSessionStore *session.Store

var UserSessionStore *session.Store

var KafkaWriter *kafka.Writer
