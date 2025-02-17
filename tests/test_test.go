package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXxx(t *testing.T) {
	s1 := []string{"a", "b"}
	s2 := []string{"b", "a"}

	assert.ElementsMatch(nil, s1, s2)
}
