package messaging

import "context"

type Producer interface {
    Produce(ctx context.Context, eventType string, data interface{}) error
    Close()
}