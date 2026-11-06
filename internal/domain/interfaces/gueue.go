

type QueueActions interface {
    Create(queue Queue) (string, error)
    GetByID(id string) (*Queue, error)
    Update(queue Queue) error
    Delete(id string) error
}