# Vegeta — Agent Guide

## Project Overview
**Vegeta** is a versatile HTTP load testing tool built in Go. It is designed to attack HTTP services with a constant request rate.

## Technology Stack
- **Language**: Go
- **Build System**: `Makefile` (wraps `go build`, `go generate`)
- **Dependencies**: Managed via `go.mod`.

## Project Structure
- `attack.go`: Core attack logic.
- `lib/`: Core library code.
- `internal/`: Internal utilities.
- `main.go`: CLI entry point.

## Build and Test Commands
- **Build**: `make vegeta` (Generates binary with version/commit info).
- **Generate**: `make generate` (Runs `go generate`).
- **Test**: `go test ./...`

## Development Conventions
- **Cross-platform**: Handles windows/non-windows builds (`*_windows.go`, `*_nonwindows.go`).
- **Static Linking**: The Makefile builds a statically linked binary by default.
- **Flags**: Uses `flags.go` for CLI argument parsing.

## Task Implementation
1. **Analyze Requirements**: Refer to `README.md` for detailed feature specifications and system design.
2. **Implementation**: Modify source code in the respective directories (e.g., `src/`, `internal/`).
3. **Verification**: Run provided build and test commands (see above) to ensure correctness.
4. **Push Changes**:
   - Commit changes: `git commit -m "feat: implement <feature>"`
   - Push to remote: `git push origin <branch-name>`
