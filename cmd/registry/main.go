package main

import (
	"os"

	"github.com/modelcontextprotocol/registry/cmd/registry/cmd"
)

// Version info for the MCP Registry application
// These variables are injected at build time via ldflags
var (
	// Version is the current version of the MCP Registry application
	Version = "dev"

	// BuildTime is the time at which the binary was built
	BuildTime = "unknown"

	// GitCommit is the git commit that was compiled
	GitCommit = "unknown"
)

func main() {
	// Pass version information to the cmd package
	cmd.Version = Version
	cmd.GitCommit = GitCommit
	cmd.BuildTime = BuildTime

	// Execute the root command
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
