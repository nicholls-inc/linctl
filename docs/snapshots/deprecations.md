Based on the provided web page content, here's a summary of the Deprecations documentation:

# Deprecations

Linear's GraphQL API approach to deprecations is unique:

## Key Points

- Linear doesn't use traditional API versioning due to GraphQL's evolving nature
- Breaking changes are taken seriously
- When significant API changes occur, Linear will:
  - Proactively reach out to developers
  - Provide ample time to make necessary adjustments
  - Sometimes leave non-functioning stubs to prevent query/mutation breakage

## Deprecation Mechanism

- Uses the `@deprecated` directive in the GraphQL schema
- API changes are logged with `[API]` prefix in the [Linear changelog](https://linear.app/changelog)

## Checking for Deprecations

To stay updated on deprecations:

1. Monitor the Linear changelog for `[API]` prefixed entries
2. Use GraphQL introspection to check for `@deprecated` directives
3. Subscribe to Linear's developer communications

## Migration Strategy

When deprecations are announced:

1. Linear will provide advance notice
2. Deprecated fields will be marked in the schema
3. Migration guides will be provided when possible
4. Non-functioning stubs may be left to prevent breaking existing queries

## Best Practices

- Regularly check the Linear changelog
- Use GraphQL introspection tools to identify deprecated fields
- Plan migration timelines when deprecations are announced
- Test applications against schema updates

The documentation does not provide specific details about currently deprecated fields or exact migration guidance. For the most up-to-date deprecation information, developers are recommended to check the Linear changelog and GraphQL schema.
