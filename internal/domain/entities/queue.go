package entities

import (
	"dario.cat/mergo"
	"sync"
	qerrors "letsgo/internal/domain/errors"
	utility "letsgo/internal/utilities"
)

type Queue struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Location           string `json:"location"`
	ConsumptionRetries int    `json:"consumptionRetries"`
	MaxConsumers       int    `json:"maxConsumers"` // at max these many consumers can consume from the queue
	messages           []*QueueMessage
	mu                 sync.Mutex
}

// guards QueueIdMap and Queues
var QueueRWlock sync.RWMutex

type QueueList []*Queue

var Queues = QueueList{}
var QueueIdMap = map[string]*Queue{}

func InitiateQueues() *QueueList {

	Queues = QueueList{}
	QueueIdMap = make(map[string]*Queue)
	MessageIdMap = make(map[string]*QueueMessage)
	return &Queues

}

func GetQueues() []*Queue {

	return Queues

}

func GetQueue(id string) (*Queue, error) {

	return CheckQueueExists(id)

}

func CreateNewQueue(queueProperties *Queue) string {

	QueueRWlock.Lock()
	defer QueueRWlock.Unlock()

	queueId := utility.GenerateUnique()
	queueProperties.ID = queueId
	queueProperties.messages = []*QueueMessage{}

	Queues = append(Queues, queueProperties)
	QueueIdMap[queueId] = queueProperties

	return queueId
}

func UpdateQueue(id string, queueProperties *Queue) (string, error) {

	queue, err := CheckQueueExists(id)

	if err != nil {
		return "", err
	}

	queue.mu.Lock()
	defer queue.mu.Unlock()

	// never let an update change the id; empty fields are skipped by mergo
	queueProperties.ID = ""
	err = mergo.Merge(queue, queueProperties, mergo.WithOverride)

	if err != nil {
		return "UPDATE_MERGE_FAILED", qerrors.UpdateMergeFailed("Queue merge failed", err)
	}

	return "UPDATE_SUCCESSFUL", nil

}

func DeleteQueue(id string) (string, error) {

	queue, err := CheckQueueExists(id)

	if err != nil {
		return "", err
	}

	queue.mu.Lock()
	defer queue.mu.Unlock()

	QueueRWlock.Lock()
	defer QueueRWlock.Unlock()

	MessageRWlock.Lock()
	defer MessageRWlock.Unlock()

	for _, msg := range queue.messages {
		delete(MessageIdMap, msg.ID)
	}

	delete(QueueIdMap, id)

	for i, q := range Queues {
		if q.ID == id {
			Queues = append(Queues[:i], Queues[i+1:]...)
			break
		}
	}

	return "DELETE_SUCCESSFUL", nil

}

func CheckQueueExists(id string) (*Queue, error) {

	QueueRWlock.RLock()
	defer QueueRWlock.RUnlock()

	queue, ok := QueueIdMap[id]

	if !ok {
		return nil, qerrors.QueueNotFound("Queue not found", nil)
	}

	return queue, nil
}
