# Contributing to PrefixOS

Thank you for your interest in contributing to **PrefixOS**!

---

## 1. Code of Conduct
Please review and adhere to our [Code of Conduct](CODE_OF_CONDUCT.md) at all times.

---

## 2. How to Contribute

1. **Fork & Clone**: Fork the repository on GitHub and clone locally.
2. **Branching**: Create a topic branch: `git checkout -b feature/my-new-feature`
3. **Coding Guidelines**: Adhere strictly to the [Style Guide](STYLE_GUIDE.md) and [Architecture Principles](docs/ARCHITECTURE_PRINCIPLES.md).
4. **Testing**: Run tests and race detection before opening a Pull Request:
   ```bash
   make lint
   make test
   make benchmark
   ```
5. **Pull Request**: Open a PR targeting the `develop` branch with clear rationale and benchmark proof.
