package rabbit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	"ppo/internal/config"
	"ppo/internal/contracts"
	"ppo/internal/messaging"
	"ppo/internal/metrics"
)

type Bus struct {
	cfg    config.Broker
	conn   *amqp.Connection
	ch     *amqp.Channel
	logger *zap.Logger
}

func New(cfg config.Broker, logger *zap.Logger) (*Bus, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("connect to broker: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}
	if cfg.Prefetch > 0 {
		if err := ch.Qos(cfg.Prefetch, 0, false); err != nil {
			_ = ch.Close()
			_ = conn.Close()
			return nil, fmt.Errorf("configure qos: %w", err)
		}
	}

	bus := &Bus{cfg: cfg, conn: conn, ch: ch, logger: logger}
	if err := bus.ensureQueues(); err != nil {
		_ = bus.Close()
		return nil, err
	}
	return bus, nil
}

func (b *Bus) ensureQueues() error {
	declared := map[string]struct{}{}
	queues := []string{
		b.cfg.GatewayToCoreQueue,
		b.cfg.CoreToGatewayQueue,
		b.cfg.CoreToDataQueue,
		b.cfg.DataToCoreQueue,
	}

	for _, q := range queues {
		if q == "" {
			continue
		}
		if _, ok := declared[q]; ok {
			continue
		}
		args := amqp.Table{}
		if b.cfg.EnableDLQ {
			args["x-dead-letter-exchange"] = ""
			args["x-dead-letter-routing-key"] = q + ".dlq"
		}
		if _, err := b.ch.QueueDeclare(
			q,
			true,  // durable
			false, // auto-delete
			false, // exclusive
			false, // no-wait
			args,
		); err != nil {
			return fmt.Errorf("declare queue %s: %w", q, err)
		}
		if b.cfg.EnableDLQ {
			if _, err := b.ch.QueueDeclare(
				q+".dlq",
				true,
				false,
				false,
				false,
				nil,
			); err != nil {
				return fmt.Errorf("declare dlq %s: %w", q+".dlq", err)
			}
		}
		declared[q] = struct{}{}
	}
	return nil
}

func (b *Bus) Publish(ctx context.Context, queue string, msg contracts.Envelope) error {
	if msg.ID == "" {
		msg.ID = uuid.NewString()
	}
	if msg.OccurredAt.IsZero() {
		msg.OccurredAt = time.Now()
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal envelope: %w", err)
	}

	if err := b.ch.PublishWithContext(ctx, "", queue, false, false, amqp.Publishing{
		ContentType:   "application/json",
		Body:          body,
		MessageId:     msg.ID,
		CorrelationId: msg.CorrelationID,
		Timestamp:     msg.OccurredAt,
		ReplyTo:       msg.ReplyTo,
	}); err != nil {
		metrics.ObserveConsume(queue, "publish_error", 0)
		return fmt.Errorf("publish to queue %s: %w", queue, err)
	}
	metrics.IncPublished(queue)
	return nil
}

func (b *Bus) Consume(ctx context.Context, queue string, handler messaging.Handler) error {
	consumer := fmt.Sprintf("%s-consumer-%s", queue, uuid.NewString())

	msgs, err := b.ch.ConsumeWithContext(ctx, queue, consumer, false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume from queue %s: %w", queue, err)
	}

	go b.consumeLoop(ctx, queue, msgs, handler)

	return nil
}

func (b *Bus) Close() error {
	if b.ch != nil {
		_ = b.ch.Close()
	}
	if b.conn != nil {
		return b.conn.Close()
	}
	return nil
}

func (b *Bus) consumeLoop(ctx context.Context, queue string, msgs <-chan amqp.Delivery, handler messaging.Handler) {
	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-msgs:
			if !ok {
				return
			}
			b.handleDelivery(ctx, queue, handler, d)
		}
	}
}

func (b *Bus) handleDelivery(ctx context.Context, queue string, handler messaging.Handler, d amqp.Delivery) {
	start := time.Now()
	var env contracts.Envelope
	if err := json.Unmarshal(d.Body, &env); err != nil {
		_ = d.Nack(false, false)
		metrics.ObserveConsume(queue, "decode_error", time.Since(start))
		return
	}
	if env.OccurredAt.IsZero() {
		env.OccurredAt = time.Now()
	}
	if err := handler(ctx, env); err != nil {
		_ = d.Nack(false, false)
		metrics.ObserveConsume(queue, "error", time.Since(start))
		if b.logger != nil {
			b.logger.Warn("handler failed", zap.String("queue", queue), zap.Error(err))
		}
		return
	}
	_ = d.Ack(false)
	metrics.ObserveConsume(queue, "ok", time.Since(start))
}

var _ messaging.Bus = (*Bus)(nil)
