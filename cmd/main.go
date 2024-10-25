package main

import (
	"hw1/internal/pkg/server"
	"hw1/internal/pkg/storage"
	"time"
)

func main() {
	s, _ := storage.NewStorage(time.Second*20, time.Second*10, "my-storage.json")
	s.LoadFromFile("my-storage.json")
	ser := server.New("localhost:8090", s)

	ser.Start()

}
