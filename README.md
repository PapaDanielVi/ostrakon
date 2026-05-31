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

## Quick Start

1. **Initialize** your vault:
   ```bash
   ostrakon init
   ```
   This will prompt you for:
   - Repository URL (e.g., `https://github.com/owner/repo` or `owner/repo`)
   - GitHub Personal Access Token (with `repo` scope)
   - Master password for encryption

2. **Add a secret**:
   ```bash
   ostrakon add secret.txt
   # or with piped data
   echo "API_KEY=abc123" | ostrakon add
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

### `init`
Initialize Ostrakon by setting up the GitHub repository and master password.

### `add <file> [-n name] [-p profile]`
Encrypt and upload a file to the vault. Reads from stdin if data is piped.
- `-n, --name`: Name for the file in the vault
- `-p, --profile`: Profile/namespace for the file

### `get <name> [-o file] [-p profile]`
Download and decrypt a secret from the vault.
- `-o, --output`: Output file (default: stdout)
- `-p, --profile`: Profile/namespace for the file

### `ls [--profile profile]`
List all secrets stored in the vault.
- `-p, --profile`: Filter by profile/namespace

### `rm <name> [-p profile]`
Delete a secret from the vault. For secure deletion with history destruction, use `shred`.
- `-p, --profile`: Profile/namespace for the file

### `shred <name> | --all`
Securely delete a secret by overwriting it with random data before deletion. This provides deniability by destroying the encrypted file's history.
- `--all`: Reset all Ostrakon data (clear keychain)

### `run <script> [-e secret]`
Execute a local script using decrypted secrets as environment variables.
- `-e, --env`: Secret name(s) to inject as environment variables

### `set-global-master <password>`
Store your master password in the system keychain to avoid repeated prompts. Use this only on trusted machines where you control the keychain.

## Profiles

Profiles provide namespacing for your secrets. Use the `-p` flag to organize secrets:

```bash
ostrakon add config.env -p production
ostrakon get config.env -p production
ostrakon ls -p production
```

## Security

- All secrets are encrypted client-side before being sent to GitHub
- The master password is never stored directly (only a hash for validation, unless you use `set-global-master`)
- Tokens and passwords are stored in the OS keychain (Keychain on macOS, Credential Manager on Windows, Secret Service on Linux)
- `shred` provides secure deletion by overwriting files before removal

## Examples

### Basic Usage

```bash
# Initialize your vault
ostrakon init
# Repository URL: https://github.com/owner/repo
# Enter your GitHub Personal Access Token (with repo scope)
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

# Delete a secret
ostrakon rm secret.txt
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

### Global Master Password

Store your master password in the system keychain to avoid repeated prompts:

```bash
# Set global master password (macOS Keychain, Windows Credential Manager, or Linux Secret Service)
ostrakon set-global-master
# Enter master password (will be stored encrypted in keyring)

# Now password prompts are skipped for add, get, and run commands
ostrakon add secret.txt  # No password prompt needed
ostrakon get secret.txt  # No password prompt needed
```

## Requirements

- Go 1.21 or later (if installing via `go install`)
- GitHub Personal Access Token with `repo` scope
- A private GitHub repository for storing secrets

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## Support

If you find Ostrakon useful, please consider [starring the repository](https://github.com/PapaDanielVi/ostrakon/stargazers) to show your support!