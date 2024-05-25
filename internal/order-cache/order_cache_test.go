package ordercache

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stsolovey/order_tracker/internal/models"
)

func TestOrderCache_Upsert(t *testing.T) {
	log := logrus.New()
	cache := New(log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	order := models.Order{
		OrderUID:    "testUID123",
		TrackNumber: "TN1234567890",
		CustomerID:  "Cust123",
		DateCreated: time.Now(),
	}

	cache.Upsert(ctx, order)
	if got, _ := cache.Get(ctx, order.OrderUID); got == nil {
		t.Errorf("Upsert() failed to add order, want order UID %s, got nothing", order.OrderUID)
	} else {
		if got.OrderUID != order.OrderUID {
			t.Errorf("Upsert() failed, want order UID %s, got %s", order.OrderUID, got.OrderUID)
		}
	}

	order.TrackNumber = "TN0987654321"
	cache.Upsert(ctx, order)
	if got, err := cache.Get(ctx, order.OrderUID); err != nil {
		t.Errorf("Upsert() failed to update order, want track number %s, got error: %v", order.TrackNumber, err)
	} else {
		if got.TrackNumber != order.TrackNumber {
			t.Errorf("Upsert() failed to update order, want track number %s, got %s",
				order.TrackNumber, got.TrackNumber)
		}
	}
}

func TestOrderCache_Delete(t *testing.T) {
	log := logrus.New()
	cache := New(log)
	orderUID := "testUID123"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cache.Upsert(ctx, models.Order{OrderUID: orderUID, CustomerID: "Cust123", DateCreated: time.Now()})

	cache.Delete(ctx, orderUID)
	if got, _ := cache.Get(ctx, orderUID); got.OrderUID == orderUID {
		t.Errorf("Delete() failed, order UID %s should have been deleted but was found", orderUID)
	}
}

func TestOrderCache_Get(t *testing.T) {
	log := logrus.New()
	cache := New(log)
	orderUID := "testUID456"
	order := models.Order{
		OrderUID:    orderUID,
		TrackNumber: "TN1234567890",
		CustomerID:  "Cust456",
		DateCreated: time.Now(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cache.Upsert(ctx, order)

	if got, _ := cache.Get(ctx, orderUID); got == nil {
		t.Errorf("Get() failed, order UID %s should have been retrieved but was not found", orderUID)
	} else {
		if got.OrderUID != order.OrderUID {
			t.Errorf("Get() failed, want order UID %s, got %s", order.OrderUID, got.OrderUID)
		}
	}
}
