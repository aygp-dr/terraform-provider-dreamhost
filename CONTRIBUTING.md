# Contributing to Terraform Provider DreamHost

Thank you for your interest in contributing to the DreamHost Terraform Provider! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to abide by our code of conduct: be respectful, inclusive, and professional.

## How to Contribute

### Reporting Issues

Before creating an issue, please check if it already exists. When creating a new issue, include:

- Clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Terraform and provider versions
- Relevant configuration (sanitized of sensitive data)
- Error messages and logs

### Suggesting Enhancements

Enhancement suggestions are welcome! Please create an issue with:

- Clear use case description
- Proposed solution
- Alternative solutions considered
- Potential implementation approach

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Follow the coding standards** (see below)
3. **Write tests** for new functionality
4. **Update documentation** as needed
5. **Ensure all tests pass**
6. **Submit a pull request** with a clear description

## Development Setup

### Prerequisites

- Go 1.19 or later
- Terraform 1.0 or later
- DreamHost API key for testing

### Local Development

1. Clone your fork:
```bash
git clone https://github.com/YOUR_USERNAME/terraform-provider-dreamhost.git
cd terraform-provider-dreamhost
```

2. Install dependencies:
```bash
go mod download
```

3. Build the provider:
```bash
make build
```

4. Install locally for testing:
```bash
make install
```

### Running Tests

Unit tests:
```bash
make test
```

Acceptance tests (requires `DREAMHOST_API_KEY`):
```bash
export DREAMHOST_API_KEY="your-api-key"
make testacc
```

### Code Quality

Format code:
```bash
gofmt -w .
```

Run linter:
```bash
make lint
```

## Coding Standards

### Go Code Style

- Follow standard Go formatting (`gofmt`)
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and small
- Handle errors explicitly
- Use table-driven tests where appropriate

### Terraform Provider Guidelines

- Resources should be immutable where the API doesn't support updates
- Use `ForceNew: true` for immutable attributes
- Implement proper error handling with helpful messages
- Add retry logic for transient failures
- Include comprehensive acceptance tests
- Document all resources and data sources

### Commit Messages

Follow conventional commits format:

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Test additions or changes
- `chore`: Maintenance tasks

Example:
```
feat(resource): add support for CAA DNS records

- Implement CAA record validation
- Add acceptance tests
- Update documentation
```

## Testing Guidelines

### Unit Tests

- Test all validation functions
- Test helper functions
- Mock external dependencies
- Aim for high coverage

### Acceptance Tests

- Test full CRUD lifecycle
- Test import functionality
- Test error conditions
- Test edge cases
- Clean up resources after tests

Example test structure:
```go
func TestAccDreamHostDNSRecord_basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:     func() { testAccPreCheck(t) },
        Providers:    testAccProviders,
        CheckDestroy: testAccCheckDNSRecordDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccDNSRecordConfig_basic(),
                Check: resource.ComposeTestCheckFunc(
                    testAccCheckDNSRecordExists("dreamhost_dns_record.test"),
                    resource.TestCheckResourceAttr("dreamhost_dns_record.test", "type", "A"),
                ),
            },
        },
    })
}
```

## Documentation

### Provider Documentation

- Update docs/ for any schema changes
- Include examples for all use cases
- Document error conditions
- Keep documentation in sync with code

### Code Documentation

- Document all exported functions
- Include examples in comments where helpful
- Explain complex logic
- Document assumptions and limitations

## Release Process

1. Update CHANGELOG.md
2. Create a new tag following semantic versioning
3. Push the tag to trigger release workflow
4. GitHub Actions will build and publish the release

## Getting Help

- Check existing issues and documentation
- Ask questions in issues with the "question" label
- Join community discussions

## Security

- Never commit API keys or secrets
- Report security vulnerabilities privately to maintainers
- Follow secure coding practices
- Validate all inputs

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

Thank you for contributing to the DreamHost Terraform Provider!