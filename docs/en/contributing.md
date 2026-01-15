---
title: Contributing Guide
layout: default
nav_order: 6
description: "How to contribute to the Double Entry Generator project"
permalink: /en/contributing/
lang: en
---

# Contributing Guide

Thank you for your interest in the Double Entry Generator project! We welcome contributions of all kinds, including code, documentation, issue reports, and more.

## Ways to Contribute

### 🐛 Reporting Issues

If you find a bug or have a feature request, please:

1. Check [existing Issues](https://github.com/deb-sig/double-entry-generator/issues) to confirm the issue hasn't been reported
2. Create a new Issue with:
   - Clear problem description
   - Steps to reproduce
   - Expected behavior
   - Actual behavior
   - Environment information (OS, Go version, etc.)

### 💻 Contributing Code

#### Development Environment Setup

```bash
# 1. Fork and clone the repository
git clone https://github.com/YOUR_USERNAME/double-entry-generator.git
cd double-entry-generator

# 2. Add upstream repository
git remote add upstream https://github.com/deb-sig/double-entry-generator.git

# 3. Install dependencies
go mod download

# 4. Run tests
make test
```

#### Development Workflow

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Write code**
   - Follow Go code conventions
   - Add necessary tests
   - Update relevant documentation

3. **Run tests**
   ```bash
   make test
   make lint
   ```

4. **Commit code**
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

5. **Push and create PR**
   ```bash
   git push origin feature/your-feature-name
   ```

#### Code Standards

- Use `gofmt` to format code
- Follow official Go code conventions
- Add appropriate comments
- Write unit tests

### 📝 Contributing Documentation

Documentation contributions include:

- Fixing errors in documentation
- Adding new usage examples
- Improving documentation structure
- Translating documentation

Documentation files are located in the `docs/` directory, using Markdown format.

### 🔧 Adding a New Provider

If you want to add support for a new data source:

1. **Create Provider directory**
   ```bash
   mkdir pkg/provider/your-provider
   ```

2. **Implement interfaces**
   - Interfaces defined in `pkg/provider/interface.go`
   - Reference existing provider implementations

3. **Add tests**
   - Add test scripts in the `test/` directory
   - Provide sample data files

4. **Update documentation**
   - Add documentation under `docs/providers/`
   - Update README and navigation

## Commit Convention

We use [Conventional Commits](https://www.conventionalcommits.org/) convention:

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation update
- `style:` Code style changes
- `refactor:` Code refactoring
- `test:` Test related
- `chore:` Build process or auxiliary tool changes

Examples:
```bash
git commit -m "feat(alipay): add support for new transaction types"
git commit -m "fix(ccb): handle empty description field"
git commit -m "docs: update installation guide"
```

## Code Review

All submitted code will be reviewed:

1. **Automated checks**
   - Code format checks
   - Unit tests
   - Integration tests

2. **Manual review**
   - Code quality
   - Functional correctness
   - Documentation completeness

## Release Process

1. Code merged into main branch
2. Automatically trigger CI/CD process
3. Generate release version
4. Update documentation website

## Community Guidelines

### Code of Conduct

- Be friendly and respectful
- Welcome contributors from different backgrounds
- Focus on what's best for the community
- Show empathy towards other community members

### Communication Channels

- **GitHub Issues**: Issue reports and feature requests
- **GitHub Discussions**: General discussions and Q&A
- **Pull Requests**: Code review and discussion

## License

By contributing code, you agree that your contributions will be released under the Apache 2.0 License.

## Acknowledgments

Thank you to all developers who have contributed to the project! Your contributions make this project better.

## Need Help?

If you have any questions, please:

1. Check the [FAQ]({{ '/en/faq/' | relative_url }})
2. Ask in [GitHub Discussions](https://github.com/deb-sig/double-entry-generator/discussions)
3. Create an [Issue](https://github.com/deb-sig/double-entry-generator/issues)
