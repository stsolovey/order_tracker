package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stsolovey/order_tracker/internal/config"
	"github.com/stsolovey/order_tracker/internal/logger"
	"github.com/stsolovey/order_tracker/internal/models"
)

const (
	numOfOrdersToGenerate     = 100
	numOfOrdersSlowGeneration = 10
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

	for i := range numOfOrdersToGenerate {
		order := generateSampleOrder(i)

		data, err := json.Marshal(order)
		if err != nil {
			log.WithError(err).Panic("Failed to marshal order")
		}

		_, err = js.Publish("orders", data)
		if err != nil {
			log.WithError(err).Panic("Failed to publish order")
		}

		log.Infof("Published order '%s'", order.OrderUID)

		if i < numOfOrdersSlowGeneration {
			time.Sleep(1 * time.Second)
		}
	}
}

func generateSampleOrder(i int) models.Order {
	orderUID := "orderUID" + strconv.Itoa(i)

	const (
		trackNumberMax       = 1000
		transactionNumberMax = 1000
		amountMax            = 10000
		chrtIDMax            = 100000
		priceMax             = 100
		nmidMax              = 1000
	)

	return models.Order{
		OrderUID:        orderUID,
		TrackNumber:     fmt.Sprintf("TRACK_%d", randInt(trackNumberMax)),
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
			Transaction: fmt.Sprintf("TXN_%d", randInt(transactionNumberMax)),
			Currency:    "USD",
			Provider:    "test_provider",
			Amount:      float64(randInt(amountMax)),
			PaymentDT:   time.Now(),
		},
		Items: []models.Item{
			{
				OrderUID:    orderUID,
				ChrtID:      randInt(chrtIDMax),
				TrackNumber: fmt.Sprintf("TRACK_%d", randInt(trackNumberMax)),
				Price:       float64(randInt(priceMax)),
				Name:        "Test Item 1",
				NMID:        randInt(nmidMax),
				Brand:       "TestBrand",
				Status:      1,
			},
			{
				OrderUID:    orderUID,
				ChrtID:      randInt(chrtIDMax),
				TrackNumber: fmt.Sprintf("TRACK_%d", randInt(trackNumberMax)),
				Price:       float64(randInt(priceMax)),
				Name:        "Test Item 2",
				NMID:        randInt(nmidMax),
				Brand:       "TestBrand",
				Status:      2, //nolint:mnd
			},
		},
	}
}

func randInt(max int) int {
	return int(time.Now().UnixNano()) % max
}
