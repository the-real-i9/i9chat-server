package helpers

import (
	"fmt"
	"testing"
	"time"
)

func TestJwt(t *testing.T) {
	token := JwtSign(map[string]any{"a": 5}, "some secret", time.Now().Add(1*time.Hour))

	data, err := JwtVerify(token, "some secret")

	if err != nil {
		t.Errorf("%s", err)
		return
	}

	fmt.Println(data)
}
