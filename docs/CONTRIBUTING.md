# 🤝 Contributing to SingerOS

Welcome to SingerOS! We're excited that you're interested in contributing to our enterprise digital workforce operating system. This guide will help you get started with contributing to our project.

## 🔍 Code of Conduct

By participating in this project, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md). We pledge to foster an open and welcoming environment.

## 🎯 Types of Contributions

We welcome various types of contributions:

### 💡 Ideas & Feedback
- Reporting bugs
- Suggesting new features
- Providing feedback on existing functionality

### 🛠️ Development
- Fixing bugs
- Implementing new features
- Improving documentation
- Writing tests

### 📚 Documentation
- Updating README files
- Creating guides and tutorials
- Improving API documentation

## 🚀 Getting Started

1. Fork the repository
2. Create a new branch for your feature/fix
3. Make your changes
4. Run tests to ensure everything works
5. Submit a pull request

## 📦 Project Structure

```
singeros/
├── control-plane/     # Governance & management components
├── data-plane/        # Runtime execution components
├── plugins/           # Plugin architecture
└── infrastructure/    # Infrastructure layer
```

## 🧪 Testing

Before submitting your contribution:
1. Run existing tests to ensure you didn't break anything
2. Add new tests for your changes if applicable
3. Ensure all tests pass

### Running Tests
```bash
# Run all tests
make test

# Run tests with coverage
make test-cover
```

## 📝 Code Style

### Go Code
- Follow Go best practices
- Use clear, descriptive variable and function names
- Maintain consistent formatting (use `gofmt`)
- Add comments for exported functions
- Write unit tests for new functionality

### Documentation
- Keep documentation clear and concise
- Use consistent terminology
- Include examples where appropriate

## 🐛 Bug Reports

When filing a bug report:

1. **Describe the issue** in detail
2. **Provide reproduction steps**
3. **Include system/environment details**
4. **Share any relevant logs or screenshots**

## 🔧 Pull Request Process

1. **Ensure any install or build dependencies** are removed before the end of the layer when doing a build
2. **Update the README.md** with details of changes to the interface
3. **Increase the version numbers** in any examples files and the README.md to the new version that this Pull Request would represent
4. **You may merge the Pull Request** once you have the sign-off of one other developer

## 📋 Contribution Checklist

- [ ] My code follows the project's style guidelines
- [ ] I have performed a self-review of my code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works

## 🌟 License

By contributing, you agree that your contributions will be licensed under the GNU General Public License v3.0.

## 📬 Contact

- GitHub Discussions: [link to discussions]
- Email: [contact email if applicable]

Thank you for contributing to SingerOS! 🐶