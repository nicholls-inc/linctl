# Agent Integration Guide

This guide provides comprehensive instructions for integrating linctl with AI agents and automated systems, optimized for OAuth actor authentication and seamless automation workflows.

## Overview

The linctl OAuth implementation is specifically designed for agent integration with:
- **JSON output** for all commands
- **Environment variable configuration** for secure credential management
- **Proper exit codes** for script automation
- **Silent operation modes** for automated workflows
- **Actor attribution** for clear audit trails

## Quick Start for Agents

### Minimal Agent Setup

```bash
# Set OAuth credentials
export LINEAR_CLIENT_ID="your-oauth-client-id"
export LINEAR_CLIENT_SECRET="your-oauth-client-secret"

# Set default actor for attribution
export LINEAR_DEFAULT_ACTOR="AI Agent"
export LINEAR_DEFAULT_AVATAR_URL="https://example.com/agent-avatar.png"

# Test authentication
linctl auth status --json
```

### Basic Agent Workflow

```bash
#!/bin/bash
set -e  # Exit on any error

# Check authentication
if ! linctl auth status --json >/dev/null 2>&1; then
    echo "Error: Not authenticated with Linear" >&2
    exit 1
fi

# Create issue with JSON output
ISSUE_RESULT=$(linctl issue create \
    --title "Automated Issue from Agent" \
    --team "ENG" \
    --description "This issue was created by an AI agent" \
    --json)

# Extract issue ID for further operations
ISSUE_ID=$(echo "$ISSUE_RESULT" | jq -r '.data.issueCreate.issue.id')

# Add comment to the issue
linctl comment create "$ISSUE_ID" \
    --body "Agent workflow completed successfully" \
    --json
```

## Environment Configuration

### Required Environment Variables

```bash
# OAuth Authentication (Required)
export LINEAR_CLIENT_ID="your-oauth-client-id"
export LINEAR_CLIENT_SECRET="your-oauth-client-secret"

# Actor Attribution (Recommended)
export LINEAR_DEFAULT_ACTOR="AI Agent"
export LINEAR_DEFAULT_AVATAR_URL="https://example.com/agent-avatar.png"

# Optional Configuration
export LINEAR_BASE_URL="https://api.linear.app"
export LINEAR_SCOPES="read,write,issues:create,comments:create"
```

### Container Environment Setup

#### Docker
```dockerfile
FROM alpine:latest

# Install linctl
RUN apk add --no-cache curl jq
RUN curl -L https://github.com/your-org/linctl/releases/latest/download/linctl-linux-amd64 -o /usr/local/bin/linctl
RUN chmod +x /usr/local/bin/linctl

# Set environment variables
ENV LINEAR_CLIENT_ID=""
ENV LINEAR_CLIENT_SECRET=""
ENV LINEAR_DEFAULT_ACTOR="Docker Agent"

# Copy agent scripts
COPY scripts/ /app/scripts/
WORKDIR /app

CMD ["./scripts/agent-workflow.sh"]
```

#### Docker Compose
```yaml
version: '3.8'
services:
  linear-agent:
    build: .
    environment:
      - LINEAR_CLIENT_ID=${LINEAR_CLIENT_ID}
      - LINEAR_CLIENT_SECRET=${LINEAR_CLIENT_SECRET}
      - LINEAR_DEFAULT_ACTOR=Compose Agent
      - LINEAR_DEFAULT_AVATAR_URL=https://example.com/compose-avatar.png
    volumes:
      - ./scripts:/app/scripts:ro
```

#### Kubernetes
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: linear-oauth-secret
type: Opaque
stringData:
  client-id: "your-oauth-client-id"
  client-secret: "your-oauth-client-secret"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: linear-agent
spec:
  replicas: 1
  selector:
    matchLabels:
      app: linear-agent
  template:
    metadata:
      labels:
        app: linear-agent
    spec:
      containers:
      - name: agent
        image: your-registry/linear-agent:latest
        env:
        - name: LINEAR_CLIENT_ID
          valueFrom:
            secretKeyRef:
              name: linear-oauth-secret
              key: client-id
        - name: LINEAR_CLIENT_SECRET
          valueFrom:
            secretKeyRef:
              name: linear-oauth-secret
              key: client-secret
        - name: LINEAR_DEFAULT_ACTOR
          value: "Kubernetes Agent"
```

## JSON Output and Parsing

### Command Output Format

All linctl commands support `--json` flag for structured output:

```bash
# Authentication status
linctl auth status --json
{
  "authenticated": true,
  "method": "oauth",
  "user": {
    "name": "Agent User",
    "email": "agent@example.com"
  },
  "token_expires_at": "2024-01-15T10:30:00Z",
  "scopes": ["read", "write", "issues:create", "comments:create"],
  "suggestions": []
}

# Issue creation
linctl issue create --title "Test" --team "ENG" --json
{
  "success": true,
  "data": {
    "issueCreate": {
      "success": true,
      "issue": {
        "id": "LIN-123",
        "title": "Test",
        "url": "https://linear.app/company/issue/LIN-123",
        "createdAt": "2024-01-01T10:00:00Z"
      }
    }
  }
}

# Error response
{
  "success": false,
  "error": {
    "code": "AUTHENTICATION_ERROR",
    "message": "OAuth token expired",
    "details": {
      "suggestion": "Run 'linctl auth refresh' to refresh your token"
    }
  }
}
```

### JSON Parsing Examples

#### Using jq (Recommended)
```bash
# Extract issue ID
ISSUE_ID=$(linctl issue create --title "Test" --team "ENG" --json | jq -r '.data.issueCreate.issue.id')

# Check if operation was successful
SUCCESS=$(linctl issue create --title "Test" --team "ENG" --json | jq -r '.success')
if [ "$SUCCESS" != "true" ]; then
    echo "Issue creation failed"
    exit 1
fi

# Extract error message
ERROR_MSG=$(linctl auth status --json | jq -r '.error.message // empty')
if [ -n "$ERROR_MSG" ]; then
    echo "Error: $ERROR_MSG"
    exit 1
fi
```

#### Using Python
```python
import json
import subprocess
import sys

def run_linctl_command(cmd):
    """Run linctl command and return parsed JSON result."""
    try:
        result = subprocess.run(
            cmd, 
            shell=True, 
            capture_output=True, 
            text=True, 
            check=True
        )
        return json.loads(result.stdout)
    except subprocess.CalledProcessError as e:
        print(f"Command failed: {e}")
        sys.exit(1)
    except json.JSONDecodeError as e:
        print(f"Failed to parse JSON: {e}")
        sys.exit(1)

# Check authentication
auth_status = run_linctl_command("linctl auth status --json")
if not auth_status.get("authenticated"):
    print("Not authenticated")
    sys.exit(1)

# Create issue
issue_result = run_linctl_command(
    "linctl issue create --title 'Python Agent Issue' --team 'ENG' --json"
)

if issue_result.get("success"):
    issue_id = issue_result["data"]["issueCreate"]["issue"]["id"]
    print(f"Created issue: {issue_id}")
else:
    print("Failed to create issue")
    sys.exit(1)
```

#### Using Node.js
```javascript
const { execSync } = require('child_process');

function runLinctlCommand(cmd) {
    try {
        const output = execSync(cmd, { encoding: 'utf8' });
        return JSON.parse(output);
    } catch (error) {
        console.error('Command failed:', error.message);
        process.exit(1);
    }
}

// Check authentication
const authStatus = runLinctlCommand('linctl auth status --json');
if (!authStatus.authenticated) {
    console.error('Not authenticated');
    process.exit(1);
}

// Create issue
const issueResult = runLinctlCommand(
    "linctl issue create --title 'Node.js Agent Issue' --team 'ENG' --json"
);

if (issueResult.success) {
    const issueId = issueResult.data.issueCreate.issue.id;
    console.log(`Created issue: ${issueId}`);
} else {
    console.error('Failed to create issue');
    process.exit(1);
}
```

## Exit Codes and Error Handling

### Exit Code Standards

linctl follows standard Unix exit code conventions:

- **0**: Success
- **1**: General error (authentication, network, etc.)
- **2**: Misuse of shell command (invalid arguments)
- **3**: Configuration error
- **4**: Permission denied
- **5**: Resource not found

### Error Handling Patterns

#### Bash Script Error Handling
```bash
#!/bin/bash
set -e  # Exit on any error
set -o pipefail  # Exit on pipe failures

# Function to handle errors
handle_error() {
    local exit_code=$?
    echo "Error occurred with exit code: $exit_code" >&2
    
    case $exit_code in
        1) echo "General error - check authentication and network" >&2 ;;
        2) echo "Invalid command arguments" >&2 ;;
        3) echo "Configuration error - check environment variables" >&2 ;;
        4) echo "Permission denied - check OAuth scopes" >&2 ;;
        5) echo "Resource not found" >&2 ;;
        *) echo "Unknown error" >&2 ;;
    esac
    
    exit $exit_code
}

# Set error trap
trap handle_error ERR

# Check authentication with proper error handling
if ! linctl auth status --json >/dev/null 2>&1; then
    echo "Authentication failed. Please run: linctl auth login --oauth" >&2
    exit 1
fi

# Perform operations with error checking
RESULT=$(linctl issue create --title "Test Issue" --team "ENG" --json)
if ! echo "$RESULT" | jq -e '.success' >/dev/null; then
    ERROR_MSG=$(echo "$RESULT" | jq -r '.error.message // "Unknown error"')
    echo "Issue creation failed: $ERROR_MSG" >&2
    exit 1
fi

echo "Operation completed successfully"
```

#### Python Error Handling
```python
import json
import subprocess
import sys
import logging

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class LinctlError(Exception):
    """Custom exception for linctl errors."""
    def __init__(self, message, exit_code=None, details=None):
        super().__init__(message)
        self.exit_code = exit_code
        self.details = details

def run_linctl_command(cmd, check_success=True):
    """Run linctl command with comprehensive error handling."""
    try:
        result = subprocess.run(
            cmd,
            shell=True,
            capture_output=True,
            text=True,
            timeout=30  # 30 second timeout
        )
        
        # Parse JSON output
        try:
            data = json.loads(result.stdout)
        except json.JSONDecodeError:
            if result.returncode != 0:
                raise LinctlError(
                    f"Command failed: {result.stderr}",
                    exit_code=result.returncode
                )
            raise LinctlError("Failed to parse JSON output")
        
        # Check for application-level errors
        if check_success and not data.get("success", True):
            error_info = data.get("error", {})
            raise LinctlError(
                error_info.get("message", "Unknown error"),
                details=error_info.get("details")
            )
        
        return data
        
    except subprocess.TimeoutExpired:
        raise LinctlError("Command timed out")
    except subprocess.CalledProcessError as e:
        raise LinctlError(f"Command failed with exit code {e.returncode}")

def main():
    try:
        # Check authentication
        auth_status = run_linctl_command("linctl auth status --json", check_success=False)
        if not auth_status.get("authenticated"):
            logger.error("Not authenticated. Run: linctl auth login --oauth")
            sys.exit(1)
        
        logger.info(f"Authenticated as: {auth_status['user']['name']}")
        
        # Create issue
        issue_result = run_linctl_command(
            "linctl issue create --title 'Python Agent Issue' --team 'ENG' --json"
        )
        
        issue_id = issue_result["data"]["issueCreate"]["issue"]["id"]
        logger.info(f"Created issue: {issue_id}")
        
    except LinctlError as e:
        logger.error(f"Linctl error: {e}")
        if e.details:
            logger.error(f"Details: {e.details}")
        sys.exit(e.exit_code or 1)
    except Exception as e:
        logger.error(f"Unexpected error: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
```

## Agent Workflow Examples

### Issue Management Workflow

```bash
#!/bin/bash
# Agent workflow for issue management

set -e

# Configuration
TEAM_ID="ENG"
ACTOR_NAME="Issue Management Agent"

# Function to create issue with error handling
create_issue() {
    local title="$1"
    local description="$2"
    
    local result=$(linctl issue create \
        --title "$title" \
        --team "$TEAM_ID" \
        --description "$description" \
        --actor "$ACTOR_NAME" \
        --json)
    
    if echo "$result" | jq -e '.success' >/dev/null; then
        echo "$result" | jq -r '.data.issueCreate.issue.id'
    else
        echo "Failed to create issue: $(echo "$result" | jq -r '.error.message')" >&2
        return 1
    fi
}

# Function to add comment
add_comment() {
    local issue_id="$1"
    local comment_body="$2"
    
    linctl comment create "$issue_id" \
        --body "$comment_body" \
        --actor "$ACTOR_NAME" \
        --json >/dev/null
}

# Main workflow
main() {
    echo "Starting issue management workflow..."
    
    # Create issue
    ISSUE_ID=$(create_issue "Automated Bug Report" "This issue was created by an automated agent")
    echo "Created issue: $ISSUE_ID"
    
    # Add initial comment
    add_comment "$ISSUE_ID" "Agent has started processing this issue"
    
    # Simulate work
    sleep 2
    
    # Add progress comment
    add_comment "$ISSUE_ID" "Analysis complete. Issue has been categorized and assigned priority."
    
    echo "Workflow completed for issue: $ISSUE_ID"
}

main "$@"
```

### Monitoring and Alerting Workflow

```python
#!/usr/bin/env python3
"""
Linear monitoring agent that creates issues for system alerts.
"""

import json
import subprocess
import sys
import time
from datetime import datetime
from typing import Dict, List, Optional

class LinearMonitoringAgent:
    def __init__(self):
        self.team_id = "OPS"
        self.actor_name = "Monitoring Agent"
        self.avatar_url = "https://example.com/monitoring-avatar.png"
    
    def run_command(self, cmd: str) -> Dict:
        """Run linctl command and return parsed result."""
        try:
            result = subprocess.run(
                cmd,
                shell=True,
                capture_output=True,
                text=True,
                check=True
            )
            return json.loads(result.stdout)
        except (subprocess.CalledProcessError, json.JSONDecodeError) as e:
            print(f"Command failed: {e}", file=sys.stderr)
            sys.exit(1)
    
    def create_alert_issue(self, alert: Dict) -> str:
        """Create an issue for a system alert."""
        title = f"🚨 {alert['severity'].upper()}: {alert['title']}"
        description = f"""
**Alert Details:**
- **Service**: {alert['service']}
- **Severity**: {alert['severity']}
- **Time**: {alert['timestamp']}
- **Description**: {alert['description']}

**Metrics:**
{self._format_metrics(alert.get('metrics', {}))}

**Runbook**: {alert.get('runbook_url', 'N/A')}
        """.strip()
        
        cmd = f"""linctl issue create \
            --title '{title}' \
            --team '{self.team_id}' \
            --description '{description}' \
            --actor '{self.actor_name}' \
            --avatar-url '{self.avatar_url}' \
            --json"""
        
        result = self.run_command(cmd)
        return result['data']['issueCreate']['issue']['id']
    
    def update_issue_status(self, issue_id: str, status: str, comment: str):
        """Update issue status and add comment."""
        # Add comment
        comment_cmd = f"""linctl comment create '{issue_id}' \
            --body '{comment}' \
            --actor '{self.actor_name}' \
            --json"""
        
        self.run_command(comment_cmd)
        
        # Update status if needed
        if status:
            update_cmd = f"""linctl issue update '{issue_id}' \
                --state '{status}' \
                --json"""
            self.run_command(update_cmd)
    
    def _format_metrics(self, metrics: Dict) -> str:
        """Format metrics for display."""
        if not metrics:
            return "No metrics available"
        
        formatted = []
        for key, value in metrics.items():
            formatted.append(f"- **{key}**: {value}")
        
        return "\\n".join(formatted)
    
    def process_alerts(self, alerts: List[Dict]):
        """Process a list of alerts."""
        for alert in alerts:
            try:
                issue_id = self.create_alert_issue(alert)
                print(f"Created alert issue: {issue_id}")
                
                # Add initial processing comment
                self.update_issue_status(
                    issue_id,
                    "In Progress",
                    "🤖 Alert received and issue created. Monitoring agent is analyzing..."
                )
                
            except Exception as e:
                print(f"Failed to process alert {alert.get('id', 'unknown')}: {e}", file=sys.stderr)

def main():
    # Example alerts (in real scenario, these would come from monitoring system)
    alerts = [
        {
            "id": "alert-001",
            "title": "High CPU Usage",
            "service": "web-server-01",
            "severity": "warning",
            "timestamp": datetime.now().isoformat(),
            "description": "CPU usage has exceeded 80% for the last 5 minutes",
            "metrics": {
                "cpu_usage": "85%",
                "memory_usage": "60%",
                "load_average": "2.5"
            },
            "runbook_url": "https://wiki.company.com/runbooks/high-cpu"
        }
    ]
    
    agent = LinearMonitoringAgent()
    agent.process_alerts(alerts)

if __name__ == "__main__":
    main()
```

### CI/CD Integration Workflow

```yaml
# GitHub Actions workflow
name: Linear Integration
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  linear-integration:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Setup linctl
      run: |
        curl -L https://github.com/your-org/linctl/releases/latest/download/linctl-linux-amd64 -o /usr/local/bin/linctl
        chmod +x /usr/local/bin/linctl
    
    - name: Configure Linear OAuth
      env:
        LINEAR_CLIENT_ID: ${{ secrets.LINEAR_CLIENT_ID }}
        LINEAR_CLIENT_SECRET: ${{ secrets.LINEAR_CLIENT_SECRET }}
        LINEAR_DEFAULT_ACTOR: "GitHub Actions"
        LINEAR_DEFAULT_AVATAR_URL: "https://github.com/github.png"
      run: |
        # Verify authentication
        linctl auth status --json
    
    - name: Create deployment issue
      if: github.ref == 'refs/heads/main'
      env:
        LINEAR_CLIENT_ID: ${{ secrets.LINEAR_CLIENT_ID }}
        LINEAR_CLIENT_SECRET: ${{ secrets.LINEAR_CLIENT_SECRET }}
        LINEAR_DEFAULT_ACTOR: "GitHub Actions"
      run: |
        ISSUE_ID=$(linctl issue create \
          --title "Deployment: ${{ github.sha }}" \
          --team "OPS" \
          --description "Automated deployment from commit ${{ github.sha }}" \
          --json | jq -r '.data.issueCreate.issue.id')
        
        echo "DEPLOYMENT_ISSUE_ID=$ISSUE_ID" >> $GITHUB_ENV
    
    - name: Update deployment status
      if: success() && github.ref == 'refs/heads/main'
      env:
        LINEAR_CLIENT_ID: ${{ secrets.LINEAR_CLIENT_ID }}
        LINEAR_CLIENT_SECRET: ${{ secrets.LINEAR_CLIENT_SECRET }}
        LINEAR_DEFAULT_ACTOR: "GitHub Actions"
      run: |
        linctl comment create "$DEPLOYMENT_ISSUE_ID" \
          --body "✅ Deployment completed successfully" \
          --json
        
        linctl issue update "$DEPLOYMENT_ISSUE_ID" \
          --state "Done" \
          --json
```

## Performance and Rate Limiting

### Rate Limiting Considerations

Linear API has rate limits that agents should respect:
- **OAuth tokens**: 1000 requests per hour per token
- **Burst limit**: 100 requests per minute
- **GraphQL complexity**: Complex queries count as multiple requests

### Agent Rate Limiting Strategies

```bash
# Simple rate limiting with sleep
rate_limited_command() {
    local cmd="$1"
    local delay="${2:-1}"  # Default 1 second delay
    
    $cmd
    sleep "$delay"
}

# Batch operations to reduce API calls
batch_issue_creation() {
    local issues=("$@")
    
    for issue in "${issues[@]}"; do
        rate_limited_command "linctl issue create --title '$issue' --team 'ENG' --json" 2
    done
}
```

### Monitoring Agent Performance

```python
import time
import logging
from functools import wraps

def monitor_performance(func):
    """Decorator to monitor function performance."""
    @wraps(func)
    def wrapper(*args, **kwargs):
        start_time = time.time()
        try:
            result = func(*args, **kwargs)
            duration = time.time() - start_time
            logging.info(f"{func.__name__} completed in {duration:.2f}s")
            return result
        except Exception as e:
            duration = time.time() - start_time
            logging.error(f"{func.__name__} failed after {duration:.2f}s: {e}")
            raise
    return wrapper

@monitor_performance
def create_issue_with_monitoring(title, team):
    """Create issue with performance monitoring."""
    return run_linctl_command(f"linctl issue create --title '{title}' --team '{team}' --json")
```

## Security Best Practices for Agents

### Credential Management

```bash
# Use secure credential storage
# Option 1: Environment variables (basic)
export LINEAR_CLIENT_ID="$(cat /secure/path/client-id)"
export LINEAR_CLIENT_SECRET="$(cat /secure/path/client-secret)"

# Option 2: External secret management
LINEAR_CLIENT_ID="$(vault kv get -field=client_id secret/linear)"
LINEAR_CLIENT_SECRET="$(vault kv get -field=client_secret secret/linear)"

# Option 3: Kubernetes secrets
LINEAR_CLIENT_ID="$(cat /var/secrets/linear/client-id)"
LINEAR_CLIENT_SECRET="$(cat /var/secrets/linear/client-secret)"
```

### Token Security

```python
import os
import stat
import tempfile

def secure_token_handling():
    """Example of secure token handling in agents."""
    # Ensure token files have restricted permissions
    token_file = os.path.expanduser("~/.config/linctl/token.json")
    if os.path.exists(token_file):
        # Set file permissions to 600 (owner read/write only)
        os.chmod(token_file, stat.S_IRUSR | stat.S_IWUSR)
    
    # Use temporary files for sensitive operations
    with tempfile.NamedTemporaryFile(mode='w', delete=True) as temp_file:
        # Write sensitive data to temporary file
        temp_file.write("sensitive data")
        temp_file.flush()
        
        # Use the temporary file
        # File is automatically deleted when context exits
```

### Network Security

```bash
# Verify SSL certificates
export LINEAR_VERIFY_SSL=true

# Use secure networks only
if [[ "$NETWORK_TYPE" != "secure" ]]; then
    echo "Error: Agent must run on secure network" >&2
    exit 1
fi

# Implement connection timeouts
export LINEAR_TIMEOUT=30
```

## Troubleshooting Agent Issues

### Common Agent Problems

#### 1. Authentication Failures
```bash
# Debug authentication
linctl auth status --json | jq '.'

# Check environment variables
env | grep LINEAR_

# Test OAuth flow
linctl auth refresh --json
```

#### 2. JSON Parsing Errors
```bash
# Validate JSON output
linctl issue list --json | jq '.' >/dev/null && echo "Valid JSON" || echo "Invalid JSON"

# Debug with verbose output
LINEAR_DEBUG=true linctl issue list --json
```

#### 3. Rate Limiting
```bash
# Check for rate limit errors
if linctl issue create --title "Test" --team "ENG" --json | grep -q "rate_limit"; then
    echo "Rate limited - waiting..."
    sleep 60
fi
```

### Agent Debugging Tools

```bash
#!/bin/bash
# Agent debugging script

debug_agent() {
    echo "=== Agent Debug Information ==="
    
    echo "Environment Variables:"
    env | grep LINEAR_ | sed 's/CLIENT_SECRET=.*/CLIENT_SECRET=***HIDDEN***/'
    
    echo -e "\nAuthentication Status:"
    linctl auth status --json | jq '.'
    
    echo -e "\nToken Information:"
    if [ -f ~/.config/linctl/token.json ]; then
        echo "Token file exists"
        ls -la ~/.config/linctl/token.json
    else
        echo "No token file found"
    fi
    
    echo -e "\nNetwork Connectivity:"
    curl -s -o /dev/null -w "%{http_code}" https://api.linear.app/graphql || echo "Connection failed"
    
    echo -e "\nLinctl Version:"
    linctl --version 2>/dev/null || echo "Version not available"
}

debug_agent
```

## Next Steps

After implementing agent integration:

1. **Monitor Performance**: Set up monitoring for agent operations using the [Monitoring and Logging](monitoring-logging.md) guide
2. **Security Review**: Follow the [Security Best Practices](security-best-practices.md) for production deployment
3. **Troubleshooting**: Refer to the [Troubleshooting Guide](troubleshooting-guide.md) for common issues
4. **API Reference**: Review the [API Reference](api-reference.md) for advanced usage patterns

## Support

For agent-specific issues:
- Check agent logs for error messages
- Verify environment variable configuration
- Test authentication separately from agent workflows
- Use debug mode for detailed troubleshooting information