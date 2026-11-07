package routes

import (
	"net/http"

	"letsgo/internal/transport/http/handlers"
)

func RegisterMessageRoutes(
	mux *http.ServeMux,
	handler *handlers.MessageHandler,
) {
	mux.HandleFunc("POST /queues/{id}/messages", handler.AddMessageHandler)
	mux.HandleFunc("GET /queues/{id}/messages", handler.GetMessageHandler)
	mux.HandleFunc("GET /queues/{id}/messages/{messageId}", handler.GetMessageByIdHandler)
	mux.HandleFunc("PATCH /queues/{id}/messages/{messageId}", handler.UpdateMessageHandler)
	mux.HandleFunc("DELETE /queues/{id}/messages/{messageId}", handler.DeleteMessageHandler)
}
