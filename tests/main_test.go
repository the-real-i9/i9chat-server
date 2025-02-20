// User-story-based testing for server applications
package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"testing"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

const HOST_URL string = "http://localhost:8000"
const WSHOST_URL string = "ws://localhost:8000"

// var dbDriver neo4j.DriverWithContext

func TestMain(m *testing.M) {
	dbDriver, err := neo4j.NewDriverWithContext(os.Getenv("NEO4J_URL"), neo4j.BasicAuth(os.Getenv("NEO4J_USER"), os.Getenv("NEO4J_PASSWORD"), ""))
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()

	defer dbDriver.Close(ctx)

	neo4j.ExecuteQuery(ctx, dbDriver, `MATCH (n) DETACH DELETE n`, nil, neo4j.EagerResultTransformer)

	cleanUpDB()

	c := m.Run()

	os.Exit(c)
}

func cleanUpDB() {

}

func reqBody(data map[string]any) (io.Reader, error) {
	dataBt, err := json.Marshal(data)

	return bytes.NewReader(dataBt), err
}

func resBody(body io.ReadCloser) ([]byte, error) {
	defer body.Close()

	return io.ReadAll(body)
}
