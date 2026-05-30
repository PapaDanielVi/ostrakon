# ostrakon

A secure CLI tool for managing secrets in a private GitHub repository with client-side encryption.

## Overview

In ancient Athens, an ostrakon was a piece of pottery used as a scrap for everyday writing, tax receipts, and secret voting. It was the ancient world's equivalent of a Gist or a pastebin.

Ostrakon provides client-side encryption, ensuring your secrets are encrypted before they leave your computer.

## Installation

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

## Requirements

- Go 1.21 or later
- GitHub Personal Access Token with `repo` scope
- A private GitHub repository for storing secrets