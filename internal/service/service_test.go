package service_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/stsolovey/order_tracker/internal/models"
	"github.com/stsolovey/order_tracker/internal/service"
)

type MockCache struct {
	UpsertFunc func(ctx context.Context, order models.Order) error
	GetFunc    func(ctx context.Context, orderUID string) (*models.Order, error)
}

func (m *MockCache) Upsert(ctx context.Context, order models.Order) error {
	if m.UpsertFunc != nil {
		return m.UpsertFunc(ctx, order)
	}
	return nil
}

func (m *MockCache) Get(ctx context.Context, orderUID string) (*models.Order, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, orderUID)
	}
	return nil, models.ErrOrderNotFound
}

type MockStorage struct {
	GetFunc    func(ctx context.Context, orderUID string) (*models.Order, error)
	GetAllFunc func(ctx context.Context) ([]models.Order, error)
	UpsertFunc func(ctx context.Context, order *models.Order) (*models.Order, error)
}

func (m *MockStorage) Get(ctx context.Context, orderUID string) (*models.Order, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, orderUID)
	}
	return nil, models.ErrOrderNotFound
}

func (m *MockStorage) GetAll(ctx context.Context) ([]models.Order, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc(ctx)
	}
	return nil, nil
}

func (m *MockStorage) Upsert(ctx context.Context, order *models.Order) (*models.Order, error) {
	if m.UpsertFunc != nil {
		return m.UpsertFunc(ctx, order)
	}
	return order, nil
}

type ServiceSuite struct {
	suite.Suite
	service     *service.Service
	mockCache   *MockCache
	mockStorage *MockStorage
	log         *logrus.Logger
}

func (s *ServiceSuite) SetupSuite() {
	s.log = logrus.New()
	s.mockCache = &MockCache{}
	s.mockStorage = &MockStorage{}
	s.service = service.New(s.log, s.mockCache, s.mockStorage)
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}

func (s *ServiceSuite) TestInit() {
	expectedOrders := []models.Order{
		{
			OrderUID:    "testUID123",
			TrackNumber: "TN1234567890",
			CustomerID:  "Cust123",
			DateCreated: time.Now(),
		},
	}

	s.mockStorage.GetAllFunc = func(ctx context.Context) ([]models.Order, error) {
		return expectedOrders, nil
	}

	s.mockCache.UpsertFunc = func(ctx context.Context, order models.Order) error {
		return nil
	}

	err := s.service.Init(context.Background())
	s.Require().NoError(err)
}

func (s *ServiceSuite) TestUpsertOrder() {
	order := models.Order{
		OrderUID:    "testUID123",
		TrackNumber: "TN1234567890",
		CustomerID:  "Cust123",
		DateCreated: time.Now(),
	}

	s.mockCache.UpsertFunc = func(ctx context.Context, order models.Order) error {
		return nil
	}

	s.mockStorage.UpsertFunc = func(ctx context.Context, order *models.Order) (*models.Order, error) {
		return order, nil
	}

	err := s.service.UpsertOrder(context.Background(), order)
	s.Require().NoError(err)
}

func (s *ServiceSuite) TestGetOrder_CacheHit() {
	order := models.Order{
		OrderUID:    "testUID123",
		TrackNumber: "TN1234567890",
		CustomerID:  "Cust123",
		DateCreated: time.Now(),
	}

	s.mockCache.GetFunc = func(ctx context.Context, orderUID string) (*models.Order, error) {
		return &order, nil
	}

	result, err := s.service.GetOrder(context.Background(), "testUID123")
	s.Require().NoError(err)
	s.Require().Equal(order.OrderUID, result.OrderUID)
}

func (s *ServiceSuite) TestGetOrder_CacheMiss() {
	order := models.Order{
		OrderUID:    "testUID123",
		TrackNumber: "TN1234567890",
		CustomerID:  "Cust123",
		DateCreated: time.Now(),
	}

	s.mockCache.GetFunc = func(ctx context.Context, orderUID string) (*models.Order, error) {
		return nil, models.ErrOrderNotFound
	}

	s.mockStorage.GetFunc = func(ctx context.Context, orderUID string) (*models.Order, error) {
		return &order, nil
	}

	s.mockCache.UpsertFunc = func(ctx context.Context, order models.Order) error {
		return nil
	}

	result, err := s.service.GetOrder(context.Background(), "testUID123")
	s.Require().NoError(err)
	s.Require().Equal(order.OrderUID, result.OrderUID)
}

func (s *ServiceSuite) TestConcurrentUpsert() {
	order := models.Order{
		OrderUID:    "testUID123",
		TrackNumber: "TN1234567890",
		CustomerID:  "Cust123",
		DateCreated: time.Now(),
	}

	s.mockCache.UpsertFunc = func(ctx context.Context, order models.Order) error {
		return nil
	}

	s.mockStorage.UpsertFunc = func(ctx context.Context, order *models.Order) (*models.Order, error) {
		return order, nil
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := s.service.UpsertOrder(context.Background(), order)
			s.Require().NoError(err)
		}()
	}
	wg.Wait()
}
