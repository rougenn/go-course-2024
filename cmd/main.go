package main

import (
	"fmt"
	"hw1/internal/pkg/argsparser"
	"hw1/internal/pkg/server"
	"hw1/internal/pkg/storage"
	"log"
	"time"
)

func main() {
	SD, CD, filename, port := argsparser.ParseArgs()
	store, err := storage.NewStorage(time.Second*time.Duration(SD), time.Second*time.Duration(CD), filename)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}

	fmt.Println("Storage created successfully")

	s := server.New(port, store)
	fmt.Println("Starting server on :8090...")
	s.Start()

	store.Wait()
}
