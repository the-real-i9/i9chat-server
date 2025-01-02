package initializers

import (
	"context"
	"i9chat/appGlobals"
	"log"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/postgres/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
	"google.golang.org/api/option"
)

func initGCSClient() error {
	stClient, err := storage.NewClient(context.Background(), option.WithAPIKey(os.Getenv("GCS_API_KEY")))
	if err != nil {
		return err
	}

	appGlobals.GCSClient = stClient

	return nil
}

func initDBPool() error {
	pool, err := pgxpool.New(context.Background(), os.Getenv("PGDATABASE_URL"))
	if err != nil {
		return err
	}

	appGlobals.DBPool = pool

	return nil
}

func initKafkaWriter() error {
	w := &kafka.Writer{
		Addr:                   kafka.TCP("localhost:9092"),
		AllowAutoTopicCreation: true,
	}

	appGlobals.KafkaWriter = w

	return nil
}

func configSessionStore() {
	getStorage := func(tableName string) *postgres.Storage {
		return postgres.New(postgres.Config{
			DB:    appGlobals.DBPool,
			Table: tableName,
		})
	}

	appGlobals.SignupSessionStore = session.New(session.Config{
		Storage:        getStorage("ongoing_signup"),
		CookiePath:     "/api/auth/signup",
		CookieDomain:   os.Getenv("APP_DOMAIN"),
		CookieHTTPOnly: true,
	})

	appGlobals.UserSessionStore = session.New(session.Config{
		Storage:        getStorage("user_session"),
		CookiePath:     "/api/app",
		CookieDomain:   os.Getenv("APP_DOMAIN"),
		CookieHTTPOnly: true,
		Expiration:     (10 * 24) * time.Hour,
	})
}

func InitApp() error {

	if os.Getenv("GO_ENV") != "production" {
		if err := godotenv.Load(".env"); err != nil {
			return err
		}
	}

	if err := initDBPool(); err != nil {
		return err
	}

	configSessionStore()

	if err := initGCSClient(); err != nil {
		return err
	}

	initKafkaWriter()

	return nil
}

func CleanUp() {
	if err := appGlobals.KafkaWriter.Close(); err != nil {
		log.Println("failed to close writer:", err)
	}

	appGlobals.DBPool.Close()
}
