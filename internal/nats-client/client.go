package natsconsumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/stsolovey/order_tracker/internal/config"
	"github.com/stsolovey/order_tracker/internal/models"
	"github.com/stsolovey/order_tracker/internal/service"
)

const queueGroupName = "order_tracker"

type Consumer struct {
	conn    *nats.Conn
	js      nats.JetStreamContext
	log     *logrus.Logger
	service service.OrderServiceInterface
}

func New(cfg *config.Config, log *logrus.Logger, svc service.OrderServiceInterface) (*Consumer, error) {
	nc, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		return nil, fmt.Errorf("natsconsumer New(...) nats.Connect(...): %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("natsconsumer New(...) nc.JetStream(...): %w", err)
	}

	client := &Consumer{
		conn:    nc,
		js:      js,
		log:     log,
		service: svc,
	}

	return client, nil
}

func (nc *Consumer) Subscribe(ctx context.Context, subject string) error {
	go func() {
		<-ctx.Done()
		nc.Close()
	}()

	_, err := nc.js.QueueSubscribe(subject, queueGroupName, func(msg *nats.Msg) {
		var order models.Order

		if err := json.Unmarshal(msg.Data, &order); err != nil {
			nc.log.WithError(err).Error("failed to unmarshal order")

			if nakErr := msg.Nak(); nakErr != nil {
				nc.log.WithError(nakErr).Error("failed to negatively acknowledge message")
			}

			return
		}

		if err := nc.service.UpsertOrder(ctx, order); err != nil {
			nc.log.WithError(err).Error("failed to upsert order")

			if nakErr := msg.Nak(); nakErr != nil {
				nc.log.WithError(nakErr).Error("failed to negatively acknowledge message")
			}

			return
		}

		nc.log.Infof("Order %s upserted successfully", order.OrderUID)

		if err := msg.AckSync(); err != nil {
			nc.log.WithError(err).Error("failed to acknowledge message")
		}
	})
	if err != nil {
		return fmt.Errorf("natsconsumer Subscribe(...): %w", err)
	}

	return nil
}

func (nc *Consumer) PublishOrder(order models.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("consumer.go PublishOrder(...) json.Marshal(order): %w", err)
	}

	_, err = nc.js.Publish("orders", data)
	if err != nil {
		return fmt.Errorf("consumer.go PublishOrder(...) nc.js.Publish(...): %w", err)
	}

	return nil
}

func (nc *Consumer) Close() {
	nc.conn.Close()
}
