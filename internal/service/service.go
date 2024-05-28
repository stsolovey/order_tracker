package service

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/stsolovey/order_tracker/internal/models"
)

type cache interface {
	Upsert(ctx context.Context, order models.Order) error
	Get(ctx context.Context, orderUID string) (*models.Order, error)
}

type storage interface {
	Get(ctx context.Context, orderUID string) (*models.Order, error)
	GetAll(ctx context.Context) ([]models.Order, error)
	Upsert(ctx context.Context, order *models.Order) (*models.Order, error)
}

type Service struct {
	log     *logrus.Logger
	cache   cache
	storage storage
}

func New(log *logrus.Logger, cache cache, storage storage) *Service {
	return &Service{
		log:     log,
		cache:   cache,
		storage: storage,
	}
}

func (s *Service) Init(ctx context.Context) error {
	orders, err := s.storage.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("service.go Init(...): %w", err)
	}

	for _, order := range orders {
		if err := s.cache.Upsert(ctx, order); err != nil {
			s.log.WithError(err).Errorf("service.go Init(...) Upsert(%s)", order.OrderUID)
		}
	}

	s.log.Infof("Initialized cache with %d orders", len(orders))

	return nil
}

func (s *Service) UpsertOrder(ctx context.Context, order models.Order) error {
	if err := s.cache.Upsert(ctx, order); err != nil {
		return fmt.Errorf("service.go UpsertOrder s.cache.Upsert(...): %w", err)
	}

	if _, err := s.storage.Upsert(ctx, &order); err != nil {
		return fmt.Errorf("service.go UpsertOrder s.storage.Upsert(...): %w", err)
	}

	return nil
}

func (s *Service) GetOrder(ctx context.Context, orderID string) (*models.Order, error) {
	order, err := s.cache.Get(ctx, orderID)
	if err != nil {
		s.log.WithError(err).Warnf("Order %s not found in cache, fetching from storage", orderID)

		order, err = s.storage.Get(ctx, orderID)
		if err != nil {
			return nil, fmt.Errorf("service.go GetOrder s.storage.Get(...): %w", err)
		}

		if err := s.cache.Upsert(ctx, *order); err != nil {
			s.log.WithError(err).Errorf("service.go GetOrder s.cache.Upsert(%s)", orderID)
		}
	}

	return order, nil
}
