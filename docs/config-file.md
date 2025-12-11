# Registry Configuration File

The MCP Registry server supports configuration via multiple sources with the following precedence (highest to lowest):

1. **Command-line flags** - Explicit flags passed to the `serve` command
2. **Environment variables** - Variables with `MCP_REGISTRY_` prefix
3. **Configuration file** - `registry.yaml` (or path specified with `--config`)
4. **Default values** - Built-in defaults

## Configuration File Format

Create a `registry.yaml` file in the current directory or `/etc/mcp-registry/`:

```yaml
server:
  address: ":8080"           # Server listen address
  version: "dev"              # Application version (typically set by build)

database:
  url: "postgres://localhost:5432/mcp-registry?sslmode=disable"  # PostgreSQL connection URL

seed:
  from: ""                    # Seed data source (file path or URL)

github:
  client_id: ""               # GitHub OAuth client ID
  client_secret: ""           # GitHub OAuth client secret

jwt:
  private_key: ""             # JWT private key (hex-encoded, 64 characters for Ed25519)

features:
  enable_anonymous_auth: false              # Enable anonymous authentication (for testing)
  enable_registry_validation: true          # Enable registry validation

oidc:
  enabled: false              # Enable OIDC authentication
  issuer: ""                  # OIDC issuer URL (e.g., https://accounts.google.com)
  client_id: ""               # OIDC client ID
  extra_claims: ""            # Additional OIDC claims to validate (JSON)
  edit_permissions: ""        # OIDC claims required for edit permissions (JSON)
  publish_permissions: ""     # OIDC claims required for publish permissions (JSON)
```

## Usage Examples

### Using a Config File

```bash
# Use default config file locations (./registry.yaml or /etc/mcp-registry/registry.yaml)
registry serve

# Specify a custom config file
registry serve --config /path/to/custom-config.yaml
```

### Using Environment Variables

All configuration values can be set via environment variables with the `MCP_REGISTRY_` prefix:

```bash
export MCP_REGISTRY_SERVER_ADDRESS=":9090"
export MCP_REGISTRY_DATABASE_URL="postgres://user:pass@localhost:5432/registry"
export MCP_REGISTRY_GITHUB_CLIENT_ID="Iv23..."
export MCP_REGISTRY_JWT_PRIVATE_KEY="bb2c6b424005..."

registry serve
```

Environment variables use underscores and uppercase. Nested configuration keys are flattened:
- `server.address` → `MCP_REGISTRY_SERVER_ADDRESS`
- `github.client_id` → `MCP_REGISTRY_GITHUB_CLIENT_ID`
- `features.enable_anonymous_auth` → `MCP_REGISTRY_FEATURES_ENABLE_ANONYMOUS_AUTH`

### Using Command-Line Flags

```bash
registry serve \
  --server-address=":9090" \
  --database-url="postgres://localhost:5432/registry" \
  --github-client-id="Iv23..." \
  --jwt-private-key="bb2c6b424005..." \
  --enable-registry-validation
```

### Combining Configuration Sources

You can combine all three sources. For example, use a config file for common settings and override specific values with environment variables:

```bash
# config.yaml contains common settings
# Override server address via environment variable
export MCP_REGISTRY_SERVER_ADDRESS=":9090"

registry serve --config config.yaml
```

## Configuration Reference

### Server Configuration

| Key | Environment Variable | Flag | Default | Description |
|-----|---------------------|------|---------|-------------|
| `server.address` | `MCP_REGISTRY_SERVER_ADDRESS` | `--server-address` | `:8080` | HTTP server listen address |
| `server.version` | `MCP_REGISTRY_SERVER_VERSION` | N/A | `dev` | Application version (set by build) |

### Database Configuration

| Key | Environment Variable | Flag | Default | Description |
|-----|---------------------|------|---------|-------------|
| `database.url` | `MCP_REGISTRY_DATABASE_URL` | `--database-url` | `postgres://localhost:5432/mcp-registry?sslmode=disable` | PostgreSQL connection URL |

### Seed Configuration

| Key | Environment Variable | Flag | Default | Description |
|-----|---------------------|------|---------|-------------|
| `seed.from` | `MCP_REGISTRY_SEED_FROM` | `--seed-from` | (empty) | Seed data source - file path or HTTP(S) URL |

### GitHub OAuth Configuration

| Key | Environment Variable | Flag | Default | Description |
|-----|---------------------|------|---------|-------------|
| `github.client_id` | `MCP_REGISTRY_GITHUB_CLIENT_ID` | `--github-client-id` | (empty) | GitHub OAuth application client ID |
| `github.client_secret` | `MCP_REGISTRY_GITHUB_CLIENT_SECRET` | `--github-client-secret` | (empty) | GitHub OAuth application client secret |

### JWT Configuration

| Key | Environment Variable | Flag | Default | Description |
|-----|---------------------|------|---------|-------------|
| `jwt.private_key` | `MCP_REGISTRY_JWT_PRIVATE_KEY` | `--jwt-private-key` | (empty) | Ed25519 private key for JWT signing (hex-encoded, 64 chars) |

### Feature Flags

| Key | Environment Variable | Flag | Default | Description |
|-----|---------------------|------|---------|-------------|
| `features.enable_anonymous_auth` | `MCP_REGISTRY_FEATURES_ENABLE_ANONYMOUS_AUTH` | `--enable-anonymous-auth` | `false` | Enable anonymous authentication (for testing) |
| `features.enable_registry_validation` | `MCP_REGISTRY_FEATURES_ENABLE_REGISTRY_VALIDATION` | `--enable-registry-validation` | `true` | Enable schema validation for published servers |

### OIDC Configuration

| Key | Environment Variable | Flag | Default | Description |
|-----|---------------------|------|---------|-------------|
| `oidc.enabled` | `MCP_REGISTRY_OIDC_ENABLED` | `--oidc-enabled` | `false` | Enable OIDC authentication |
| `oidc.issuer` | `MCP_REGISTRY_OIDC_ISSUER` | `--oidc-issuer` | (empty) | OIDC issuer URL |
| `oidc.client_id` | `MCP_REGISTRY_OIDC_CLIENT_ID` | `--oidc-client-id` | (empty) | OIDC client ID |
| `oidc.extra_claims` | `MCP_REGISTRY_OIDC_EXTRA_CLAIMS` | `--oidc-extra-claims` | (empty) | Additional claims to validate (JSON) |
| `oidc.edit_permissions` | `MCP_REGISTRY_OIDC_EDIT_PERMISSIONS` | `--oidc-edit-permissions` | (empty) | Claims required for edit operations (JSON) |
| `oidc.publish_permissions` | `MCP_REGISTRY_OIDC_PUBLISH_PERMISSIONS` | `--oidc-publish-permissions` | (empty) | Claims required for publish operations (JSON) |

## Security Best Practices

1. **Never commit secrets to version control** - Use environment variables or secure secret management systems for sensitive values like:
   - `github.client_secret`
   - `jwt.private_key`
   - Database passwords in `database.url`

2. **Use file permissions** - If storing secrets in a config file, ensure it has restrictive permissions:
   ```bash
   chmod 600 registry.yaml
   ```

3. **Use environment-specific configs** - Maintain separate config files for different environments:
   - `registry-dev.yaml`
   - `registry-staging.yaml`
   - `registry-prod.yaml`

4. **Validate configuration** - The registry will validate configuration on startup and fail fast if required values are missing or invalid.

## Docker and Container Deployments

When running in Docker, configuration is typically provided via environment variables:

```yaml
# docker-compose.yml
services:
  registry:
    image: ghcr.io/modelcontextprotocol/registry:latest
    environment:
      - MCP_REGISTRY_DATABASE_URL=postgres://user:pass@postgres:5432/registry
      - MCP_REGISTRY_GITHUB_CLIENT_ID=${GITHUB_CLIENT_ID}
      - MCP_REGISTRY_GITHUB_CLIENT_SECRET=${GITHUB_CLIENT_SECRET}
      - MCP_REGISTRY_JWT_PRIVATE_KEY=${JWT_PRIVATE_KEY}
    ports:
      - "8080:8080"
```

Alternatively, mount a config file:

```yaml
services:
  registry:
    image: ghcr.io/modelcontextprotocol/registry:latest
    volumes:
      - ./registry.yaml:/app/registry.yaml:ro
    command: serve --config /app/registry.yaml
```

## Troubleshooting

### Config file not found

If you see messages about config file not being found, this is normal - the registry will fall back to environment variables and defaults.

### Invalid configuration values

The registry validates configuration on startup. Check the error messages for specific issues:
- JWT private key must be valid hex-encoded Ed25519 key (64 characters)
- Database URL must be a valid PostgreSQL connection string
- Port numbers must be valid (e.g., `:8080` not `8080`)

### Environment variables not being read

Ensure environment variables have the `MCP_REGISTRY_` prefix and use uppercase with underscores:
- ✅ `MCP_REGISTRY_SERVER_ADDRESS`
- ❌ `SERVER_ADDRESS`
- ❌ `mcp_registry_server_address`
