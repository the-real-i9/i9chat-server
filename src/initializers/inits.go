package initializers

import (
	"context"
	"i9chat/src/appGlobals"
	"i9chat/src/backgroundWorkers"
	"i9chat/src/helpers"
	"os"

	"cloud.google.com/go/storage"
	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
)

func initGCSClient() error {
	stClient, err := storage.NewClient(context.Background())
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

	ctx := context.Background()

	if os.Getenv("GO_ENV") == "test" {
		_, err := neo4j.ExecuteQuery(ctx, driver, `MATCH (n) DETACH DELETE n`, nil, neo4j.EagerResultTransformer)
		if err != nil {
			return err
		}
	}

	sess := driver.NewSession(ctx, neo4j.SessionConfig{})

	defer func() {
		if err := sess.Close(ctx); err != nil {
			helpers.LogError(err)
		}
	}()

	_, err2 := sess.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		var err error

		_, err = tx.Run(ctx, `/* cypher */ CREATE CONSTRAINT unique_username IF NOT EXISTS FOR (u:User) REQUIRE u.username IS UNIQUE`, nil)
		if err != nil {
			return nil, err
		}

		_, err = tx.Run(ctx, `/* cypher */ CREATE CONSTRAINT unique_email IF NOT EXISTS FOR (u:User) REQUIRE u.email IS UNIQUE`, nil)
		if err != nil {
			return nil, err
		}

		_, err = tx.Run(ctx, `/* cypher */ CREATE CONSTRAINT unique_direct_chat IF NOT EXISTS FOR (dc:DirectChat) REQUIRE (dc.owner_username, dc.partner_username) IS UNIQUE`, nil)
		if err != nil {
			return nil, err
		}

		_, err = tx.Run(ctx, `/* cypher */ CREATE CONSTRAINT unique_direct_msg IF NOT EXISTS FOR (dm:DirectMessage) REQUIRE dm.id IS UNIQUE`, nil)
		if err != nil {
			return nil, err
		}

		_, err = tx.Run(ctx, `/* cypher */ CREATE CONSTRAINT unique_direct_msg_rxn IF NOT EXISTS FOR (dmrxn:DirectMessageReaction) REQUIRE (dmrxn.reactor_username, dmrxn.message_id) IS UNIQUE`, nil)
		if err != nil {
			return nil, err
		}

		_, err = tx.Run(ctx, `/* cypher */ CREATE CONSTRAINT unique_group IF NOT EXISTS FOR (g:Group) REQUIRE g.id IS UNIQUE`, nil)
		if err != nil {
			return nil, err
		}

		_, err = tx.Run(ctx, `/* cypher */ CREATE CONSTRAINT unique_group_chat IF NOT EXISTS FOR (gc:GroupChat) REQUIRE (gc.owner_username, gc.group_id) IS UNIQUE`, nil)
		if err != nil {
			return nil, err
		}

		_, err = tx.Run(ctx, `/* cypher */ CREATE CONSTRAINT unique_group_msg IF NOT EXISTS FOR (gm:GroupMessage) REQUIRE gm.id IS UNIQUE`, nil)
		if err != nil {
			return nil, err
		}

		_, err = tx.Run(ctx, `/* cypher */ CREATE CONSTRAINT unique_group_msg_rxn IF NOT EXISTS FOR (gmrxn:GroupMessageReaction) REQUIRE (gmrxn.reactor_username, gmrxn.message_id) IS UNIQUE`, nil)
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

func initRedisClient() error {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASS"),
		DB:       0,

		// Explicitly disable maintenance notifications
		// This prevents the client from sending CLIENT MAINT_NOTIFICATIONS ON
		MaintNotificationsConfig: &maintnotifications.Config{
			Mode: maintnotifications.ModeDisabled,
		},
	})

	if os.Getenv("GO_ENV") == "test" {
		err := client.FlushDB(context.Background()).Err()
		if err != nil {
			return err
		}
	}

	appGlobals.RedisClient = client

	backgroundWorkers.Start(client)

	return nil
}

func InitApp() error {

	if os.Getenv("GO_ENV") == "" {
		if err := godotenv.Load(".env"); err != nil {
			return err
		}
	}

	if os.Getenv("GO_ENV") == "test" {
		if err := godotenv.Load(".env.test"); err != nil {
			return err
		}
	}

	if err := initNeo4jDriver(); err != nil {
		return err
	}

	if err := initGCSClient(); err != nil {
		return err
	}

	if err := initRedisClient(); err != nil {
		return err
	}

	return nil
}

func CleanUp() {
	if err := appGlobals.Neo4jDriver.Close(context.Background()); err != nil {
		helpers.LogError(err)
	}

	if err := appGlobals.RedisClient.Close(); err != nil {
		helpers.LogError(err)
	}
}
