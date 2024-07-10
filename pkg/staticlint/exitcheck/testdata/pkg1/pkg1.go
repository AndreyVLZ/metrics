// прямой вызов os.Exit в main.
// err.
package main

import (
	"os"
)

func main() {
	os.Exit(1) // want "вызов os.Exit"
}
