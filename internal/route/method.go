package route

import "github.com/gofiber/fiber/v2"

const (
	GET_METHOD    = "GET"
	POST_METHOD   = "POST"
	PUT_METHOD    = "PUT"
	DELETE_METHOD = "DELETE"
)

func GET(app fiber.Router, relativePath string, f fiber.Handler) {
	route(app, GET_METHOD, relativePath, f)
}

func POST(app fiber.Router, relativePath string, f fiber.Handler) {
	route(app, POST_METHOD, relativePath, f)
}

func PUT(app fiber.Router, relativePath string, f fiber.Handler) {
	route(app, PUT_METHOD, relativePath, f)
}

func DELETE(app fiber.Router, relativePath string, f fiber.Handler) {
	route(app, DELETE_METHOD, relativePath, f)
}

func route(app fiber.Router, method string, relativePath string, f fiber.Handler) {
	switch method {
	case POST_METHOD:
		app.Post(relativePath, f)
	case GET_METHOD:
		app.Get(relativePath, f)
	case PUT_METHOD:
		app.Put(relativePath, f)
	case DELETE_METHOD:
		app.Delete(relativePath, f)
	}
}
