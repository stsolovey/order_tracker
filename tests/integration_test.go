package integrationtest

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // Importing `pgx/v5/stdlib` is necessary for `sql.Open("pgx", s.dsn)`.
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/stsolovey/order_tracker/internal/config"
	"github.com/stsolovey/order_tracker/internal/logger"
	natsconsumer "github.com/stsolovey/order_tracker/internal/nats-consumer"
	ordercache "github.com/stsolovey/order_tracker/internal/order-cache"
	"github.com/stsolovey/order_tracker/internal/service"
	"github.com/stsolovey/order_tracker/internal/storage"
)

const (
// natsURL = "nats://localhost:4222"
)

type IntegrationTestSuite struct {
	suite.Suite
	log        *logrus.Logger
	db         *storage.Storage
	cfg        *config.Config
	natsConn   *nats.Conn
	natsClient *natsconsumer.Consumer
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.cfg = config.New("../.env")
	s.log = logger.New(s.cfg.LogLevel)

	var err error
	s.db, err = storage.NewStorage(context.Background(), s.log, s.cfg.DatabaseURL)
	s.Require().NoError(err, "should connect to database without error")

	err = s.db.Migrate()
	s.Require().NoError(err, "should migrate without error")

	orderCache := ordercache.New(s.log)
	app := service.New(s.log, orderCache, s.db)

	err = app.Init(context.Background())
	s.Require().NoError(err, "should initialize app without error")

	s.natsConn, err = nats.Connect(s.cfg.NATSURL)
	s.Require().NoError(err, "should connect to NATS without error")

	js, err := s.natsConn.JetStream()
	s.Require().NoError(err, "should get JetStream context without error")

	streamName := "ORDERS"
	streamSubjects := []string{"orders"}
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: streamSubjects,
	})
	if err != nil && err != nats.ErrStreamNameAlreadyInUse {
		s.Require().NoError(err, "should create or reuse an existing stream without error")
	}

	s.natsClient, err = natsconsumer.New(s.cfg, s.log, app)
	s.Require().NoError(err, "should initialize NATS client without error")

	err = s.natsClient.Subscribe(context.Background(), "orders")
	s.Require().NoError(err, "should subscribe to NATS subject without error")
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.natsClient.Close()
	s.natsConn.Close()
	s.truncateTables()
	// s.db.Close()
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
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

func (s *IntegrationTestSuite) TestConsumer() {
	/*
		order := &models.Order{
			OrderUID:        "uniqueOrderID123",
			TrackNumber:     "TN1234567890",
			CustomerID:      "Cust123",
			DateCreated:     time.Now(),
			DeliveryService: "TestService",
			Locale:          "en",
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
					ChrtID:      1,
					OrderUID:    "uniqueOrderID123",
					TrackNumber: "TN1234567890",
					Price:       100.00,
					Name:        "Test Item 1",
					NMID:        1001,
					Brand:       "TestBrand",
					Status:      1,
				},
				{
					ChrtID:      2,
					OrderUID:    "uniqueOrderID123",
					TrackNumber: "TN1234567890",
					Price:       50.00,
					Name:        "Test Item 2",
					NMID:        2002,
					Brand:       "BrandTest",
					Status:      2,
				},
			},
		}*/
}
