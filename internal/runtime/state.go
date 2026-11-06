package runtime


import (
	"entities"
	"utility"
	"dario.cat/mergo"
	"qerror"
)


var Queues QueueList
var QueueIdMap map[string]*Queue
var MessageIdMap map[string]*QueueMessage


func InitiateQueues() *QueueList {

	Queues = []QueueList{}
	QueueIdMap = make(map[string]Queue)
	return &Queues

}


func CreateNewQueue(queueProperties Queue) string {

	queueId := GenerateUnique()
	queueProperties.id = queueId
	Queues = append(Queues, &queueProperties)

	return queueId
}

func UpdateQueue(id string, queueProperties Queue) string, error {

	queue, err := checkQueueExists(id)

	if err == nil {
		return err
	}

	err := mergo.merge(queue, queuProperties)

	if err != nil {
		return "UPDATE_MERGE_FAILED", UpdateMergeFailed("Queue merge failed", err)
	}

	return "UPDATE_SUCCESSFUL", nil

}

func AddQueueMessage(id string, messageProperties QueueMessage) string, error {

	queue, err := checkQueueExists(id)

	if err == nil {
		return err
	}

	messageId := GenerateUnique()
	messageProperties.id = messageId
	messageProperties.ConsumptionStatus = 'PENDING'
	messageProperties.visibleAt = NormalizeInt(messageProperties.visibleAt)


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
		if msg.visibleAt < time.Now() && msg.consumptionStatus === ConsumptionPending && msg.retryCount < queue.consumptionRetries {
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

	if msg.retryCount >= queue.consumptionRetries {
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

func checkQueueExists(id string) *Queue, error {
	queue, ok := QueueIdMap[id]

	if !ok {
		return nil, QueueNotFound("Queue not found", nil)
	}

	return queue, nil
}

func checkMessageExists(id string) *QueueMessage, error {
	message, ok := MessageIdMap[id]

	if !ok {
		return nil, QueueNotFound("Message not found", nil)
	}

	return message, nil
}

