# Filtering in Linear's GraphQL API

## Overview
Filtering allows you to retrieve specific data by applying various conditions to queries. Most paginated results support filtering.

## Comparators

### General Comparators (String, Numeric, Date)
- `eq`: Equals the given value
- `neq`: Does not equal the given value
- `in`: Value is in the given collection
- `nin`: Value is not in the given collection

### Numeric and Date Additional Comparators
- `lt`: Less than the given value
- `lte`: Less than or equal to the given value
- `gt`: Greater than the given value
- `gte`: Greater than or equal to the given value

### String-Specific Comparators
- `eqIgnoreCase`: Case-insensitive equals
- `startsWith`: Starts with the given value
- `endsWith`: Ends with the given value
- `contains`: Contains the given value
- `containsIgnoreCase`: Case-insensitive contains

## Logical Operators
- Default is logical AND
- Use `or` keyword to switch to logical OR

## Query Examples

### Basic Priority Filtering
```graphql
query HighPriorityIssues {
  issues(filter: { 
    priority: { lte: 2, neq: 0 }
  }) {
    nodes {
      id, title, priority
    }
  }
}
```

### Filtering by Relationship
```graphql
query AssignedIssues {
  issues(filter: { 
    assignee: { email: { eq: "john@linear.app" } }
  }) {
    nodes {
      id, title
    }
  }
}
```

### Complex Filtering
```graphql
query Issues {
  projects(filter: { 
    lead: { name: { startsWith: "John" } } 
  }) {
    nodes {
      issues(filter: { 
        labels: { name: { in: ["Bug", "Defect"] } } 
      }) {
        nodes {
          id, title
        }
      }
    }
  }
}
```

## Relative Time Filtering

Use relative time expressions for dynamic date filtering:

```graphql
query RecentIssues {
  issues(filter: { 
    createdAt: { gte: "-7d" }  # Issues created in last 7 days
  }) {
    nodes {
      id, title, createdAt
    }
  }
}
```

## Null Value Filtering

Check for null or non-null values:

```graphql
query UnassignedIssues {
  issues(filter: { 
    assignee: { null: true }
  }) {
    nodes {
      id, title
    }
  }
}
```