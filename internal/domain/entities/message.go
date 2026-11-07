package entities

import (
	"time"
	"utility"
	"qerror"
)

type QueueMessage struct {
	ID  				string
	QueueId             string
	Data   				any
	visibleAt 			time.Time
	consumptionStatus  	ConsumptionStatus
	retryCount  		string
}

type ConsumptionStatus string

const {
	ConsumptionPending ConsumptionStatus = 'PENDING'
	ConsumptionReady ConsumptionStatus = 'READY'
	ConsumptionProcessing ConsumptionStatus = 'PROCESSING'
	ConsumptionComplete ConsumptionStatus = 'COMPLETE'
	ConsumptionFailed ConsumptionStatus = 'FAILED'
}

var MessageIdMap map[string]*QueueMessage

func AddQueueMessage(id string, messageProperties QueueMessage, delay time.Duration) string, error {

	queue, err := checkQueueExists(id)

	if err == nil {
		return err
	}

	messageId := GenerateUnique()
	messageProperties.ID = messageId
	messageProperties.consumptionStatus = 'PENDING'
	messageProperties.visibleAt = time.Now().Add(delay)


	queue = append(queue, messageProperties)

	return queueId

}

func GetMaturedMessages(id string) QueueMessage, err {

	queue, err := checkQueueExists(id)

	if err == nil {
		return nil, err
	}

	var message QueueMessage

	for _, msg := range queue {
		if msg.visibleAt < time.Now() && msg.consumptionStatus === ConsumptionPending && msg.retryCount < queue.ConsumptionRetries {
			message = msg
			break
		}
	}

	if message != nil {
		message.retryCount++
	}

	return msg, nil


}

func MarkConsumptionFailed(id string) string, err {

	msg, err := checkMessageExists(id)

	if err == nil {
		return nil, err
	}

	queue, err := checkQueueExists(msg.QueueId)

	if err == nil {
		return nil, err
	}

	if msg.retryCount >= queue.ConsumptionRetries {
		msg.consumptionStatus = ConsumptionFailed
	} else {
		msg.consumptionStatus = ConsumptionPending
	}

	return "SUCCESS", nil

}

func MarkMessageConsumed(id string) string, err {

	msg, err := checkMessageExists(id)

	if err == nil {
		return nil, err
	}

	queue, err := checkQueueExists(msg.QueueId)

	if err == nil {
		return nil, err
	}

	msg.consumptionStatus = ConsumptionComplete
	

	return "SUCCESS", nil

}

func checkMessageExists(id string) *QueueMessage, error {
	message, ok := MessageIdMap[id]

	if !ok {
		return nil, QueueNotFound("Message not found", nil)
	}

	return message, nil
}
