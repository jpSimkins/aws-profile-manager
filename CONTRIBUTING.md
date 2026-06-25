# Contributing to AWS Profile Manager

Thank you for your interest in contributing to AWS Profile Manager! This document provides guidelines and instructions for contributing.

## Table of Contents
- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Submitting Changes](#submitting-changes)
- [Reporting Issues](#reporting-issues)
- [Community](#community)

---

## Code of Conduct

This project is licensed under the MIT License, which means the code is free for everyone to use, modify, and distribute. By contributing, you agree that your contributions will be licensed under the same MIT License.

We are committed to providing a welcoming and inclusive environment. By participating in this project, you agree to:

- **Be respectful** - Treat everyone with respect and consideration
- **Be collaborative** - Work together constructively
- **Be patient** - Remember that everyone has different skill levels
- **Be thoughtful** - Consider how your words and actions affect others

## Getting Started

### Prerequisites

Before contributing, ensure you have:

- **Go 1.26+** - For building and testing
- **Node.js/npm** - For convenience scripts
- **Git** - For version control
- **System dependencies** - For GUI development (installed via `make deps-system`)

### Initial Setup

1. **Fork the repository** on GitHub

2. **Clone your fork:**
   ```bash
   git clone https://github.com/YOUR-USERNAME/aws-profile-manager.git
   cd aws-profile-manager
   ```

3. **Add upstream remote:**
   ```bash
   git remote add upstream https://github.com/jpSimkins/aws-profile-manager.git
   ```

4. **Run setup:**
   ```bash
   npm run setup
   ```

5. **Verify everything works:**
   ```bash
   make test
   make build
   ```

See [DEVELOPER.md](DEVELOPER.md) for complete development setup and workflows.

## Development Workflow

### Creating a Branch

Always create a new branch for your changes:

```bash
# Update your main branch
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feature/your-feature-name
# Or for bug fixes:
git checkout -b fix/issue-description
```

**Branch Naming Convention:**
- Features: `feature/descriptive-name`
- Fixes: `fix/issue-description`
- Docs: `docs/what-changed`
- Refactors: `refactor/what-changed`

### Making Changes

1. **Follow the coding standards** (see below)
2. **Write tests** for new functionality
3. **Update documentation** as needed
4. **Test your changes** thoroughly

### Before Committing

**Run this checklist (IN ORDER):**

```bash
make fmt                  # Format code
make vet                  # Run go vet
make lint                 # MUST be 100% clean (zero errors)
make test-coverage        # Verify tests pass (target: 95%+)
make build                # Verify build succeeds
```

**Critical Requirements:**
- ✅ Lint MUST show zero errors
- ✅ All tests MUST pass
- ✅ Test coverage should maintain or improve current levels
- ✅ Documentation MUST be thorough (all exported symbols documented)

### Commit Messages

Write clear, descriptive commit messages:

**Format:**
```
Short summary (50 chars or less)

More detailed explanation if needed. Wrap at 72 characters.

- Bullet points are fine
- Explain WHAT changed and WHY, not HOW

Fixes #123
```

**Good examples:**
```
Add AssumeRole support to profile generator

Implements support for AWS AssumeRole chains in the profile
generator, allowing users to configure role assumption profiles
with MFA, external IDs, and custom session names.

Fixes #42
```

**Bad examples:**
```
fix bug          ❌ Too vague
Updated files    ❌ Doesn't explain what/why
WIP             ❌ Never commit work-in-progress
```

## Coding Standards

### Critical Rules

**MANDATORY - Read Before Making ANY Changes:**

See [`.github/copilot-instructions.md`](.github/copilot-instructions.md) for complete coding standards. Key requirements:

1. **Logging** - ALWAYS use `logging.Log.*()`, NEVER `fmt.Print*` or `log.Print*`
2. **Naming** - ALWAYS PascalCase (HttpClient, not HTTPClient)
3. **Test Isolation** - ALWAYS use `test.SetupTestEnvironment(t)` for file I/O tests
4. **Task Package** - NEVER use `exec.Command` or raw goroutines, ALWAYS use `task.*Task`
5. **Dependency Injection** - Business logic NEVER imports `settings` package
6. **Documentation** - ALL exported symbols MUST be documented (godoc format)

### Package Patterns

**CLI/GUI** (thin presentation layers):
```go
func runCommand(cmd *cobra.Command, args []string) error {
    // 1. Parse flags (CLI responsibility)
    // 2. Call ONE package function (all logic in package)
    // 3. Display results (CLI responsibility)
}
```

**Business Logic** (packages):
```go
func DoWork(ctx context.Context, cfg WorkConfig, reporter task.Reporter) (*Result, error) {
    // Accept config, not settings
    // Use task.Reporter for progress
    // Return result struct
}
```

### Testing Standards

**Every `.go` file MUST have `_test.go` companion**

```go
func TestMyFunction(t *testing.T) {
    test.SetupTestEnvironment(t)  // REQUIRED for file I/O
    
    schema := schematest.NewManagedSsoSingle()  // Use test fixtures
    
    result, err := MyFunction(schema)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    // Assertions...
}
```

**Coverage target: 95%+**

## Submitting Changes

### Pull Request Process

1. **Update your branch with latest upstream:**
   ```bash
   git checkout main
   git pull upstream main
   git checkout your-feature-branch
   git rebase main
   ```

2. **Push to your fork:**
   ```bash
   git push origin your-feature-branch
   ```

3. **Create Pull Request on GitHub:**
   - Use a clear, descriptive title
   - Fill out the PR template completely
   - Link any related issues
   - Add screenshots for UI changes
   - Mark as draft if not ready for review

### Pull Request Requirements

**Before submitting, ensure:**

- ✅ All tests pass (`make test-coverage`)
- ✅ Lint is 100% clean (`make lint`)
- ✅ Code is formatted (`make fmt`)
- ✅ Documentation is complete
- ✅ CHANGELOG.md is updated (if applicable)
- ✅ No merge conflicts with main branch

### Pull Request Template

```markdown
## Description
Brief description of changes and why they're needed.

## Type of Change
- [ ] Bug fix (non-breaking change fixing an issue)
- [ ] New feature (non-breaking change adding functionality)
- [ ] Breaking change (fix or feature causing existing functionality to change)
- [ ] Documentation update

## Testing
- [ ] All existing tests pass
- [ ] New tests added for new functionality
- [ ] Manual testing completed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex logic
- [ ] Documentation updated
- [ ] No new warnings generated
- [ ] Tests provide adequate coverage

## Related Issues
Fixes #(issue number)
```

### Review Process

1. **Automated checks** will run (tests, lint, build)
2. **Code review** by maintainers
3. **Address feedback** if requested
4. **Approval and merge** by maintainers

**Be patient** - reviews may take time. We'll get to your PR!

## Reporting Issues

### Before Opening an Issue

1. **Search existing issues** - Your issue may already exist
2. **Try latest version** - Issue may be fixed in newer release
3. **Check documentation** - Answer may be in docs

### Issue Template

**For bugs:**
```markdown
**Description**
Clear description of the bug

**Steps to Reproduce**
1. Step one
2. Step two
3. ...

**Expected Behavior**
What should happen

**Actual Behavior**
What actually happens

**Environment**
- OS: [Linux/macOS/Windows]
- Version: [e.g., v1.0.0]
- Go version: [e.g., 1.22.1]

**Additional Context**
Any other relevant information
```

**For features:**
```markdown
**Problem Statement**
What problem does this solve?

**Proposed Solution**
How should this work?

**Alternatives Considered**
Other approaches you've thought about

**Additional Context**
Examples, mockups, etc.
```

## Community

### Getting Help

- **Issues** - Report bugs and request features
- **Discussions** - Ask questions and share ideas
- **Documentation** - Check [docs/](docs/) for guides

### Recognition

Contributors are recognized in [CONTRIBUTORS.md](CONTRIBUTORS.md). Your contributions are valued!

### Licensing Your Contributions

By submitting a contribution, you agree to license your work under the MIT License, the same license as the project. You retain copyright to your contributions, but grant everyone the same rights to use them as specified in the MIT License.

---

## Additional Resources

- **[DEVELOPER.md](DEVELOPER.md)** - Complete development guide
- **[README.md](README.md)** - Project overview
- **[docs/](docs/)** - Documentation guides
- **[.github/copilot-instructions.md](.github/copilot-instructions.md)** - Complete coding standards

---

Thank you for contributing to AWS Profile Manager! 🎉
