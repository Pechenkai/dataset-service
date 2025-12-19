package messaging

import (
	"context"

	"ppo/internal/contracts"
)

type Handler func(context.Context, contracts.Envelope) error

type Bus interface {
	Publish(ctx context.Context, queue string, msg contracts.Envelope) error
	Consume(ctx context.Context, queue string, handler Handler) error
	Close() error
}
