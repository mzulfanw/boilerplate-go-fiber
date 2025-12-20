MIGRATIONS_DIR ?= ./migrations
MIGRATIONS_PATH ?= $(abspath $(MIGRATIONS_DIR))
MIGRATIONS_SOURCE ?= file://$(MIGRATIONS_PATH)
MIGRATE ?= migrate
DB_DSN ?= $(POSTGRES_DSN)
SWAG ?= swag
SWAGGER_OUT ?= ./docs
SWAGGER_DIRS ?= ./internal/transport/http

.PHONY: run build clean test fmt vet lint swagger check-db-dsn check-migrations-dir migrate-create migrate-up migrate-down migrate-force migrate-version

run:
	go mod tidy && go run ./cmd/main.go

build:
	go build -o bin/app ./cmd/main.go

clean:
	rm -rf bin

test:
	go test ./...

fmt:
	gofmt -w $(shell find . -name '*.go' -not -path './vendor/*')

vet:
	go vet ./...

lint:
	golangci-lint run

swagger:
	rm -rf $(SWAGGER_OUT)
	$(SWAG) init \
		-g cmd/main.go \
		-o $(SWAGGER_OUT) \
		--parseInternal \
		--parseDependency


check-db-dsn:
	@if [ -z "$(DB_DSN)" ]; then echo "DB_DSN is required (set POSTGRES_DSN or DB_DSN)"; exit 1; fi

check-migrations-dir:
	@if [ ! -d "$(MIGRATIONS_PATH)" ]; then echo "MIGRATIONS_PATH not found: $(MIGRATIONS_PATH)"; exit 1; fi

migrate-create: check-migrations-dir
	@if [ -z "$(name)" ]; then echo "name is required (e.g. make migrate-create name=create_users)"; exit 1; fi
	$(MIGRATE) create -ext sql -dir "$(MIGRATIONS_PATH)" -seq $(name)

migrate-up: check-db-dsn check-migrations-dir
	@if [ -n "$(n)" ]; then \
		$(MIGRATE) -database "$(DB_DSN)" -source "$(MIGRATIONS_SOURCE)" up $(n); \
	else \
		$(MIGRATE) -database "$(DB_DSN)" -source "$(MIGRATIONS_SOURCE)" up; \
	fi

migrate-down: check-db-dsn check-migrations-dir
	@if [ -n "$(n)" ]; then \
		$(MIGRATE) -database "$(DB_DSN)" -source "$(MIGRATIONS_SOURCE)" down $(n); \
	else \
		$(MIGRATE) -database "$(DB_DSN)" -source "$(MIGRATIONS_SOURCE)" down 1; \
	fi

migrate-force: check-db-dsn check-migrations-dir
	@if [ -z "$(version)" ]; then echo "version is required (e.g. make migrate-force version=2)"; exit 1; fi
	$(MIGRATE) -database "$(DB_DSN)" -source "$(MIGRATIONS_SOURCE)" force $(version)

migrate-version: check-db-dsn check-migrations-dir
	$(MIGRATE) -database "$(DB_DSN)" -source "$(MIGRATIONS_SOURCE)" version
