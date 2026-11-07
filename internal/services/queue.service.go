package service

import (
	"letsgo/internal/domain/entities"
)

type CreateQueueRequest struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Location           string `json:"location"`
	ConsumptionRetries int    `json:"consumptionRetries"`
	MaxConsumers       int    `json:"maxConsumers"`
}

func CreateQueueService(qp CreateQueueRequest) (string, error) {

	queue := entities.Queue{
		Name:               qp.Name,
		Location:           qp.Location,
		ConsumptionRetries: qp.ConsumptionRetries,
		MaxConsumers:       qp.MaxConsumers,
	}

	queueId := entities.CreateNewQueue(queue)

	return queueId, nil

}

func UpdateQueueService(id string, qp CreateQueueRequest) (string, error) {

	queue := entities.Queue{
		Name:               qp.Name,
		Location:           qp.Location,
		ConsumptionRetries: qp.ConsumptionRetries,
		MaxConsumers:       qp.MaxConsumers,
	}

	result, err := entities.UpdateQueue(id, queue)

	return result, err

}

func GetQueuesService() ([]*entities.Queue, error) {

	return entities.GetQueues(), nil

}

func GetQueueService(id string) (*entities.Queue, error) {

	return entities.GetQueue(id)

}

func DeleteQueueService(id string) (string, error) {

	return entities.DeleteQueue(id)

}
