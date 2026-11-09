package handlers

import (
	"net/http"

	service "letsgo/internal/services"
	utility "letsgo/internal/utilities"
)

type MessageHandler struct {
}

func NewMessageHandler() *MessageHandler {
	return &MessageHandler{}
}

// add a message to a queue
func (h *MessageHandler) AddMessageHandler(w http.ResponseWriter, r *http.Request) {

	data, err := utility.DecodeRequest[service.AddMessageRequest](r, "id")

	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	messageId, err := service.AddMessageService(data.Params["id"], data.Body)

	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"messageId": messageId,
	})
}

// list all messages in a queue
func (h *MessageHandler) GetMessageHandler(w http.ResponseWriter, r *http.Request) {

	messages, err := service.GetMessagesService(r.PathValue("id"))

	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, messages)
}

// get one message
func (h *MessageHandler) GetMessageByIdHandler(w http.ResponseWriter, r *http.Request) {

	message, err := service.GetMessageService(r.PathValue("id"), r.PathValue("messageId"))

	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, message)
}

// replace a message's data
func (h *MessageHandler) UpdateMessageHandler(w http.ResponseWriter, r *http.Request) {

	data, err := utility.DecodeRequest[service.UpdateMessageRequest](r, "id", "messageId")

	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	result, err := service.UpdateMessageService(data.Params["id"], data.Params["messageId"], data.Body)

	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": result,
	})
}

// delete a message
func (h *MessageHandler) DeleteMessageHandler(w http.ResponseWriter, r *http.Request) {

	result, err := service.DeleteMessageService(r.PathValue("id"), r.PathValue("messageId"))

	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": result,
	})
}

// take the next matured message; 204 when nothing is ready
func (h *MessageHandler) ReceiveMessageHandler(w http.ResponseWriter, r *http.Request) {

	message, err := service.GetMaturedMessageService(r.PathValue("id"))

	if err != nil {
		writeError(w, err)
		return
	}

	if message == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	writeJSON(w, http.StatusOK, message)
}

// acknowledge successful consumption
func (h *MessageHandler) AckMessageHandler(w http.ResponseWriter, r *http.Request) {

	result, err := service.MarkMessageConsumedService(r.PathValue("messageId"))

	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": result,
	})
}

// report failed consumption: message is retried, or FAILED once retries are exhausted
func (h *MessageHandler) FailMessageHandler(w http.ResponseWriter, r *http.Request) {

	result, err := service.MarkConsumptionFailedService(r.PathValue("messageId"))

	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": result,
	})
}
