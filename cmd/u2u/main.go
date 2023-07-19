package main

import (
	"fmt"
	"os"

	"github.com/unicornultrafoundation/go-u2u/cmd/u2u/launcher"
)

func main() {
	if err := launcher.Launch(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
