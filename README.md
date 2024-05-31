# WB Tech Order Tracker Service

## Overview

The Order Tracker Service is a Golang-based service designed to track and display order details using a JSON-based data model. The service leverages PostgreSQL for persistent storage, NATS JetStream for message streaming, and an in-memory cache for fast data retrieval. It also features a basic HTTP server to serve order data via a REST API.

### Features
- **PostgreSQL Setup and Data Storage**: Persistent storage of order data.
- **NATS JetStream Integration**: Subscription to updates via NATS JetStream.
- **In-Memory Caching**: Fast access to order data with automatic cache recovery on service restart.
- **HTTP Server**: API for retrieving order data by ID.
- **Publisher Script**: Script for publishing data to NATS for testing subscription.
- **Automated Tests**: Unit and integration tests to ensure service reliability.
- **Stress Testing**: WRK and Vegeta tools for performance testing.

## Project Structure

The project is organized as follows:

```
order_tracker/
├── api/                    # API definition in Swagger format
├── cmd/                    # Main application entry point
│   └── order_service/
│       └── main.go
├── deploy/                 # Deployment configurations
│   ├── deployments/
│   │   └── Dockerfile
│   └── local/
│       └── docker-compose.yml
├── internal/               # Core internal logic
│   ├── config/             # Configuration loading
│   ├── logger/             # Logging setup
│   ├── models/             # Data models
│   ├── nats-client/        # NATS client setup
│   ├── order-cache/        # In-memory cache
│   ├── server/             # HTTP server setup
│   ├── service/            # Business logic
│   └── storage/            # Database interactions
├── tests/                  # Integration tests
├── .env                    # Environment variables
├── .env.example            # Example environment variables
├── .gitignore              # Git ignore file
├── .golangci.yml           # GolangCI-Lint configuration
├── Makefile                # Makefile for running tasks
├── README.md               # Project documentation
├── go.mod                  # Go module dependencies
└── go.sum                  # Go module checksums
```

## Components and How They Work

### 1. Configuration
- **`internal/config/config.go`**: Loads environment variables from the `.env` file to configure the service.

### 2. Logging
- **`internal/logger/logger.go`**: Sets up structured logging using `logrus`.

### 3. Models
- **`internal/models/order.go`**: Defines data structures for orders, deliveries, payments, and items.

### 4. NATS Client
- **`internal/nats-client/client.go`**: Manages the connection to NATS JetStream, subscribes to subjects, and processes messages.

### 5. In-Memory Cache
- **`internal/order-cache/order_cache.go`**: Provides an in-memory cache for orders, with methods to get, upsert, and delete orders.

### 6. Service Layer
- **`internal/service/service.go`**: Contains the core business logic for handling orders, including initialization, upsertion, and retrieval.

### 7. HTTP Server
- **`internal/server/server.go`**: Sets up an HTTP server using the `chi` router to handle API requests for order data.

### 8. Storage Layer
- **`internal/storage/`**:
    - **`storage.go`**: Manages database interactions and migrations.
    - **`storage_get.go`**: Retrieves individual orders.
    - **`storage_get_all.go`**: Retrieves all orders.
    - **`storage_upsert.go`**: Upserts orders, deliveries, payments, and items into the database.

### 9. Testing
- **`tests/`**: Contains integration tests to ensure the service components work together as expected.

### 10. Deployment
- **`deploy/`**:
    - **`Dockerfile`**: Docker configuration for building the service image.
    - **`docker-compose.yml`**: Docker Compose configuration for setting up PostgreSQL, NATS, and the service.

## Quick Start

### Requirements
- Golang 1.22+
- PostgreSQL
- NATS JetStream Server

### Setup and Run

1. **Clone the Repository**
   ```bash
   git clone https://github.com/stsolovey/order_tracker.git
   cd order_tracker
   ```

2. **Set Up Environment Variables**
Copy the example `.env.example` to `.env` and fill in the required values.

3. **Build and Run the Service**
Start the service along with PostgreSQL and NATS:
      ```bash
      make up
      ```

4. **Run Tests**
Execute the tests to verify the service:
      ```bash
      make test
      ```

### Testing the Service
Script in the `cmd/order_service` directory for publication sample order data to the NATS subject for testing.
```bash
make gen
```

### Stress Testing

WRK and Vegeta perform stress testing and evaluate the performance of the service.

```bash
make stress-wrk
make stress-vegeta
```

## Contact Information
Feel free to reach out via Telegram: [@duckever](https://t.me/duckever).

