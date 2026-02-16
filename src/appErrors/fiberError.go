package appErrors

type HTTPError struct {
	Code    int    `msgpack:"code"`
	Message string `msgpack:"message"`
}
