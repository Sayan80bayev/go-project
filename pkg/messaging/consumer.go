package messaging

import "context"

type Consumer interface {
    Start(ctx context.Context)
    Close()
}