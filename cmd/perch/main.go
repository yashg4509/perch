package main

import (
	"os"

	"github.com/yashg4509/perch/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
