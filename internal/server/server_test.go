// server_test.go

package server_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/stsolovey/order_tracker/internal/config"
	"github.com/stsolovey/order_tracker/internal/logger"
	"github.com/stsolovey/order_tracker/internal/models"
	"github.com/stsolovey/order_tracker/internal/server"
)

type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) Init(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockOrderService) UpsertOrder(ctx context.Context, order models.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderService) GetOrder(ctx context.Context, orderID string) (*models.Order, error) {
	args := m.Called(ctx, orderID)
	if obj := args.Get(0); obj != nil {
		return obj.(*models.Order), args.Error(1)
	}
	return nil, args.Error(1)
}

type ServerTestSuite struct {
	suite.Suite
	srv      *server.Server
	log      *logrus.Logger
	service  *MockOrderService
	recorder *httptest.ResponseRecorder
	ctx      context.Context
	router   *chi.Mux
}

func (s *ServerTestSuite) SetupTest() {
	cfg := &config.Config{
		AppPort: "8081",
	}
	s.log = logger.New("debug")
	s.service = &MockOrderService{}
	s.srv = server.CreateServer(cfg, s.log, s.service)
	s.recorder = httptest.NewRecorder()
	s.ctx = context.Background()
	s.router = chi.NewRouter()
	server.ConfigureRoutes(s.router, s.service, s.log)
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}

func (s *ServerTestSuite) TestGetOrder_Success() {
	orderUID := "testUID123"
	order := &models.Order{
		OrderUID:    orderUID,
		TrackNumber: "TN1234567890",
		CustomerID:  "Cust123",
		DateCreated: time.Now(),
	}

	s.service.On("GetOrder", mock.Anything, orderUID).Return(order, nil)

	req := httptest.NewRequest(http.MethodGet, "/order/get?id="+orderUID, nil)
	s.router.ServeHTTP(s.recorder, req)

	require.Equal(s.T(), http.StatusOK, s.recorder.Code)
	require.Contains(s.T(), s.recorder.Body.String(), orderUID)
}

func (s *ServerTestSuite) TestGetOrder_NotFound() {
	orderUID := "nonExistentUID"
	s.service.On("GetOrder", mock.Anything, orderUID).Return(nil, models.ErrOrderNotFound)

	req := httptest.NewRequest(http.MethodGet, "/order/get?id="+orderUID, nil)
	s.router.ServeHTTP(s.recorder, req)

	require.Equal(s.T(), http.StatusNotFound, s.recorder.Code)
}

func (s *ServerTestSuite) TestGetOrder_BadRequest() {
	req := httptest.NewRequest(http.MethodGet, "/order/get", nil)
	s.router.ServeHTTP(s.recorder, req)

	require.Equal(s.T(), http.StatusBadRequest, s.recorder.Code)
}

func (s *ServerTestSuite) TestStartServer() {
	orderUID := "testUID123"
	order := &models.Order{
		OrderUID:    orderUID,
		TrackNumber: "TN1234567890",
		CustomerID:  "Cust123",
		DateCreated: time.Now(),
	}

	s.service.On("GetOrder", mock.Anything, orderUID).Return(order, nil)

	go func() {
		err := s.srv.Start(s.ctx)
		require.NoError(s.T(), err)
	}()

	time.Sleep(time.Second)

	req, err := http.NewRequest(http.MethodGet, "http://localhost:8081/order/get?id="+orderUID, nil)
	require.NoError(s.T(), err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	var responseOrder models.Order
	err = json.NewDecoder(resp.Body).Decode(&responseOrder)
	require.NoError(s.T(), err)

	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	require.Equal(s.T(), orderUID, responseOrder.OrderUID)
}
