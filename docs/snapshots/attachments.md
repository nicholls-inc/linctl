Here's the Attachments documentation extracted as markdown:

# Attachments in Linear

## Overview
Attachments allow linking external resources to issues, similar to GitHub Pull Requests. They are designed for API developers and integrations.

## Key Concepts
- Unique URLs are core to attachments
- Attachments are "idempotent" - creating an attachment with the same URL on the same issue updates the original
- Can query attachments and associated issues by URL

## Example Use Cases
- Customer support software creating Linear issues
- Release bots attaching version information to issues

## Authentication and Icons
- Recommended to create attachments through OAuth authentication
- Application icon used by default
- Can specify custom icon URL (png or jpg)

## Metadata Support
- Supports key-value metadata
- Metadata can store integration-related information
- Currently only exposed via API

## API Examples

### Create Attachment
```graphql
mutation {
  attachmentCreate(input:{
    issueId: "590a1127-f98b-49fc-ba74-2df8751c089e"
    title: "Exception"
    subtitle: "Open"
    url: "http://exception.com/123"
    iconUrl: "https://exception.com/assets/icon.png"
    metadata: {exceptionId: "exc-123"}
  }) {
    success
    attachment {
      id
    }
  }
}
```

### Update Attachment
```graphql
mutation {
  attachmentUpdate(id: "47e14163-404c-4a34-b775-5c536d67760a", input: {
    title: "Exception"
    subtitle: "Resolved"
    metadata: {exceptionId: "exc-123"}
  }) {
    success
    attachment {
      id
    }
  }
}
```

## Rich Metadata Features
Supports advanced metadata rendering, including:
- Title
- Messages (with optional subject, body, timestamp)
- Attributes list

## Date Formatting
Supports dynamic date formatting in subtitles, like:
- `{variableName__since}`: "2 days ago"
- `{variableName__relativeTimestamp}`: Relative timestamp formatting

## Querying Attachments

### Find by URL
```graphql
query {
  attachments(filter: { url: { eq: "http://exception.com/123" } }) {
    nodes {
      id
      title
      subtitle
      issue {
        id
        title
      }
    }
  }
}
```

### Issue Attachments
```graphql
query {
  issue(id: "issue-id") {
    attachments {
      nodes {
        id
        title
        url
        metadata
      }
    }
  }
}
```
