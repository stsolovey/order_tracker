# Пути
CMD_SERVER_PATH=./cmd/order_service/
BIN_PATH=./bin/
SERVER_EXECUTABLE=order_service

# Запуск окружения и сервиса
up: up-deps run_server

# Запуск сервера
run_server:
	go run $(CMD_SERVER_PATH)main.go

# Запуск окружения (небольшая пауза - иначе падение на первом старте)
up-deps:
	docker compose --env-file ./.env -f ./deploy/local/docker-compose.yml up -d
	sleep 3

# Запуск окружения с (надёжной) паузой
up-deps-ci:
	docker compose --env-file ./.env -f ./deploy/local/docker-compose.yml up -d
	sleep 10

# Остановка окружения
down-deps:
	docker compose --env-file ./.env -f ./deploy/local/docker-compose.yml down

# Компиляция проекта
build:
	mkdir -p $(BIN_PATH)
	go build -o $(BIN_PATH)$(SERVER_EXECUTABLE) $(CMD_SERVER_PATH)main.go

# Запуск приложения в терминале
run-app: build
	$(BIN_PATH)$(SERVER_EXECUTABLE)

# Запуск приложения в фоне (сохраняем pid чтобы кикнуть позже)
run-app-background: build
	$(BIN_PATH)$(SERVER_EXECUTABLE) & echo $$! > $(BIN_PATH)PID

# Остановка приложения по pid
stop-app:
	if [ -f $(BIN_PATH)PID ]; then \
		kill `cat $(BIN_PATH)PID` || true; \
		rm $(BIN_PATH)PID; \
	fi

# View output (компоуза)
logs:
	docker compose --env-file ./.env -f ./deploy/local/docker-compose.yml logs

# Остановка всего: приложения и зависимостей
down: stop-app down-deps

# Remove volumes
remove-volumes:
	docker compose --env-file ./.env -f ./deploy/local/docker-compose.yml down -v --remove-orphans


# Тестирование: старт окружения и приложения, тест, стоп
test: up-deps run-app-background
	sleep 1 # Даём приложению время для запуска
	go test ./... -count=1; result=$$?; \
	make stop-app; \
	make down-deps; \
	exit $$result

test-ci: up-deps-ci run-app-background
	sleep 1 # Даём приложению время для запуска
	go test ./... -count=1; result=$$?; \
	make stop-app; \
	make down-deps; \
	exit $$result

testv: up-deps run-app-background
	sleep 1 # Даём приложению время для запуска
	go test ./... -count=1 -v; result=$$?; \
	make stop-app; \
	make down-deps; \
	exit $$result

itest: up-deps run-app-background
	sleep 1 # Allow time for the application to start
	go test ./tests/... -count=1; result=$$?; \
	make stop-app; \
	make down-deps; \
	exit $$result

itestv: up-deps run-app-background
	sleep 1 # Allow time for the application to start
	go test ./tests/... -count=1 -v; result=$$?; \
	make stop-app; \
	make down-deps; \
	exit $$result

tidy:
	gofumpt -w .
	gci write . --skip-generated -s standard -s default
	go mod tidy

lint: tidy
	golangci-lint run ./...

tools:
	go install mvdan.cc/gofumpt@latest
	go install github.com/daixiang0/gci@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

help:
	@echo "Available commands:"
	@echo "  up                   - Start dependencies and the service"
	@echo "  run_server           - Start the server using 'go run'"
	@echo "  up-deps              - Start environment using Docker Compose"
	@echo "  down-deps            - Stop environment using Docker Compose"
	@echo "  build                - Compile the project"
	@echo "  logs                 - View Docker Compose logs"
	@echo "  run-app              - Run the compiled application"
	@echo "  run-app-background   - Run the compiled application in the background"
	@echo "  stop-app             - Stop the application using its PID"
	@echo "  down                 - Stop application and dependencies"
	@echo "  remove-volumes       - Stop environment (if it's UP) and remove volumes in Docker Compose"
	@echo "  test                 - Start environment, run tests, and clean up"
	@echo "  testv                - Start environment, run tests verbosely, and clean up"
	@echo "  itest                - Start environment, run integration tests, and clean up"
	@echo "  itestv               - Start environment, run integration tests verbosely, and clean up"
	@echo "  tidy                 - Format and tidy up the Go code"
	@echo "  lint                 - Lint and format the project code"
	@echo "  tools                - Install necessary tools"
