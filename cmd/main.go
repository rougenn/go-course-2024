package main

import (
	"fmt"
	"hw1/internal/pkg/parseduration"
	"hw1/internal/pkg/server"
	"hw1/internal/pkg/storage"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	SD, CD, filename, port := parseduration.ParseDuration()
	store, err := storage.NewStorage(time.Second*time.Duration(SD), time.Second*time.Duration(CD), filename)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}

	if err := store.LoadFromPostgres(); err != nil {
		log.Fatalf("Ошибка загрузки состояния из базы данных: %v", err)
	}

	fmt.Println("Storage created successfully")

	s := server.New(port, store)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	go func() {
		sig := <-sigChan
		fmt.Printf("Получен сигнал: %s\n", sig)

		store.Stop()
		os.Exit(0)
	}()

	fmt.Println("Starting server on " + port)
	s.Start()
}
