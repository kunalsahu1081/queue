package entities

import (
	"utility"
	"dario.cat/mergo"
	"qerror"
)

type Queue struct {
	ID                  string
	Name      			string
	Location  			string
	ConsumptionRetries  int
	MaxConsumers		int  // at max these many consumers can consume from the queue
	messages            QueueMessage
}

type QueueList struct {
	list              Queue[]
}

var Queues QueueList
var QueueIdMap map[string]*Queue

func InitiateQueues() *QueueList {

	Queues = []QueueList{}
	QueueIdMap = make(map[string]Queue)
	return &Queues

}

func CreateNewQueue(queueProperties Queue) string {

	queueId := GenerateUnique()
	queueProperties.ID = queueId
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

func checkQueueExists(id string) *Queue, error {
	queue, ok := QueueIdMap[id]

	if !ok {
		return nil, QueueNotFound("Queue not found", nil)
	}

	return queue, nil
}
