package main

import (
	"fmt"
	"hw1/internal/pkg/storage"
)

func main() {
	my_storage := storage.NewStorage()

	my_storage.Set("1", "123")
	my_storage.Set("2", "abcd")
	fmt.Println(*my_storage.Get("1"))
	fmt.Println(*my_storage.Get("2"))
	fmt.Println(my_storage.Get("3"))
	fmt.Println(*my_storage.GetKind("1"))
	fmt.Println(*my_storage.GetKind("2"))
}
