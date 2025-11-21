package appGlobals

import (
	"cloud.google.com/go/storage"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/redis/go-redis/v9"
)

var GCSClient *storage.Client

var Neo4jDriver neo4j.DriverWithContext

var RedisClient *redis.Client
