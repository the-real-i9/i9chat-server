package db

import (
	"context"
	"i9chat/appGlobals"
	"log"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func Query(ctx context.Context, cypher string, params map[string]any) (*neo4j.EagerResult, error) {
	return neo4j.ExecuteQuery(ctx, appGlobals.Neo4jDriver, cypher, params, neo4j.EagerResultTransformer)
}

func MultiQuery(ctx context.Context, work func(tx neo4j.ManagedTransaction) (any, error)) (any, error) {
	sess := appGlobals.Neo4jDriver.NewSession(ctx, neo4j.SessionConfig{})

	defer func() {
		if err := sess.Close(ctx); err != nil {
			log.Println("db.go: error closing session:", err)
		}
	}()

	return sess.ExecuteWrite(ctx, work)
}
