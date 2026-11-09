package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	qerrors "letsgo/internal/domain/errors"
	service "letsgo/internal/services"
	utility "letsgo/internal/utilities"
)

type QueueHandler struct {
	
}

func NewQueueHandler() *QueueHandler {
	return &QueueHandler{
		
	}
}

// list all queues
func (h *QueueHandler) GetQueues(w http.ResponseWriter, r *http.Request) {

	queues, err := service.GetQueuesService()

	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, queues)
}

// get one queue
func (h *QueueHandler) GetQueue(w http.ResponseWriter, r *http.Request) {

	queue, err := service.GetQueueService(r.PathValue("id"))

	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, queue)
}

//  create a new queue
func (h *QueueHandler) CreateNewQueue(w http.ResponseWriter, r *http.Request) {

	data, err := utility.DecodeRequest[service.CreateQueueRequest](r)

	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	queueId, err := service.CreateQueueService(data.Body)

	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"queueId": queueId,
	})
}

// update a queue
func (h *QueueHandler) UpdateQueue(w http.ResponseWriter, r *http.Request) {

	data, err := utility.DecodeRequest[service.CreateQueueRequest](r, "id")

	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	result, err := service.UpdateQueueService(data.Params["id"], data.Body)

	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": result,
	})
}

// delete a queue and all its messages
func (h *QueueHandler) DeleteQueue(w http.ResponseWriter, r *http.Request) {

	result, err := service.DeleteQueueService(r.PathValue("id"))

	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": result,
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return
	}
}

// maps domain errors to http status codes
func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError

	var appErr *qerrors.AppError
	if errors.As(err, &appErr) && appErr.Code == qerrors.CodeNotFound {
		status = http.StatusNotFound
	}

	writeJSON(w, status, map[string]string{"error": err.Error()})
}
