package http

import (
	"net/http"

	"letsgo/internal/transport/http/handlers"
	"letsgo/internal/transport/http/routes"
)

func NewRouter() *http.ServeMux {

	mux := http.NewServeMux()

	routes.RegisterQueueRoutes(mux, handlers.NewQueueHandler())
	routes.RegisterMessageRoutes(mux, handlers.NewMessageHandler())

	return mux
}
