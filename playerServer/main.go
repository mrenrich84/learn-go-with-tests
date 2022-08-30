package main

import (
	"log"
	"net/http"
	"os"
)

const dbName = "game.db.json"

func main() {
	db, err := os.OpenFile(dbName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("problem opening %s %v", dbName, err)
	}

	store, err := NewFileSystemPlayerStore(db)
	if err != nil {
		log.Fatalf("problem creating file system player store, %v", err)
	}

	server := NewPlayerServer(store)
	log.Fatal(http.ListenAndServe(":5000", server))
}