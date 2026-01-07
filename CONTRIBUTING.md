# Contributing to OpenNHP

Thank you for your interest in contributing to OpenNHP! This document provides guidelines and information for contributors.

## Code of Conduct

Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md) to maintain a welcoming and inclusive community.

## Getting Started

### Prerequisites

- **Go 1.24+** - Required for building the project
- **Make** - Build automation
- **Git** - Version control
- **golangci-lint** (optional) - For code linting

### Clone and Build

```bash
# Clone the repository
git clone https://github.com/OpenNHP/opennhp.git
cd opennhp

# Initialize dependencies
make init

# Build all binaries
make
```

### Project Structure

```
opennhp/
├── nhp/                 # Core NHP protocol library
│   ├── core/            # Protocol implementation
│   ├── common/          # Shared types and utilities
│   ├── utils/           # Helper functions
│   ├── plugins/         # Plugin system
│   └── test/            # Unit tests
├── endpoints/           # Network daemon implementations
│   ├── agent/           # NHP Agent (client)
│   ├── server/          # NHP Server
│   ├── ac/              # Access Controller
│   ├── db/              # Database component
│   └── kgc/             # Key Generation Center
├── examples/            # Example implementations
│   ├── server_plugin/   # Example server plugin
│   └── client_sdk/      # SDK usage examples
├── docs/                # Documentation
├── docker/              # Docker configurations
└── release/             # Build outputs
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with race detection
make test-race

# Run tests with coverage
make coverage
```

### Code Style

- Use `gofmt` for code formatting: `make fmt`
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Write clear, descriptive commit messages
- Add tests for new functionality
- Document exported functions and types

## Development Workflow

### Making Changes

1. **Fork** the repository on GitHub
2. **Clone** your fork locally
3. **Create a branch** for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```
4. **Make your changes** and commit them
5. **Test** your changes:
   ```bash
   make test
   make fmt
   ```
6. **Push** to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```
7. **Open a Pull Request** on GitHub

### Commit Messages

Write clear, concise commit messages that explain the "why" behind your changes:

```
feat(server): add health check endpoint

Add /health and /ready endpoints for Kubernetes liveness
and readiness probes.
```

Use conventional commit prefixes:
- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `test:` - Test additions or modifications
- `refactor:` - Code refactoring
- `ci:` - CI/CD changes
- `build:` - Build system changes

### Pull Request Guidelines

- Keep PRs focused on a single change
- Update documentation if needed
- Add tests for new functionality
- Ensure all tests pass
- Request review from maintainers

## Reporting Issues

### Bug Reports

When reporting bugs, please include:

- Go version (`go version`)
- Operating system and version
- Steps to reproduce the issue
- Expected vs actual behavior
- Relevant logs or error messages

### Feature Requests

For feature requests, please:

- Check existing issues to avoid duplicates
- Clearly describe the use case
- Explain the proposed solution

## Security Issues

For security vulnerabilities, please see our [Security Policy](SECURITY.md) and report issues privately.

## Getting Help

- **GitHub Issues** - For bugs and feature requests
- **Discussions** - For questions and community support
- **Documentation** - See the [docs/](docs/) directory

## License

By contributing to OpenNHP, you agree that your contributions will be licensed under the [Apache 2.0 License](LICENSE).
