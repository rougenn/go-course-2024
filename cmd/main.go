package main

import "fmt"

type A struct {
	x int
	Y int
}

func newA(x, y int) *A {
	var Ne = A{
		x: x,
		Y: y,
	}
	return &Ne
}

func main() {
	// fmt.Print("Hellowordl))SDF\n")

	a := newA(5123, 10)
	fmt.Printf("x: %d, Y: %d\n", a.x, a.Y)
}
