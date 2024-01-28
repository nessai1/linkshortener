package main

import (
	"fmt"
	"math/rand"
	"os"
)

func main() {
	fmt.Println("some hello message")

	if iHaveLuckyDay() == 0 {
		os.Exit(0) // want "Must not contain os.Exit expression"
	}

	os.Exit(1) // want "Must not contain os.Exit expression"
}

func iHaveLuckyDay() int {
	return rand.Int() % 2
}

func someFunc() {
	os.Exit(1)
}
