package main

import (
	// "hw1/internal/pkg/server"
	"fmt"
	"hw1/internal/pkg/storage"
)

func main() {
	// store, err = server.New("/8080")

	s := storage.NewStorage()
	s.LoadFromFile("ex.json")
	// example 1
	// s.Rpush("list1", 1, 2, 3)
	fmt.Println(s.Lpop("list1", 0, -1))
	// fmt.Println(s.Lpop("list1", 0, 1))

	// example 2
	// s.Rpush("list1", 1, 2, 3)
	// s.Raddtoset("list1", 3, 5, 8, 4, 8)
	// fmt.Println(s.Lpop("list1", 0, -1))

	// example 3
	// s.Rpush("list1", 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	// fmt.Println(s.Lpop("list1", 2))
	// fmt.Println(s.Lpop("list1", 2, -2))
	// fmt.Println(s.Lpop("list1"))

	// example 4
	// s.Rpush("list2", 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	// fmt.Println(s.Rpop("list1", 2))
	// fmt.Println(s.Rpop("list1", 2, -2))
	// fmt.Println(s.Rpop("list1"))

	// example 5
	// s.Rpush("list3", 0, 1, 2, 3, 4, 5, 6, 7, 8, 9)
	// fmt.Println(s.Lset("list1", 3, 30))
	// fmt.Println(s.Lget("list1", 3))
	// fmt.Println(s.Lset("list1", 20, 2))

	s.SaveToFile("ex.json")
}
