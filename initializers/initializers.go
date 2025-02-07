package initializers

import (
	"context"
	"i9chat/appGlobals"
	"log"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gofiber/fiber/v2/middleware/session"
	neo4jstore "github.com/gofiber/storage/neo4j"
	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
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

func initNeo4jDriver() error {
	driver, err := neo4j.NewDriverWithContext(os.Getenv("NEO4J_URL"), neo4j.BasicAuth(os.Getenv("NEO4J_USER"), os.Getenv("NEO4J_PASSWORD"), ""))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sess := driver.NewSession(ctx, neo4j.SessionConfig{})

	_, err2 := sess.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx, `CREATE CONSTRAINT unique_username IF NOT EXISTS FOR (u:User) REQUIRE u.username IS UNIQUE`, nil)
		if err != nil {
			return nil, err
		}

		_, err2 := tx.Run(ctx, `CREATE CONSTRAINT unique_email IF NOT EXISTS FOR (u:User) REQUIRE u.email IS UNIQUE`, nil)
		if err2 != nil {
			return nil, err
		}

		return nil, nil
	})

	if err2 != nil {
		return err2
	}

	if err := sess.Close(ctx); err != nil {
		return err
	}

	appGlobals.Neo4jDriver = driver

	return nil
}

func initKafkaWriter() error {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(os.Getenv("KAFKA_ADDRESS")),
		AllowAutoTopicCreation: true,
	}

	appGlobals.KafkaWriter = w

	return nil
}

func initSessionStore() {
	getStorage := func(nodeName string) *neo4jstore.Storage {
		return neo4jstore.New(neo4jstore.Config{
			DB:   appGlobals.Neo4jDriver,
			Node: nodeName,
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

	if err := initNeo4jDriver(); err != nil {
		return err
	}

	initSessionStore()

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

	if err := appGlobals.Neo4jDriver.Close(context.TODO()); err != nil {
		log.Println("error closing neo4j driver", err)
	}
}
