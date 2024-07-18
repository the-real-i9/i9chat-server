package helpers

import (
	"fmt"
	"testing"
	"time"
)

func TestJwt(t *testing.T) {
	token := JwtSign(struct{ A int }{A: 5}, "some secret", time.Now().Add(1*time.Hour))

	data, err := JwtVerify[struct{ A int }]("Bearer "+token, "some secret")

	b := map[string]any{"auth": data}

	if err != nil {
		t.Errorf("%s", err)
		return
	}

	x := b["auth"].(*struct{ A int })

	fmt.Println(x)
}
