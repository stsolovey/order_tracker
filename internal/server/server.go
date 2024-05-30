package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"github.com/stsolovey/order_tracker/internal/config"
	"github.com/stsolovey/order_tracker/internal/models"
	"github.com/stsolovey/order_tracker/internal/service"
)

const (
	readHeaderTimeoutDuration = 10 * time.Second
	readTimeoutDuration       = 15 * time.Second
	writeTimeoutDuration      = 15 * time.Second
	idleTimeoutDuration       = 60 * time.Second

	shutdownTimeoutDuration = 5 * time.Second
)

type Server struct {
	config *config.Config
	logger *logrus.Logger
	server *http.Server
}

func CreateServer(
	cfg *config.Config,
	log *logrus.Logger,
	orderService service.OrderServiceInterface,
) *Server {
	r := chi.NewRouter()

	ConfigureRoutes(r, orderService, log)

	s := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           r,
		ReadHeaderTimeout: readHeaderTimeoutDuration,
		ReadTimeout:       readTimeoutDuration,
		WriteTimeout:      writeTimeoutDuration,
		IdleTimeout:       idleTimeoutDuration,
	}

	return &Server{
		config: cfg,
		logger: log,
		server: s,
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("Starting HTTP server...")

	go func() {
		<-ctx.Done()
		s.logger.Info("HTTP server shutdown initiated.")

		ctxShutdown, cancel := context.WithTimeout(context.Background(), shutdownTimeoutDuration)
		defer cancel()

		if err := s.Shutdown(ctxShutdown); err != nil { //nolint:contextcheck
			s.logger.WithError(err).Error("http server shutdown failed")
		}
	}()

	s.logger.Infof("HTTP server is running on %s", s.server.Addr)

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("HTTP server listen and serve: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server...")

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("HTTP server shutdown failed: %w", err)
	}

	return nil
}

func ConfigureRoutes(r chi.Router, orderService service.OrderServiceInterface, log *logrus.Logger) {
	r.Route("/orders", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "Missing order ID", http.StatusBadRequest)
		})
		r.Get("/{uid}", func(w http.ResponseWriter, req *http.Request) {
			getOrder(w, req, orderService, log)
		})
	})
}

func getOrder(w http.ResponseWriter, r *http.Request, app service.OrderServiceInterface, log *logrus.Logger) {
	orderID := chi.URLParam(r, "uid")
	if orderID == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)

		return
	}

	ctx := r.Context()

	order, err := app.GetOrder(ctx, orderID)
	if err != nil {
		if errors.Is(err, models.ErrOrderNotFound) {
			http.Error(w, "Order not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	response, err := json.Marshal(order)
	if err != nil {
		http.Error(w, "Failed to serialize the order", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(response)
	if err != nil {
		log.Infof("Failed to write response: %s", err)
	}
}
