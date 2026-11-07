package entities

import (
	"time"

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
}

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

func AddQueueMessage(id string, messageProperties QueueMessage, delay time.Duration) (string, error) {

	queue, err := checkQueueExists(id)

	if err != nil {
		return "", err
	}

	messageId := utility.GenerateUnique()
	messageProperties.ID = messageId
	messageProperties.QueueId = id
	messageProperties.consumptionStatus = ConsumptionPending
	messageProperties.visibleAt = time.Now().Add(delay)

	queue.messages = append(queue.messages, &messageProperties)
	MessageIdMap[messageId] = &messageProperties

	return messageId, nil

}

func GetQueueMessages(queueId string) ([]*QueueMessage, error) {

	queue, err := checkQueueExists(queueId)

	if err != nil {
		return nil, err
	}

	return queue.messages, nil

}

func GetQueueMessage(queueId string, messageId string) (*QueueMessage, error) {

	return checkMessageInQueue(queueId, messageId)

}

func UpdateQueueMessage(queueId string, messageId string, data any) (string, error) {

	msg, err := checkMessageInQueue(queueId, messageId)

	if err != nil {
		return "", err
	}

	msg.Data = data

	return "UPDATE_SUCCESSFUL", nil

}

func DeleteQueueMessage(queueId string, messageId string) (string, error) {

	msg, err := checkMessageInQueue(queueId, messageId)

	if err != nil {
		return "", err
	}

	queue, err := checkQueueExists(msg.QueueId)

	if err != nil {
		return "", err
	}

	for i, m := range queue.messages {
		if m.ID == messageId {
			queue.messages = append(queue.messages[:i], queue.messages[i+1:]...)
			break
		}
	}

	delete(MessageIdMap, messageId)

	return "DELETE_SUCCESSFUL", nil

}

// ---------- consumption (yours to finish) ----------

func GetMaturedMessages(id string) (*QueueMessage, error) {

	queue, err := checkQueueExists(id)

	if err != nil {
		return nil, err
	}

	var message *QueueMessage

	for _, msg := range queue.messages {
		if !time.Now().Before(msg.visibleAt) && msg.consumptionStatus == ConsumptionPending && msg.retryCount < queue.ConsumptionRetries {
			message = msg
			break
		}
	}

	if message != nil {
		message.retryCount++
	}

	return message, nil

}

func MarkConsumptionFailed(id string) (string, error) {

	msg, err := checkMessageExists(id)

	if err != nil {
		return "", err
	}

	queue, err := checkQueueExists(msg.QueueId)

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

	_, err = checkQueueExists(msg.QueueId)

	if err != nil {
		return "", err
	}

	msg.consumptionStatus = ConsumptionComplete

	return "SUCCESS", nil

}

func checkMessageExists(id string) (*QueueMessage, error) {
	message, ok := MessageIdMap[id]

	if !ok {
		return nil, qerrors.MessageNotFound("Message not found", nil)
	}

	return message, nil
}

// a message id is only valid under the queue it belongs to
func checkMessageInQueue(queueId string, messageId string) (*QueueMessage, error) {

	if _, err := checkQueueExists(queueId); err != nil {
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
