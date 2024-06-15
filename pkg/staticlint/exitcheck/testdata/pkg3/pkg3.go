// alias для пакета os.
package main

import fos "os"

func main() {
	fos.Exit(1) // want "прямой вызов os.Exit в функции main"
}
