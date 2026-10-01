#!/usr/bin/env python3
"""Go test execution utilities.

Provides functions for running Go unit tests and race detection.
"""

import os
import subprocess
import sys
from pathlib import Path
from typing import List

# Packages per `go test` invocation in run_unit_tests.
BATCH_SIZE = 25
# Concurrent package builds/test runs within a batch. Each test binary links
# the msgraph beta SDK; the runner default (one per CPU) exhausts memory on
# ubuntu-24.04-arm and the runner is shut down mid-batch.
PARALLELISM = 2


def run_unit_tests(packages: List[str], output_dir: str = "coverage") -> Path:
    """Run Go unit tests with coverage profiling.
    
    Args:
        packages: List of package paths to test.
        output_dir: Directory for coverage output files.
    
    Returns:
        Path to merged coverage file.
    """
    print("\n" + "="*60)
    print("📊 Running Unit Tests with Coverage")
    print("="*60)
    
    coverage_dir = Path(output_dir)
    coverage_dir.mkdir(parents=True, exist_ok=True)
    
    merged_file = coverage_dir / "unit-coverage.txt"
    coverage_files = []
    failed_batches = []
    
    # Test packages in batches: one `go test` per batch builds and runs the
    # batch's packages in parallel. One invocation per package serialises
    # every compile and link, which pushes large PRs past the job timeout.
    batches = [packages[i:i + BATCH_SIZE] for i in range(0, len(packages), BATCH_SIZE)]

    for idx, batch in enumerate(batches, 1):
        coverage_file = coverage_dir / f"batch_{idx}.out"

        print(f"\n[{idx}/{len(batches)}] Testing {len(batch)} package(s):")
        for package in batch:
            print(f"   - {package}")

        cmd = [
            "go", "test", "-v",
            "-p", str(PARALLELISM),
            # No symbol table / DWARF: much less linker memory and time.
            "-ldflags=-s -w",
            # HashiCorp runs TestAcc* whenever TF_ACC is non-empty, even "0".
            "-skip=^TestAcc",
            f"-coverprofile={coverage_file}",
            "-covermode=atomic",
            *[f"./{package}" for package in batch]
        ]

        result = subprocess.run(
            cmd,
            env={"TF_ACC": "0", **os.environ},
            check=False
        )
        if result.returncode != 0:
            failed_batches.append(idx)

        if coverage_file.exists():
            coverage_files.append(coverage_file)
            print(f"✅ Coverage generated for batch {idx}")
        else:
            print(f"⚠️  No coverage file for batch {idx}")
    
    # Merge coverage files
    print(f"\n📊 Merging {len(coverage_files)} coverage file(s)...")
    _merge_coverage_files(coverage_files, merged_file)
    
    print(f"✅ Merged coverage file: {merged_file}")
    if failed_batches:
        # Fork: propagate test failures instead of passing on coverage alone.
        sys.exit(f"❌ go test failed in batch(es) {failed_batches}; see the --- FAIL lines above")
    return merged_file


def run_race_detection(packages: List[str]) -> int:
    """Run Go race detector on specified packages.
    
    Args:
        packages: List of package paths to test.
    
    Returns:
        Exit code: 0 if all tests pass, 1 if any fail.
    """
    print("\n" + "="*60)
    print("🔍 Running Race Detection Tests")
    print("="*60)
    
    has_failures = False
    
    for idx, package in enumerate(packages, 1):
        print(f"\n[{idx}/{len(packages)}] Testing: {package}")
        
        cmd = ["go", "test", "-v", "-race", f"./{package}"]
        
        result = subprocess.run(
            cmd,
            env={"TF_ACC": "0", **os.environ},
            check=False
        )
        
        if result.returncode != 0:
            has_failures = True
            print(f"❌ Race detection failed in {package}")
        else:
            print(f"✅ Race detection passed in {package}")
    
    return 1 if has_failures else 0


def _merge_coverage_files(coverage_files: List[Path], output_file: Path) -> None:
    """Merge multiple Go coverage files into one.
    
    Args:
        coverage_files: List of coverage file paths.
        output_file: Path for merged output file.
    """
    with open(output_file, 'w', encoding='utf-8') as out_f:
        out_f.write("mode: atomic\n")
        for cov_file in coverage_files:
            with open(cov_file, 'r', encoding='utf-8') as in_f:
                for line in in_f:
                    if not line.startswith('mode:'):
                        out_f.write(line)
