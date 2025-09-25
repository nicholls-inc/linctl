Here's the TypeScript SDK documentation extracted as markdown:

# Linear TypeScript SDK: Getting Started

## Installation

```bash
npm install @linear/sdk
```

## Authentication Methods

Two primary authentication approaches:

1. Personal API Key Authentication
```typescript
import { LinearClient } from '@linear/sdk'

const client1 = new LinearClient({
  apiKey: YOUR_PERSONAL_API_KEY
})
```

2. OAuth2 Authentication
```typescript
const client2 = new LinearClient({
  accessToken: YOUR_OAUTH_ACCESS_TOKEN
})
```

## Basic Usage

### Querying Issues (Async/Await)
```typescript
async function getMyIssues() {
  const me = await linearClient.viewer;
  const myIssues = await me.assignedIssues();

  if (myIssues.nodes.length) {
    myIssues.nodes.map(issue =>
      console.log(`${me.displayName} has issue: ${issue.title}`)
    );
  } else {
    console.log(`${me.displayName} has no issues`);
  }
}
```

### Querying Issues (Promise)
```typescript
linearClient.viewer.then(me => {
  return me.assignedIssues().then(myIssues => {
    if (myIssues.nodes.length) {
      myIssues.nodes.map(issue =>
        console.log(`${me.displayName} has issue: ${issue.title}`)
      );
    } else {
      console.log(`${me.displayName} has no issues`);
    }
  });
});
```

## Creating Issues

```typescript
async function createIssue() {
  const team = await linearClient.team("team-id");

  const issuePayload = await linearClient.createIssue({
    teamId: team.id,
    title: "New Issue",
    description: "Issue description"
  });

  if (issuePayload.success) {
    console.log(`Created issue: ${issuePayload.issue?.title}`);
  }
}
```

## Error Handling

```typescript
try {
  const me = await linearClient.viewer;
  console.log(`Hello ${me.displayName}!`);
} catch (error) {
  console.error("Failed to fetch viewer:", error);
}
```

## Key Features

The SDK exposes Linear's GraphQL schema through strongly typed models and operations, compatible with both TypeScript and JavaScript environments.

- Fully typed GraphQL operations
- Built-in pagination handling
- Real-time subscriptions support
- Comprehensive error handling
- Built on top of Linear's GraphQL API
