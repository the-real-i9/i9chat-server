package initializers

import (
	"context"
	"i9chat/src/appGlobals"
	"log"
	"os"

	"cloud.google.com/go/storage"
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

	defer func() {
		if err := sess.Close(ctx); err != nil {
			log.Println("initializers.go: initNeo4jDriver: sess.Close:", err)
		}
	}()

	_, err2 := sess.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		var err error

		_, err = tx.Run(ctx, `CREATE CONSTRAINT unique_username IF NOT EXISTS FOR (u:User) REQUIRE u.username IS UNIQUE`, nil)
		if err != nil {
			return nil, err
		}

		_, err = tx.Run(ctx, `CREATE CONSTRAINT unique_email IF NOT EXISTS FOR (u:User) REQUIRE u.email IS UNIQUE`, nil)
		if err != nil {
			return nil, err
		}

		_, err = tx.Run(ctx, `CREATE CONSTRAINT unique_phone IF NOT EXISTS FOR (u:User) REQUIRE u.phone IS UNIQUE`, nil)
		if err != nil {
			return nil, err
		}

		_, err = tx.Run(ctx, `CREATE CONSTRAINT unique_dm_chat IF NOT EXISTS FOR (dmc:DMChat) REQUIRE (dmc.owner_username, dmc.partner_username) IS UNIQUE`, nil)
		if err != nil {
			return nil, err
		}

		_, err = tx.Run(ctx, `CREATE CONSTRAINT unique_dm_msg IF NOT EXISTS FOR (dmm:DMMessage) REQUIRE dmm.id IS UNIQUE`, nil)
		if err != nil {
			return nil, err
		}

		_, err = tx.Run(ctx, `CREATE CONSTRAINT unique_group IF NOT EXISTS FOR (g:Group) REQUIRE g.id IS UNIQUE`, nil)
		if err != nil {
			return nil, err
		}

		_, err = tx.Run(ctx, `CREATE CONSTRAINT unique_group_chat IF NOT EXISTS FOR (gc:GroupChat) REQUIRE (gc.owner_username, gc.group_id) IS UNIQUE`, nil)
		if err != nil {
			return nil, err
		}

		_, err = tx.Run(ctx, `CREATE CONSTRAINT unique_group_msg IF NOT EXISTS FOR (gm:GroupMessage) REQUIRE gm.id IS UNIQUE`, nil)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	if err2 != nil {
		return err2
	}

	appGlobals.Neo4jDriver = driver

	return nil
}

func initKafkaWriter() error {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(os.Getenv("KAFKA_BROKER_ADDRESS")),
		AllowAutoTopicCreation: true,
	}

	appGlobals.KafkaWriter = w

	return nil
}

func InitApp() error {

	if os.Getenv("GO_ENV") == "" {
		if err := godotenv.Load(".env"); err != nil {
			return err
		}
	}

	if err := initNeo4jDriver(); err != nil {
		return err
	}

	if err := initGCSClient(); err != nil {
		return err
	}

	initKafkaWriter()

	return nil
}

func CleanUp() {
	if err := appGlobals.KafkaWriter.Close(); err != nil {
		log.Println("failed to close kafka writer:", err)
	}

	if err := appGlobals.Neo4jDriver.Close(context.TODO()); err != nil {
		log.Println("error closing neo4j driver", err)
	}
}
