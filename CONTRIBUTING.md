# Contributing to ZeroBloat 🤝

Thank you for showing interest in contributing to **ZeroBloat**! We welcome and appreciate contributions of all kinds, whether it's reporting bugs, suggesting features, improving documentation, or writing code.

Please take a moment to review this document to make the contribution process smooth and effective for everyone involved.

---

## 📜 Code of Conduct

By participating in this project, you agree to abide by our code of conduct:
- Be respectful and welcoming to other contributors.
- Focus on what is best for the community and project.
- Gracefully accept constructive criticism.

---

## 🛠️ Getting Started

### 1. Fork the Repository
First, fork the project to your own GitHub account and clone it locally:

```bash
git clone https://github.com/YOUR_USERNAME/ZeroBloat.git
cd ZeroBloat
```

### 2. Set Up Prerequisites
Ensure you have the following installed on your development machine:
- **Go** (1.21+)
- **Node.js** (18+)
- **Wails v2 CLI** (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- **ADB Tools** (configured in your PATH)

### 3. Spin Up Development Server
Run the local live-reload Wails development environment:

```bash
wails dev
```

---

## 🌿 Branching Strategy

To keep the git history clean and manageable, please follow this branching format:
- Bug fixes: `fix/issue-description`
- New features: `feature/feature-name`
- Documentation: `docs/documentation-update`
- Code cleanup: `refactor/refactor-target`

Always create your branch from the main `main` branch:

```bash
git checkout -b feature/your-awesome-feature
```

---

## 💻 Coding Guidelines

- **Go Code**:
  - Keep functions focused and cohesive.
  - Run `go fmt ./...` before committing.
  - Return explicit errors and handle them properly.
- **Frontend Code (Svelte & CSS)**:
  - Write clean Svelte code.
  - Avoid inline styles where possible; use the vanilla CSS custom variables defined in `App.svelte` for themes.
  - Ensure any new CSS class added is actually used to prevent build-time Vite warnings.

---

## 🚀 Submitting a Pull Request (PR)

Once you've tested your changes and are ready to submit:

1. **Commit your changes** with a clear and descriptive commit message:
   ```bash
   git commit -m "feat: add support for XYZ feature"
   ```
2. **Push** your branch to your forked repository:
   ```bash
   git push origin feature/your-awesome-feature
   ```
3. **Open a Pull Request** on GitHub against our `main` branch.
4. **Fill out the PR template** describing:
   - What changes were made.
   - Why they were made.
   - Any testing you did.
5. Wait for review! We will review your PR and provide feedback as soon as possible.

---

## ✉️ Need Help?

If you have any questions or get stuck at any point, feel free to open an issue or contact the project maintainer:
- **Author**: Adhwaith AS ([adhwaithas2007@gmail.com](mailto:adhwaithas2007@gmail.com))
