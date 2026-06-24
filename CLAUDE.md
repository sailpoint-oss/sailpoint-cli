# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SailPoint CLI (`sail`) is a Go CLI tool for interacting with SailPoint Identity Security Cloud (ISC) tenants. Built with [cobra](https://github.com/spf13/cobra) for command structure and [viper](https://github.com/spf13/viper) for configuration management. The binary is named `sail`.

## Build & Test Commands

```bash
# Build and install locally
make install          # builds `sail` binary and installs to /usr/local/bin/sail

# Run all tests
make test             # go test -v -count=1 ./...

# Run a single test
go test -v -count=1 -run TestNewConnCreateCmd ./cmd/connector/

# Run tests with coverage report
make test-report

# Run tests with race detection
make test-race

# Regenerate mocks (requires mockgen)
make mocks

# Clean build artifacts
make clean
```

## Architecture

### Entry Point

`main.go` → initializes config via `internal/config.InitConfig()`, then creates and executes the root cobra command from `cmd/root/root.go`.

### Command Structure

Each command lives in `cmd/<command>/` as its own package. Commands follow a consistent pattern:

- **Parent command**: a `New<Name>Command()` function returns `*cobra.Command`, adds subcommands
- **Subcommands**: unexported `new<Action>Command()` functions (e.g., `newListCommand()`, `newCreateCommand()`)
- **Help text**: many commands embed `.md` files via `//go:embed` and parse them with `util.ParseHelp()` which extracts `==Long==...====` and `==Example==...====` sections

Two API client patterns coexist:
1. **SailPoint Go SDK** (`sailpoint-oss/golang-sdk/v2`): used by most commands (workflow, search, cluster, spconfig, etc.) via `config.InitAPIClient(experimental bool)`
2. **Internal HTTP client** (`internal/client/client.go`): used by the `connector` and `api` commands. The `connector` package injects this client via function parameters for testability.

### Internal Packages

- `internal/config/` — config management via viper; reads `~/.sailpoint/config.yaml`; manages environments, auth types (PAT/OAuth), token lifecycle
- `internal/auth/` — authentication logic (PAT login, OAuth flow, token caching via keyring)
- `internal/keyring/` — secrets storage abstraction (system keyring with fallback)
- `internal/client/` — low-level HTTP client (`Client` interface) with auth token injection
- `internal/tui/` — interactive prompts using [charmbracelet/huh](https://github.com/charmbracelet/huh) (confirm, input, password, list selection)
- `internal/templates/` — search, export, and report template loading (built-in + user-defined from `~/.sailpoint/`)
- `internal/mocks/` — generated gomock mocks for `Client` and `Terminal` interfaces

### Testing Pattern

Tests use `gomock` for mocking. The connector tests demonstrate the standard pattern:
1. Create `gomock.Controller`
2. Create `mocks.NewMockClient(ctrl)` with expected calls
3. Build the command with the mock injected
4. Execute with `cmd.Execute()` and assert output

When adding a new root-level subcommand, update `numRootSubcommands` in `cmd/root/root_test.go`.

### Configuration

User config lives at `~/.sailpoint/config.yaml`. Supports multiple named environments with an `activeenvironment` key. Auth types: `pat` (Personal Access Token) or `oauth`. Environment variables prefixed with `SAIL_` are auto-bound by viper.

### Release

Uses GoReleaser (`.goreleaser.yaml`). Version is set in `cmd/root/root.go` and injected via ldflags at build time. Builds for Linux, macOS, and Windows. Distributed via Homebrew tap, deb/rpm packages, and zip/tar.gz archives.
