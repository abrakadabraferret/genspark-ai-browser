package main

// genspark: entry point package for the mini demo

import (
	"log"
	"genspark-mini/internal/server"
)

func main() {
	// genspark: start the HTTP server
	if err := server.Serve(":8787"); err != nil {
		log.Fatal(err)
	}
}
