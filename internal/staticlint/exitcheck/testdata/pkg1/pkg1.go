package pkg1

import (
	"fmt"
	"os"
)

func myBeautifulFunc(s int) {
	if s > 42 {
		os.Exit(1)
	} else if s < 0 {
		os.Exit(0)
	}

	Exit()
	fmt.Println("We continue work")
}

// SomeFunc make some functionality for function
func SomeFunc() func() {
	wtf := func() {
		fmt.Println("wtf???? Goodbye!")
		os.Exit(3)
	}

	return wtf
}

// Exit some exit function
func Exit() {
	fmt.Println("I exit from some expression")
}
