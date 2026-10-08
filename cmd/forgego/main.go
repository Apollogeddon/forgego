// Command forgego scaffolds linting, testing, CI/CD, Docker and Debian packaging into
// Go projects.
package main

import (
	"os"

	"github.com/apollogeddon/forgego/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
