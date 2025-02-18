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

var dbDriver neo4j.DriverWithContext

func TestMain(m *testing.M) {
	driver, err := neo4j.NewDriverWithContext(os.Getenv("NEO4J_URL"), neo4j.BasicAuth(os.Getenv("NEO4J_USER"), os.Getenv("NEO4J_PASSWORD"), ""))
	if err != nil {
		log.Fatalln(err)
	}

	dbDriver = driver

	ctx := context.Background()

	defer dbDriver.Close(ctx)

	c := m.Run()

	cleanUpDB()

	os.Exit(c)
}

func cleanUpDB() {
	neo4j.ExecuteQuery(context.Background(), dbDriver, `MATCH (n) DETACH DELETE n`, nil, neo4j.EagerResultTransformer)
}

func reqBody(data map[string]any) (io.Reader, error) {
	dataBt, err := json.Marshal(data)

	return bytes.NewReader(dataBt), err
}

func resBody(body io.ReadCloser) (string, error) {
	defer body.Close()

	bt, err := io.ReadAll(body)
	return string(bt), err
}
