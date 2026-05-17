# ZeroBloat 🚀

ZeroBloat is a modern, lightweight, and user-friendly desktop application designed to help you declutter and safely debloat your Android devices via ADB (Android Debug Bridge). Built with a blazing fast Go backend and a highly responsive Svelte frontend powered by Wails, ZeroBloat gives you granular control over what runs on your phone.

---

![License](https://img.shields.io/github/license/AdhwaithAS/ZeroBloat?style=flat-square)
![Wails](https://img.shields.io/badge/built%20with-Wails%20v2-00add8?style=flat-square&logo=go)
![Svelte](https://img.shields.io/badge/frontend-Svelte%20%2B%20Vite-ff3e00?style=flat-square&logo=svelte)
![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey?style=flat-square)

---

## ✨ Features

- **⚡ Instant ADB Detection**: Auto-detects connected devices over USB debugging instantly without manual configuration.
- **🗂️ Card Grid Interface**: Say goodbye to boring tables! Enjoy a gorgeous, modern card grid layout with app icons and metadata.
- **🏷️ Intelligent App Categorization**:
  - 🔴 **Safe to Remove**: Automatically detects known manufacturer bloatware.
  - 🟢 **System Core**: Identifies essential system-level apps.
  - 🔵 **Third Party**: Displays user-installed applications.
- **🔍 Advanced Filtering & Search**: Filter packages by category (All, Safe to Remove, System Core, Third Party) and query them in real-time.
- **📌 Reactive Pinning**: Selected packages are instantly pinned and sorted to the top of the grid for easy review before deletion.
- **🛡️ Smart Safety Modal**: Deep confirmation system that pops up warnings when trying to delete crucial **System Core** apps, preventing accidental phone bricking.
- **🌗 Native Theme Toggle**: Switch effortlessly between premium Light Mode and Dark Mode.
- **🔧 ADB Driver Helper**: Built-in helper to quickly download ADB platform tools.

---

## 🛠️ Tech Stack

- **Backend**: Go (Golang)
- **Frontend**: Svelte & Vanilla CSS
- **Bundler & Dev Server**: Vite
- **Desktop Runtime Framework**: [Wails v2](https://wails.io/)

---

## 🚀 Getting Started

### Prerequisites

To build ZeroBloat from source, you'll need:

1. **Go** (v1.21 or higher) -> [Download Go](https://golang.org/dl/)
2. **Node.js** (v18 or higher) & npm -> [Download Node](https://nodejs.org/)
3. **Wails CLI** -> Install via:
   ```bash
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```
4. **ADB (Android Debug Bridge)** -> Set up in your system `PATH`. If missing, download platform tools directly using the in-app helper link.

### Development Mode

Run the live development server:

```bash
wails dev
```

This starts a local Wails window and watches for code changes on both Go backend and Svelte frontend, instantly reloading them.

### Building for Production

Compile a highly-optimized, single binary executable for your host platform:

```bash
wails build
```

The compiled installer/executable will be generated inside the `build/bin/` folder.

---

## 🤝 Contributing

Contributions make the open-source community an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

Please refer to [CONTRIBUTING.md](./CONTRIBUTING.md) for details on our code of conduct and how to submit pull requests.

---

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.

---

## 👤 Author

**Adhwaith AS**
- Email: [adhwaithas2007@gmail.com](mailto:adhwaithas2007@gmail.com)
- GitHub: [@AdhwaithAS](https://github.com/AdhwaithAS)
