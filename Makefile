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