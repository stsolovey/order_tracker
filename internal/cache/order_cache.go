package cache

import (
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/stsolovey/order_tracker/internal/models"
)

type OrderCache struct {
	sync.RWMutex
	m map[string]models.Order
}

func NewOrderCache() *OrderCache {
	return &OrderCache{
		m: make(map[string]models.Order),
	}
}

func (oc *OrderCache) Upsert(log *logrus.Logger, order models.Order) {
	oc.Lock()
	defer oc.Unlock()
	oc.m[order.OrderUID] = order
	log.Infoln("Order upserted:", order.OrderUID)
}

func (oc *OrderCache) Delete(log *logrus.Logger, orderUID string) {
	oc.Lock()
	defer oc.Unlock()

	if _, found := oc.m[orderUID]; found {
		delete(oc.m, orderUID)
		log.Infoln("Order deleted:", orderUID)
	} else {
		log.Infoln("Order not found:", orderUID)
	}
}

func (oc *OrderCache) Get(orderUID string) (models.Order, bool) {
	oc.RLock()
	defer oc.RUnlock()
	order, found := oc.m[orderUID]

	return order, found
}
