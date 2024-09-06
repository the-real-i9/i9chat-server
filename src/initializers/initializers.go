package initializers

import (
	"context"
	"i9chat/src/appGlobals"
	"os"

	"cloud.google.com/go/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func initGCSClient() error {
	stClient, err := storage.NewClient(context.Background())
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

	if err := godotenv.Load(".env"); err != nil {
		return err
	}

	if err := initDBPool(); err != nil {
		return err
	}

	if err := initGCSClient(); err != nil {
		return err
	}

	return nil
}
