package testdata

var SignupSessionJwt = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImVtYWlsIjoib2d1bnJpbm9sYS5rZWhpbmRlQHlhaG9vLmNvbSIsInNlc3Npb25JZCI6ImY4Yzk1NGNlLTFkOTktNGFhZS1iMjE4LWI5YWUxZDY2NDhkZiJ9LCJleHAiOiIyMDI0LTA1LTEwVDE2OjIzOjM5LjIxNjI2NDQ2NFoifQ.IN183bXQtnEg7E+UbxrpR15CqaFqEHmju6pj33nZfWY="

var I9xAuthJwt = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7InVzZXJJZCI6NiwidXNlcm5hbWUiOiJpOXgifSwiZXhwIjoiMjAyNS0wNS0xMFQxNToxODozMC4xODUwMjkyMzJaIn0.+R7pvrJIDchu/dA7xIwSTTNd46j4hphO++fy6AD4xdQ="
var DollypAuthJwt = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7InVzZXJJZCI6NywidXNlcm5hbWUiOiJkb2xseXAifSwiZXhwIjoiMjAyNS0wNS0xMFQxNTo0NDozMy42MDc2MzkzNzZaIn0.AApxlWhC7Wg28qisI8jCJIAuvixi1vPnETUhfNcmIJA="

type clientData struct {
	Id       int
	Username string
}

var I9xClientData = clientData{Id: 6, Username: "i9x"}
var DollypClientData = clientData{Id: 7, Username: "dollyp"}
