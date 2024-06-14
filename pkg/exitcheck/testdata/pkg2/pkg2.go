package main

import e "os"

type E int

func (e E) Exit(i int) {
	_ = i
}

func main() {
	var os E

	// допустимо
	os.Exit(1)
	e.Environ()
}
