# Linear OAuth Actor Authentication Implementation Plan - Walking Skeleton Approach

Based on comprehensive review of the documentation and source code, this document outlines an incremental development plan using the Walking Skeleton approach for implementing Linear OAuth Actor Authentication in linctl.

## Analysis Summary

### Current State
- **linctl** currently uses Personal API Key authentication only (`pkg/auth/auth.go`)
- **webhook-server** has a complete OAuth Actor implementation (`internal/services/linear/oauth.go`)
- The goal is to implement OAuth Actor authentication in linctl to enable app-level attribution

### Key Findings
1. **Existing OAuth Implementation**: The webhook-server already has a robust OAuth client credentials implementation with actor authorization
2. **linctl Architecture**: Clean CLI structure with separate auth, API, and command packages
3. **Integration Opportunity**: Can leverage the existing OAuth patterns from webhook-server

## Walking Skeleton Approach

The Walking Skeleton approach ensures we have a minimal end-to-end working system early, then incrementally add features. Each phase builds on the previous one with a working, testable system.

## Phase 1: Walking Skeleton - Basic OAuth Flow (Week 1)

**Goal**: Establish minimal end-to-end OAuth authentication that can obtain and use tokens

### Deliverables

#### 1.1 Minimal OAuth Package Structure
```
/workspaces/linctl/pkg/oauth/
├── client.go          # Basic OAuth client credentials implementation
└── types.go           # Essential OAuth response types
```

#### 1.2 Core OAuth Client (`pkg/oauth/client.go`)
Minimal implementation based on webhook-server patterns:
```go
type OAuthClient struct {
    clientID     string
    clientSecret string
    baseURL      string
    httpClient   *http.Client
}

// Core method - client credentials flow only
func (c *OAuthClient) GetAccessToken(ctx context.Context, scopes []string) (*TokenResponse, error)

// Basic validation
func (c *OAuthClient) ValidateToken(ctx context.Context, accessToken string) error
```

#### 1.3 Essential Types (`pkg/oauth/types.go`)
```go
type TokenResponse struct {
    AccessToken string `json:"access_token"`
    TokenType   string `json:"token_type"`
    ExpiresIn   int    `json:"expires_in"`
    Scope       string `json:"scope"`
}
```

#### 1.4 Basic Auth Integration (`pkg/auth/auth.go`)
Extend existing auth to support OAuth alongside API key:
```go
type AuthConfig struct {
    APIKey     string `json:"api_key,omitempty"`
    OAuthToken string `json:"oauth_token,omitempty"`
}

// New OAuth login function
func LoginWithOAuth(plaintext, jsonOut bool) error

// Updated to handle both auth methods
func GetAuthHeader() (string, error)
```

#### 1.5 Simple OAuth Command (`cmd/auth.go`)
Add basic OAuth authentication:
```bash
linctl auth login --oauth    # New OAuth flow
linctl auth status           # Updated to show OAuth status
```

#### 1.6 Walking Skeleton Test
- OAuth client can obtain token from Linear
- Token can be stored and retrieved
- `linctl auth status` shows OAuth authentication
- One simple API call (e.g., get viewer) works with OAuth token

**Success Criteria**: Complete OAuth authentication flow working end-to-end with basic token usage

## Phase 2: Token Management & Persistence (Week 2)

**Goal**: Add robust token storage, refresh, and automatic token management

### Deliverables

#### 2.1 Token Storage (`pkg/oauth/token.go`)
```go
type TokenStore struct {
    configPath string
}

func (ts *TokenStore) SaveToken(token *TokenResponse) error
func (ts *TokenStore) LoadToken() (*TokenResponse, error)
func (ts *TokenStore) ClearToken() error
func (ts *TokenStore) IsTokenExpired(token *TokenResponse) bool
```

#### 2.2 Enhanced OAuth Client
Add token refresh and automatic management:
```go
func (c *OAuthClient) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)
func (c *OAuthClient) GetValidToken(ctx context.Context, scopes []string) (*TokenResponse, error)
```

#### 2.3 Configuration Management (`pkg/oauth/config.go`)
```go
type Config struct {
    ClientID     string
    ClientSecret string
    BaseURL      string
    Scopes       []string
}

func LoadFromEnvironment() (*Config, error)
func (c *Config) Validate() error
```

#### 2.4 Enhanced API Client (`pkg/api/client.go`)
Update to handle OAuth tokens with automatic refresh:
```go
type Client struct {
    httpClient  *http.Client
    authHeader  string
    oauthClient *oauth.OAuthClient // New field
}

// Enhanced to handle 401 errors with token refresh
func (c *Client) makeRequest(ctx context.Context, req *http.Request) (*http.Response, error)
```

#### 2.5 Environment Variable Support
Support OAuth configuration via environment variables:
```bash
LINEAR_CLIENT_ID=your-oauth-client-id
LINEAR_CLIENT_SECRET=your-oauth-client-secret
```

**Success Criteria**: Tokens are persistently stored, automatically refreshed, and OAuth works from environment variables

## Phase 3: Actor Authorization Support (Week 3)

**Goal**: Add actor authorization parameters to enable app-level attribution

### Deliverables

#### 3.1 Actor Support in GraphQL Operations
Update API queries and mutations to support actor fields:
```go
type IssueCreateInput struct {
    Title          string  `json:"title"`
    TeamID         string  `json:"teamId"`
    CreateAsUser   *string `json:"createAsUser,omitempty"`
    DisplayIconURL *string `json:"displayIconUrl,omitempty"`
}

type CommentCreateInput struct {
    IssueID        string  `json:"issueId"`
    Body           string  `json:"body"`
    CreateAsUser   *string `json:"createAsUser,omitempty"`
    DisplayIconURL *string `json:"displayIconUrl,omitempty"`
}
```

#### 3.2 Actor Command Parameters
Add actor authorization flags to relevant commands:
```bash
# Issue creation with actor attribution
linctl issue create --title "Bug fix" --team ENG --actor "AI Agent" --avatar-url "https://example.com/agent.png"

# Comment creation with actor attribution
linctl comment create LIN-123 --body "Working on this" --actor "AI Agent"
```

#### 3.3 Enhanced Command Implementation
Update issue and comment commands to use actor parameters:
```go
// In cmd/issue.go
var createCmd = &cobra.Command{
    Use: "create",
    Run: func(cmd *cobra.Command, args []string) {
        input := api.IssueCreateInput{
            Title:          title,
            TeamID:         teamID,
            CreateAsUser:   actor,      // New field
            DisplayIconURL: avatarURL,  // New field
        }
        // ... rest of implementation
    },
}

func init() {
    createCmd.Flags().StringVar(&actor, "actor", "", "Actor name for attribution")
    createCmd.Flags().StringVar(&avatarURL, "avatar-url", "", "Avatar URL for actor")
}
```

#### 3.4 Default Actor Configuration
Support default actor configuration via environment variables:
```bash
LINEAR_DEFAULT_ACTOR="AI Agent"
LINEAR_DEFAULT_AVATAR_URL="https://example.com/agent.png"
```

**Success Criteria**: All Linear operations can be attributed to custom actors, with app-level attribution working

## Phase 4: Enhanced Authentication UX (Week 4)

**Goal**: Improve authentication user experience with focus on core value and simplicity

### Design Philosophy
- **Automatic Priority System**: OAuth takes priority over API key, no manual switching needed
- **Seamless Fallback**: Graceful degradation from OAuth to API key
- **Clear Guidance**: Help users understand their auth status and suggest improvements
- **Minimal Complexity**: Avoid feature bloat, focus on essential functionality

### Deliverables

#### 4.1 Simplified Auth Commands
```bash
linctl auth login --oauth          # OAuth setup flow
linctl auth login                  # API key setup (existing, default for backward compatibility)
linctl auth refresh                # Refresh OAuth token
linctl auth logout                 # Clear all credentials
linctl auth status                 # Enhanced status with guidance
```

#### 4.2 Enhanced Status Command
Intelligent status reporting with user guidance:
```bash
# OAuth authenticated
$ linctl auth status
✅ Authenticated via OAuth
👤 User: John Doe (john@example.com)
🔑 Token expires: 2024-01-15 10:30 UTC (in 23 hours)
📋 Scopes: read, write, issues:create, comments:create

# API key authenticated
$ linctl auth status
✅ Authenticated via API Key
👤 User: John Doe (john@example.com)
💡 Consider upgrading to OAuth for enhanced features: linctl auth login --oauth

# JSON output for automation
$ linctl auth status --json
{
  "authenticated": true,
  "method": "oauth",
  "user": {"name": "John Doe", "email": "john@example.com"},
  "token_expires_at": "2024-01-15T10:30:00Z",
  "scopes": ["read", "write", "issues:create", "comments:create"],
  "suggestions": []
}
```

#### 4.3 Smart OAuth Login
Enhanced OAuth login with existing auth detection:
```bash
$ linctl auth login --oauth
ℹ️  Detected existing API key authentication
🔄 Setting up OAuth (API key will remain as fallback)
🌐 Opening browser for Linear OAuth authorization...
✅ OAuth setup complete! Future commands will use OAuth automatically.
```

#### 4.4 Robust Token Management
Automatic token refresh with clear error handling:
```go
// Enhanced token refresh with user-friendly errors
func RefreshOAuthTokenWithFeedback() error {
    // Attempt refresh
    // On failure, provide clear guidance:
    // "OAuth token expired and refresh failed. Please re-authenticate: linctl auth login --oauth"
}

// Smart auth header with automatic fallback
func GetAuthHeader() (string, error) {
    // 1. Try OAuth with automatic refresh
    // 2. Fall back to API key
    // 3. Provide clear error with next steps
}
```

#### 4.5 Clear Error Messages and Guidance
User-friendly error messages with actionable guidance:
```bash
# Token expired
$ linctl issue list
❌ OAuth token expired and refresh failed
💡 Please re-authenticate: linctl auth login --oauth

# No authentication
$ linctl issue list
❌ Not authenticated
💡 Set up authentication: linctl auth login --oauth (recommended) or linctl auth login

# Network issues
$ linctl issue list
❌ Authentication failed: network error
💡 Check your internet connection and try again
```

#### 4.6 Environment Variable Documentation
Clear guidance on environment-based configuration:
```bash
# OAuth configuration
export LINEAR_CLIENT_ID="your-oauth-client-id"
export LINEAR_CLIENT_SECRET="your-oauth-client-secret"

# API key configuration (fallback)
export LINEAR_API_KEY="your-api-key"

# Actor defaults (works with both auth methods)
export LINEAR_DEFAULT_ACTOR="AI Agent"
export LINEAR_DEFAULT_AVATAR_URL="https://example.com/agent.png"
```

### Implementation Details

#### 4.7 Enhanced Auth Status Logic
```go
type AuthStatus struct {
    Authenticated bool              `json:"authenticated"`
    Method        string            `json:"method"`        // "oauth", "api_key", or "none"
    User          *User             `json:"user,omitempty"`
    TokenExpiry   *time.Time        `json:"token_expires_at,omitempty"`
    Scopes        []string          `json:"scopes,omitempty"`
    Suggestions   []string          `json:"suggestions,omitempty"`
}

func GetAuthStatus() (*AuthStatus, error) {
    // Determine current auth method and status
    // Add helpful suggestions based on current state
    // Return comprehensive status information
}
```

#### 4.8 Automatic Priority System (Already Implemented)
The existing `GetAuthHeader()` function already implements smart priority:
1. **OAuth with automatic refresh** (highest priority)
2. **Stored OAuth token** (medium priority)
3. **API key fallback** (lowest priority)
4. **Clear error** with guidance (no auth found)

### Removed Features (Simplified Approach)

#### ❌ `linctl auth switch` - Not Needed
- **Why removed**: Automatic priority system makes manual switching unnecessary
- **Alternative**: Users control method via environment variables or login commands
- **Benefit**: Reduces complexity, eliminates user confusion

#### ❌ `linctl auth migrate` - Not Essential
- **Why removed**: Smart OAuth login handles migration naturally
- **Alternative**: Enhanced `auth status` suggests OAuth upgrade
- **Benefit**: Simpler implementation, less maintenance burden

#### ❌ Complex Method Detection - Over-Engineering
- **Why removed**: Current implementation already handles detection well
- **Alternative**: Clear status reporting shows current method
- **Benefit**: Focus on user experience over technical complexity

**Success Criteria**:
- Clear authentication status with helpful guidance
- Seamless OAuth setup experience
- Automatic token management with graceful fallback
- User-friendly error messages with actionable next steps
- Zero breaking changes to existing workflows

## Phase 5: Production Readiness (Week 5)

**Goal**: Production-ready implementation with comprehensive error handling and testing

### Deliverables

#### 5.1 Comprehensive Error Handling
- Network error handling with retries
- Token expiry and refresh error handling
- Clear error messages for common issues
- Graceful fallback mechanisms

#### 5.2 Rate Limiting & Performance
```go
type RateLimitedClient struct {
    *Client
    limiter *rate.Limiter
}

func (c *RateLimitedClient) makeRequest(ctx context.Context, req *http.Request) (*http.Response, error)
```

#### 5.3 Comprehensive Testing
- Unit tests for OAuth client implementation
- Integration tests for complete OAuth flow
- End-to-end tests with Linear API
- Backward compatibility tests for API key authentication
- Error scenario testing

#### 5.4 Security Enhancements
- Secure token storage with appropriate file permissions
- Token encryption at rest (optional)
- Secure handling of client secrets
- Input validation and sanitization

#### 5.5 Monitoring & Logging
```go
type Logger interface {
    Debug(msg string, fields ...interface{})
    Info(msg string, fields ...interface{})
    Warn(msg string, fields ...interface{})
    Error(msg string, fields ...interface{})
}

// OAuth operations logging
func (c *OAuthClient) logOperation(operation string, success bool, duration time.Duration)
```

**Success Criteria**: Production-ready OAuth implementation with comprehensive error handling and testing

## Phase 6: Documentation & Agent Integration (Week 6)

**Goal**: Complete documentation and optimize for agent workflows

### Deliverables

#### 6.1 Comprehensive Documentation
- OAuth setup guide for administrators
- Agent integration guide with environment variable configuration
- Migration guide from API key to OAuth
- Troubleshooting guide for common issues
- API reference for OAuth-related functions

#### 6.2 Agent Optimization
- Ensure all commands work seamlessly with `--json` flag
- Proper exit codes for script automation
- Environment variable configuration validation
- Silent operation modes for automated workflows

#### 6.3 Example Configurations
```bash
# Agent environment setup
export LINEAR_CLIENT_ID="your-client-id"
export LINEAR_CLIENT_SECRET="your-client-secret"
export LINEAR_DEFAULT_ACTOR="AI Agent"
export LINEAR_DEFAULT_AVATAR_URL="https://example.com/agent.png"

# Test agent setup
linctl auth status --json
linctl issue create --title "Test issue" --team ENG --json
```

#### 6.4 Integration Examples
- Docker container setup with OAuth
- CI/CD pipeline integration examples
- Agent workflow templates
- Monitoring and alerting examples

**Success Criteria**: Complete documentation with seamless agent integration and clear setup instructions

## Implementation Strategy

### Walking Skeleton Benefits
1. **Early Validation**: Working OAuth flow from Phase 1
2. **Risk Mitigation**: Issues discovered early in simple implementation
3. **Incremental Value**: Each phase adds working functionality
4. **Testability**: Each phase can be thoroughly tested before moving forward
5. **Feedback Loop**: Early user feedback on OAuth experience

### Backward Compatibility Strategy
- Maintain API key authentication as default method
- OAuth as opt-in feature initially
- Clear migration path with helper tools
- No breaking changes to existing commands
- Gradual transition support

### Code Reuse Strategy
- Adapt OAuth patterns from webhook-server implementation
- Maintain linctl's existing CLI patterns and user experience
- Consistent error handling and logging patterns
- Shared configuration patterns where possible

### Testing Strategy
Each phase includes:
- Unit tests for new functionality
- Integration tests for OAuth flow
- Backward compatibility verification
- End-to-end testing with Linear API
- Performance and error scenario testing

## Risk Mitigation

### Technical Risks
- **OAuth Token Expiry**: Automatic refresh with buffer time
- **Network Issues**: Retry logic with exponential backoff
- **Configuration Errors**: Comprehensive validation and clear error messages
- **API Changes**: Robust error handling and version compatibility

### User Experience Risks
- **Authentication Confusion**: Clear status reporting and automatic priority system
- **Backward Compatibility**: Maintain existing workflows unchanged
- **Agent Integration**: Thorough testing of automated workflows
- **OAuth Setup Complexity**: Smart login flow with clear guidance

## Success Metrics

### Technical Metrics
- OAuth token refresh success rate > 99.9%
- Linear API operation success rate > 99.5%
- Command execution time < 2 seconds average
- Zero authentication-related failures in agent workflows

### User Experience Metrics
- Successful OAuth setup rate > 95%
- Authentication error resolution rate > 90% (users can fix issues based on error messages)
- Support ticket reduction for authentication issues
- Positive feedback on simplified OAuth experience

## Future Enhancements

### Advanced Features (Post-Phase 6)
- Multi-workspace OAuth support
- Advanced permission scoping
- Token sharing between multiple tools
- OAuth token caching and optimization
- Advanced actor management features

### Integration Opportunities
- Integration with other Linear tools
- Webhook server shared OAuth token usage
- Centralized OAuth token management
- Advanced monitoring and analytics

## Conclusion

This Walking Skeleton approach ensures we have a working OAuth implementation early while incrementally building toward a comprehensive, production-ready solution. Each phase delivers working functionality that can be tested and validated, reducing risk and ensuring high-quality implementation.

The plan maintains backward compatibility while providing a clear path to OAuth actor authorization, enabling proper app-level attribution for all Linear operations performed through linctl.
