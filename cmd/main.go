package main

import (
	"hw1/internal/pkg/storage"
	"time"
)

func main() {
	store, _ := storage.NewStorage(time.Second*5, time.Second*60, "my-storage.json")
	// s := server.New("localhost:8090", store)

	// s.Start()
	// s := storage.NewStorage()
	// s.LoadFromFile("storage.json")
	// fmt.Println(s.Get("abcd"))
	// s.SaveToFile("storage.json")

	store.Set("asdf", "myval", 9)
	store.Set("domtdelete", 1234, 0)
	store.Lpush("arr", 23, 12, 32, 43)
	// store.SaveToFile("my-storage.json")
	// time.Sleep(10 * time.Second)
	// store.Get("asdf")
	// store.Get("domtdelete")

}
