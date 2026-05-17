# ZeroBloat

ZeroBloat is a lightweight desktop utility designed to uninstall unwanted pre-installed system apps, carrier bloatware, and trackers from Android devices via ADB (Android Debug Bridge). No root access is required.

Built using a Go backend and a Svelte frontend powered by Wails, it provides a clean graphical interface to safely manage your device's packages.

---

![License](https://img.shields.io/github/license/AdhwaithAS/ZeroBloat?style=flat-square)
![Wails](https://img.shields.io/badge/built%20with-Wails%20v2-00add8?style=flat-square&logo=go)
![Svelte](https://img.shields.io/badge/frontend-Svelte%20%2B%20Vite-ff3e00?style=flat-square&logo=svelte)
![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey?style=flat-square)

---

## 📸 Screenshots

![ZeroBloat Dashboard](./screenshots/image.png)

![ZeroBloat About Modal](./screenshots/image2.png)

---

## ⚙️ Features

* **Automatic Device Detection**: Instantly recognizes connected Android devices via USB debugging.
* **Background ADB Execution**: Runs all ADB operations silently in the background. Command prompt (CMD) or terminal windows will not flash on the screen.
* **Intelligent Categorization**:
  * **Safe to Remove**: Identifies known carrier and manufacturer bloatware.
  * **System Core**: Displays sensitive Android operating system components.
  * **Third Party**: Lists standard user-installed applications.
* **Search & Filters**: Filter packages by category and search in real-time by package ID or app title.
* **Reactive Pinning**: Selected packages are automatically pinned and sorted to the top of the list.
* **Safety Confirmation**: Triggers double-confirmation warnings when attempting to delete System Core applications.
* **Community Integrations**: Simple Info panel linking directly to the GitHub repository and Buy Me a Coffee page.

---

## 🚀 Installation & Usage

You do not need to compile the code or set up a developer environment. Pre-compiled, single-file executables are available for download.

### Steps to Run:

1. Enable **USB Debugging** on your phone:
   * Go to **Settings** -> **About Phone** and tap **Build Number** 7 times.
   * Go back, open **Developer Options**, and toggle **USB Debugging** on.
2. Download the pre-built zip for your operating system from the [Releases](https://github.com/AdhwaithAS/ZeroBloat/releases) page.
3. Extract the downloaded folder and launch:
   * **Windows**: Run `ZeroBloat.exe`.
   * **macOS**: Move `ZeroBloat.app` to your Applications folder and open it.
   * **Linux**: Mark `ZeroBloat.AppImage` as executable (`chmod +x ZeroBloat-Linux.AppImage`) and launch it.
4. Connect your Android device to your computer via a USB cable. The application will auto-detect the device and retrieve the package list.

---

## 🛠️ Building From Source (Developers)

If you wish to compile or modify ZeroBloat locally, follow these steps:

### Prerequisites

The following tools must be installed on your machine:
1. **Go** (v1.21 or higher) -> [Download Go](https://golang.org/dl/)
2. **Node.js** (v18 or higher) & npm -> [Download Node](https://nodejs.org/)
3. **Wails CLI** -> Install via terminal:
   ```bash
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```

### Setup Instructions:

1. **Clone the repository**:
   ```bash
   git clone https://github.com/AdhwaithAS/ZeroBloat.git
   cd ZeroBloat
   ```

2. **Run in Development Mode**:
   Spins up a local development window with hot-reloading for Go backend and Svelte frontend changes:
   ```bash
   wails dev
   ```

3. **Build the Production Executable**:
   Compiles a highly-optimized single-file binary for your host operating system inside the `build/bin/` folder:
   ```bash
   wails build
   ```

---

## 📄 License

Distributed under the MIT License. See `LICENSE` for more details.

---

## 👤 Author & Support

**Adhwaith AS**
* **GitHub**: [@AdhwaithAS](https://github.com/AdhwaithAS)
* **Email**: [adhwaithas2007@gmail.com](mailto:adhwaithas2007@gmail.com)
* Consider supporting the project here: ☕ [Buy Me a Coffee](https://buymeacoffee.com/adhwaithas)
