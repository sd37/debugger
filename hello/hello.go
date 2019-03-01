package main

import (
	"fmt"
)

func main() {
	i := 0x3335
	f1(i)
	fmt.Println("bye bye bye")
}

func f1(i int) int {
	return f2(i, 0x2222)
}

func f2(i, j int) int {
	return i * j
}
