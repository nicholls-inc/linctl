# Rate Limiting

Linear uses a [leaky bucket algorithm](https://en.wikipedia.org/wiki/Leaky_bucket) for rate limiting, with tokens refilled at a constant rate.

## API Request Limits

### Authenticated Requests (API Key)
- **Limit**: 1,500 requests per hour
- Requests are associated with the authenticated user

### OAuth App
- **Limit**: 1,200 requests per hour per user/app

### Unauthenticated Requests
- **Limit**: 60 requests per hour
- Requests are associated with the originating IP address

## Complexity Limits

### Authenticated Requests (API Key)
- **Limit**: 250,000 complexity points per hour

### Unauthenticated Requests
- **Limit**: 10,000 complexity points per hour

### Maximum Query Complexity
- **Limit**: 10,000 points per single query

## Rate Limit Headers

- `X-RateLimit-Requests-Limit`: Maximum requests per hour
- `X-RateLimit-Requests-Remaining`: Remaining requests in current window
- `X-RateLimit-Requests-Reset`: Time window reset in UTC epoch milliseconds

## Avoiding Rate Limits

Best practices:
1. Avoid polling the API
2. Use filtering to fetch only needed data
3. Sort data by updated timestamp
4. Write custom, specific queries
5. Use webhooks for updates

## Handling Rate Limit Errors

When rate limits are exceeded, the API returns an error response with:
```json
{
  "errors": [
    {
      "message": "...",
      "extensions": {
        "code": "RATELIMITED"
      }
    }
  ]
}
```
