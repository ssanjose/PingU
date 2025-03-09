package main

import (
	"log"

	"github.com/ssanjose/PingU/internal/db"
	"github.com/ssanjose/PingU/internal/env"
	"github.com/ssanjose/PingU/internal/store"
)

func main() {
	addr := env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/pingu?sslmode=disable")
	conn, err := db.New(addr, 3, 3, "15m")
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	store := store.NewStorage(conn)

	db.Seed(store)
}
