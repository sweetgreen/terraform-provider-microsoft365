# Running GitHub Actions Locally with act

This guide explains how to test GitHub Actions workflows locally using [act](https://github.com/nektos/act), saving time and avoiding the need to push changes to GitHub for testing.

## Prerequisites

### Install act

#### macOS (using Homebrew)
```bash
brew install act
```

#### Linux
```bash
curl -s https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash
```

#### Windows (using Chocolatey)
```powershell
choco install act-cli
```

## Configuration

### First-time Setup

When running act for the first time, it will ask you to choose a default Docker image. For compatibility:

```bash
act --container-architecture linux/amd64
```

### Docker Requirements

Ensure Docker is running before using act:
```bash
docker ps
```

## Usage

### Basic Commands

#### List available workflows
```bash
act -l
```

#### Run a specific workflow
```bash
act -W .github/workflows/unit-tests.yml
```

#### Run a specific job
```bash
act -j test
```

#### Run workflows triggered by specific events
```bash
# Pull request event
act pull_request

# Push event
act push

# Workflow dispatch
act workflow_dispatch
```

### Advanced Usage

#### Dry run (show what would be executed)
```bash
act -n
```

#### Use specific Docker image
```bash
act --container-architecture linux/amd64 -P ubuntu-latest=catthehacker/ubuntu:act-latest
```

#### Pass environment variables
```bash
act -e .env.local
```

#### Pass secrets
```bash
act -s GITHUB_TOKEN="your-token-here"
```

## Makefile Integration

Use the provided makefile commands for common act operations:

```bash
# Run unit tests locally
make act-test

# Run linting locally
make act-lint

# Run all CI checks locally
make act-ci

# Dry run to see what would execute
make act-dry-run
```

## Container Architecture

The project uses AMD64 architecture with catthehacker images for better compatibility:

```bash
act --container-architecture linux/amd64 -P ubuntu-latest=catthehacker/ubuntu:act-latest
```

The `.actrc` file in the project root configures these defaults automatically.

## Troubleshooting

### Common Issues

1. **Docker not running**
   ```
   Error: Cannot connect to the Docker daemon
   ```
   Solution: Start Docker Desktop or Docker daemon

2. **Architecture mismatch**
   ```
   exec format error
   ```
   Solution: Use `--container-architecture linux/amd64` flag

3. **Large workflow files**
   ```
   Error: Workflow too large
   ```
   Solution: Use `-W` to specify a single workflow file

4. **Missing secrets**
   ```
   Error: Required secret not found
   ```
   Solution: Create `.secrets` file or use `-s` flag

### Resource Limits

For resource-intensive workflows, increase Docker resources:

```bash
# Increase memory limit
act --container-options "--memory=8g"

# Use multiple CPUs
act --container-options "--cpus=4"
```

## Best Practices

1. **Test locally first**: Always run `act` before pushing to verify workflow syntax
2. **Use dry run**: Run with `-n` flag to preview actions without execution
3. **Match CI environment**: Use the same runner images as defined in workflows
4. **Cache dependencies**: Use act's caching to speed up repeated runs
5. **Keep secrets secure**: Never commit `.secrets` files to version control

## Example Workflows

### Testing Unit Tests Locally
```bash
# Run unit tests with act
act -W .github/workflows/unit-tests.yml -e .env.test

# With specific Go version
act -W .github/workflows/unit-tests.yml --env GOVERSION=1.22.5
```

### Testing Release Workflow
```bash
# Dry run release workflow
act -W .github/workflows/provider-release.yml -n

# Run with secrets
act -W .github/workflows/provider-release.yml -s GPG_PRIVATE_KEY="$GPG_KEY"
```

## Integration with VS Code

Add to `.vscode/tasks.json`:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Act: Run Unit Tests",
      "type": "shell",
      "command": "make act-test",
      "group": "test",
      "presentation": {
        "reveal": "always",
        "panel": "new"
      }
    }
  ]
}
```

## Additional Resources

- [act Documentation](https://github.com/nektos/act)
- [act Architecture Support](https://github.com/nektos/act/issues/1178)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)