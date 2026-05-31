# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Professional README badges for CI, release, license, and Go version
- Repository topics for improved SEO discoverability

## [0.0.2-beta] - 2026-05-31

### Added
- Global master password support via `set-global-master` command
- Password prompts use `golang.org/x/term.ReadPassword` for secure input
- Keychain storage for passwords and tokens (macOS Keychain, Windows Credential Manager, Linux Secret Service)

### Changed
- Improved goreleaser configuration for multi-platform builds
- Enhanced shell completion scripts for bash, zsh, fish, and pwsh

## [0.0.1-beta] - 2026-05-30

### Added
- Initial release of Ostrakon
- `init` command to set up GitHub repository and master password
- `add` command to encrypt and upload secrets
- `get` command to download and decrypt secrets
- `ls` command to list stored secrets
- `rm` command to delete secrets
- `shred` command for secure deletion with history destruction
- `run` command to execute scripts with decrypted secrets as environment variables
- Profile support for namespacing secrets
- Argon2id key derivation for strong encryption
- AES-256-GCM authenticated encryption
- Homebrew installation support