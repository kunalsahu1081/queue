package routes

import (
	"net/http"

	"letsgo/internal/transport/http/handlers"
)

func RegisterQueueRoutes(
	mux *http.ServeMux,
	handler *handlers.QueueHandler,
) {
	mux.HandleFunc("GET /queues", handler.GetQueues)
	mux.HandleFunc("POST /queues", handler.CreateNewQueue)
	mux.HandleFunc("GET /queues/{id}", handler.GetQueue)
	mux.HandleFunc("PATCH /queues/{id}", handler.UpdateQueue)
	mux.HandleFunc("DELETE /queues/{id}", handler.DeleteQueue)
}
