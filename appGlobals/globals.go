package appGlobals

import (
	"cloud.google.com/go/storage"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/segmentio/kafka-go"
)

var GCSClient *storage.Client

var Neo4jDriver neo4j.DriverWithContext

var SessionStore *session.Store

var KafkaWriter *kafka.Writer
