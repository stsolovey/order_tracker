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
	// upsert get getall (database)
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

func (s *Service) Init(_ context.Context) error {
	// t0d0 init (db->cache)
	return nil
}

func (s *Service) UpsertOrder(ctx context.Context, order models.Order) error {
	if err := s.cache.Upsert(ctx, order); err != nil {
		return fmt.Errorf("Service UpsertOrder; %w", err)
	}

	return nil
}

func (s *Service) GetOrder(ctx context.Context, orderID string) (*models.Order, error) {
	order, err := s.cache.Get(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf(".. %w", err)
	}

	return order, nil
}
