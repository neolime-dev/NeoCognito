# Contributing to NeoCognito

First of all, thank you for your interest in contributing to NeoCognito! We are happy to have you here. All contributions are welcome, from fixing a simple typo to implementing a new feature.

> Also available in: [Português (pt-BR)](CONTRIBUTING.pt-BR.md)

## Getting Started

1. **Fork the Repository:** Click the "Fork" button in the top-right corner of the repository page on GitHub to create your own copy.
2. **Clone your Fork:** `git clone https://github.com/YOUR-USERNAME/NeoCognito.git`
3. **Create a Branch:** `git checkout -b my-awesome-feature` (use a descriptive name).
4. **Make your Changes:** Implement your feature or bug fix.
5. **Commit your Changes:** `git commit -m "feat: Add awesome feature"` (see our commit message guide below).
6. **Push to your Fork:** `git push origin my-awesome-feature`
7. **Open a Pull Request:** On GitHub, go to the original repository and open a Pull Request with a clear description of your changes.

## Pull Request Guide

- **Keep it focused:** Prefer small, focused pull requests. Do not bundle multiple features into a single PR.
- **Write a good description:** Explain the *what* and the *why* of your changes. If your PR resolves an existing issue, reference it (e.g. `Closes #123`).
- **Clean code:** Make sure your code follows Go conventions. Run `go fmt` and `go vet` before committing.
- **Be patient:** We will do our best to review your PR as quickly as possible.

## Commit Message Convention

We follow a lightweight conventional-commit style to keep the history clean and readable:

```
<type>: <subject>
```

**Common types:**

| Type | When to use |
|------|-------------|
| `feat` | A new feature |
| `fix` | A bug fix |
| `docs` | Documentation changes only |
| `style` | Formatting, missing semicolons, etc. — no logic change |
| `refactor` | Code change that is neither a fix nor a feature |
| `test` | Adding or fixing tests |
| `chore` | Build process, tooling, CI changes |

**Example:** `feat: Add custom theme support to config.toml`

Thank you again for your contribution!
