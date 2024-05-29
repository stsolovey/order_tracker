package ordercache

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/stsolovey/order_tracker/internal/models"
)

type OrderCache struct {
	log *logrus.Logger
	mu  sync.RWMutex

	m map[string]models.Order
}

func New(log *logrus.Logger) *OrderCache {
	return &OrderCache{
		log: log,
		m:   make(map[string]models.Order),
	}
}

func (oc *OrderCache) Get(_ context.Context, orderUID string) (*models.Order, error) {
	oc.mu.RLock()
	defer oc.mu.RUnlock()

	order, found := oc.m[orderUID]
	if !found {
		return nil, models.ErrOrderNotFound
	}

	return &order, nil
}

func (oc *OrderCache) Upsert(_ context.Context, order models.Order) error {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	oc.m[order.OrderUID] = order
	oc.log.Debugf("order_cache.go Upsert(...), order upserted: %s", order.OrderUID)

	return nil
}

func (oc *OrderCache) Delete(_ context.Context, orderUID string) {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	if _, found := oc.m[orderUID]; found {
		delete(oc.m, orderUID)
		oc.log.Debug("Order deleted:", orderUID)
	} else {
		oc.log.Debug("Order not found:", orderUID)
	}
}
