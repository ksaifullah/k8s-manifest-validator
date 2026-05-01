package main

import (
	"fmt"
	"os"

	"github.com/ksaifullah/go-cli-k8s-manifest-label-validator/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
