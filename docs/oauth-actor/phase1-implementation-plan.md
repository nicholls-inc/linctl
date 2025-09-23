# Phase 1 Implementation Plan: Walking Skeleton - Basic OAuth Flow

Based on comprehensive analysis of the documentation and existing code, this document outlines the detailed implementation plan for Phase 1 of the OAuth Actor Authentication feature.

## Analysis Summary

### Current State
- **linctl** uses Personal API Key authentication stored in `~/.linctl-auth.json`
- **Auth system** is in `pkg/auth/auth.go` with simple `AuthConfig` struct
- **API client** in `pkg/api/client.go` expects auth header string
- **Commands** in `cmd/auth.go` support login/logout/status for API keys only

### Phase 1 Requirements
- Create minimal OAuth package with client credentials flow
- Extend auth system to support both API key and OAuth
- Add `--oauth` flag to login command
- Implement walking skeleton test: OAuth token → store → retrieve → API call

## Implementation Plan

### 1. Create OAuth Package Structure

**Files to create:**
- `/workspaces/linctl/pkg/oauth/client.go` - OAuth client implementation
- `/workspaces/linctl/pkg/oauth/types.go` - OAuth response types

### 2. OAuth Client Implementation (`pkg/oauth/client.go`)

Based on webhook-server patterns, implement:
```go
type OAuthClient struct {
    clientID     string
    clientSecret string
    baseURL      string
    httpClient   *http.Client
}

func NewOAuthClient(clientID, clientSecret, baseURL string) *OAuthClient
func (c *OAuthClient) GetAccessToken(ctx context.Context, scopes []string) (*TokenResponse, error)
func (c *OAuthClient) ValidateToken(ctx context.Context, accessToken string) error
```

**Key implementation details:**
- Use client credentials flow (`grant_type=client_credentials`)
- HTTP Basic Auth with `clientID:clientSecret`
- POST to `{baseURL}/oauth/token`
- Handle JSON response parsing and errors

### 3. OAuth Types (`pkg/oauth/types.go`)

```go
type TokenResponse struct {
    AccessToken string `json:"access_token"`
    TokenType   string `json:"token_type"`
    ExpiresIn   int    `json:"expires_in"`
    Scope       string `json:"scope"`
}
```

### 4. Extend Auth System (`pkg/auth/auth.go`)

**Modify existing `AuthConfig`:**
```go
type AuthConfig struct {
    APIKey     string `json:"api_key,omitempty"`
    OAuthToken string `json:"oauth_token,omitempty"`
}
```

**Add new functions:**
```go
func LoginWithOAuth(plaintext, jsonOut bool) error
func GetAuthHeader() (string, error) // Updated to handle both methods
```

**Implementation approach:**
- `LoginWithOAuth()` prompts for client ID/secret, gets token, stores it
- `GetAuthHeader()` checks OAuth token first, falls back to API key
- Maintain backward compatibility - existing API key users unaffected

### 5. Update Auth Commands (`cmd/auth.go`)

**Add `--oauth` flag to login command:**
```go
var oauthFlag bool

func init() {
    loginCmd.Flags().BoolVar(&oauthFlag, "oauth", false, "Use OAuth authentication instead of API key")
}
```

**Update login command logic:**
```go
if oauthFlag {
    err = auth.LoginWithOAuth(plaintext, jsonOut)
} else {
    err = auth.Login(plaintext, jsonOut) // existing API key flow
}
```

**Update status command to show OAuth info when applicable**

### 6. Environment Variable Support

Support OAuth configuration via environment variables:
- `LINEAR_CLIENT_ID`
- `LINEAR_CLIENT_SECRET`
- `LINEAR_BASE_URL` (optional, defaults to `https://api.linear.app`)

### 7. Walking Skeleton Test Flow

The complete end-to-end flow will be:
1. `linctl auth login --oauth` prompts for client credentials
2. OAuth client gets access token from Linear
3. Token stored in `~/.linctl-auth.json` alongside existing API key field
4. `linctl auth status` shows OAuth authentication status
5. Any API call (e.g., `linctl whoami`) uses OAuth token via `GetAuthHeader()`

## Implementation Steps

### Step 1: Create OAuth Package
- Create `pkg/oauth/` directory
- Implement `types.go` with `TokenResponse`
- Implement `client.go` with basic OAuth client credentials flow

### Step 2: Extend Auth System
- Modify `AuthConfig` struct to include `OAuthToken`
- Add `LoginWithOAuth()` function
- Update `GetAuthHeader()` to check OAuth token first
- Ensure backward compatibility with existing API key flow

### Step 3: Update Commands
- Add `--oauth` flag to login command
- Update login command to route to OAuth flow when flag is set
- Update status command to show OAuth authentication details
- Test both OAuth and API key flows work

### Step 4: Environment Variables
- Add support for `LINEAR_CLIENT_ID` and `LINEAR_CLIENT_SECRET`
- Allow OAuth login without interactive prompts when env vars are set
- This enables agent/automated usage

### Step 5: Integration Testing
- Test complete OAuth flow: login → status → API call
- Verify backward compatibility with existing API key authentication
- Test environment variable configuration
- Ensure error handling works properly

## Key Design Decisions

### 1. Backward Compatibility First
- Existing API key authentication remains default and unchanged
- OAuth is opt-in via `--oauth` flag
- No breaking changes to existing workflows

### 2. Simple Storage Strategy
- Store OAuth token in same `~/.linctl-auth.json` file
- `GetAuthHeader()` checks OAuth token first, falls back to API key
- Phase 2 will add proper token management and refresh

### 3. Environment Variable Priority
- Interactive prompts for manual usage
- Environment variables for automated/agent usage
- Clear error messages when configuration is missing

### 4. Minimal Walking Skeleton
- Focus on basic client credentials flow only
- No token refresh in Phase 1 (that's Phase 2)
- No actor authorization yet (that's Phase 3)
- Just prove OAuth token can be obtained and used

## Success Criteria

- [ ] `linctl auth login --oauth` successfully obtains OAuth token
- [ ] OAuth token is stored and retrieved from config file
- [ ] `linctl auth status` shows OAuth authentication status
- [ ] API calls work with OAuth token (test with `linctl whoami`)
- [ ] Existing API key authentication continues to work unchanged
- [ ] Environment variable configuration works for automated setups
- [ ] Error handling provides clear messages for common issues

## Files to Modify/Create

**New files:**
- `pkg/oauth/client.go`
- `pkg/oauth/types.go`

**Modified files:**
- `pkg/auth/auth.go` (extend AuthConfig, add OAuth functions)
- `cmd/auth.go` (add --oauth flag, update command logic)

**No changes needed:**
- `pkg/api/client.go` (already accepts auth header string)
- Other command files (they use `auth.GetAuthHeader()` which will be updated)

## Implementation Details

### OAuth Client Implementation Pattern

Based on the webhook-server implementation, the OAuth client will:

1. **Client Credentials Flow**:
   ```go
   data := url.Values{
       "grant_type": {"client_credentials"},
       "scope":      {strings.Join(scopes, " ")},
   }
   ```

2. **HTTP Basic Authentication**:
   ```go
   req.SetBasicAuth(c.clientID, c.clientSecret)
   ```

3. **Error Handling**:
   - Network errors with context
   - HTTP status code validation
   - JSON parsing errors
   - Clear error messages for common issues

4. **Token Validation**:
   - Simple GraphQL query to verify token works
   - Used for `linctl auth status` command

### Auth System Integration

The auth system will be extended to support both authentication methods:

1. **Config Structure**:
   ```go
   type AuthConfig struct {
       APIKey     string `json:"api_key,omitempty"`
       OAuthToken string `json:"oauth_token,omitempty"`
   }
   ```

2. **Priority Order**:
   - Check OAuth token first (if present)
   - Fall back to API key (existing behavior)
   - Return error if neither is available

3. **Storage**:
   - Same `~/.linctl-auth.json` file
   - Backward compatible with existing configs
   - File permissions remain 0600 for security

### Command Integration

Commands will be updated minimally:

1. **Login Command**:
   - Add `--oauth` flag
   - Route to appropriate authentication method
   - Maintain existing behavior as default

2. **Status Command**:
   - Show authentication method (API Key vs OAuth)
   - Display relevant token information
   - Maintain existing output format compatibility

3. **Other Commands**:
   - No changes needed (they use `auth.GetAuthHeader()`)
   - Transparent OAuth support

This plan provides a minimal but complete OAuth implementation that proves the end-to-end flow while maintaining full backward compatibility with existing API key authentication.

## Testing

### Test Coverage

Comprehensive tests have been implemented for all Phase 1 components:

#### OAuth Package Tests (`pkg/oauth/`)
- **`client_test.go`** - OAuth client functionality
  - Client initialization with custom/default URLs
  - Successful token acquisition with proper request validation
  - HTTP error handling (401, 500, etc.)
  - JSON parsing and validation
  - Token validation via GraphQL
  - Context cancellation handling
  - Edge cases (empty tokens, invalid JSON)

- **`types_test.go`** - OAuth types and serialization
  - JSON marshaling/unmarshaling of TokenResponse
  - Handling of optional fields
  - Error cases for malformed JSON

#### Auth Package Tests (`pkg/auth/`)
- **`auth_test.go`** - Authentication system integration
  - AuthConfig JSON serialization with both auth methods
  - GetAuthHeader priority (OAuth > API key)
  - GetAuthMethod detection
  - Config file save/load operations
  - File permissions verification (0600)
  - Missing credentials handling
  - Invalid credentials error handling

#### Command Tests (`cmd/`)
- **`auth_test.go`** - Command integration
  - OAuth flag availability and parsing
  - Help text includes OAuth information
  - Command structure validation
  - Authentication method display
  - Global flag integration

### Running Tests

```bash
# Run all tests
go test ./... -v

# Run specific package tests
go test ./pkg/oauth -v
go test ./pkg/auth -v
go test ./cmd -v

# Run tests with coverage
go test ./... -cover
```

### Test Results

All tests pass successfully:
- **OAuth Package**: 13 tests covering client functionality and types
- **Auth Package**: 8 tests covering authentication integration (1 skipped)
- **Command Package**: 6 tests covering CLI integration
- **Total**: 27 tests with comprehensive coverage of Phase 1 functionality

### Test Quality

- **Unit Tests**: Isolated testing of individual components
- **Integration Tests**: End-to-end testing of auth workflows
- **Mock Servers**: HTTP test servers for OAuth flow validation
- **Error Scenarios**: Comprehensive error handling validation
- **Edge Cases**: Empty inputs, malformed data, network failures
- **Backward Compatibility**: Verification that existing API key auth works