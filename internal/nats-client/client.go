package natsclient

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

type Client struct {
	conn    *nats.Conn
	js      nats.JetStreamContext
	log     *logrus.Logger
	service service.OrderServiceInterface
}

func New(cfg *config.Config, log *logrus.Logger, svc service.OrderServiceInterface) (*Client, error) {
	nc, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		return nil, fmt.Errorf("natsclient New(...) nats.Connect(...): %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("natsclient New(...) nc.JetStream(...): %w", err)
	}

	client := &Client{
		conn:    nc,
		js:      js,
		log:     log,
		service: svc,
	}

	return client, nil
}

func (nc *Client) Subscribe(ctx context.Context, subject string) error {
	go func() {
		<-ctx.Done()
		nc.Close()
	}()

	_, err := nc.js.QueueSubscribe(subject, queueGroupName, func(msg *nats.Msg) {
		var order models.Order

		if err := json.Unmarshal(msg.Data, &order); err != nil {
			nc.log.WithError(err).Error("failed to unmarshal order")

			if err := msg.Nak(); err != nil {
				nc.log.WithError(err).Error("failed to negatively acknowledge message")
			}

			return
		}

		if err := nc.service.UpsertOrder(ctx, order); err != nil {
			nc.log.WithError(err).Error("failed to upsert order")

			return
		}

		nc.log.Infof("Order %s upserted successfully", order.OrderUID)

		if err := msg.Ack(); err != nil {
			nc.log.WithError(err).Error("failed to acknowledge message")
		}
	})
	if err != nil {
		return fmt.Errorf("natsclient Subscribe(...): %w", err)
	}

	return nil
}

func (nc *Client) PublishOrder(order models.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("client.go PublishOrder(...) json.Marshal(order): %w", err)
	}

	_, err = nc.js.Publish("orders", data)
	if err != nil {
		return fmt.Errorf("client.go PublishOrder(...) nc.js.Publish(...): %w", err)
	}

	return nil
}

func (nc *Client) Close() {
	nc.conn.Close()
}
