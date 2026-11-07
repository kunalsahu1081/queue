package interfaces

import "letsgo/internal/domain/entities"

type QueueActions interface {
	Create(queue entities.Queue) (string, error)
	GetByID(id string) (*entities.Queue, error)
	Update(queue entities.Queue) error
	Delete(id string) error
}
