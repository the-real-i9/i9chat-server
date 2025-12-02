// User-story-based testing for server applications
package tests

import (
	"bytes"
	"i9chat/src/initializers"
	"i9chat/src/routes/appRoutes"
	"i9chat/src/routes/authRoutes"
	"io"
	"log"
	"os"
	"testing"
	"time"

	"github.com/goccy/go-json"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
	"github.com/gofiber/fiber/v2/middleware/helmet"
)

const HOST_URL string = "http://localhost:8000"
const WSHOST_URL string = "ws://localhost:8000"
const wsPath = WSHOST_URL + "/api/app/ws"

const signupPath = HOST_URL + "/api/auth/signup"
const signinPath = HOST_URL + "/api/auth/signin"

const userPath = HOST_URL + "/api/app/user"
const signoutPath = userPath + "/signout"

const directChatPath = HOST_URL + "/api/app/dm_chat"
const groupChatPath = HOST_URL + "/api/app/group_chat"

type UserGeolocation struct {
	X float64
	Y float64
}

type UserT struct {
	Email          string
	Username       string
	Password       string
	Geolocation    UserGeolocation
	SessionCookie  string
	WSConn         *websocket.Conn
	ServerEventMsg chan map[string]any
}

func TestMain(m *testing.M) {
	if err := initializers.InitApp(); err != nil {
		log.Fatal(err)
	}

	defer initializers.CleanUp()

	app := fiber.New()

	app.Use(helmet.New())
	app.Use(cors.New())

	app.Use(encryptcookie.New(encryptcookie.Config{
		Key: os.Getenv("COOKIE_SECRET"),
	}))

	app.Route("/api/auth", authRoutes.Route)
	app.Route("/api/app", appRoutes.Route)

	var PORT string

	if os.Getenv("GO_ENV") != "production" {
		PORT = "8000"
	} else {
		PORT = os.Getenv("PORT")
	}

	go func() {
		app.Listen("0.0.0.0:" + PORT)
	}()

	waitReady := time.NewTimer(2 * time.Second)
	<-waitReady.C

	c := m.Run()

	waitFinish := time.NewTimer(2 * time.Second)
	<-waitFinish.C

	app.Shutdown()

	os.Exit(c)
}

func makeReqBody(data map[string]any) (io.Reader, error) {
	dataBt, err := json.Marshal(data)

	return bytes.NewReader(dataBt), err
}

func succResBody[T any](body io.ReadCloser) (T, error) {
	var d T

	defer body.Close()

	bt, err := io.ReadAll(body)
	if err != nil {
		return d, err
	}

	if err := json.Unmarshal(bt, &d); err != nil {
		return d, err
	}

	return d, nil
}

func errResBody(body io.ReadCloser) (string, error) {
	defer body.Close()

	bt, err := io.ReadAll(body)
	if err != nil {
		return "", err
	}

	return string(bt), nil
}
