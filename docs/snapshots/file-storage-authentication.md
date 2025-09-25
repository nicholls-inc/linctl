# File Storage Authentication

Files uploaded to Linear are stored in their private cloud storage and require authentication to access. These files are available from the `https://uploads.linear.app` hostname.

## Authorization Header

You can use the same access token and authorization header as used for Linear's GraphQL API when requesting files from storage. An example request looks like:

```bash
curl https://uploads.linear.app/6db02bb9-fba2-473b-8f9d-f38188e84813/d20adbea-186d-4643-ad07-004bda7d099d  \
  -X GET \
  -H 'Authorization: Bearer 00a21d8b0c4e2375114e49c067dfb81eb0d2076f48354714cd5df984d87b67cc'
```

## Request Signed URLs

When using the GraphQL API, you can request file storage URLs include a signature for temporary access by passing the `public-file-urls-expire-in` header with an integer representing signature expiration in seconds.

With the TypeScript SDK, you can set this header directly on the client:

```javascript
const client = new LinearClient({
  apiKey: process.env.LINEAR_API_KEY,
  headers: {
    "public-file-urls-expire-in": "60",
  }
});
```

This allows you to generate signed URLs valid for a specified duration.

## Security Considerations

- Files are stored in private cloud storage
- Access requires proper authentication
- Signed URLs provide temporary access without exposing credentials
- Always use HTTPS for file requests
- Consider expiration times for signed URLs based on your use case
