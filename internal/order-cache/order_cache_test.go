package ordercache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/stsolovey/order_tracker/internal/models"
)

type OrderCacheSuite struct {
	suite.Suite
	cache  *OrderCache
	ctx    context.Context
	cancel context.CancelFunc
}

func (s *OrderCacheSuite) SetupTest() { // run before
	logger := logrus.New()
	s.cache = New(logger)
	s.ctx, s.cancel = context.WithCancel(context.Background())
}

func (s *OrderCacheSuite) TearDownTest() { // run after
	s.cancel()
}

func TestOrderCacheSuite(t *testing.T) {
	suite.Run(t, new(OrderCacheSuite))
}

func (s *OrderCacheSuite) TestUpsert() {
	order := models.Order{
		OrderUID:    "testUID123",
		TrackNumber: "TN1234567890",
		CustomerID:  "Cust123",
		DateCreated: time.Now(),
	}

	s.Run("inserting a new order", func() {
		err := s.cache.Upsert(s.ctx, order)
		s.Require().NoError(err)

		retrievedOrder, err := s.cache.Get(s.ctx, order.OrderUID)
		s.Require().NoError(err, "Get should not return an error after Upsert")
		s.Require().NotNil(retrievedOrder, "Retrieved order should not be nil")
		s.Require().Equal(order.OrderUID, retrievedOrder.OrderUID)
	})

	updatedOrder := order
	updatedOrder.TrackNumber = "TN0987654321"

	s.Run("updating an existing order", func() {
		err := s.cache.Upsert(s.ctx, updatedOrder)
		s.Require().NoError(err)

		retrievedUpdatedOrder, err := s.cache.Get(s.ctx, updatedOrder.OrderUID)
		s.Require().NoError(err)
		s.Require().NotNil(retrievedUpdatedOrder)
		s.Require().Equal(updatedOrder.TrackNumber, retrievedUpdatedOrder.TrackNumber)
	})
}

func (s *OrderCacheSuite) TestGet() {
	orderUID := "testUID456"
	order := models.Order{
		OrderUID:    orderUID,
		TrackNumber: "TN1234567890",
		CustomerID:  "Cust456",
		DateCreated: time.Now(),
	}

	err := s.cache.Upsert(s.ctx, order)
	s.Require().NoError(err, "Upsert should not fail")

	s.Run("retrieve existing order", func() {
		retrievedOrder, err := s.cache.Get(s.ctx, orderUID)
		s.Require().NoError(err, "Error returned")
		s.Require().NotNil(retrievedOrder, "Order is nil!")
		s.Require().Equal(order.OrderUID, retrievedOrder.OrderUID, "Wrong OrderUID's!")
	})
}

func (s *OrderCacheSuite) TestDelete() {
	orderUID := "testUID123"
	order := models.Order{
		OrderUID:    orderUID,
		CustomerID:  "Cust123",
		DateCreated: time.Now(),
	}

	err := s.cache.Upsert(s.ctx, order)
	s.Require().NoError(err, "Upsert should not fail before delete")

	s.cache.Delete(s.ctx, orderUID)

	retrievedOrder, err := s.cache.Get(s.ctx, orderUID)
	s.Run("verify deletion", func() {
		s.Require().Error(err, "Get should return an error after Delete")
		s.Require().True(errors.Is(err, models.ErrOrderNotFound), "Error should be 'ErrOrderNotFound'")
		s.Require().Nil(retrievedOrder, "Retrieved order should be nil after deletion")
	})
}
