package appTypes

import (
	"github.com/goccy/go-json"
)

type ClientUser struct {
	Username string `msgpack:"username"`
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
	Event string `msgpack:"event"`
	Data  any    `msgpack:"data"`
}

type UserGeolocation struct {
	X float64 `msgpack:"x"`
	Y float64 `msgpack:"y"`
}
