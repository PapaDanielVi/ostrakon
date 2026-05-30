# Security Policy

## Reporting a Vulnerability

Ostrakon takes security seriously. If you discover a security vulnerability, please report it responsibly.

**Please DO NOT open a public GitHub issue for security vulnerabilities.**

Instead, please report them directly to the maintainer:

- Email: [Email the maintainer](mailto:papadanielvi@users.noreply.github.com)
- GitHub: [Open a private security advisory](https://github.com/PapaDanielVi/ostrakon/security/advisories/new)

Please include the following information:
- A clear description of the vulnerability
- Steps to reproduce the issue
- Potential impact of the vulnerability
- Any suggested fixes (optional)

## Response Timeline

- **Acknowledgment**: Within 48 hours
- **Initial Assessment**: Within 7 days
- **Fix Timeline**: We will work with you to determine the best timeline for a fix

## Security Best Practices

When using Ostrakon, follow these security best practices:

1. **Use a strong master password**: Use a unique, strong password that you don't use elsewhere
2. **Protect your PAT**: Your GitHub Personal Access Token should be treated like a password
3. **Enable2FA on GitHub**: This adds an extra layer of protection to your account
4. **Regularly rotate secrets**: If a secret may have been compromised, rotate it immediately
5. **Use `ostrakon shred` for sensitive deletions**: This overwrites files before deletion

## Encryption Details

Ostrakon uses:
- **Argon2id** for key derivation (memory-hard, resistant to GPU/ASIC attacks)
- **AES-256-GCM** for authenticated encryption
- Keys are never stored - only a validation hash of your password

Even if the vault repository is made public or the PAT is stolen, your secrets remain encrypted and secure.