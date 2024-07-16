package tests

import (
	"i9chat/utils/helpers"
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if err := helpers.LoadEnv(); err != nil {
		log.Fatal(err)
	}

	if err := helpers.InitDBPool(); err != nil {
		log.Fatal(err)
	}

	if err := helpers.InitGCSClient(); err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	os.Exit(code)
}
