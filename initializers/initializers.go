package initializers

import (
	"context"
	"i9chat/appGlobals"
	"os"

	"cloud.google.com/go/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
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

func InitApp() error {

	if os.Getenv("GO_ENV") != "production" {
		if err := godotenv.Load(".env"); err != nil {
			return err
		}
	}

	if err := initDBPool(); err != nil {
		return err
	}

	if err := initGCSClient(); err != nil {
		return err
	}

	return nil
}
