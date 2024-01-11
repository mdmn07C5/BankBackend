postgres:
	docker run --name postgres16 --network bank-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:alpine

createdb:
	docker exec -it postgres16 createdb --username=root --owner=root bank

dropdb:
	docker exec -it postgres16 dropdb bank

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/bank?sslmode=disable" -verbose up

migrateup1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/bank?sslmode=disable" -verbose up 1

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/bank?sslmode=disable" -verbose down

migratedown1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/bank?sslmode=disable" -verbose down 1

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

testAPI:
	go test -v github.com/mdmn07C5/bank/api

proto:
	rm -f pb/*go
	rm -f doc/swagger/*.swagger.json
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
	--openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=bank_rpc_doc \
    proto/*.proto

evans: 
	evans --host localhost --port 9090 -r repl

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test server mock testAPI migrateup1 migratedown1 proto evans