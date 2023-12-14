# this variable should be handled in more secure way
TRADING_DB_DSN=postgres://trading:pa55w0rd@localhost/trading?sslmode=disable

.PHONY: docker/pull
docker/pull:
	docker pull postgres:13-alpine

.PHONY: docker/create/container
docker/create/container:
	docker run --name postgres13-tradingEngine -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:13-alpine
	
.PHONY: docker/create/db
docker/create/db:
	docker exec -it postgres13-tradingEngine createdb --username=root --owner=root trading 

.PHONY: docker/create/db/user
docker/create/db/user:
	docker exec -it postgres13-tradingEngine psql -U root -d trading -c "CREATE ROLE trading WITH LOGIN PASSWORD 'pa55w0rd';"

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