# Release-Please Setup Guide

This repository uses [release-please](https://github.com/googleapis/release-please) for automated releases via the [googleapis/release-please-action](https://github.com/googleapis/release-please-action). This guide explains how to configure the necessary permissions.

## GitHub Actions Permission Error

If you see this error:
```
Error: release-please failed: GitHub Actions is not permitted to create or approve pull requests.
```

You need to configure repository permissions. Choose one of the solutions below:

## Solution 1: Enable GitHub Actions Permissions (Recommended)

1. Go to your repository on GitHub
2. Navigate to **Settings** → **Actions** → **General**
3. Under "Workflow permissions":
   - Select **"Read and write permissions"**
   - Check **"Allow GitHub Actions to create and approve pull requests"**
4. Click **Save**

This allows the default `GITHUB_TOKEN` to create release PRs.

## Solution 2: Use Personal Access Token

If you prefer not to enable broad GitHub Actions permissions:

1. **Create a Personal Access Token**:
   - Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
   - Click "Generate new token (classic)"
   - Select scopes: `repo` (full control of private repositories)
   - Copy the generated token

2. **Add as Repository Secret**:
   - Go to your repository → Settings → Secrets and variables → Actions
   - Click "New repository secret"
   - Name: `RELEASE_PLEASE_TOKEN`
   - Value: Your PAT from step 1
   - Click "Add secret"

The workflow is already configured to use `RELEASE_PLEASE_TOKEN` if available, falling back to `GITHUB_TOKEN`.

## Solution 3: GitHub App (Advanced)

For organizations preferring app-based authentication:

1. Create a GitHub App with these permissions:
   - Repository permissions: Contents (write), Pull requests (write), Metadata (read)
2. Install the app on your repository
3. Configure the workflow to use app authentication

## How Release-Please Works

Once configured, release-please will:

1. **Monitor commits** on the main branch for conventional commit messages
2. **Create release PRs** automatically when it detects releasable changes
3. **Generate changelogs** from commit messages
4. **Create GitHub releases** when release PRs are merged
5. **Trigger builds** to upload release assets

## Conventional Commits

Use these commit prefixes to trigger releases:

- `feat:` - New feature (minor version bump)
- `fix:` - Bug fix (patch version bump)
- `feat!:` or `fix!:` - Breaking change (major version bump)
- `chore:`, `docs:`, `style:`, `refactor:`, `test:` - No version bump

## Testing the Setup

After configuration, test by pushing a commit:

```bash
git commit -m "feat: add new feature"
git push origin main
```

Release-please should create a PR within a few minutes.

## Troubleshooting

- **No PR created**: Check that conventional commit format is used
- **Permission errors**: Verify GitHub Actions permissions are enabled
- **Token issues**: Ensure `RELEASE_PLEASE_TOKEN` has correct scopes
- **Workflow not running**: Check that the workflow file is on the main branch

## Current Configuration

- **Release type**: Go
- **Current version**: 0.2.0
- **Config file**: `.release-please-config.json`
- **Manifest file**: `.release-please-manifest.json`
- **Changelog sections**: Features, Bug Fixes, Performance, etc.
