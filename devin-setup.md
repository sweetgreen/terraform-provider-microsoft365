# Devin Setup Instructions for terraform-provider-microsoft365

## Step 2: Configure Secrets
**Skip this step** - No secrets needed for initial setup. Authentication credentials will be configured later via environment variables.

## Step 3: Install Dependencies

```bash
# Install Go 1.24.1 (required by go.mod)
wget https://go.dev/dl/go1.24.1.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.24.1.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc

# Verify Go installation
go version

# Install required tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest

# Install Terraform (if not already installed)
wget -O- https://apt.releases.hashicorp.com/gpg | sudo gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/hashicorp.list
sudo apt update && sudo apt install terraform

# Install mitmproxy (optional, for API debugging)
# Note: mitmproxy is optional and only needed for API debugging
# If you encounter installation issues, you can skip this tool
sudo apt-get update
sudo apt-get install -y python3-pip

# Update pip first to avoid compatibility issues
pip3 install --upgrade pip setuptools

# Try installing mitmproxy (skip if it fails)
pip3 install mitmproxy || echo "Warning: mitmproxy installation failed. This is optional for API debugging."
```

## Step 4: Maintain Dependencies

```bash
cd ~/repos/terraform-provider-microsoft365 && make deps
```

## Step 5: Setup Lint

```bash
# Run linter using the Makefile (recommended)
make lint

# Note: If you encounter "sort-results should be 'true' to use sort-order" error,
# the configuration has already been fixed in the repository.
# The error was caused by missing 'sort-results: true' in .golangci.yml

# For quick lint checks without building
golangci-lint run

# If you still see configuration errors, ensure you have the latest version:
golangci-lint --version
```

## Step 6: Setup Tests

```bash
# Run unit tests (no real API calls)
make unittest

# Run specific unit test
make unittest TEST=UserResource

# Run acceptance tests (requires credentials, see additional notes)
# TF_ACC=1 make acctest TEST=UserResource
```

## Step 7: Setup Local App

This is a Terraform provider, not a traditional app. To use it locally:

```bash
# Build and install the provider
make install

# The provider will be installed to $GOPATH/bin/terraform-provider-microsoft365
# You'll need to configure Terraform to use the local provider by creating a .terraformrc file
```

## Step 8: Additional Notes

### Important Information
- This is a **Terraform Provider** for Microsoft 365, not a web application
- It requires Go 1.24.1 (as specified in go.mod)
- The provider uses Microsoft Graph API (both v1.0 and beta endpoints)
- Always run `make precommit` before submitting changes - this runs clean, build, lint, unittest, docs generation, and terraform formatting

### Development Workflow
1. **Building**: Use `make build` to compile the provider to `./bin/`
2. **Installing**: Use `make install` to install locally for testing
3. **Testing**: 
   - Unit tests: `make unittest` (mocked, no API calls)
   - Acceptance tests: `make acctest` (real API calls, requires auth)
4. **Linting**: Use `make lint` to run golangci-lint
5. **Documentation**: Use `make userdocs` to generate Terraform docs

### Key Makefile Commands
- `make deps` - Install/update Go dependencies
- `make build` - Build provider binary
- `make install` - Build and install provider locally
- `make clean` - Clean build artifacts
- `make unittest` - Run unit tests
- `make acctest` - Run acceptance tests (requires auth)
- `make lint` - Run golangci-lint
- `make userdocs` - Generate documentation
- `make precommit` - Full pre-commit check (recommended before commits)

### Authentication Setup (for acceptance tests)
To run acceptance tests, you'll need to set up Microsoft 365 authentication:

1. Register an application in Azure AD
2. Set environment variables:
   ```bash
   export ARM_TENANT_ID="your-tenant-id"
   export ARM_CLIENT_ID="your-client-id"
   export ARM_CLIENT_SECRET="your-client-secret"
   # Or use other auth methods (certificate, managed identity, etc.)
   ```

### Resource Development Pattern
When developing new resources, follow the established pattern:
- `resource.go` - Resource metadata and schema
- `crud.go` - Create, Read, Update, Delete operations
- `model.go` - Terraform state models
- `construct.go` - Build API requests
- `state.go` - Map API responses to state

### Testing Your Changes
1. Build and install: `make install`
2. Create a test Terraform configuration
3. Configure local provider override in `~/.terraformrc`:
   ```hcl
   provider_installation {
     dev_overrides {
       "deploymenttheory/microsoft365" = "/home/ubuntu/go/bin"
     }
     direct {}
   }
   ```
4. Run `terraform plan` and `terraform apply` with your test configuration

### Debugging
- Use `TF_LOG=DEBUG` for verbose Terraform output
- Use `make netdump` to start mitmproxy for API debugging
- Check Graph X-Ray (https://graphxray.merill.net/) to understand Microsoft API calls

### Important Files
- `Makefile` - All build and test commands
- `go.mod` - Go module dependencies
- `internal/provider/` - Provider initialization
- `internal/services/` - Resource implementations
- `internal/client/` - API client setup
- `docs/development/guide.md` - Development guide

### Before Making Changes
1. Always pull latest changes: `git pull`
2. Run `make deps` to ensure dependencies are up to date
3. Create a new branch for your work
4. Follow the existing code patterns and conventions
5. Add tests for new functionality
6. Run `make precommit` before committing

### Common Issues
- If you see Go version errors, ensure Go 1.24.1 is installed
- For "command not found" errors, ensure `$GOPATH/bin` is in your PATH
- For auth errors in tests, check your environment variables
- For API errors, use Graph X-Ray to understand the correct API format