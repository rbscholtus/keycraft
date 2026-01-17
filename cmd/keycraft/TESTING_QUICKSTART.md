# Testing Quick Start Guide

## Running Tests

### All Tests
```bash
go test ./cmd/keycraft/
```

### Verbose Output
```bash
go test ./cmd/keycraft/ -v
```

### Specific Test Categories

**Flag Tests:**
```bash
go test ./cmd/keycraft/ -run "TestAllShared|TestSharedFlag|TestCommandSpecificFlags|TestFlagDefaults"
```

**Configuration Loading Tests:**
```bash
go test ./cmd/keycraft/ -run "TestTargetLoads|TestWeights|TestCorpus_|TestLoadLayout"
```

**Command Tests:**
```bash
go test ./cmd/keycraft/ -run "TestCorpusCommand|TestViewCommand|TestAnalyseCommand|TestRankCommand|TestFlipCommand|TestOptimiseCommand|TestGenerateCommand"
```

**Individual Command:**
```bash
go test ./cmd/keycraft/ -run "TestCorpusCommand"
go test ./cmd/keycraft/ -run "TestGenerateCommand"
```

### Coverage Reports

**Basic Coverage:**
```bash
go test ./cmd/keycraft/ -cover
```

**Detailed Coverage:**
```bash
go test ./cmd/keycraft/ -coverprofile=coverage.out
go tool cover -func=coverage.out
```

**HTML Coverage Report:**
```bash
go test ./cmd/keycraft/ -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Test Statistics

Current status:
- **48** test functions
- **72+** individual test cases
- **75%** pass rate
- **100%** flag coverage
- **100%** command coverage

## Quick Test Results

```bash
# Get summary
go test ./cmd/keycraft/ | tail -3

# Count pass/fail
go test ./cmd/keycraft/ -v 2>&1 | grep -E "^--- (PASS|FAIL):" | sort | uniq -c
```

## Test Files

| File | Purpose | Tests |
|------|---------|-------|
| [testing_utils.go](testing_utils.go) | Shared utilities and fixtures | N/A |
| [flags_test.go](flags_test.go) | Flag validation | 6 |
| [config_loading_test.go](config_loading_test.go) | Config hierarchy | 11 |
| [commands_test.go](commands_test.go) | Command validation | 15 |
| [helpers_test.go](helpers_test.go) | Helper functions (existing) | 16 |

## Debugging Failed Tests

**View detailed failure output:**
```bash
go test ./cmd/keycraft/ -v -run TestTargetLoads_FromFile
```

**Focus on a specific sub-test:**
```bash
go test ./cmd/keycraft/ -v -run "TestGenerateCommand_BuildInput/rowstag"
```

## Adding New Tests

### 1. Add to appropriate file
- Flags → [flags_test.go](flags_test.go)
- Config → [config_loading_test.go](config_loading_test.go)
- Commands → [commands_test.go](commands_test.go)

### 2. Use test utilities
```go
func TestNewFeature(t *testing.T) {
    // Setup test directories
    origLayoutDir, origCorpusDir, origConfigDir := setupTestDirs(t)
    defer restoreTestDirs(origLayoutDir, origCorpusDir, origConfigDir)

    // Create test files
    writeTestCorpus(t, corpusDir, "default.txt")
    writeTestLayout(t, layoutDir, "test.klf", minimalLayoutContent)

    // Your test logic here
}
```

### 3. Use table-driven tests
```go
tests := []struct {
    name    string
    input   string
    wantErr bool
}{
    {"valid input", "value", false},
    {"invalid input", "", true},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // Test logic
    })
}
```

## CI Integration

Add to `.github/workflows/test.yml`:
```yaml
- name: Run CLI tests
  run: go test ./cmd/keycraft/ -v -cover

- name: Generate coverage
  run: |
    go test ./cmd/keycraft/ -coverprofile=coverage.out
    go tool cover -func=coverage.out
```

## Common Issues

### Issue: "invalid memory address"
**Cause:** Command context not properly initialized
**Fix:** Use full `app.Run()` instead of calling validation functions directly

### Issue: "no such file or directory"
**Cause:** Test files not created in temp directories
**Fix:** Call `writeTestLayout()` or `writeTestCorpus()` before testing

### Issue: "invalid metric"
**Cause:** Using non-existent metric names
**Fix:** Use actual metric names: SFB, LSB, FSB, etc.

## Documentation

For detailed information, see:
- [TESTING_PLAN.md](TESTING_PLAN.md) - Comprehensive testing strategy
- [TEST_IMPLEMENTATION_SUMMARY.md](TEST_IMPLEMENTATION_SUMMARY.md) - Implementation details and results
