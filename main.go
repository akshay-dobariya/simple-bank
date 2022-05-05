package main

import (
	"database/sql"
	"log"

	"github.com/akshay-dobariya/simple-bank/api"
	db "github.com/akshay-dobariya/simple-bank/db/sqlc"
	"github.com/akshay-dobariya/simple-bank/util"
	_ "github.com/lib/pq"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load configurations:", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatalf("Cannot connect to the DB error: '%s'", err)
	}
	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("Cannot start server:", err)
	}
}
