// Command packetlab is the headless entry point for the simulator.
package main

import (
	"os"

	"github.com/HalxDocs/packetlab/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
