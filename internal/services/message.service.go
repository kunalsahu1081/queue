package service

import (
	"time"

	"letsgo/internal/domain/entities"
)

type AddMessageRequest struct {
	Data         any `json:"data"`
	DelaySeconds int `json:"delaySeconds"`
}

type UpdateMessageRequest struct {
	Data any `json:"data"`
}

func AddMessageService(queueId string, mp AddMessageRequest) (string, error) {

	message := entities.QueueMessage{
		Data: mp.Data,
	}

	delay := time.Duration(mp.DelaySeconds) * time.Second

	return entities.AddQueueMessage(queueId, message, delay)

}

func GetMessagesService(queueId string) ([]*entities.QueueMessage, error) {

	return entities.GetQueueMessages(queueId)

}

func GetMessageService(queueId string, messageId string) (*entities.QueueMessage, error) {

	return entities.GetQueueMessage(queueId, messageId)

}

func UpdateMessageService(queueId string, messageId string, mp UpdateMessageRequest) (string, error) {

	return entities.UpdateQueueMessage(queueId, messageId, mp.Data)

}

func DeleteMessageService(queueId string, messageId string) (string, error) {

	return entities.DeleteQueueMessage(queueId, messageId)

}
