package integrationtest

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/stsolovey/order_tracker/internal/models"
)

func (s *IntegrationTestSuite) TestNATSIntegration() {
	order := models.Order{
		OrderUID:    "uniqueOrderID123",
		TrackNumber: "TN1234567890",
		Entry:       "WBIL",
		Delivery: models.Delivery{
			OrderUID: "uniqueOrderID123",
			Name:     "John Doe",
			Phone:    "+1234567890",
			City:     "TestCity",
			Address:  "123 Test St",
		},
		Payment: models.Payment{
			OrderUID:    "uniqueOrderID123",
			Transaction: "TX1234567890",
			Currency:    "USD",
			Provider:    "TestProvider",
			Amount:      150.00,
			PaymentDT:   time.Now(),
		},
		Items: []models.Item{
			{
				OrderUID:    "uniqueOrderID123",
				ChrtID:      1,
				TrackNumber: "TN1234567890",
				Price:       100.00,
				Name:        "Test Item 1",
				NMID:        1001,
				Brand:       "TestBrand",
				Status:      1,
			},
			{
				OrderUID:    "uniqueOrderID123",
				ChrtID:      2,
				TrackNumber: "TN1234567890",
				Price:       50.00,
				Name:        "Test Item 2",
				NMID:        2002,
				Brand:       "BrandTest",
				Status:      2,
			},
		},
		Locale:          "en",
		CustomerID:      "Cust123",
		DeliveryService: "TestService",
		DateCreated:     time.Now(),
	}

	s.Run("Successful natsClient Publishing", func() {
		err := s.natsClient.PublishOrder(order)
		time.Sleep(2 * time.Second)
		s.Require().NoError(err, "should publish without error")
	})

	time.Sleep(2 * time.Second)

	s.Run("Successful from database direct retrieving", func() {
		err := s.natsClient.PublishOrder(order)
		s.Require().NoError(err, "should publish without error")

		var fetchedOrder *models.Order
		timeout := time.After(20 * time.Second)
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()

		found := false
		for !found {
			select {
			case <-timeout:
				s.Require().Fail("timeout waiting for the order to be stored")
				return
			case <-ticker.C:
				fetchedOrder, err = s.db.Get(context.Background(), order.OrderUID)
				if err == nil && fetchedOrder != nil {
					found = true
				}
			}
		}

		s.Require().NoError(err, "should fetch order from the database without error")
		s.Require().NotNil(fetchedOrder, "fetched order should not be nil")
	})

	s.Run("Successful cache retrieving", func() {
		err := s.natsClient.PublishOrder(order)
		s.Require().NoError(err, "should publish without error")

		var cachedOrder *models.Order
		timeout := time.After(30 * time.Second)
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()

		found := false
		for !found {
			select {
			case <-timeout:
				s.Require().Fail("timeout waiting for the order to be stored")
				return
			case <-ticker.C:
				cachedOrder, err = s.orderCache.Get(context.Background(), order.OrderUID)
				if err == nil && cachedOrder != nil {
					found = true
				}
			}
		}

		s.Require().NoError(err, "should fetch order from the cache without error")
		s.Require().NotNil(cachedOrder, "cached order should not be nil")
	})

	s.Run("Successful service layer retrieving", func() {
		err := s.natsClient.PublishOrder(order)
		s.Require().NoError(err, "should publish without error")

		var retrievedOrder *models.Order
		timeout := time.After(20 * time.Second)
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()

		found := false
		for !found {
			select {
			case <-timeout:
				s.Require().Fail("timeout waiting for the order to be stored")
				return
			case <-ticker.C:
				retrievedOrder, err = s.app.GetOrder(context.Background(), order.OrderUID)
				if err == nil && retrievedOrder != nil {
					found = true
				}
			}
		}

		s.Require().NoError(err, "should GetOrder(...) work without error")
		s.Require().NotNil(retrievedOrder, "retrieved with GetOrder(...) should not be nil")
		s.Require().Equal(order.OrderUID, retrievedOrder.OrderUID, "sent and got OrderUID should match")
		s.Require().Equal(order.TrackNumber, retrievedOrder.TrackNumber, "sent and got TrackNumber should match")
		s.Require().Equal(order.Delivery.Name, retrievedOrder.Delivery.Name, "sent and got DeliveryName should match")
		s.Require().Equal(order.Payment.Transaction, retrievedOrder.Payment.Transaction, "payment transaction should match")
		s.Require().Equal(len(order.Items), len(retrievedOrder.Items), "sent and got number of Items should match")
	})

	s.Run("Successful httpServer retrieving", func() {
		err := s.natsClient.PublishOrder(order)
		s.Require().NoError(err, "should publish without error")

		var resp *http.Response
		timeout := time.After(30 * time.Second)
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()

		url := "http://localhost:" + s.cfg.AppPort + "/api/v1/orders/" + order.OrderUID

		found := false
		for !found {
			select {
			case <-timeout:
				s.Require().Fail("timeout waiting for the response to be get")
				return
			case <-ticker.C:
				resp, err = http.Get(url)
				if err == nil && resp != nil {
					found = true
				}
			}
		}

		s.Require().NoError(err, "should fetch order through HTTP server without error")
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				s.log.Info("failed to close response body")
			}
		}(resp.Body)

		s.Require().Equal(http.StatusOK, resp.StatusCode, "HTTP status should be 200 OK")

		var httpOrder models.Order
		err = json.NewDecoder(resp.Body).Decode(&httpOrder)

		s.Require().NoError(err, "should decode HTTP response without error")
		s.Require().Equal(order.OrderUID, httpOrder.OrderUID, "sent and got OrderUID should match")
		s.Require().Equal(order.TrackNumber, httpOrder.TrackNumber, "sent and got TrackNumber should match")
		s.Require().Equal(order.Delivery.Name, httpOrder.Delivery.Name, "sent and got DeliveryName should match")
		s.Require().Equal(order.Payment.Transaction, httpOrder.Payment.Transaction, "payment transaction should match")
		s.Require().Equal(len(order.Items), len(httpOrder.Items), "sent and got number of Items should match")
	})
}
