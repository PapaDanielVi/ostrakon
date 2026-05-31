# Ostrakon - Secure Secret Management CLI

A Go CLI tool for managing encrypted secrets in a private GitHub repository using client-side encryption with Argon2id key derivation and AES-256-GCM.

## Project Structure

- `cmd/ostrakon/` - Main CLI entry point and Cobra command definitions
  - `commands/` - Individual command implementations (add, get, init, ls, run, shred, setglobalmaster)
- `pkg/crypto/` - Encryption/decryption logic (Argon2id + AES-256-GCM)
- `pkg/keyring/` - OS keychain integration (Keychain/Credential Manager/Secret Service)
- `pkg/config/` - Configuration management and repository settings
- `pkg/vault/` - Vault abstraction for storage operations
- `pkg/github/` - GitHub API client for repository operations
- `pkg/mocks/` - Mock implementations for testing

## Commands

| Command | Description |
|---------|-------------|
| `init` | Initialize vault with GitHub repo URL and master password |
| `add <file>` | Encrypt and upload a file to the vault |
| `get <name>` | Download and decrypt a secret from the vault |
| `ls` | List all secrets in the vault |
| `shred <name>` | Securely delete a secret (overwrite before deletion) |
| `run <script>` | Execute a script with secrets as environment variables |
| `set-global-master` | Store master password in OS keychain |

## Architecture

- Uses Cobra framework for CLI
- Fine-grained GitHub tokens with Contents: Read/Write permission (preferred over classic `repo` scope)
- All secrets encrypted client-side before upload to GitHub
- Master password stored in OS keychain when `set-global-master` is used

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

## 5. Implementation Notes

### Password Masking (golang.org/x/term)
- Master password prompts now use `golang.org/x/term.ReadPassword` for secure input
- Password is read from stdin without echoing to terminal
- Requires network access to download the module; use `GOSUMDB=off` when building behind restrictive proxies

### Command Structure
- Commands are defined in `cmd/ostrakon/commands/` as separate files
- Shared helper functions `readPassword()`, `readPasswordPrompt()`, `getPassword()` are in `init.go`
- `getPassword()` automatically uses stored global master password if available

### Global Master Password
- Stored in OS keychain under key `global_master_password`
- Managed via `set-global-master` command
- When set, password prompts are skipped for `add`, `get`, and `run` commands
