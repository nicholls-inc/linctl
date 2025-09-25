#!/bin/sh

# GitHub App Configuration - UPDATE THESE
APP_ID="1897420"                                    # Your App ID
INSTALLATION_ID="84347412"                         # Your Installation ID
PRIVATE_KEY_FILE="/usr/local/secrets/ONA_GH_BOT_APP_PRIVATE_KEY.pem"  # File path

# Generate installation token and run gh command
generate_and_run() {
    # Validate configuration
    if [ -z "$APP_ID" ] || [ -z "$INSTALLATION_ID" ]; then
        echo "Error: GITHUB_APP_ID and GITHUB_INSTALLATION_ID must be set" >&2
        exit 1
    fi

    if [ ! -f "$PRIVATE_KEY_FILE" ] || [ ! -r "$PRIVATE_KEY_FILE" ]; then
        echo "Error: Private key file not found or unreadable at $PRIVATE_KEY_FILE" >&2
        exit 1
    fi

    # Generate JWT
    header='{"alg":"RS256","typ":"JWT"}'
    now=$(date +%s)
    iat=$((now - 60))
    exp=$((now + 600))
    payload="{\"iss\":\"$APP_ID\",\"iat\":$iat,\"exp\":$exp}"

    header_b64=$(printf '%s' "$header" | base64 | tr -d '=' | tr '/+' '_-' | tr -d '\n')
    payload_b64=$(printf '%s' "$payload" | base64 | tr -d '=' | tr '/+' '_-' | tr -d '\n')
    signature=$(printf "%s.%s" "$header_b64" "$payload_b64" | \
                openssl dgst -sha256 -sign "$PRIVATE_KEY_FILE" -binary | \
                base64 | tr -d '=' | tr '/+' '_-' | tr -d '\n')

    jwt="$header_b64.$payload_b64.$signature"

    # Get installation token
    token=$(curl -s -X POST \
        -H "Accept: application/vnd.github.v3+json" \
        -H "Authorization: Bearer $jwt" \
        "https://api.github.com/app/installations/$INSTALLATION_ID/access_tokens" | \
        jq -r '.token')

    if [ -z "$token" ] || [ "$token" = "null" ]; then
        echo "Error: Failed to generate installation access token" >&2
        exit 1
    fi

    # Authenticate and run command
    printf '%s\n' "$token" | gh auth login --with-token --hostname github.com
    gh "$@"
}

generate_and_run "$@"
