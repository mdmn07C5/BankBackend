postgres:
	docker run --name postgres16 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:alpine

createdb:
	docker exec -it postgres16 createdb --username=root --owner=root bank

dropdb:
	docker exec -it postgres16 dropdb bank

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/bank?sslmode=disable" -verbose down

sqlc:
	sqlc generate

gotest:
	go test -timeout 30s github.com/mdmn07C5/bank/db/sqlc -run ^TestMain$

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/mdmn07C5/bank/db/sqlc Store

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test server mock