# Linear OAuth Actor Authentication Documentation

This directory contains comprehensive documentation for the Linear OAuth Actor Authentication implementation in linctl. The OAuth actor authentication enables proper app-level attribution for all Linear operations, ensuring actions appear as coming from the application rather than individual users.

## Documentation Structure

### Setup and Configuration
- [**OAuth Setup Guide**](oauth-setup-guide.md) - Complete guide for setting up OAuth authentication with Linear
- [**Environment Configuration**](environment-configuration.md) - Environment variable setup and configuration options
- [**Migration Guide**](migration-guide.md) - Step-by-step migration from API key to OAuth authentication

### Usage and Integration
- [**Agent Integration Guide**](agent-integration-guide.md) - Optimized workflows for AI agents and automated systems
- [**CLI Usage Guide**](cli-usage-guide.md) - Complete command reference and usage examples
- [**API Reference**](api-reference.md) - Technical reference for OAuth-related functions and types

### Troubleshooting and Maintenance
- [**Troubleshooting Guide**](troubleshooting-guide.md) - Common issues and solutions
- [**Security Best Practices**](security-best-practices.md) - Security considerations and recommendations
- [**Monitoring and Logging**](monitoring-logging.md) - Monitoring OAuth operations and debugging

## Quick Start

### For Administrators
1. Follow the [OAuth Setup Guide](oauth-setup-guide.md) to configure Linear OAuth application
2. Set up environment variables as described in [Environment Configuration](environment-configuration.md)
3. Test the setup using the [CLI Usage Guide](cli-usage-guide.md)

### For Agents and Automation
1. Configure environment variables for OAuth authentication
2. Use the [Agent Integration Guide](agent-integration-guide.md) for optimized workflows
3. Implement error handling as described in the [Troubleshooting Guide](troubleshooting-guide.md)

### For Developers
1. Review the [API Reference](api-reference.md) for technical implementation details
2. Follow [Security Best Practices](security-best-practices.md) for secure implementation
3. Set up monitoring using the [Monitoring and Logging](monitoring-logging.md) guide

## Key Features

- **OAuth Client Credentials Flow**: Server-to-server authentication without user interaction
- **Actor Authorization**: All actions attributed to the application with custom actor names
- **Automatic Token Management**: Transparent token refresh and error handling
- **Agent Optimization**: JSON output, proper exit codes, and environment variable configuration
- **Backward Compatibility**: Seamless fallback to API key authentication
- **Production Ready**: Comprehensive error handling, rate limiting, and monitoring

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    linctl OAuth Architecture                    │
│                                                                 │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  │   OAuth Client  │    │  Token Manager  │    │  API Client     │
│  │                 │    │                 │    │                 │
│  │ - Client creds  │◄──►│ - Token storage │◄──►│ - GraphQL ops   │
│  │ - Token refresh │    │ - Auto refresh  │    │ - Actor auth    │
│  │ - Error handling│    │ - Validation    │    │ - Rate limiting │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘
│                                    │                              │
└────────────────────────────────────┼──────────────────────────────┘
                                     │
                    ┌────────────────┼────────────────┐
                    │                │                │
                    ▼                ▼                ▼
        ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
        │   CLI Commands  │ │  Agent Scripts  │ │  Integrations   │
        │                 │ │                 │ │                 │
        │ - Interactive   │ │ - JSON output   │ │ - CI/CD         │
        │ - Status info   │ │ - Exit codes    │ │ - Monitoring    │
        │ - Error msgs    │ │ - Env config    │ │ - Webhooks      │
        └─────────────────┘ └─────────────────┘ └─────────────────┘
```

## Implementation Status

This documentation covers **Phase 6: Documentation & Agent Integration** of the OAuth Actor implementation plan. The implementation includes:

✅ **Complete OAuth Implementation**
- Client credentials flow with automatic token management
- Actor authorization for all Linear operations
- Comprehensive error handling and retry logic

✅ **Production-Ready Features**
- Secure token storage with appropriate file permissions
- Rate limiting and performance optimization
- Comprehensive testing and validation

✅ **Agent Integration**
- JSON output for all commands
- Environment variable configuration
- Silent operation modes for automation
- Proper exit codes for script integration

## Support and Contributing

For issues, questions, or contributions:
1. Check the [Troubleshooting Guide](troubleshooting-guide.md) for common issues
2. Review existing documentation for answers
3. Follow security best practices when reporting issues
4. Include relevant logs and configuration (without secrets) when seeking help

## License

This documentation is part of the linctl project and follows the same licensing terms.
