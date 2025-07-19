# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is the Community Terraform Provider for Microsoft 365, enabling Infrastructure as Code (IaC) management of Microsoft 365 environments including Intune, Teams, Defender, and related services. The provider uses the Microsoft Graph API (both v1.0 and beta endpoints) through the Kiota-generated SDK.

## Development Commands

### Building and Installing
```bash
make deps         # Install/update Go dependencies
make build        # Build provider binary to ./bin/
make install      # Build and install provider locally
make clean        # Clean build artifacts and test cache
```

### Testing
```bash
# Unit tests (mocked, no real API calls)
make unittest                        # Run all unit tests
make unittest TEST=MyTest           # Run specific unit test

# Acceptance tests (real API calls, requires credentials)
make acctest                        # Run all acceptance tests
make acctest TEST=MyTest           # Run specific acceptance test

# Coverage and full test suite
make test                          # Run all tests (unit + acceptance)
make coverage                      # Generate test coverage report

# Running single tests directly
go test -v -run TestUnitUserResource_Create ./internal/services/resources/users/graph_beta/user
TF_ACC=1 go test -v -timeout 30m -run TestAccUserResource_Create ./internal/services/resources/users/graph_beta/user
```

### Development Workflow
```bash
make precommit      # Full pre-commit check (clean, build, lint, test, docs, format)
make lint           # Run golangci-lint
make userdocs       # Generate Terraform registry documentation
make terraformfmt   # Format all .tf files
make netdump        # Start mitmproxy for API debugging
```

## High-Level Architecture

### Directory Structure
- `/internal/client/` - API client setup, authentication, cloud configuration
- `/internal/provider/` - Provider initialization and resource registration
- `/internal/services/` - Resources organized by Microsoft service domain:
  - `device_management/` - Intune device configuration
  - `device_and_app_management/` - Intune app management
  - `groups/`, `users/` - Azure AD resources
  - `identity_and_access/` - Conditional access policies
  - `windows_365/` - Windows 365 Cloud PC
  - `microsoft_teams/` - Teams policies (uses PowerShell)

### Resource Implementation Pattern

Each resource follows a consistent file structure:
- `resource.go` - Resource metadata, schema, and configuration
- `crud.go` - Create, Read, Update, Delete operations
- `model.go` - Terraform state model definitions
- `construct.go` - Build API request objects from Terraform state
- `state.go` - Map API responses to Terraform state
- `modify_plan.go` - Plan modifiers for computed fields (optional)

### API Client Architecture

The provider uses a dual-client approach:
1. **Kiota SDK Clients** - Microsoft's official Graph SDK
   - `KiotaGraphV1Client` - For stable v1.0 endpoints
   - `KiotaGraphBetaClient` - For beta endpoints

2. **Raw HTTP Clients** - For operations not supported by SDK
   - `GraphV1Client` - Raw HTTP for v1.0
   - `GraphBetaClient` - Raw HTTP for beta

### Key Development Patterns

1. **Resource Naming**: `graph_beta_[service]_[resource]` or `graph_[service]_[resource]`

2. **CRUD Functions Should**:
   - Include descriptive comments for logic flow
   - Delegate request construction to separate functions (`constructResource`)
   - Use `crud.HandleTimeout` for operation timeouts
   - Use `ReadWithRetry` after create/update for eventual consistency
   - Handle errors with `errors.HandleGraphError`

3. **Testing Conventions**:
   - Unit tests: `TestUnit[ResourceName]_[Operation]_[Scenario]`
   - Acceptance tests: `TestAcc[ResourceName]_[Operation]_[Scenario]`
   - Test configs in `mocks/terraform/resource_minimal.tf` and `resource_maximal.tf`

4. **Authentication Methods**: Supports 10+ auth methods including client secret, certificate, managed identity, OIDC

5. **Error Handling**: Centralized error handling with permission-aware messages

### Special Considerations

1. **Graph API Versions**: Choose between v1.0 and beta based on feature availability, following Microsoft's portal usage patterns

2. **Settings Catalog**: Intune's settings catalog has dynamic schema requiring special handling

3. **PowerShell Integration**: Some resources (Teams policies) use PowerShell where Graph API isn't available

4. **Assignment Resources**: Many resources have separate assignment resources for managing target groups/users

5. **File Uploads**: Complex scenarios like app package uploads to Azure Storage are supported

### Development Tips

- Use [Graph X-Ray](https://graphxray.merill.net/) to observe Microsoft 365 portal API calls
- Always test with both minimal and maximal configurations
- Follow the existing resource patterns for consistency
- Check both v1.0 and beta endpoints for API availability
- Use the provider's test helpers and mocking utilities