package entities

import (
	"time"
	"sync"
	qerrors "letsgo/internal/domain/errors"
	utility "letsgo/internal/utilities"
)

type QueueMessage struct {
	ID                string `json:"id"`
	QueueId           string `json:"queueId"`
	Data              any    `json:"data"`
	visibleAt         time.Time
	consumptionStatus ConsumptionStatus
	retryCount        int
	mu                sync.Mutex // guards Data only; status fields are guarded by the queue lock
}

// guards MessageIdMap
var MessageRWlock sync.RWMutex

type ConsumptionStatus string

const (
	ConsumptionPending    ConsumptionStatus = "PENDING"
	ConsumptionReady      ConsumptionStatus = "READY"
	ConsumptionProcessing ConsumptionStatus = "PROCESSING"
	ConsumptionComplete   ConsumptionStatus = "COMPLETE"
	ConsumptionFailed     ConsumptionStatus = "FAILED"
)

var MessageIdMap = map[string]*QueueMessage{}

// ---------- CRUD ----------

func AddQueueMessage(id string, messageProperties *QueueMessage, delay time.Duration) (string, error) {

	queue, err := CheckQueueExists(id)

	if err != nil {
		return "", err
	}

	queue.mu.Lock()
	defer queue.mu.Unlock()

	MessageRWlock.Lock()
	defer MessageRWlock.Unlock()

	messageId := utility.GenerateUnique()
	messageProperties.ID = messageId
	messageProperties.QueueId = id
	messageProperties.consumptionStatus = ConsumptionPending
	messageProperties.visibleAt = time.Now().Add(delay)

	queue.messages = append(queue.messages, messageProperties)
	MessageIdMap[messageId] = messageProperties

	return messageId, nil

}

func GetQueueMessages(queueId string) ([]*QueueMessage, error) {

	queue, err := CheckQueueExists(queueId)

	if err != nil {
		return nil, err
	}

	return queue.messages, nil

}

func GetQueueMessage(queueId string, messageId string) (*QueueMessage, error) {

	msg, err := checkMessageInQueue(queueId, messageId)

	if err != nil {
		return nil, err
	}

	return copyMessage(msg), nil

}

func UpdateQueueMessage(queueId string, messageId string, data any) (string, error) {

	msg, err := checkMessageInQueue(queueId, messageId)

	if err != nil {
		return "", err
	}

	msg.mu.Lock()
	defer msg.mu.Unlock()

	msg.Data = data

	return "UPDATE_SUCCESSFUL", nil

}

func DeleteQueueMessage(queueId string, messageId string) (string, error) {

	queue, err := CheckQueueExists(queueId)

	if err != nil {
		return "", err
	}

	queue.mu.Lock()
	defer queue.mu.Unlock()

	_, err = checkMessageInQueue(queueId, messageId)

	if err != nil {
		return "", err
	}

	for i, m := range queue.messages {
		if m.ID == messageId {
			queue.messages = append(queue.messages[:i], queue.messages[i+1:]...)
			break
		}
	}

	MessageRWlock.Lock()
	defer MessageRWlock.Unlock()

	delete(MessageIdMap, messageId)

	return "DELETE_SUCCESSFUL", nil

}



func GetMaturedMessages(id string) (*QueueMessage, error) {

	
	queue, err := CheckQueueExists(id)

	if err != nil {
		return nil, err
	}

	queue.mu.Lock()
	defer queue.mu.Unlock()

	var message *QueueMessage

	for _, msg := range queue.messages {
		if !time.Now().Before(msg.visibleAt) && msg.consumptionStatus == ConsumptionPending && msg.retryCount < queue.ConsumptionRetries {
			message = msg
			break
		}
	}

	if message != nil {
		message.retryCount++
		message.consumptionStatus = ConsumptionProcessing
		return copyMessage(message), nil
	}

	return nil, nil

}

func MarkConsumptionFailed(id string) (string, error) {

	msg, err := checkMessageExists(id)

	if err != nil {
		return "", err
	}

	queue, err := CheckQueueExists(msg.QueueId)

	if err != nil {
		return "", err
	}
	queue.mu.Lock()
	defer queue.mu.Unlock()

	msg, err = checkMessageExists(id)

	if err != nil {
		return "", err
	}


	if msg.retryCount >= queue.ConsumptionRetries {
		msg.consumptionStatus = ConsumptionFailed
	} else {
		msg.consumptionStatus = ConsumptionPending
	}

	return "SUCCESS", nil

}

func MarkMessageConsumed(id string) (string, error) {

	msg, err := checkMessageExists(id)

	if err != nil {
		return "", err
	}

	queue, err := CheckQueueExists(msg.QueueId)

	if err != nil {
		return "", err
	}
	queue.mu.Lock()
	defer queue.mu.Unlock()

	msg, err = checkMessageExists(id)

	if err != nil {
		return "", err
	}

	msg.consumptionStatus = ConsumptionComplete

	return "SUCCESS", nil

}

// snapshot of the exported fields; Data is read under the message lock
func copyMessage(msg *QueueMessage) *QueueMessage {

	msg.mu.Lock()
	defer msg.mu.Unlock()

	return &QueueMessage{ID: msg.ID, QueueId: msg.QueueId, Data: msg.Data}

}

func checkMessageExists(id string) (*QueueMessage, error) {

	MessageRWlock.RLock()
	defer MessageRWlock.RUnlock()

	message, ok := MessageIdMap[id]

	if !ok {
		return nil, qerrors.MessageNotFound("Message not found", nil)
	}

	return message, nil
}

// a message id is only valid under the queue it belongs to
func checkMessageInQueue(queueId string, messageId string) (*QueueMessage, error) {

	if _, err := CheckQueueExists(queueId); err != nil {
		return nil, err
	}

	msg, err := checkMessageExists(messageId)

	if err != nil {
		return nil, err
	}

	if msg.QueueId != queueId {
		return nil, qerrors.MessageNotFound("Message not found", nil)
	}

	return msg, nil
}
