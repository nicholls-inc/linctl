# OAuth 2.0 Authentication for Linear

## Create an OAuth2 Application
1. Create a new OAuth2 Application at https://linear.app/settings/api/applications/new
2. Configure redirect callback URLs for your application

## Authorization Request Parameters

| Parameter | Description | Required | Example |
|-----------|-------------|----------|---------|
| `client_id` | Client ID from OAuth2 Application | Yes | |
| `redirect_uri` | Redirect URI | Yes | |
| `response_type` | Must be `code` | Yes | |
| `scope` | Comma-separated list of access scopes | Yes | |
| `state` | Prevents CSRF attacks | Optional | |
| `prompt` | Force consent screen | Optional | `consent` |
| `actor` | Define resource creation method | Optional | `user` or `app` |

### Scope Options
- `read`: Default read access
- `write`: Write access
- `issues:create`: Create issues and attachments
- `comments:create`: Create issue comments
- `timeSchedule:write`: Modify time schedules
- `admin`: Full admin-level access

### Example Authorization Request
```
GET https://linear.app/oauth/authorize?response_type=code&client_id=YOUR_CLIENT_ID&redirect_uri=YOUR_REDIRECT_URL&state=SECURE_RANDOM&scope=read,write
```

## Token Exchange
Send a POST request to `https://api.linear.app/oauth/token` with:
- `code`: Authorization code
- `redirect_uri`: Same as authorization request
- `client_id`: Application's client ID
- `client_secret`: Application's client secret
- `grant_type`: Must be `authorization_code`

## Making API Requests
After obtaining an access token, you can:
1. Initialize Linear Client with token
2. Use token in Authorization header

### Example Client Initialization
```javascript
const client = new LinearClient({ accessToken: response.access_token })
const me = await client.viewer
```

## Revoking Access Token
POST to `https://api.linear.app/oauth/revoke` with the access token