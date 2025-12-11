package cmd

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/registry/internal/api"
	v0 "github.com/modelcontextprotocol/registry/internal/api/handlers/v0"
	"github.com/modelcontextprotocol/registry/internal/config"
	"github.com/modelcontextprotocol/registry/internal/database"
	"github.com/modelcontextprotocol/registry/internal/importer"
	"github.com/modelcontextprotocol/registry/internal/service"
	"github.com/modelcontextprotocol/registry/internal/telemetry"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MCP Registry server",
	Long: `Starts the MCP Registry HTTP server.

The server provides APIs for:
  - Publishing MCP servers
  - Discovering available servers
  - Authentication and authorization
  - Server metadata management`,
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Server configuration flags
	serveCmd.Flags().String("server-address", "", "Server listen address (e.g., :8080)")
	serveCmd.Flags().String("database-url", "", "PostgreSQL connection URL")
	serveCmd.Flags().String("seed-from", "", "Seed data source (file path or URL)")

	// GitHub OAuth flags
	serveCmd.Flags().String("github-client-id", "", "GitHub OAuth client ID")
	serveCmd.Flags().String("github-client-secret", "", "GitHub OAuth client secret")

	// JWT flags
	serveCmd.Flags().String("jwt-private-key", "", "JWT private key (hex-encoded)")

	// Feature flags
	serveCmd.Flags().Bool("enable-anonymous-auth", false, "Enable anonymous authentication")
	serveCmd.Flags().Bool("enable-registry-validation", true, "Enable registry validation")

	// OIDC flags
	serveCmd.Flags().Bool("oidc-enabled", false, "Enable OIDC authentication")
	serveCmd.Flags().String("oidc-issuer", "", "OIDC issuer URL")
	serveCmd.Flags().String("oidc-client-id", "", "OIDC client ID")
	serveCmd.Flags().String("oidc-extra-claims", "", "Additional OIDC claims to validate")
	serveCmd.Flags().String("oidc-edit-permissions", "", "OIDC claims required for edit permissions")
	serveCmd.Flags().String("oidc-publish-permissions", "", "OIDC claims required for publish permissions")

	// Bind flags to Viper (will be available via getViper())
	viper.BindPFlag("server.address", serveCmd.Flags().Lookup("server-address"))
	viper.BindPFlag("database.url", serveCmd.Flags().Lookup("database-url"))
	viper.BindPFlag("seed.from", serveCmd.Flags().Lookup("seed-from"))
	viper.BindPFlag("github.client_id", serveCmd.Flags().Lookup("github-client-id"))
	viper.BindPFlag("github.client_secret", serveCmd.Flags().Lookup("github-client-secret"))
	viper.BindPFlag("jwt.private_key", serveCmd.Flags().Lookup("jwt-private-key"))
	viper.BindPFlag("features.enable_anonymous_auth", serveCmd.Flags().Lookup("enable-anonymous-auth"))
	viper.BindPFlag("features.enable_registry_validation", serveCmd.Flags().Lookup("enable-registry-validation"))
	viper.BindPFlag("oidc.enabled", serveCmd.Flags().Lookup("oidc-enabled"))
	viper.BindPFlag("oidc.issuer", serveCmd.Flags().Lookup("oidc-issuer"))
	viper.BindPFlag("oidc.client_id", serveCmd.Flags().Lookup("oidc-client-id"))
	viper.BindPFlag("oidc.extra_claims", serveCmd.Flags().Lookup("oidc-extra-claims"))
	viper.BindPFlag("oidc.edit_permissions", serveCmd.Flags().Lookup("oidc-edit-permissions"))
	viper.BindPFlag("oidc.publish_permissions", serveCmd.Flags().Lookup("oidc-publish-permissions"))
}

func runServe(cmd *cobra.Command, args []string) error {
	log.Printf("Starting MCP Registry Application v%s (commit: %s)", Version, GitCommit)

	// Initialize Viper with config file
	var err error
	v, err = config.InitViper(cfgFile)
	if err != nil {
		log.Printf("Failed to initialize configuration: %v", err)
		return err
	}

	// Create config from Viper
	cfg := config.NewConfigFromViper(v)
	cfg.Version = Version // Override version with build-time value

	var (
		registryService service.RegistryService
		db              database.Database
	)

	// Create a context with timeout for PostgreSQL connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to PostgreSQL
	db, err = database.NewPostgreSQL(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Printf("Failed to connect to PostgreSQL: %v", err)
		return err
	}

	// Store the PostgreSQL instance for later cleanup
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing PostgreSQL connection: %v", err)
		} else {
			log.Println("PostgreSQL connection closed successfully")
		}
	}()

	registryService = service.NewRegistryService(db, cfg)

	// Import seed data if seed source is provided
	if cfg.SeedFrom != "" {
		log.Printf("Importing data from %s...", cfg.SeedFrom)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		importerService := importer.NewService(registryService)
		if err := importerService.ImportFromPath(ctx, cfg.SeedFrom); err != nil {
			log.Printf("Failed to import seed data: %v", err)
		}
	}

	shutdownTelemetry, metrics, err := telemetry.InitMetrics(cfg.Version)
	if err != nil {
		log.Printf("Failed to initialize metrics: %v", err)
		return err
	}

	defer func() {
		if err := shutdownTelemetry(context.Background()); err != nil {
			log.Printf("Failed to shutdown telemetry: %v", err)
		}
	}()

	// Prepare version information
	versionInfo := &v0.VersionBody{
		Version:   Version,
		GitCommit: GitCommit,
		BuildTime: BuildTime,
	}

	// Initialize HTTP server
	server := api.NewServer(cfg, registryService, metrics, versionInfo)

	// Start server in a goroutine so it doesn't block signal handling
	go func() {
		if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Failed to start server: %v", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create context with timeout for shutdown
	sctx, scancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer scancel()

	// Gracefully shutdown the server
	if err := server.Shutdown(sctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
	return nil
}
