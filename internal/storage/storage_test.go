package storage_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/stsolovey/order_tracker/internal/config"
	"github.com/stsolovey/order_tracker/internal/logger"
	"github.com/stsolovey/order_tracker/internal/models"
	"github.com/stsolovey/order_tracker/internal/storage"
)

type StorageSuite struct {
	suite.Suite
	storage *storage.Storage
	ctx     context.Context
	cancel  context.CancelFunc
}

func (s *StorageSuite) SetupSuite() {
	cfg := config.New("../../.env") // если '.env' в корне
	log := logger.New(cfg.LogLevel)
	dsn := cfg.DatabaseURL
	ctx, cancel := context.WithCancel(context.Background())
	stor, err := storage.NewStorage(ctx, log, dsn)
	if err != nil {
		s.T().Fatal(err)
	}
	s.storage = stor
	s.ctx = ctx
	s.cancel = cancel

	err = s.cleanDatabase()
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *StorageSuite) TearDownSuite() {
	s.cancel()
	err := s.cleanDatabase()
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *StorageSuite) cleanDatabase() error {
	queries := []string{
		"DELETE FROM items;",
		"DELETE FROM payment;",
		"DELETE FROM delivery;",
		"DELETE FROM orders;",
	}

	for _, query := range queries {
		if _, err := s.storage.DB().Exec(context.Background(), query); err != nil {
			return err
		}
	}
	return nil
}

func TestStorageSuite(t *testing.T) {
	suite.Run(t, new(StorageSuite))
}

func (s *StorageSuite) TestUpsertOrder() {
	order := &models.Order{
		OrderUID:          "testUID123",
		TrackNumber:       "TN1234567890",
		CustomerID:        "Cust123",
		DateCreated:       time.Now(),
		DeliveryService:   "TestService",
		InternalSignature: "",
		Locale:            "en",
		Shardkey:          "9",
		SMID:              99,
	}

	s.Run("Insertion of a new order", func() {
		insertedOrder, err := s.storage.UpsertOrder(s.ctx, s.storage.DB(), order)
		s.Require().NoError(err, "Upsert should not fail on insertion")
		s.Require().NotNil(insertedOrder, "Inserted order should not be nil")
		s.Require().Equal(order.OrderUID, insertedOrder.OrderUID, "Order UID should match")
	})

	s.Run("Updating the existing order", func() {
		order.TrackNumber = "TN0987654321"
		updatedOrder, err := s.storage.UpsertOrder(s.ctx, s.storage.DB(), order)
		s.Require().NoError(err, "Upsert should not fail on update")
		s.Require().NotNil(updatedOrder, "Updated order should not be nil")
		s.Require().Equal("TN0987654321", updatedOrder.TrackNumber, "Track number should be updated")
	})
}

func (s *StorageSuite) TestUpsertDelivery() {
	delivery := &models.Delivery{
		OrderUID: "testUID123",
		Name:     "John Doe",
		Phone:    "+1234567890",
		Zip:      "12345",
		City:     "TestCity",
		Address:  "123 Test St",
		Region:   "TestRegion",
		Email:    "john.doe@example.com",
	}

	s.Run("Insertion of a new delivery", func() {
		insertedDelivery, err := s.storage.UpsertDelivery(s.ctx, s.storage.DB(), delivery)
		s.Require().NoError(err, "Upsert should not fail on insertion")
		s.Require().NotNil(insertedDelivery, "Inserted delivery should not be nil")
		s.Require().Equal(delivery.OrderUID, insertedDelivery.OrderUID, "Order UID should match for delivery")
	})

	s.Run("Updating the existing delivery", func() {
		delivery.Phone = "+0987654321"
		updatedDelivery, err := s.storage.UpsertDelivery(s.ctx, s.storage.DB(), delivery)
		s.Require().NoError(err, "Upsert should not fail on update")
		s.Require().NotNil(updatedDelivery, "Updated delivery should not be nil")
		s.Require().Equal("+0987654321", updatedDelivery.Phone, "Phone number should be updated")
	})
}

func (s *StorageSuite) TestUpsertPayment() {
	payment := &models.Payment{
		OrderUID:     "testUID123",
		Transaction:  "TX1234567890",
		RequestID:    "RQ1234567890",
		Currency:     "USD",
		Provider:     "TestProvider",
		Amount:       100.50,
		PaymentDT:    time.Now(),
		Bank:         "TestBank",
		DeliveryCost: 5.00,
		GoodsTotal:   95.50,
		CustomFee:    0.00,
	}

	s.Run("Insertion of a new payment", func() {
		insertedPayment, err := s.storage.UpsertPayment(s.ctx, s.storage.DB(), payment)
		s.Require().NoError(err, "Upsert should not fail on insertion")
		s.Require().NotNil(insertedPayment, "Inserted payment should not be nil")
		s.Require().Equal(payment.OrderUID, insertedPayment.OrderUID, "Order UID should match for payment")
	})

	s.Run("Updating the existing payment", func() {
		payment.Amount = 200.00
		updatedPayment, err := s.storage.UpsertPayment(s.ctx, s.storage.DB(), payment)
		s.Require().NoError(err, "Upsert should not fail on update")
		s.Require().NotNil(updatedPayment, "Updated payment should not be nil")
		s.Require().Equal(200.00, updatedPayment.Amount, "Payment amount should be updated")
	})
}

func (s *StorageSuite) TestUpsertItems() {
	items := &[]models.Item{
		{
			OrderUID:    "testUID123",
			ChrtID:      101,
			TrackNumber: "TN1234567890",
			Price:       29.99,
			RID:         "RID1234567890",
			Name:        "Test Item 1",
			Sale:        10,
			Size:        "M",
			TotalPrice:  26.99,
			NMID:        1001,
			Brand:       "TestBrand",
			Status:      0,
		},
		{
			OrderUID:    "testUID123",
			ChrtID:      102,
			TrackNumber: "TN0987654321",
			Price:       39.99,
			RID:         "RID0987654321",
			Name:        "Test Item 2",
			Sale:        0,
			Size:        "L",
			TotalPrice:  39.99,
			NMID:        1002,
			Brand:       "TestBrand",
			Status:      1,
		},
	}

	s.Run("Insertion of new items", func() {
		insertedItems, err := s.storage.UpsertItems(s.ctx, s.storage.DB(), *items)
		s.Require().NoError(err, "Upsert should not fail on insertion")
		s.Require().NotNil(insertedItems, "Inserted items should not be nil")
		s.Require().Len(*insertedItems, 2, "Should insert two items")
	})

	s.Run("Updating the existing items", func() {
		(*items)[0].Price = 19.99
		(*items)[0].TotalPrice = 17.99
		(*items)[0].Sale = 20

		updatedItems, err := s.storage.UpsertItems(s.ctx, s.storage.DB(), *items)
		s.Require().NoError(err, "Upsert should not fail on update")
		s.Require().NotNil(updatedItems, "Updated items should not be nil")
		s.Require().Len(*updatedItems, 2, "Should maintain two items")

		s.Require().Equal(19.99, (*updatedItems)[0].Price, "Item price should be updated")
		s.Require().Equal(17.99, (*updatedItems)[0].TotalPrice, "Total price should be updated")
		s.Require().Equal(20, (*updatedItems)[0].Sale, "Sale percentage should be updated")
	})
}

func (s *StorageSuite) TestGetOrder() {
	order := &models.Order{
		OrderUID:          "testUID123",
		TrackNumber:       "TN1234567890",
		CustomerID:        "Cust123",
		DateCreated:       time.Now(),
		DeliveryService:   "TestService",
		InternalSignature: "Signature123",
		Locale:            "en",
		Shardkey:          "Shard9",
		SMID:              99,
		OOFShard:          "OOFShard1",
	}

	_, err := s.storage.UpsertOrder(s.ctx, s.storage.DB(), order)
	s.Require().NoError(err, "Insertion for test setup should not fail")

	s.Run("Retrieve existing order", func() {
		retrievedOrder, err := s.storage.GetOrder(s.ctx, s.storage.DB(), order.OrderUID)
		s.Require().NoError(err)
		s.Require().NotNil(retrievedOrder)
		s.Require().Equal(order.OrderUID, retrievedOrder.OrderUID)
		s.Require().Equal(order.TrackNumber, retrievedOrder.TrackNumber)
		s.Require().Equal(order.CustomerID, retrievedOrder.CustomerID)
	})

	s.Run("Non-existent order", func() {
		_, err := s.storage.GetOrder(s.ctx, s.storage.DB(), "nonExistentUID123")
		s.Require().Error(err, "Should return an error")
		s.Require().True(errors.Is(err, models.ErrOrderNotFound), "The error should be 'ErrOrderNotFound'")
	})
}

func (s *StorageSuite) TestGetDelivery() {
	delivery := &models.Delivery{
		OrderUID: "testUID123",
		Name:     "John Doe",
		Phone:    "+1234567890",
		Zip:      "12345",
		City:     "TestCity",
		Address:  "123 Test St",
		Region:   "TestRegion",
		Email:    "john.doe@example.com",
	}

	order := &models.Order{
		OrderUID:          delivery.OrderUID,
		TrackNumber:       "TN1234567890",
		CustomerID:        "Cust123",
		DateCreated:       time.Now(),
		DeliveryService:   "TestService",
		InternalSignature: "Signature123",
		Locale:            "en",
		Shardkey:          "Shard9",
		SMID:              99,
		OOFShard:          "OOFShard1",
	}

	_, err := s.storage.UpsertOrder(s.ctx, s.storage.DB(), order)
	s.Require().NoError(err, "TestGetDelivery models.Order insertion")

	_, err = s.storage.UpsertDelivery(s.ctx, s.storage.DB(), delivery)
	s.Require().NoError(err, "TestGetDelivery models.Delivery insertion")

	delivery.Phone = "+0987654321"
	_, err = s.storage.UpsertDelivery(s.ctx, s.storage.DB(), delivery)
	s.Require().NoError(err, "TestGetDelivery models.Delivery update")

	s.Run("Retrieve existing delivery", func() {
		retrievedDelivery, err := s.storage.GetDelivery(s.ctx, s.storage.DB(), delivery.OrderUID)
		s.Require().NoError(err)
		s.Require().NotNil(retrievedDelivery)
		s.Require().Equal(delivery.Name, retrievedDelivery.Name, "Names should match")
		s.Require().Equal(delivery.Phone, retrievedDelivery.Phone, "Phone numbers should match")
	})
}

func (s *StorageSuite) TestGetPayment() {
	payment := &models.Payment{
		OrderUID:     "testUID123",
		Transaction:  "TX1234567890",
		RequestID:    "RQ1234567890",
		Currency:     "USD",
		Provider:     "TestProvider",
		Amount:       100.50,
		PaymentDT:    time.Now(),
		Bank:         "TestBank",
		DeliveryCost: 5.00,
		GoodsTotal:   95.50,
		CustomFee:    0.00,
	}

	order := &models.Order{
		OrderUID:          payment.OrderUID,
		TrackNumber:       "TN1234567890",
		CustomerID:        "Cust123",
		DateCreated:       time.Now(),
		DeliveryService:   "TestService",
		InternalSignature: "Signature123",
		Locale:            "en",
		Shardkey:          "Shard9",
		SMID:              99,
		OOFShard:          "OOFShard1",
	}

	_, err := s.storage.UpsertOrder(s.ctx, s.storage.DB(), order)
	s.Require().NoError(err, "Insertion for test setup should not fail for order")

	_, err = s.storage.UpsertPayment(s.ctx, s.storage.DB(), payment)
	s.Require().NoError(err, "Insertion for test setup should not fail for payment")

	s.Run("Retrieve existing payment", func() {
		retrievedPayment, err := s.storage.GetPayment(s.ctx, s.storage.DB(), payment.OrderUID)
		s.Require().NoError(err)
		s.Require().NotNil(retrievedPayment)
		s.Require().Equal(payment.Transaction, retrievedPayment.Transaction, "Transaction IDs should match")
		s.Require().Equal(payment.Amount, retrievedPayment.Amount, "Amounts should match")
	})

	s.Run("Non-existent payment", func() {
		_, err := s.storage.GetPayment(s.ctx, s.storage.DB(), "nonExistentUID123")
		s.Require().Error(err, "Should return an error for a non-existent payment")
		s.Require().True(errors.Is(err, models.ErrPaymentNotFound), "Error should be ErrPaymentNotFound")
	})
}

func (s *StorageSuite) TestGetItems() {
	items := []models.Item{
		{
			OrderUID:    "testUID123",
			ChrtID:      1001,
			TrackNumber: "TN1234567890",
			Price:       299.99,
			RID:         "RID123456",
			Name:        "Widget A",
			Sale:        10,
			Size:        "M",
			TotalPrice:  269.99,
			NMID:        501,
			Brand:       "BrandX",
			Status:      1,
		},
		{
			OrderUID:    "testUID123",
			ChrtID:      1002,
			TrackNumber: "TN0987654321",
			Price:       159.49,
			RID:         "RID654321",
			Name:        "Widget B",
			Sale:        15,
			Size:        "L",
			TotalPrice:  135.57,
			NMID:        502,
			Brand:       "BrandY",
			Status:      1,
		},
	}

	order := &models.Order{
		OrderUID:          "testUID123",
		TrackNumber:       "TN1234567890",
		CustomerID:        "Cust123",
		DateCreated:       time.Now(),
		DeliveryService:   "TestService",
		InternalSignature: "Signature123",
		Locale:            "en",
		Shardkey:          "Shard9",
		SMID:              99,
	}

	_, err := s.storage.UpsertOrder(s.ctx, s.storage.DB(), order)
	s.Require().NoError(err, "Insertion for test setup should not fail")

	_, err = s.storage.UpsertItems(s.ctx, s.storage.DB(), items)
	s.Require().NoError(err, "Insertion for test setup of items should not fail")

	s.Run("Retrieve existing items", func() {
		retrievedItems, err := s.storage.GetItems(s.ctx, s.storage.DB(), "testUID123")
		s.Require().NoError(err)
		s.Require().NotNil(retrievedItems)
		s.Require().Len(retrievedItems, 2, "Should retrieve two items")
	})

	s.Run("Non-existent items", func() {
		_, err := s.storage.GetItems(s.ctx, s.storage.DB(), "nonExistentUID123")
		s.Require().Error(err, "Should return an error for a non-existent items query")
		s.Require().ErrorIs(err, models.ErrItemsNotFound, "Error should be ErrItemsNotFound for non-existent items")
	})
}
