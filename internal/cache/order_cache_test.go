package cache

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stsolovey/order_tracker/internal/models"
)

func TestOrderCache_Upsert(t *testing.T) {
	log := logrus.New()
	cache := NewOrderCache()

	order := models.Order{
		OrderUID:    "testUID123",
		TrackNumber: "TN1234567890",
		CustomerID:  "Cust123",
		DateCreated: time.Now(),
	}

	cache.Upsert(log, order)
	if got, found := cache.Get(order.OrderUID); !found {
		t.Errorf("Upsert() failed to add order, want order UID %s, got nothing", order.OrderUID)
	} else {
		if got.OrderUID != order.OrderUID {
			t.Errorf("Upsert() failed, want order UID %s, got %s", order.OrderUID, got.OrderUID)
		}
	}

	order.TrackNumber = "TN0987654321"
	cache.Upsert(log, order)
	if got, found := cache.Get(order.OrderUID); !found {
		t.Errorf("Upsert() failed to update order, want track number %s, got nothing", order.TrackNumber)
	} else {
		if got.TrackNumber != order.TrackNumber {
			t.Errorf("Upsert() failed to update order, want track number %s, got %s", order.TrackNumber, got.TrackNumber)
		}
	}
}

func TestOrderCache_Delete(t *testing.T) {
	log := logrus.New()
	cache := NewOrderCache()
	orderUID := "testUID123"

	cache.Upsert(log, models.Order{OrderUID: orderUID, CustomerID: "Cust123", DateCreated: time.Now()})

	cache.Delete(log, orderUID)
	if _, found := cache.Get(orderUID); found {
		t.Errorf("Delete() failed, order UID %s should have been deleted but was found", orderUID)
	}
}

func TestOrderCache_Get(t *testing.T) {
	log := logrus.New()
	cache := NewOrderCache()
	orderUID := "testUID456"
	order := models.Order{
		OrderUID:    orderUID,
		TrackNumber: "TN1234567890",
		CustomerID:  "Cust456",
		DateCreated: time.Now(),
	}

	cache.Upsert(log, order)

	if got, found := cache.Get(orderUID); !found {
		t.Errorf("Get() failed, order UID %s should have been retrieved but was not found", orderUID)
	} else {
		if got.OrderUID != order.OrderUID {
			t.Errorf("Get() failed, want order UID %s, got %s", order.OrderUID, got.OrderUID)
		}
	}
}
