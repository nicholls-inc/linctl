# Pagination

All list responses from Linear's GraphQL API return paginated results using a Relay-style cursor-based pagination model.

## Pagination Arguments

- `first`: Number of items to retrieve initially
- `after`: Cursor for retrieving the next set of results
- `last`: Number of items to retrieve from the end
- `before`: Cursor for retrieving the previous set of results

## Example Query

```graphql
query Issues {
  issues(first: 10) {
    edges {
      node {
        id
        title
      }
      cursor
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
```

## Pagination Best Practices

- The first 50 results are returned by default without query arguments
- Use `pageInfo.endCursor` as the `after` parameter for subsequent requests
- Continue paginating while `pageInfo.hasNextPage` is true

## Alternative Syntax

You can also use a simpler syntax similar to GitHub's GraphQL API:

```graphql
query Teams {
  teams {
    nodes {
      id
      name
    }
  }
}
```

## Ordering Results

By default, results are ordered by `createdAt`. To get most recently updated resources, use `orderBy: updatedAt`:

```graphql
query Issues {
  issues(orderBy: updatedAt) {
    nodes {
      id
      identifier
      title
      createdAt
      updatedAt
    }
  }
}
```