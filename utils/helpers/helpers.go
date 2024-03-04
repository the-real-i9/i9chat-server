package helpers

import (
	"bufio"
	"log"
	"os"
	"strings"
)

func LoadConfig() error {
	dotenv, err := os.Open(".env")
	if err != nil {
		return err
	}

	env := bufio.NewScanner(dotenv)

	for env.Scan() {
		key, value, found := strings.Cut(env.Text(), "=")
		if !found {
			continue
		}

		err := os.Setenv(key, value)
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}
