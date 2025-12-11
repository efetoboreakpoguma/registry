package config

import (
	"strings"

	"github.com/spf13/viper"
)

// InitViper initializes a Viper instance with MCP Registry configuration
// It sets up environment variable reading with the MCP_REGISTRY_ prefix
// and optionally loads configuration from a YAML file
func InitViper(configFile string) (*viper.Viper, error) {
	v := viper.New()

	// Set all defaults from the current Config struct tags
	v.SetDefault("server.address", ":8080")
	v.SetDefault("database.url", "postgres://localhost:5432/mcp-registry?sslmode=disable")
	v.SetDefault("seed.from", "")
	v.SetDefault("server.version", "dev")
	v.SetDefault("github.client_id", "")
	v.SetDefault("github.client_secret", "")
	v.SetDefault("jwt.private_key", "")
	v.SetDefault("features.enable_anonymous_auth", false)
	v.SetDefault("features.enable_registry_validation", true)

	// OIDC defaults
	v.SetDefault("oidc.enabled", false)
	v.SetDefault("oidc.issuer", "")
	v.SetDefault("oidc.client_id", "")
	v.SetDefault("oidc.extra_claims", "")
	v.SetDefault("oidc.edit_permissions", "")
	v.SetDefault("oidc.publish_permissions", "")

	// Environment variables with MCP_REGISTRY_ prefix
	v.SetEnvPrefix("MCP_REGISTRY")
	// Replace . with _ for nested keys (e.g., server.address -> MCP_REGISTRY_SERVER_ADDRESS)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Config file support (optional)
	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		v.SetConfigName("registry")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")              // Look in current directory
		v.AddConfigPath("/etc/mcp-registry/") // Look in /etc
	}

	// Try to read config file, but config file not found is OK
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error occurred
			return nil, err
		}
		// Config file not found is acceptable - will use env vars and defaults
	}

	return v, nil
}

// NewConfigFromViper creates a Config struct from a Viper instance
func NewConfigFromViper(v *viper.Viper) *Config {
	return &Config{
		ServerAddress:            v.GetString("server.address"),
		DatabaseURL:              v.GetString("database.url"),
		SeedFrom:                 v.GetString("seed.from"),
		Version:                  v.GetString("server.version"),
		GithubClientID:           v.GetString("github.client_id"),
		GithubClientSecret:       v.GetString("github.client_secret"),
		JWTPrivateKey:            v.GetString("jwt.private_key"),
		EnableAnonymousAuth:      v.GetBool("features.enable_anonymous_auth"),
		EnableRegistryValidation: v.GetBool("features.enable_registry_validation"),
		OIDCEnabled:              v.GetBool("oidc.enabled"),
		OIDCIssuer:               v.GetString("oidc.issuer"),
		OIDCClientID:             v.GetString("oidc.client_id"),
		OIDCExtraClaims:          v.GetString("oidc.extra_claims"),
		OIDCEditPerms:            v.GetString("oidc.edit_permissions"),
		OIDCPublishPerms:         v.GetString("oidc.publish_permissions"),
	}
}
