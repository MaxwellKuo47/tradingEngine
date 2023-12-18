# Only for testing purposes
# This variable should be handled in more secure way
TRADING_DB_DSN=postgres://trading:pa55w0rd@localhost/trading?sslmode=disable

# ==================================================================================== #
# DOCKER POSTGRESQL
# ==================================================================================== #
.PHONY: docker/pull/postgres
docker/pull/postgres:
	docker pull postgres:13-alpine

.PHONY: docker/create/container/postgresql
docker/create/container/postgresql:
	docker run --name postgres13-tradingEngine -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:13-alpine
	
.PHONY: docker/create/db
docker/create/db:
	docker exec -it postgres13-tradingEngine createdb --username=root --owner=root trading 

.PHONY: docker/create/dbExtension
docker/create/dbExtension: 
	docker exec -it postgres13-tradingEngine psql -U root -d trading -c "CREATE EXTENSION IF NOT EXISTS citext;"

.PHONY: docker/create/db/user
docker/create/db/user:
	docker exec -it postgres13-tradingEngine psql -U root -d trading -c "CREATE ROLE trading WITH LOGIN PASSWORD 'pa55w0rd';"

# ==================================================================================== #
# DOCKER REDIS
# ==================================================================================== #
.PHONY: docker/pull/redis
docker/pull/redis:
	docker pull redis:7-alpine

.PHONY: docker/create/container/redis
docker/create/container/redis:
	docker run --name redis-tradingEngine -p 6379:6379 -d redis:7-alpine


# ==================================================================================== #
# DB MIGRATION
# ==================================================================================== #

.PHONY: db/migrate/genFile
db/migrate/genFile:
	migrate create -seq -ext=.sql -dir=./migrations $(name)

.PHONY: db/migrate/checkVer
db/migrate/checkVer:
	migrate -path=./migrations -database=$(TRADING_DB_DSN) version

.PHONY: db/migrate/up
db/migrate/up:
	migrate -path=./migrations -database=$(TRADING_DB_DSN) up

.PHONY: db/migrate/down
db/migrate/down:
	migrate -path=./migrations -database=$(TRADING_DB_DSN) down

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## audit: tidy and vendor dependencies and format, vet and test all code
.PHONY: audit
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

## vendor: tidy and vendor dependencies
.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor

# ==================================================================================== #
# BUILD
# ==================================================================================== #

## build/api: build the cmd/api application
.PHONY: build/api
build/api: build/api-linux build/api-mac build/api-win

build/api-linux:
	@echo 'Building cmd/api for Linux...'
	GOOS=linux GOARCH=amd64 go build -o=./bin/linux_amd64/api ./cmd/api

build/api-mac:
	@echo 'Building cmd/api for macOS...'
	GOOS=darwin GOARCH=amd64 go build -o=./bin/darwin_amd64/api ./cmd/api

build/api-win:
	@echo 'Building cmd/api for Windows...'
	GOOS=windows GOARCH=amd64 go build -o=./bin/windows_amd64/api.exe ./cmd/api
