# Ostrakon - Secure Secret Management CLI

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

A Go CLI tool for managing encrypted secrets in a private GitHub repository using client-side encryption with Argon2id key derivation and AES-256-GCM.

## Project Structure

- `cmd/ostrakon/` - Main CLI entry point and Cobra command definitions
  - `commands/` - Individual command implementations (add, get, init, ls, run, shred, write, edit)
- `pkg/crypto/` - Encryption/decryption logic (Argon2id + AES-256-GCM)
- `pkg/keyring/` - OS keychain integration (Keychain/Credential Manager/Secret Service)
- `pkg/config/` - Configuration management and repository settings
- `pkg/vault/` - Vault abstraction defining the `Provider` interface
- `pkg/github/` - GitHub API client implementing the vault provider interface
- `pkg/gitlab/` - GitLab API client implementing the vault provider interface
- `pkg/provider/` - Factory for creating vault provider clients (GitHub/GitLab)

## Development Commands

### Build and Run

```bash
make build           # Build the ostrakon binary
make test            # Run all tests with race detection
go test -v ./...     # Run tests verbosely across all packages
go test -v -run TestEncrypt ./pkg/crypto/  # Run a single test
make lint            # Run golangci-lint
make fmt             # Format code with gofmt
make vet             # Run go vet
make all             # Run fmt, vet, lint, test, and build
make mock            # Regenerate mocks for interfaces
```

### Testing

```bash
go test ./pkg/crypto/... -v -race    # Test crypto package
go test -v -count=1 ./...            # Fresh test run (no cache)
```

## Big-Picture Architecture

### Encryption Flow (Add/Get)

1. **Add**: File content → `crypto.Encrypt(plaintext, password)` → base64-encoded ciphertext → upload to GitHub via `github.Client.UploadFile()` to `contents/<path>`
2. **Get**: Download from GitHub via `github.Client.DownloadFile()` → `crypto.Decrypt(encryptedB64, password)` → plaintext output

The `crypto` package handles all encryption using Argon2id key derivation and AES-256-GCM. Encrypted files are stored in the `contents/` subdirectory of the repository.

### Authentication and Configuration

- `config` package stores GitHub token and repo info in OS keyring under service name "ostrakon"
- Master password hash is stored for validation (actual password stored in keyring when using default mode)
- `github.Client` wraps `go-github` client with OAuth2 token authentication
- All vault operations go through the `vault.Provider` interface, with `github.Client` as the implementation

### Password Keyring Behavior

- During `init`: Master password stored in keyring by default (unless `--no-keyring`)
- `add`/`shred`/`edit`: Use keyring silently (no prompt) for convenience on write operations
- `get`/`run`/`write`: Always prompt for password (keyring ignored for security on read operations)
- This design ensures "zero-knowledge" - the master password is never sent to GitHub

### Command Pattern

Each command file in `cmd/ostrakon/commands/` follows the Cobra pattern:

- Define a `cobra.Command` struct with `Use`, `Short`, `Long`, `RunE`
- Use `config` package to access stored token, repo info, and passwords
- Use `vault.Provider` interface (via `github.Client`) for storage operations

### Security Model

- Fine-grained GitHub tokens with Contents: Read/Write permission (preferred over classic `repo` scope)
- All secrets encrypted client-side before upload to GitHub
- Uses OWASP-recommended Argon2id parameters (3 iterations, 64MB memory, 4 threads)
- AES-256-GCM for authenticated encryption with random nonce per encryption

## 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

## 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

## 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

## 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.
