package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/mdmn07C5/bank/api"
	db "github.com/mdmn07C5/bank/db/sqlc"
)

// need to connect to db and create store first
const (
	dbDriver      = "postgres"
	dbSource      = "postgresql://root:secret@localhost:5432/bank?sslmode=disable"
	serverAddress = "0.0.0.0:8080"
)

func main() {
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start(serverAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
