package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/modelcontextprotocol/registry/internal/api/handlers/v0/auth"
	"github.com/modelcontextprotocol/registry/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestRegisterGitHubOIDCEndpoint_Enabled(t *testing.T) {
	// Create a new test API
	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("Test API", "1.0.0"))

	// Create config with GitHub OIDC enabled (default)
	cfg := &config.Config{
		EnableGitHubOIDC: true,
		JWTPrivateKey:    "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	}

	// Register the GitHub OIDC endpoint
	auth.RegisterGitHubOIDCEndpoint(api, "/v0", cfg)

	// Create a test request to the endpoint
	// We expect the endpoint to exist (even if the request fails due to missing body)
	req := httptest.NewRequest(http.MethodPost, "/v0/auth/github-oidc", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Serve the request
	mux.ServeHTTP(w, req)

	// The endpoint should be registered, so we should NOT get a 404
	// We expect a 400 or 422 (bad request/unprocessable entity) due to missing body
	assert.NotEqual(t, http.StatusNotFound, w.Code, "Endpoint should be registered when EnableGitHubOIDC is true")
}

func TestRegisterGitHubOIDCEndpoint_Disabled(t *testing.T) {
	// Create a new test API
	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("Test API", "1.0.0"))

	// Create config with GitHub OIDC disabled
	cfg := &config.Config{
		EnableGitHubOIDC: false,
		JWTPrivateKey:    "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	}

	// Register the GitHub OIDC endpoint (should be a no-op)
	auth.RegisterGitHubOIDCEndpoint(api, "/v0", cfg)

	// Create a test request to the endpoint
	req := httptest.NewRequest(http.MethodPost, "/v0/auth/github-oidc", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Serve the request
	mux.ServeHTTP(w, req)

	// The endpoint should NOT be registered, so we should get a 404
	assert.Equal(t, http.StatusNotFound, w.Code, "Endpoint should not be registered when EnableGitHubOIDC is false")
}

func TestRegisterGitHubOIDCEndpoint_DefaultConfig(t *testing.T) {
	// Create a new test API
	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("Test API", "1.0.0"))

	// Create a default config - EnableGitHubOIDC should default to true
	cfg := &config.Config{
		JWTPrivateKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	}

	// Note: In Go, bool defaults to false, but our config uses envDefault:"true"
	// This test verifies the behavior when the flag is explicitly not set
	// The actual default from env parsing would be true, but here we're testing
	// the explicit false case which is the Go zero value
	auth.RegisterGitHubOIDCEndpoint(api, "/v0", cfg)

	// Create a test request to the endpoint
	req := httptest.NewRequest(http.MethodPost, "/v0/auth/github-oidc", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Serve the request
	mux.ServeHTTP(w, req)

	// With Go's default bool value (false), the endpoint should not be registered
	assert.Equal(t, http.StatusNotFound, w.Code, "Endpoint should not be registered when EnableGitHubOIDC is false (Go default)")
}

func TestRegisterGitHubOIDCEndpoint_MultiplePathPrefixes(t *testing.T) {
	// Test that the endpoint works with different path prefixes (v0 and v0.1)
	tests := []struct {
		name       string
		pathPrefix string
		enabled    bool
		expectCode int
	}{
		{
			name:       "v0 prefix enabled",
			pathPrefix: "/v0",
			enabled:    true,
			expectCode: http.StatusUnprocessableEntity, // Endpoint exists but request is invalid
		},
		{
			name:       "v0.1 prefix enabled",
			pathPrefix: "/v0.1",
			enabled:    true,
			expectCode: http.StatusUnprocessableEntity, // Endpoint exists but request is invalid
		},
		{
			name:       "v0 prefix disabled",
			pathPrefix: "/v0",
			enabled:    false,
			expectCode: http.StatusNotFound, // Endpoint not registered
		},
		{
			name:       "v0.1 prefix disabled",
			pathPrefix: "/v0.1",
			enabled:    false,
			expectCode: http.StatusNotFound, // Endpoint not registered
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			api := humago.New(mux, huma.DefaultConfig("Test API", "1.0.0"))

			cfg := &config.Config{
				EnableGitHubOIDC: tt.enabled,
				JWTPrivateKey:    "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			}

			auth.RegisterGitHubOIDCEndpoint(api, tt.pathPrefix, cfg)

			req := httptest.NewRequest(http.MethodPost, tt.pathPrefix+"/auth/github-oidc", nil)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			// For enabled endpoints, we accept any non-404 status (could be 400, 422, etc.)
			if tt.enabled {
				assert.NotEqual(t, http.StatusNotFound, w.Code, "Endpoint should be registered when enabled")
			} else {
				assert.Equal(t, http.StatusNotFound, w.Code, "Endpoint should not be registered when disabled")
			}
		})
	}
}
