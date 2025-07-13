# Linear GraphQL API Documentation

## Endpoint

The GraphQL endpoint is:
```
https://api.linear.app/graphql
```

## Authentication Methods

### OAuth 2.0
- Recommended for building applications for others
- Requires obtaining an access token
- Pass token with header: `Authorization: Bearer <ACCESS_TOKEN>`

### Personal API Keys
- Best for personal scripts
- Created in "Security & access" settings
- Pass key with header: `Authorization: <API_KEY>`

## Getting Started

### Basic Query Example: Viewer Information
```graphql
query Me {
  viewer {
    id
    name
    email
  }
}
```

### Fetching Team Issues
```graphql
query Team {
  team(id: "TEAM_ID") {
    id
    name
    issues {
      nodes {
        id
        title
        description
        assignee {
          id
          name
        }
        createdAt
        archivedAt
      }
    }
  }
}
```

### Creating an Issue
```graphql
mutation IssueCreate {
  issueCreate(
    input: {
      title: "New exception"
      description: "More detailed error report"
      teamId: "TEAM_ID"
    }
  ) {
    success
    issue {
      id
      title
    }
  }
}
```

## Additional Features

- Supports introspection
- Explorable via [Apollo Studio](https://studio.apollographql.com/public/Linear-API/schema/reference?variant=current)
- Provides TypeScript SDK for easier interaction

## Best Practices

- Use webhooks for real-time updates
- Filter data in GraphQL queries
- Avoid polling individual resources

## Support

- Customer Slack channel
- Email: hello@linear.app