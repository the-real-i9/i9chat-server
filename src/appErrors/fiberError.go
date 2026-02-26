package appErrors

type HTTPError struct {
	Code    int    `msgpack:"code" json:"code"`
	Message string `msgpack:"message" json:"message"`
}
