# Contributing to Blockbusterr

Thank you for your interest in contributing to Blockbusterr! This guide will help you get started.

## Ways to Contribute

- **Report bugs** - Found something broken? Let us know!
- **Suggest features** - Have an idea? We'd love to hear it!
- **Submit pull requests** - Code contributions are always welcome
- **Improve documentation** - Help others understand Blockbusterr better
- **Share your config** - Show others how you're using Blockbusterr

## Getting Started

### Prerequisites

- Go 1.25.13 or higher
- Node.js 24 or higher (only when changing frontend assets or documentation)
- Docker (optional, for testing)
- Git

### Development Setup

1. **Fork and clone the repository**
   ```bash
   git clone https://github.com/YOUR_USERNAME/blockbusterr.git
   cd blockbusterr
   ```

2. **Install dependencies**
   ```bash
   go mod download
   npm ci
   ```

3. **Copy config template**
   ```bash
   cp config/config.example.yaml config/config.yaml
   ```

4. **Build and run**
   ```bash
   make build
   make dev
   ```

   The application will be available at `http://localhost:9090`

### Development Commands

```bash
make help           # Show all available commands
make build          # Build the application binary
make dev            # Run the application in development mode
make test           # Run all tests
make test-verbose   # Run tests with verbose output
make lint           # Run linter
make fmt            # Format code
make check          # Run fmt, vet, and lint
make assets         # Rebuild committed CSS and browser libraries
make assets-check   # Verify committed assets match their pinned sources
```

The application serves compiled files from `web/static` and does not run Node.js in production. Ordinary Go development uses the committed assets. When templates, shared frontend JavaScript, Tailwind configuration, or frontend dependencies change, run `make assets` and commit the generated files.

## Pull Request Process

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Write clear, concise commit messages
   - Follow existing code style and conventions
   - Add tests for new functionality
   - Update documentation if needed

3. **Test your changes**
   ```bash
   make test
   make check
   ```

4. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat: add awesome new feature"
   ```

5. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```

6. **Open a Pull Request**
   - Provide a clear description of the changes
   - Reference any related issues
   - Include screenshots for UI changes
   - Ensure all checks pass

## Reporting Bugs

When reporting bugs, please include:

- **Environment details** (OS, Docker version, Go version)
- **Steps to reproduce** the issue
- **Expected behavior** vs **actual behavior**
- **Logs** (if applicable)
- **Configuration** (sanitize sensitive data!)

Use the [Bug Report template](https://github.com/Mahcks/blockbusterr/issues/new?template=bug_report.yml) when creating an issue.

## Feature Requests

We love new ideas! When requesting features:

- **Check existing issues** to avoid duplicates
- **Describe the problem** you're trying to solve
- **Explain your proposed solution**
- **Consider alternatives** you've thought about
- **Describe any additional context** that might be helpful

Use the [Feature Request template](https://github.com/Mahcks/blockbusterr/issues/new?template=feature_request.md) when creating an issue.

## Project Structure

```
blockbusterr/
├── cmd/                    # Application entrypoints
│   └── app/                # Main application
├── internal/               # Internal application code
│   ├── database/           # Database interactions
│   ├── filters/            # Rule evaluation
│   ├── integrations/       # External service integrations
│   ├── rest/               # REST API handlers
│   ├── scoring/            # Content scoring algorithms
│   └── services/           # Background jobs and services
├── config/                 # Configuration files
├── web/templates/          # HTML templates
└── docs/                   # Documentation site
```

## Testing

- Write tests for new functionality
- Ensure existing tests pass: `make test`
- Run tests with coverage: `make test-coverage`
- Test edge cases and error conditions

## Documentation

If your PR adds or changes functionality:

- Update the relevant documentation in `/docs/src/content/docs/`
- Update code comments for public APIs
- Update the README if needed
- Add examples where appropriate

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers
- Accept constructive criticism
- Focus on what's best for the community
- Show empathy towards others

## Questions?

- Open a [Discussion](https://github.com/mahcks/blockbusterr/discussions)
- Check existing [Issues](https://github.com/mahcks/blockbusterr/issues)
- Read the [Documentation](https://blockbusterr.dev)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for contributing to Blockbusterr!**
