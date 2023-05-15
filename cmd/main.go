package main

import (
	"log"
	"math/rand"
	"time"
)

func main() {

	rand.Seed(time.Now().UnixNano())

	server, err := newServer()
	if err != nil {
		log.Fatalf("can not create server: %s", err)
	}
	log.Fatal(server.ListenAndServe())

}
