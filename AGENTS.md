# Ostrakon - Agent Guide

## Development Commands

```bash
make build     # Build binary (ldflags set Version from git)
make test      # Run all tests with -race
make lint      # golangci-lint
make fmt       # gofmt
make vet       # go vet
make all       # fmt → vet → lint → test → build
make mock      # Regenerate mocks with `go tool mockgen`
```

Single test: `go test -v -run TestName ./pkg/...`

## Architecture

- `cmd/ostrakon/commands/` - Cobra commands (add, get, init, ls, run, shred, write, edit)
- `pkg/vault/` - `Provider` interface for vault operations
- `pkg/github/` - GitHub client implementing `Provider`
- `pkg/crypto/` - Argon2id + AES-256-GCM encryption
- `pkg/keyring/` - OS keychain (service: "ostrakon")

Encryption flow: plaintext → `crypto.Encrypt(password)` → base64 → upload to `contents/` path.

## Keyring Behavior (Easy to Miss)

- `init`: password stored in keyring by default
- `add`/`shred`/`edit`: silent keyring access (no prompt)
- `get`/`run`/`write`: **always prompt** (keyring ignored for security)

## Go Code Standards

- `context.Context` as first parameter on all functions
- Use constants for values used more than once
- `errors.New` without formatting; `fmt.Errorf` only with placeholders
- Never ignore errors
- Test files use external package: `<package>_test`
- Use `t.Parallel()` in table-driven tests