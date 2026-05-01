package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/ksaifullah/go-cli-k8s-manifest-label-validator/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		if errors.Is(err, cmd.ErrValidationFailed) {
			// Violations were already printed; just exit with code 1.
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
