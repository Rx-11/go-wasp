package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go-wasp-runner <wasm-file> <func-name>")
		return
	}

}
