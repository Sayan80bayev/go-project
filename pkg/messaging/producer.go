package messaging

type Producer interface {
	Produce(evenType string, data interface{}) error
	Close()
}
