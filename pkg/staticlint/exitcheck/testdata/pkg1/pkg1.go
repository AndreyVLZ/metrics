// прямой вызов os.Exit в main.
// err.
package main

import (
	"fmt"
	"os"
)

func main() {
	Valid()

	os.Exit(1) // want "прямой вызов os.Exit в функции main"
	fmt.Println("vim-go")
}

func Valid() {
	// допустимо
	os.Exit(1)
}
