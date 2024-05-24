# WB Tech Order Tracker Service

## Overview

This project implements a Golang-based service for displaying order details from a JSON-based data model. The service utilizes PostgreSQL for data storage, in-memory caching for quick data retrieval, and a basic HTTP server for serving the order data.

### Features
- PostgreSQL setup and data storage
- Subscription to updates via NATS Streaming
- In-memory caching of order data
- Automatic cache recovery from PostgreSQL on service restart
- Simple web interface to display order data by ID
- Script for publishing data to NATS for testing subscription
- Stress testing with WRK and Vegeta tools

## Quick Start

### Requirements
- Golang 1.22+
- PostgreSQL
- NATS Streaming Server

### Run project
Run the tracker with:
```bash
make up
```

### Testing
Run the tests with:
```bash
make test
```