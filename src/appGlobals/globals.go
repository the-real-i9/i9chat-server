package appGlobals

import (
	"cloud.google.com/go/storage"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var GCSClient *storage.Client

var Neo4jDriver neo4j.DriverWithContext
