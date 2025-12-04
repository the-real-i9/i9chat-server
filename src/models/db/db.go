package db

import (
	"context"
	"i9chat/src/appGlobals"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func Query(ctx context.Context, cypher string, params map[string]any) (*neo4j.EagerResult, error) {
	return neo4j.ExecuteQuery(ctx, appGlobals.Neo4jDriver, cypher, params, neo4j.EagerResultTransformer)
}
