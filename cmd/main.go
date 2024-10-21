package main

import (
	"hw1/internal/pkg/server"
	"hw1/internal/pkg/storage"
)

func main() {
	store := storage.NewStorage()
	s := server.New("localhost:8090", store)

	s.Start()
	// s := storage.NewStorage()
	// s.LoadFromFile("storage.json")

	// s.SaveToFile("storage.json")

}
