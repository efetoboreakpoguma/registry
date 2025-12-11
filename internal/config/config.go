package config

// Config holds the application configuration
// See .env.example for more documentation
type Config struct {
	ServerAddress            string
	DatabaseURL              string
	SeedFrom                 string
	Version                  string
	GithubClientID           string
	GithubClientSecret       string
	JWTPrivateKey            string
	EnableAnonymousAuth      bool
	EnableRegistryValidation bool

	// OIDC Configuration
	OIDCEnabled      bool
	OIDCIssuer       string
	OIDCClientID     string
	OIDCExtraClaims  string
	OIDCEditPerms    string
	OIDCPublishPerms string
}

// NewConfig creates a new configuration with default values
// This function maintains backward compatibility by using Viper internally
func NewConfig() *Config {
	v, err := InitViper("")
	if err != nil {
		panic(err)
	}
	return NewConfigFromViper(v)
}
