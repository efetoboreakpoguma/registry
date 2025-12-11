package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version info injected at build time via ldflags
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

var (
	cfgFile string
	v       *viper.Viper
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "registry",
	Short: "MCP Registry Server",
	Long: `MCP Registry provides MCP clients with a list of MCP servers,
like an app store for MCP servers.

The registry server manages server metadata, handles authentication,
and provides APIs for publishing and discovering MCP servers.`,
	Version: Version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default searches for registry.yaml in current directory and /etc/mcp-registry/)")

	// Set custom version template
	rootCmd.SetVersionTemplate(fmt.Sprintf(`MCP Registry %s
Git commit: %s
Build time: %s
`, Version, GitCommit, BuildTime))

	// Make serve the default command (runs when no subcommand is specified)
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		// If no subcommand provided, run serve
		return serveCmd.RunE(cmd, args)
	}
}

// getViper returns the initialized Viper instance
// This is called by commands that need access to configuration
func getViper() *viper.Viper {
	if v == nil {
		// This shouldn't happen in normal flow, but provide a fallback
		newV := viper.New()
		newV.SetEnvPrefix("MCP_REGISTRY")
		newV.AutomaticEnv()
		return newV
	}
	return v
}
