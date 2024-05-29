package integrationtest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/stsolovey/order_tracker/internal/config"
	"github.com/stsolovey/order_tracker/internal/logger"
	"github.com/stsolovey/order_tracker/internal/models"
	natsclient "github.com/stsolovey/order_tracker/internal/nats-client"
	ordercache "github.com/stsolovey/order_tracker/internal/order-cache"
	"github.com/stsolovey/order_tracker/internal/service"
	"github.com/stsolovey/order_tracker/internal/storage"
)

type IntegrationTestSuite struct {
	suite.Suite
	log        *logrus.Logger
	db         *storage.Storage
	cfg        *config.Config
	natsConn   *nats.Conn
	natsClient *natsclient.Client
	app        *service.Service
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.cfg = config.New("../.env")
	s.log = logger.New(s.cfg.LogLevel)

	var err error
	s.db, err = storage.NewStorage(context.Background(), s.log, s.cfg.DatabaseURL)
	s.Require().NoError(err, "should connect to database without error")

	orderCache := ordercache.New(s.log)
	s.app = service.New(s.log, orderCache, s.db)

	err = s.app.Init(context.Background())
	s.Require().NoError(err, "should initialize app without error")

	s.natsConn, err = nats.Connect(s.cfg.NATSURL)
	s.Require().NoError(err, "should connect to NATS without error")

	_, err = s.natsConn.JetStream()
	s.Require().NoError(err, "should get JetStream context without error")

	s.natsClient, err = natsclient.New(s.cfg, s.log, s.app)
	s.Require().NoError(err, "should initialize NATS client without error")

	err = s.natsClient.Subscribe(context.Background(), "orders")
	s.Require().NoError(err, "should subscribe to NATS subject without error")
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.natsClient.Close()
	s.natsConn.Close()
	s.truncateTables()
	s.db.DB().Close()
}

func (s *IntegrationTestSuite) truncateTables() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tables := []string{"items", "payment", "delivery", "orders"}
	for _, table := range tables {
		_, err := s.db.DB().Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
		s.Require().NoError(err, fmt.Sprintf("should truncate table %s without error", table))
	}
	s.log.Infof("Tables truncated successfully")
}

func (s *IntegrationTestSuite) TestNATSIntegration() {
	order := models.Order{
		OrderUID:    "uniqueOrderID123",
		TrackNumber: "TN1234567890",
		Entry:       "WBIL",
		Delivery: models.Delivery{
			OrderUID: "uniqueOrderID123",
			Name:     "John Doe",
			Phone:    "+1234567890",
			City:     "TestCity",
			Address:  "123 Test St",
		},
		Payment: models.Payment{
			OrderUID:    "uniqueOrderID123",
			Transaction: "TX1234567890",
			Currency:    "USD",
			Provider:    "TestProvider",
			Amount:      150.00,
			PaymentDT:   time.Now(),
		},
		Items: []models.Item{
			{
				OrderUID:    "uniqueOrderID123",
				ChrtID:      1,
				TrackNumber: "TN1234567890",
				Price:       100.00,
				Name:        "Test Item 1",
				NMID:        1001,
				Brand:       "TestBrand",
				Status:      1,
			},
			{
				OrderUID:    "uniqueOrderID123",
				ChrtID:      2,
				TrackNumber: "TN1234567890",
				Price:       50.00,
				Name:        "Test Item 2",
				NMID:        2002,
				Brand:       "BrandTest",
				Status:      2,
			},
		},
		Locale:          "en",
		CustomerID:      "Cust123",
		DeliveryService: "TestService",
		DateCreated:     time.Now(),
	}

	err := s.natsClient.PublishOrder(order)
	s.Require().NoError(err, "should publish without error")

	time.Sleep(time.Second)

	retrievedOrder, err := s.app.GetOrder(context.Background(), order.OrderUID)
	s.Require().NoError(err, "should GetOrder(...) work without error")
	s.Require().NotNil(retrievedOrder, "retrieved with GetOrder(...) should not be nil")
	s.Require().Equal(order.OrderUID, retrievedOrder.OrderUID, "sent and got OrderUID should match")
	s.Require().Equal(order.TrackNumber, retrievedOrder.TrackNumber, "sent and got TrackNumber should match")
	s.Require().Equal(order.Delivery.Name, retrievedOrder.Delivery.Name, "sent and got DeliveryName should match")
	s.Require().Equal(order.Payment.Transaction, retrievedOrder.Payment.Transaction, "payment transaction should match")
	s.Require().Equal(len(order.Items), len(retrievedOrder.Items), "sent and got number of Items should match")
}
