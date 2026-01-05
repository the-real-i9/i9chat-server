package appTypes

import (
	"github.com/goccy/go-json"
)

type ClientUser struct {
	Username string `json:"username"`
}

func (c ClientUser) MarshalBinary() ([]byte, error) {
	return json.Marshal(c)
}

type BinableMap map[string]any

func (c BinableMap) MarshalBinary() ([]byte, error) {
	return json.Marshal(c)
}

type BinableSlice []any

func (c BinableSlice) MarshalBinary() ([]byte, error) {
	return json.Marshal(c)
}

type ServerEventMsg struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

type UserGeolocation struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}
