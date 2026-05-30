# Contributing to Ostrakon

Thank you for your interest in contributing to Ostrakon!

## Development Setup

### Prerequisites

- Go 1.22 or later
- Git
- A GitHub account with a Personal Access Token (PAT) with `repo` scope

### Clone and Setup

```bash
# Clone the repository
git clone https://github.com/PapaDanielVi/ostrakon.git
cd ostrakon

# Install dependencies
go mod tidy

# Build the project
go build ./...

# Run tests
go test ./...
```

### Running Locally

```bash
# Build the binary
go build -o ostrakon ./cmd/ostrakon

# Initialize Ostrakon (first time only)
./ostrakon init

# Add a secret
./ostrakon add mysecret.txt

# List secrets
./ostrakon ls

# Get a secret
./ostrakon get mysecret.txt
```

## Code Quality

We enforce the following code quality standards:

### Linting

We use `golangci-lint` for linting. Run it before submitting:

```bash
golangci-lint run ./...
```

### Testing

All new features should include tests. Run the test suite:

```bash
go test -v -cover ./...
```

### Security Scanning

We use `gosec` for security scanning:

```bash
gosec ./...
```

## Pull Request Process

1. **Fork the repository** and create a branch from `main`
2. **Name your branch descriptively**: e.g., `feat/add-tls-support`, `fix/crypto-key-derivation`
3. **Make your changes** following the code style guidelines
4. **Add tests** for any new functionality
5. **Run the linter and tests** locally to ensure they pass
6. **Submit a pull request** with a clear description of the changes
7. **Wait for review** - we aim to review within 7 days

## Commit Message Format

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types:
- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `test`: Adding or updating tests
- `chore`: Changes to the build process or auxiliary tools

Example:
```
feat(crypto): add support for Argon2id key derivation

Implements Argon2id for stronger key derivation to protect against
GPU/ASIC attacks. Uses standard parameters: 64MB memory, 3 iterations.

Closes #42
```

## Reporting Issues

When reporting issues, please include:

- **Go version**: `go version`
- **Operating system**: macOS, Linux, Windows
- **Ostrakon version**: `./ostrakon --version` (if built)
- **Steps to reproduce**: Clear steps to reproduce the issue
- **Expected behavior**: What you expected to happen
- **Actual behavior**: What actually happened
- **Error messages**: Any error messages or logs

## Code of Conduct

Please be respectful and constructive in all interactions. We follow the [Contributor Covenant](https://www.contributor-covenant.org/).

## Questions?

Feel free to open an issue for questions, or reach out to the maintainer directly.