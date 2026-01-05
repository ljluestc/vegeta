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

## AI Agent Workflow

### 1. Requirements Discovery
- **Primary Source**: `PRD.md` (Always prioritize this if present).
- **Secondary**: `requirements.txt`, `README.md`, or specific task files.
- **Goal**: Understand the full scope before writing code.

### 2. Implementation Protocol
- **Branching**: Work on a dedicated feature branch (e.g., `feat/implementation-details`).
- **Development**:
  - Analyze code structure.
  - Implement changes in `src/` or relevant directories.
  - Adhere to existing code style.
- **Verification**:
  - Run build commands (see above).
  - Run test suite (see above).
  - Ensure no regressions.

### 3. Delivery
- **Commit**: Use conventional commits (e.g., `feat: ...`, `fix: ...`).
- **PR Creation**:
  - Push branch: `git push -u origin <branch-name>`
  - Create a Pull Request against the main branch.
  - Summary: Link to `PRD.md` requirements solved.

## Task Implementation
1. **Analyze Requirements**: Refer to `README.md` for detailed feature specifications and system design.
2. **Implementation**: Modify source code in the respective directories (e.g., `src/`, `internal/`).
3. **Verification**: Run provided build and test commands (see above) to ensure correctness.
4. **Push Changes**:
   - Commit changes: `git commit -m "feat: implement <feature>"`
   - Push to remote: `git push origin <branch-name>`
