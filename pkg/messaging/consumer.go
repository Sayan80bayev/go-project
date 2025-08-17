package messaging

type Consumer interface {
	Start()
	Close()
}
