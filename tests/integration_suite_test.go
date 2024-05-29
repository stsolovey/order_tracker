package integrationtest

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/stsolovey/order_tracker/internal/config"
	"github.com/stsolovey/order_tracker/internal/logger"
	natsclient "github.com/stsolovey/order_tracker/internal/nats-client"
	ordercache "github.com/stsolovey/order_tracker/internal/order-cache"
	"github.com/stsolovey/order_tracker/internal/server"
	"github.com/stsolovey/order_tracker/internal/service"
	"github.com/stsolovey/order_tracker/internal/storage"
)

type IntegrationTestSuite struct { //nolint:revive
	suite.Suite
	log              *logrus.Logger
	orderCache       *ordercache.OrderCache
	db               *storage.Storage
	cfg              *config.Config
	natsConn         *nats.Conn
	natsClient       *natsclient.Client
	app              *service.Service
	httpServer       *server.Server
	httpServerCancel context.CancelFunc
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

	s.orderCache = ordercache.New(s.log)
	s.app = service.New(s.log, s.orderCache, s.db)

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

	s.httpServer = server.CreateServer(s.cfg, s.log, s.app)

	var ctx context.Context
	ctx, s.httpServerCancel = signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	err = s.httpServer.Start(ctx)
	s.Require().NoError(err, "should start httpServer without error")
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.natsClient.Close()
	s.natsConn.Close()
	s.truncateTables()
	s.db.DB().Close()
	s.httpServerCancel()
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
