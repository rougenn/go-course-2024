package main

import (
	"fmt"
	"hw1/internal/pkg/server"
	"hw1/internal/pkg/storage"
	"log"
	"time"
)

func main() {
	s, err := storage.NewStorage(time.Second*20, time.Second*10, "my-storage.json")
	if err != nil {
		log.Fatalf("Error initializing storage: %v", err)
	}

	err = s.LoadFromFile("my-storage.json")
	if err != nil {
		log.Printf("Warning: Could not load data from file - %v", err)
	}

	ser := server.New("localhost:4000", s)

	fmt.Println("Starting server on localhost:8090")
	ser.Start()

}
