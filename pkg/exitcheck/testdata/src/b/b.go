// alias 'os' для другого пакета с функцией Exit.
package main

import os "a"

func main() {
	// допустимо
	os.Exit(1)
}
