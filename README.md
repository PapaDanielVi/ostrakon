# ostrakon

[![Go Report Card](https://goreportcard.com/badge/github.com/PapaDanielVi/ostrakon)](https://goreportcard.com/report/github.com/PapaDanielVi/ostrakon)
[![Go Version](https://img.shields.io/github/go-mod/go-version/PapaDanielVi/ostrakon?color=blue&logo=go&logoColor=white)](https://github.com/PapaDanielVi/ostrakon)
[![Release](https://img.shields.io/github/v/release/PapaDanielVi/ostrakon?color=brightgreen&logo=github&label=release)](https://github.com/PapaDanielVi/ostrakon/releases)
[![License](https://img.shields.io/github/license/PapaDanielVi/ostrakon?color=orange&logo=mit)](https://github.com/PapaDanielVi/ostrakon/blob/main/LICENSE)
[![CI](https://github.com/PapaDanielVi/ostrakon/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/PapaDanielVi/ostrakon/actions/workflows/ci.yml)

A secure CLI tool for managing secrets in a private GitHub repository with client-side encryption.

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [GitHub Token Setup](#github-token-setup)
- [Quick Start](#quick-start)
- [Commands](#commands)
- [Profiles](#profiles)
- [Security](#security)
- [Examples](#examples)
- [Requirements](#requirements)
- [Contributing](#contributing)

## Overview

In ancient Athens, an ostrakon was a piece of pottery used as a scrap for everyday writing, tax receipts, and secret voting. It was the ancient world's equivalent of a Gist or a pastebin.

Ostrakon provides client-side encryption, ensuring your secrets are encrypted before they leave your computer. This approach provides several key advantages:

- **Zero-knowledge architecture**: Your master password is never sent to GitHub
- **End-to-end encryption**: All secrets are encrypted locally before upload
- **Password-derived keys**: Uses Argon2id for secure key derivation
- **Authenticated encryption**: AES-256-GCM ensures integrity and confidentiality

## Installation

### Homebrew (macOS)

```bash
brew tap PapaDanielVi/homebrew-tap
brew install ostrakon
```

### Go Install

```bash
go install github.com/PapaDanielVi/ostrakon@latest
```

## GitHub Token Setup

Before initializing, create a [GitHub Fine-Grained Personal Access Token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens/creating-a-fine-grained-personal-access-token) with read and write permissions for the target repository:

1. Go to **Settings → Developer settings → Personal access tokens → Fine-grained tokens** in GitHub
2. Click **Generate new token**
3. Configure the token:
   - **Token name**: Give it a descriptive name (e.g., "ostrakon-secrets")
   - **Token expiration**: Set an appropriate expiration (recommend 90 days or less)
   - **Repository permissions**:
     - Select **Only select repositories** and choose your vault repository
     - **Contents**: Read and write
4. Click **Generate token** and copy the token immediately (you won't see it again)

> **Note**: Fine-grained tokens with "Contents: Read and write" are preferred over classic tokens with `repo` scope because they provide more limited access to just the specific repository.

## Quick Start

1. **Initialize** your vault:
   ```bash
   ostrakon init
   ```
   This will prompt you for:
   - Repository URL (e.g., `https://github.com/owner/repo` or `owner/repo`)
   - GitHub Personal Access Token (with Contents read/write permission)
   - Master password for encryption

2. **Add a secret**:
   ```bash
   ostrakon add secret.txt
   # or with piped data
   echo "abc123" | ostrakon add API_KEY
   ```

3. **List secrets**:
   ```bash
   ostrakon ls
   ```

4. **Get a secret**:
   ```bash
   ostrakon get secret.txt
   ```

## Commands

### `init [--no-keyring]`
Initialize Ostrakon by setting up the GitHub repository and master password.
- `--no-keyring`: Do not store master password in keyring (will prompt for password on each operation)

The master password is automatically stored in the OS keyring during init for convenience. Use `--no-keyring` to opt out of this behavior.

### `add <file> [-n name] [-p profile]`
Encrypt and upload a file to the vault. Reads from stdin if data is piped.
- `-n, --name`: Name for the file in the vault
- `-p, --profile`: Profile/namespace for the file

The master password is retrieved from the keyring (stored during init). If not in keyring, you'll be prompted.

### `get <name> [-o file] [-p profile]`
Download and decrypt a secret from the vault.
- `-o, --output`: Output file (default: stdout)
- `-p, --profile`: Profile/namespace for the file

**Note**: Always prompts for master password for security (keyring is ignored for read operations).

### `ls [<path>] [--tree] [--profile profile]`
List all secrets stored in the vault.
- `<path>`: Filter by path prefix (supports hierarchical paths like `prod/db/`)
- `-t, --tree`: Show secrets as a tree structure
- `-p, --profile`: Filter by profile/namespace (deprecated, use path instead)

### `shred <name> | --all`
Securely delete a secret by overwriting it with random data before deletion. This provides deniability by destroying the encrypted file's history.
- `--all`: Reset all Ostrakon data (clear keychain)

### `write <name> [-o file]`
Download and decrypt a secret to a file. Always prompts for master password.
- `-o, --output`: Output file (default: uses the secret name as filename)

### `edit <name>`
Download, decrypt, edit in `$EDITOR`, and re-encrypt a secret. Always prompts for master password before editing.

### `run <script> [-e secret]`
Execute a local script using decrypted secrets as environment variables.
- `-e, --env`: Secret name(s) to inject as environment variables

## Profiles

Profiles provide namespacing for your secrets. Use the `-p` flag to organize secrets:

```bash
ostrakon add config.env -p production
ostrakon get config.env -p production
ostrakon ls -p production
```

## Security

- All secrets are encrypted client-side before being sent to GitHub
- Master password is stored in the OS keyring during `init` for convenience on write operations (`add`, `shred`)
- For read operations (`get`, `run`), the master password is **always prompted** for maximum security
- Use `ostrakon init --no-keyring` to opt out of keyring storage for master password
- Tokens and passwords are stored in the OS keychain (Keychain on macOS, Credential Manager on Windows, Secret Service on Linux)
- `shred` provides secure deletion by overwriting files before removal

## Examples

### Basic Usage

```bash
# Initialize your vault
ostrakon init
# Repository URL: https://github.com/owner/repo
# Enter your GitHub Fine-Grained Personal Access Token (with Contents: Read and write permission)
# Enter master password

# Add a secret file
ostrakon add secret.txt

# Add with a custom name
ostrakon add -n myapp.env config.env

# Add with piped data (useful for env files)
echo "DATABASE_URL=postgres://localhost:5432/mydb" | ostrakon add db.env

# List all secrets
ostrakon ls

# Get and decrypt a secret
ostrakon get secret.txt
ostrakon get secret.txt -o output.txt

# Write secret to a file
ostrakon write secret.txt
ostrakon write prod/db/password -o ./secrets/db.env

# Delete a secret
ostrakon shred secret.txt
```

### Hierarchical Paths (Recommended)

Secrets can be organized using paths with slashes:

```bash
# Add secrets with hierarchical paths
ostrakon add prod/database/password
ostrakon add prod/api/key
ostrakon add staging/database/password

# List all secrets
ostrakon ls

# List only production secrets
ostrakon ls prod/

# List as a tree
ostrakon ls --tree

# Get a secret by its full path
ostrakon get prod/database/password
```

### Using Profiles

Profiles provide namespacing for different environments:

```bash
# Add production secrets
ostrakon add -p production database.env
ostrakon add -p production api-key.txt

# Add development secrets
ostrakon add -p development database.env

# List production secrets only
ostrakon ls -p production

# Get a production secret
ostrakon get database.env -p production
```

### Running Scripts with Secrets

Execute scripts with decrypted secrets as environment variables:

```bash
# Create a script that uses secrets
cat > deploy.sh << 'EOF'
#!/bin/bash
echo "Deploying to $ENVIRONMENT with API key: $API_KEY"
# Your deployment logic here
EOF

# Run the script with secrets injected
ostrakon run deploy.sh -e api-key.txt -e environment.env
```

### Master Password Behavior

The master password workflow is designed for security and convenience:

- **During init**: You're prompted for a master password, which is stored in the OS keyring by default
- **On `add` / `shred` / `edit`**: Password is retrieved from keyring silently (no prompt)
- **On `get` / `run` / `write`**: Always prompts for password (keyring is ignored for security)

### Tree View and Search

List secrets with different views:

```bash
# List all secrets
ostrakon ls

# List with tree view
ostrakon ls --tree

# Filter by path prefix
ostrakon ls prod/

# Search for secrets containing a pattern
ostrakon ls --search api
```

### Writing and Editing Secrets

```bash
# Write a secret to a file
ostrakon write prod/db/password

# Write with custom output filename
ostrakon write prod/db/password -o ./secrets/db.env

# Edit a secret in your editor
ostrakon edit prod/db/password
# Opens $EDITOR (or vim) to edit the decrypted secret
```

```bash
# Initialize with keyring storage (default)
ostrakon init
# Password will be stored in keyring for add/shred operations

# Initialize without keyring storage
ostrakon init --no-keyring
# You'll be prompted for password on each operation

# Adding secrets uses keyring silently
ostrakon add secret.txt  # No password prompt (uses keyring)

# Getting secrets always asks for password
ostrakon get secret.txt
# Enter master password: [always prompts]
```

## Requirements

- Go 1.21 or later (if installing via `go install`)
- GitHub Fine-Grained Personal Access Token with **Contents: Read and write** permission for your vault repository
- A private GitHub repository for storing secrets

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## Support

If you find Ostrakon useful, please consider [starring the repository](https://github.com/PapaDanielVi/ostrakon/stargazers) to show your support!
