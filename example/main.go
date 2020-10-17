package main

import (
	"fmt"

	"github.com/charithe/go-ast-refactoring/example/example"
)

func main() {
	wc := example.WibbleClient{}
	wcw := example.WibbleClientWrapper{WibbleClient: wc}

	fmt.Printf("WibbleClient: %d\n", wc.Wibble(10))
	fmt.Printf("WibbleClientWrapper: %d\n", wcw.Wibble(10))
}
