// Command fakeclaude prints its arguments, one on each line, and exits with code 3.
package main

import (
	"fmt"
	"os"
)

func main() {
	for _, arg := range os.Args[1:] {
		fmt.Println(arg)
	}
	os.Exit(3)
}
