package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/mdmn07C5/bank/api"
	db "github.com/mdmn07C5/bank/db/sqlc"
	"github.com/mdmn07C5/bank/util"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("unable to load configuration:", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	store := db.NewStore(conn)
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
