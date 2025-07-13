Here's a comprehensive markdown guide to Linear's Webhooks based on the documentation:

# Linear Webhooks Documentation

## Overview
Linear provides webhooks to receive HTTP push notifications when data is created or updated, enabling integrations and custom workflows.

## Key Features
- Webhooks are specific to an Organization
- Can be configured for all public teams or a single team
- Configurable via Linear's API settings
- Only workspace admins or OAuth applications with `admin` scope can create/read webhooks

## Supported Models
Data change webhooks are available for:
- Issues
- Issue attachments
- Issue comments
- Issue labels
- Comment reactions
- Projects
- Project updates
- Cycles

## Additional Webhook Types
- Issue SLA
- OAuthApp revoked

## Webhook Consumer Requirements
- Publicly accessible HTTPS endpoint
- Responds to Linear's webhook push with HTTP 200 status
- Must handle potential retries (max 3 attempts)

## Payload Structure
Webhook payloads include:

### HTTP Headers
- `Linear-Delivery`: Unique payload ID
- `Linear-Event`: Entity type (e.g., "Issue")
- `Linear-Signature`: HMAC signature

### Payload Fields
- `action`: Type of action (create/update/remove)
- `type`: Entity type
- `createdAt`: Timestamp of action
- `data`: Serialized entity details
- `url`: Entity URL
- `updatedFrom`: Previous values for updates

## Security
To verify webhooks:
1. Validate `Linear-Signature` using webhook's signing secret
2. Check `webhookTimestamp` is within recent timeframe
3. Optionally verify source IP addresses:
   - 35.231.147.226
   - 35.243.134.228
   - 34.140.253.14
   - 34.38.87.206

## Example Webhook Consumer (Netlify Function)

```javascript
const { createHmac } = require('node:crypto');

export default async (request) => {
  const payload = await request.text();
  const { action, data, type } = JSON.parse(payload);

  // Verify signature
  const signature = createHmac('sha256', process.env.LINEAR_WEBHOOK_SECRET)
    .update(payload)
    .digest('hex');
  
  if (signature !== request.headers['linear-signature']) {
    return new Response('Invalid signature', { status: 401 });
  }

  // Process the webhook
  console.log(`Received ${action} action for ${type}`);
  
  return new Response('OK', { status: 200 });
};
```