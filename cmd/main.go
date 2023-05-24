package main

import (
	"fmt"
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

	if server.TLSConfig != nil {
		fmt.Println("serving tls")
		log.Fatal(server.ListenAndServeTLS())
	}

	log.Fatal(server.ListenAndServe())

}
