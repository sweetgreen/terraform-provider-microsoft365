# GitHub Actions Runner Optimization

This document outlines the runner optimization strategy for ARM compatibility and improved performance.

## Runner Selection Strategy

### 1. Release Workflows (High CPU/Memory)
- **Runner**: `ubuntu-latest-8-cores` 
- **Use Case**: GoReleaser builds, cross-platform compilation
- **Rationale**: 8-core ARM64 runners provide optimal performance for multi-architecture builds with parallelism=8
- **Timeout**: 120 minutes (extended for large builds)

### 2. Testing Workflows (Moderate CPU)
- **Runner**: `ubuntu-latest-4-cores`
- **Use Case**: Unit tests, integration tests, linting
- **Rationale**: 4-core ARM64 provides balanced performance for parallel test execution
- **Timeout**: 30-60 minutes (reduced from previous values)

### 3. Validation Workflows (Standard)
- **Runner**: `ubuntu-latest`
- **Use Case**: Documentation generation, PR validation, lightweight tasks
- **Rationale**: Standard runners sufficient for non-intensive tasks
- **Timeout**: 15-30 minutes

## Key Optimizations

### ARM64 Architecture Support
- All runners configured for ARM64 compatibility
- Environment variables set for consistent architecture targeting:
  ```yaml
  env:
    GOOS: linux
    GOARCH: arm64
  ```

### Go Build Optimizations
- **CGO_ENABLED=0**: Static linking for cross-platform compatibility
- **GOMAXPROCS**: Set to match runner core count (4 or 8)
- **Cache optimization**: Enabled with `cache-dependency-path: 'go.sum'`
- **Parallelism**: Increased to match available cores

### Resource Allocation
- **Release builds**: 8 cores, 120min timeout, parallelism=8
- **Unit tests**: 4 cores, 60min timeout, parallel test execution
- **Linting**: 4 cores, standard timeout, optimized cache usage
- **Documentation**: Standard runner, lightweight processing

## Performance Improvements

### Before Optimizations
- Mixed Windows/Ubuntu runners (Windows deprecated)
- Inconsistent resource allocation
- Cache disabled in release workflows
- Fixed parallelism (4) regardless of runner size

### After Optimizations
- Consistent ARM64 architecture across all workflows
- Resource allocation matched to workload requirements
- Optimized caching and build flags
- Dynamic parallelism based on runner capabilities

## Build Matrix Considerations

### Cross-Platform Support
The GoReleaser configuration builds for:
- Linux: amd64, 386, arm, arm64
- Darwin: amd64, arm64
- FreeBSD: amd64, 386, arm, arm64

### ARM64 Benefits
- Better price/performance ratio
- Consistent with modern MacOS development (M-series chips)
- Improved energy efficiency
- Native ARM64 compilation for Apple Silicon

## Environment Variables

### Global Optimizations
```yaml
env:
  CGO_ENABLED: 0          # Static linking
  GOOS: linux             # Target OS
  GOARCH: arm64           # Target architecture
  GOMAXPROCS: {cores}     # Match runner cores
```

### Workflow-Specific
- **Release**: Extended timeout, maximum parallelism
- **Tests**: Race detection enabled (CGO_ENABLED=1)
- **Validation**: Standard configuration

## Runner Labels by Job Type

| Job Type | Runner | Cores | Memory | Use Case |
|----------|--------|-------|--------|----------|
| Release | ubuntu-latest-8-cores | 8 | 32GB | GoReleaser, cross-compilation |
| Testing | ubuntu-latest-4-cores | 4 | 16GB | Unit/integration tests |
| Linting | ubuntu-latest-4-cores | 4 | 16GB | Static analysis |
| Validation | ubuntu-latest | 2 | 7GB | Pre-release checks |
| Documentation | ubuntu-latest | 2 | 7GB | Doc generation |

## Cost Optimization

- Reduced timeouts prevent runaway jobs
- Appropriate runner sizing prevents over-provisioning
- Caching reduces redundant operations
- Parallel execution reduces wall-clock time

## Reliability Improvements

- ARM64 compatibility ensures consistent execution
- Extended timeouts for complex builds prevent false failures
- Proper resource allocation reduces resource contention
- Static linking improves cross-platform reliability