# OAuth Setup Guide

This guide provides step-by-step instructions for setting up Linear OAuth authentication for linctl, enabling proper app-level attribution for all Linear operations.

## Prerequisites

Before starting, ensure you have:
- Admin access to a Linear workspace
- linctl installed and available in your environment
- Basic understanding of OAuth 2.0 concepts
- Secure environment for storing OAuth credentials

## Step 1: Create OAuth Application in Linear

### 1.1 Access Linear OAuth Settings

1. Log in to your Linear workspace
2. Navigate to **Settings** → **API** → **OAuth Applications**
3. Click **"Create OAuth Application"**

### 1.2 Configure OAuth Application

Fill in the application details:

**Application Name**: `linctl CLI Tool`
**Description**: `OAuth-enabled Linear CLI for automated operations and agent workflows`
**Website URL**: `https://github.com/your-org/linctl` (optional)
**Callback URLs**:
- `http://localhost:8080/oauth/callback` (for local development)
- `https://your-domain.com/oauth/callback` (for production, if needed)

**Important Settings**:
- ✅ Enable **"Client Credentials"** - This is required for server-to-server authentication
- ✅ Enable **"Actor Authorization"** - This allows actions to be attributed to the app
- ✅ Select appropriate scopes:
  - `read` - Read access to Linear data
  - `write` - Write access for creating/updating
  - `issues:create` - Create issues
  - `comments:create` - Create comments
  - `admin` - Admin operations (if needed)

### 1.3 Save OAuth Credentials

After creating the application, Linear will provide:
- **Client ID**: A public identifier for your application
- **Client Secret**: A private secret for authentication (keep secure!)

**⚠️ Security Note**: Store the client secret securely. Never commit it to version control or share it publicly.

## Step 2: Configure Environment Variables

### 2.1 Required Environment Variables

Set the following environment variables in your system:

```bash
# OAuth Configuration (Required)
export LINEAR_CLIENT_ID="your-oauth-client-id"
export LINEAR_CLIENT_SECRET="your-oauth-client-secret"

# Optional Configuration
export LINEAR_BASE_URL="https://api.linear.app"  # Default value
export LINEAR_SCOPES="read,write,issues:create,comments:create"  # Default scopes
```

### 2.2 Actor Configuration (Optional)

Configure default actor attribution for automated operations:

```bash
# Default actor settings
export LINEAR_DEFAULT_ACTOR="AI Agent"
export LINEAR_DEFAULT_AVATAR_URL="https://example.com/agent-avatar.png"
```

### 2.3 Environment Setup Methods

#### Option A: Shell Profile (Persistent)
Add to your `~/.bashrc`, `~/.zshrc`, or equivalent:

```bash
# Linear OAuth Configuration
export LINEAR_CLIENT_ID="your-oauth-client-id"
export LINEAR_CLIENT_SECRET="your-oauth-client-secret"
export LINEAR_DEFAULT_ACTOR="AI Agent"
```

Then reload your shell:
```bash
source ~/.bashrc  # or ~/.zshrc
```

#### Option B: Environment File
Create a `.env` file (ensure it's in `.gitignore`):

```bash
# .env file
LINEAR_CLIENT_ID=your-oauth-client-id
LINEAR_CLIENT_SECRET=your-oauth-client-secret
LINEAR_DEFAULT_ACTOR=AI Agent
LINEAR_DEFAULT_AVATAR_URL=https://example.com/agent-avatar.png
```

Load the environment file:
```bash
source .env
```

#### Option C: Docker/Container Environment
For containerized environments:

```dockerfile
# Dockerfile
ENV LINEAR_CLIENT_ID=your-oauth-client-id
ENV LINEAR_CLIENT_SECRET=your-oauth-client-secret
ENV LINEAR_DEFAULT_ACTOR="AI Agent"
```

Or with docker-compose:
```yaml
# docker-compose.yml
services:
  app:
    environment:
      - LINEAR_CLIENT_ID=your-oauth-client-id
      - LINEAR_CLIENT_SECRET=your-oauth-client-secret
      - LINEAR_DEFAULT_ACTOR=AI Agent
```

## Step 3: Test OAuth Authentication

### 3.1 Verify Configuration

Test that your OAuth configuration is working:

```bash
# Check authentication status
linctl auth status

# Expected output for successful OAuth setup:
# ✅ Authenticated via OAuth
# 👤 User: Your Name (your@email.com)
# 🔑 Token expires: 2024-01-15 10:30 UTC (in 29 days)
# 📋 Scopes: read, write, issues:create, comments:create
```

### 3.2 Test OAuth Login

If not already authenticated, perform OAuth login:

```bash
# OAuth authentication
linctl auth login --oauth

# Expected flow:
# 🔐 Linear OAuth Authentication
# 🌐 Obtaining OAuth token...
# ✅ Successfully authenticated with Linear!
```

### 3.3 Test Basic Operations

Verify OAuth authentication works with Linear operations:

```bash
# Test basic read operation
linctl team list --json

# Test issue creation (if you have appropriate permissions)
linctl issue create --title "OAuth Test Issue" --team "ENG" --json
```

## Step 4: Advanced Configuration

### 4.1 Multiple Environment Setup

For different environments (development, staging, production):

```bash
# Development
export LINEAR_CLIENT_ID="dev-client-id"
export LINEAR_CLIENT_SECRET="dev-client-secret"
export LINEAR_DEFAULT_ACTOR="Dev Agent"

# Production
export LINEAR_CLIENT_ID="prod-client-id"
export LINEAR_CLIENT_SECRET="prod-client-secret"
export LINEAR_DEFAULT_ACTOR="Production Agent"
```

### 4.2 Scope Configuration

Configure specific scopes based on your needs:

```bash
# Minimal read-only access
export LINEAR_SCOPES="read"

# Full access for automation
export LINEAR_SCOPES="read,write,issues:create,comments:create,admin"

# Custom scope combination
export LINEAR_SCOPES="read,write,issues:create"
```

### 4.3 Custom Base URL (for Enterprise)

If using Linear Enterprise with custom domain:

```bash
export LINEAR_BASE_URL="https://your-company.linear.app"
```

## Step 5: Verification and Testing

### 5.1 Complete Authentication Test

Run a comprehensive authentication test:

```bash
# Test authentication status
linctl auth status --json

# Expected JSON output:
{
  "authenticated": true,
  "method": "oauth",
  "user": {
    "name": "Your Name",
    "email": "your@email.com"
  },
  "token_expires_at": "2024-01-15T10:30:00Z",
  "scopes": ["read", "write", "issues:create", "comments:create"],
  "suggestions": []
}
```

### 5.2 Token Refresh Test

Test automatic token refresh:

```bash
# Force token refresh
linctl auth refresh

# Expected output:
# 🔄 Refreshing OAuth token...
# ✅ OAuth token refreshed successfully
```

### 5.3 Actor Attribution Test

Test that actions are properly attributed to your app:

```bash
# Create an issue with actor attribution
linctl issue create \
  --title "OAuth Actor Test" \
  --team "ENG" \
  --actor "Test Agent" \
  --avatar-url "https://example.com/test-avatar.png" \
  --json
```

Check in Linear that the issue appears as created by "Test Agent (via linctl CLI Tool)".

## Troubleshooting

### Common Issues

#### 1. "OAuth request failed (401): invalid_client"
- **Cause**: Incorrect client ID or client secret
- **Solution**: Verify your `LINEAR_CLIENT_ID` and `LINEAR_CLIENT_SECRET` environment variables

#### 2. "OAuth request failed (400): unsupported_grant_type"
- **Cause**: Client credentials not enabled in Linear OAuth app
- **Solution**: Enable "Client Credentials" in your Linear OAuth application settings

#### 3. "Access token expired and refresh failed"
- **Cause**: Token expired and automatic refresh failed
- **Solution**: Re-authenticate with `linctl auth login --oauth`

#### 4. "Not authenticated"
- **Cause**: No valid authentication found
- **Solution**: Run `linctl auth login --oauth` to set up OAuth authentication

### Debug Mode

Enable debug logging for troubleshooting:

```bash
# Enable debug logging
export LINEAR_DEBUG=true
linctl auth status

# Check token information
linctl auth status --json | jq '.'
```

### Validation Checklist

- [ ] Linear OAuth application created with correct settings
- [ ] Client credentials enabled in Linear OAuth app
- [ ] Environment variables set correctly
- [ ] `linctl auth status` shows OAuth authentication
- [ ] Basic Linear operations work (team list, issue creation)
- [ ] Actor attribution appears correctly in Linear
- [ ] Token refresh works automatically

## Security Considerations

### 1. Client Secret Protection
- Never commit client secrets to version control
- Use secure environment variable management
- Rotate client secrets periodically
- Restrict access to production credentials

### 2. Token Storage
- Tokens are stored securely with appropriate file permissions
- Token files are excluded from backups and version control
- Automatic token cleanup on logout

### 3. Network Security
- All OAuth communications use HTTPS
- Validate SSL certificates in production
- Use secure networks for OAuth setup

### 4. Access Control
- Use minimal required scopes
- Regularly review OAuth application permissions
- Monitor OAuth token usage and errors

## Next Steps

After successful OAuth setup:

1. **Agent Integration**: Follow the [Agent Integration Guide](agent-integration-guide.md) for automated workflows
2. **CLI Usage**: Review the [CLI Usage Guide](cli-usage-guide.md) for complete command reference
3. **Migration**: If migrating from API keys, see the [Migration Guide](migration-guide.md)
4. **Monitoring**: Set up monitoring using the [Monitoring and Logging](monitoring-logging.md) guide

## Support

For additional help:
- Check the [Troubleshooting Guide](troubleshooting-guide.md) for common issues
- Review Linear's OAuth documentation
- Ensure your Linear workspace has the necessary permissions for OAuth applications
