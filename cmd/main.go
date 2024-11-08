package main

import (
	"fmt"
	"hw1/internal/pkg/server"
	"hw1/internal/pkg/storage"
	"log"
	"time"
)

func main() {
	store, err := storage.NewStorage(time.Minute*20, time.Minute*60, "my-storage.json")
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	fmt.Println("Storage created successfully")

	s := server.New(":8090", store)
	fmt.Println("Starting server on :8090...")
	s.Start()
}
