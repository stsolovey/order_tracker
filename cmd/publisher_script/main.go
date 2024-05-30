package main

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/stsolovey/order_tracker/internal/config"
	"github.com/stsolovey/order_tracker/internal/logger"
	"github.com/stsolovey/order_tracker/internal/models"
)

func main() {
	cfg := config.New("./.env")
	log := logger.New(cfg.LogLevel)

	nc, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		log.WithError(err).Panic("Failed to connect to NATS")
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		log.WithError(err).Panic("Failed to get JetStream context")
	}

	for i := range 100 {
		order := generateSampleOrder(log, i)

		data, err := json.Marshal(order)
		if err != nil {
			log.WithError(err).Panic("Failed to marshal order")
		}

		_, err = js.Publish("orders", data)
		if err != nil {
			log.WithError(err).Panic("Failed to publish order")
		}

		log.Infof("Published order '%s'", order.OrderUID)
		time.Sleep(1 * time.Second)
	}
}

func generateSampleOrder(log *logrus.Logger, i int) models.Order {
	orderUID := "orderUID" + strconv.Itoa(i)

	const (
		randomMax100000 = 100000
		randomMax10000  = 10000
		randomMax1000   = 1000
		randomMax100    = 100
	)

	return models.Order{
		OrderUID:        orderUID,
		TrackNumber:     fmt.Sprintf("TRACK_%d", secureRandomInt(log, randomMax1000)),
		Entry:           "WBIL",
		Locale:          "en",
		CustomerID:      "test_customer",
		DeliveryService: "test_delivery",
		DateCreated:     time.Now(),
		Delivery: models.Delivery{
			OrderUID: orderUID,
			Name:     "John Doe",
			Phone:    "+1234567890",
			City:     "Test City",
			Address:  "123 Test St",
		},
		Payment: models.Payment{
			OrderUID:    orderUID,
			Transaction: fmt.Sprintf("TXN_%d", secureRandomInt(log, randomMax1000)),
			Currency:    "USD",
			Provider:    "test_provider",
			Amount:      float64(secureRandomInt(log, randomMax10000)),
			PaymentDT:   time.Now(),
		},
		Items: []models.Item{
			{
				OrderUID:    orderUID,
				ChrtID:      secureRandomInt(log, randomMax100000),
				TrackNumber: fmt.Sprintf("TRACK_%d", secureRandomInt(log, randomMax1000)),
				Price:       float64(secureRandomInt(log, randomMax100)),
				Name:        "Test Item 1",
				NMID:        secureRandomInt(log, randomMax1000),
				Brand:       "TestBrand",
				Status:      1,
			},
			{
				OrderUID:    orderUID,
				ChrtID:      secureRandomInt(log, randomMax100000),
				TrackNumber: fmt.Sprintf("TRACK_%d", secureRandomInt(log, randomMax1000)),
				Price:       float64(secureRandomInt(log, randomMax100)),
				Name:        "Test Item 2",
				NMID:        secureRandomInt(log, randomMax1000),
				Brand:       "TestBrand",
				Status:      2, //nolint:mnd
			},
		},
	}
}

func secureRandomInt(log *logrus.Logger, max int) int {
	var b [8]byte

	_, err := rand.Read(b[:])
	if err != nil {
		log.WithError(err).Panic("Failed to generate secure random number")
	}

	return int(binary.BigEndian.Uint64(b[:]) % uint64(max))
}
