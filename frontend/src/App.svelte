<script>
  import { onMount } from "svelte";
  import {
    CheckAdb,
    GetDevices,
    GetDeviceInfo,
    GetPackages,
    UninstallPackage,
    GetAppIcon,
  } from "../wailsjs/go/main/App.js";
  import { EventsOn, BrowserOpenURL } from "../wailsjs/runtime/runtime.js";
  import logo from "./assets/logo.png";

  let adbAvailable = false;
  /** @type {string[]} */
  let devices = [];
  let selectedDevice = "";
  /** @type {any} */
  let deviceInfo = {};
  /** @type {any[]} */
  let packages = [];
  /** @type {any[]} */
  let filteredPackages = [];
  let searchQuery = "";
  /** @type {string[]} */
  let selectedPackages = [];
  let loading = false;
  let statusMessage = "";
  let showAboutModal = false;

  onMount(async () => {
    adbAvailable = await CheckAdb();
    if (adbAvailable) {
      try {
        devices = (await GetDevices()) || [];
        if (devices.length > 0) {
          selectedDevice = devices[0];
          await loadDeviceData(selectedDevice);
        }
      } catch (e) {
        statusMessage = "Error starting device scan: " + e;
      }

      // Listen for auto-detected device changes
      EventsOn("devices_updated", async (updatedDevices) => {
        devices = updatedDevices || [];
        if (devices.length > 0) {
          if (!selectedDevice || !devices.includes(selectedDevice)) {
            selectedDevice = devices[0];
            await loadDeviceData(selectedDevice);
          }
        } else {
          selectedDevice = "";
          deviceInfo = {};
          packages = [];
          filteredPackages = [];
          statusMessage = "Device disconnected.";
        }
      });
    } else {
      statusMessage =
        "ADB not found. Please install ADB and add it to your PATH.";
    }
  });

  /**
   * @param {string} serial
   */
  async function loadDeviceData(serial) {
    loading = true;
    try {
      deviceInfo = await GetDeviceInfo(serial);
      packages = await GetPackages(serial);
      selectedPackages = [];
      statusMessage = `Connected to ${deviceInfo.model}`;
    } catch (e) {
      statusMessage = "Error loading device data: " + e;
    } finally {
      loading = false;
    }
  }

  let activeFilter = "All";

  $: {
    let result = packages || [];
    if (searchQuery) {
      const q = searchQuery.toLowerCase();
      result = result.filter(
        (p) =>
          p.appName.toLowerCase().includes(q) ||
          p.packageName.toLowerCase().includes(q),
      );
    }
    if (activeFilter === "Safe to Remove") {
      result = result.filter((p) => p.isBloat);
    } else if (activeFilter === "System Core") {
      result = result.filter((p) => p.isSystem && !p.isBloat);
    } else if (activeFilter === "Third Party") {
      result = result.filter((p) => !p.isSystem && !p.isBloat);
    }
    filteredPackages = result;
  }

  /** @type {any[]} */
  let displayPackages = [];
  $: displayPackages = [...filteredPackages].sort((a, b) => {
    const aSelected = selectedPackages.includes(a.packageName);
    const bSelected = selectedPackages.includes(b.packageName);
    if (aSelected && !bSelected) return -1;
    if (!aSelected && bSelected) return 1;
    return a.appName.localeCompare(b.appName);
  });

  async function rescanDevices() {
    loading = true;
    try {
      devices = (await GetDevices()) || [];
      if (devices.length > 0) {
        selectedDevice = devices[0];
        await loadDeviceData(selectedDevice);
      } else {
        selectedDevice = "";
        deviceInfo = {};
        packages = [];
        filteredPackages = [];
        statusMessage = "Device disconnected.";
      }
    } catch (e) {
      statusMessage = "Error scanning for devices: " + e;
    } finally {
      loading = false;
    }
  }

  let showConfirmModal = false;
  /** @type {any[]} */
  let confirmPackages = [];
  let isSystemWarning = false;

  /**
   * @param {any} pkg
   */
  function promptDeletePackage(pkg) {
    confirmPackages = [pkg];
    isSystemWarning = pkg.isSystem && !pkg.isBloat;
    showConfirmModal = true;
  }

  function promptDeleteSelected() {
    if (selectedPackages.length === 0) return;
    const pkgsToConfirm = packages.filter((p) =>
      selectedPackages.includes(p.packageName),
    );
    confirmPackages = pkgsToConfirm;
    isSystemWarning = pkgsToConfirm.some((p) => p.isSystem && !p.isBloat);
    showConfirmModal = true;
  }

  async function confirmDeletion() {
    showConfirmModal = false;
    loading = true;
    let successCount = 0;
    let failCount = 0;

    for (const pkg of confirmPackages) {
      try {
        await UninstallPackage(selectedDevice, pkg.packageName);
        successCount++;
      } catch (e) {
        failCount++;
      }
    }

    if (confirmPackages.length === 1) {
      if (failCount > 0) {
        statusMessage = `Failed to uninstall ${confirmPackages[0].appName}`;
      } else {
        statusMessage = `Successfully uninstalled ${confirmPackages[0].appName}. Refreshing...`;
      }
    } else {
      statusMessage = `Done. Success: ${successCount}, Failed: ${failCount}. Refreshing...`;
      selectedPackages = [];
    }

    setTimeout(async () => {
      await loadDeviceData(selectedDevice);
      loading = false;
    }, 1000);
  }

  async function autoSelectBloatware() {
    selectedPackages = packages
      .filter((pkg) => pkg.isBloat)
      .map((pkg) => pkg.packageName);
    statusMessage = `Selected ${selectedPackages.length} bloatware packages.`;
  }

  function toggleSelectAll() {
    const visiblePkgNames = filteredPackages.map((p) => p.packageName);
    const allVisibleSelected =
      visiblePkgNames.length > 0 &&
      visiblePkgNames.every((p) => selectedPackages.includes(p));

    if (allVisibleSelected) {
      // Deselect all visible items
      selectedPackages = selectedPackages.filter(
        (p) => !visiblePkgNames.includes(p),
      );
    } else {
      // Select all visible items (add them if not present)
      const newSelections = visiblePkgNames.filter(
        (p) => !selectedPackages.includes(p),
      );
      selectedPackages = [...selectedPackages, ...newSelections];
    }
  }

  /**
   * @param {string} pkgName
   * @param {boolean} checked
   */
  function handleCheckboxChange(pkgName, checked) {
    if (checked) {
      if (!selectedPackages.includes(pkgName)) {
        selectedPackages = [...selectedPackages, pkgName];
      }
    } else {
      selectedPackages = selectedPackages.filter((p) => p !== pkgName);
    }
  }

  /**
   * @param {string} name
   */
  function getAvatarColor(name) {
    const colors = [
      "#5E81AC", // Blue
      "#8FBCBB", // Cyan
      "#A3BE8C", // Green
      "#D08770", // Orange
      "#B48EAD", // Purple
      "#4C566A", // Slate
    ];
    let hash = 0;
    for (let i = 0; i < name.length; i++) {
      hash = name.charCodeAt(i) + ((hash << 5) - hash);
    }
    const index = Math.abs(hash) % colors.length;
    return colors[index];
  }

  /** @type {Record<string, string>} */
  let iconCache = {};
  /**
   * @param {string} pkgName
   */
  async function getIcon(pkgName) {
    if (iconCache[pkgName]) return iconCache[pkgName];
    try {
      const url = await GetAppIcon(pkgName);
      if (url) {
        iconCache[pkgName] = url;
        return url;
      }
    } catch (e) {
      console.error(e);
    }
    return "";
  }

  // Since we are using light theme only, we don't need toggleTheme.
</script>

<main class="container">
  <header class="navbar">
    <div class="nav-left">
      <div class="logo">
        <img src={logo} class="logo-img" alt="ZeroBloat" />
        <span class="brand-name">ZeroBloat</span>
      </div>
    </div>

    <div class="nav-right">
      <button class="rescan-btn" on:click={rescanDevices}>
        <svg class="icon-refresh" class:spinning={loading} viewBox="0 0 24 24"
          ><path
            d="M17.65 6.35A7.958 7.958 0 0 0 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08A5.99 5.99 0 0 1 12 18c-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z"
          /></svg
        >
        Rescan
      </button>
      <button
        class="icon-btn about-toggle"
        title="About ZeroBloat"
        on:click={() => (showAboutModal = true)}
      >
        <svg viewBox="0 0 24 24">
          <path
            d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-6h2v6zm0-8h-2V7h2v2z"
          />
        </svg>
      </button>
    </div>
  </header>

  <div class="content-area">
    {#if !adbAvailable}
      <div class="connection-state">
        <div class="error-card">
          <div class="error-icon">⚠️</div>
          <h2>ADB Not Found</h2>
          <p>{statusMessage}</p>
          <button class="btn primary" on:click={() => window.location.reload()}
            >Retry Initialization</button
          >
        </div>
      </div>
    {:else if devices.length === 0}
      <div class="connection-state">
        <div class="waiting-header">
          <div class="usb-icon-wrapper">
            <svg viewBox="0 0 24 24"
              ><path
                d="M15 7v4h1v2h-3V5h2l-3-4-3 4h2v8H8v-2.07c.7-.37 1.2-1.08 1.2-1.93 0-1.21-.99-2.2-2.2-2.2-1.21 0-2.2.99-2.2 2.2 0 .85.5 1.56 1.2 1.93V13c0 1.11.89 2 2 2h3v3.05c-.71.37-1.2 1.1-1.2 1.95 0 1.21.99 2.2 2.2 2.2 1.21 0 2.2-.99 2.2-2.2 0-.85-.49-1.58-1.2-1.95V15h3c1.11 0 2-.89 2-2v-2h1V7h-4z"
              /></svg
            >
          </div>
          <h2>Waiting for USB Connection...</h2>
          <p class="subtitle">
            Ensure USB Debugging is enabled on your Android device to begin
            managing applications and system logs.
          </p>
          <div class="scanning-status">
            <svg class="icon-refresh small spinning" viewBox="0 0 24 24"
              ><path
                d="M17.65 6.35A7.958 7.958 0 0 0 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08A5.99 5.99 0 0 1 12 18c-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z"
              /></svg
            >
            <span>SCANNING FOR DEVICES...</span>
          </div>
        </div>

        <div class="steps-grid">
          <div class="step-card">
            <div class="step-number">1</div>
            <h3>Enable Debugging</h3>
            <p>
              Navigate to <strong>Settings > About Phone</strong> and tap
              <strong>Build Number</strong> 7 times.
            </p>
            <div class="step-visual placeholder">
              <div class="mock-ui">
                <div class="mock-line"></div>
                <div class="mock-line highlight">USB Debugging [ON]</div>
                <div class="mock-line"></div>
              </div>
            </div>
          </div>

          <div class="step-card">
            <div class="step-number">2</div>
            <h3>Verify Cable</h3>
            <p>
              Use a high-quality USB 3.0 data cable. Avoid charging-only cables
              or loose ports that may disrupt the ADB handshake.
            </p>
            <div class="status-badge success">
              <svg viewBox="0 0 24 24"
                ><path
                  d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"
                /></svg
              >
              Data Sync Supported
            </div>
          </div>

          <div class="step-card">
            <div class="step-number">3</div>
            <h3>Authorize PC</h3>
            <p>
              Look for the "Allow USB Debugging?" prompt on your device screen
              and select "Always allow from this computer".
            </p>
            <div class="terminal-snippet">
              <div class="terminal-header">
                <span class="dot"></span><span class="dot"></span><span
                  class="dot"
                ></span>
              </div>
              <code>
                <span class="prompt">$</span> adb devices<br />
                List of devices attached<br />
                <span class="warning">unauthorized</span>
              </code>
            </div>
          </div>
        </div>

        <div class="footer-actions">
          <button
            class="btn footer-btn secondary"
            on:click={() =>
              BrowserOpenURL(
                "https://developer.android.com/studio/releases/platform-tools",
              )}
          >
            <svg class="btn-icon-svg" viewBox="0 0 24 24"
              ><path d="M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z" /></svg
            >
            Download ADB Drivers
          </button>
          <button
            class="btn footer-btn primary action-main"
            on:click={rescanDevices}
          >
            <svg
              class="btn-icon-svg"
              class:spinning={loading}
              viewBox="0 0 24 24"
              ><path
                d="M17.65 6.35A7.958 7.958 0 0 0 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08A5.99 5.99 0 0 1 12 18c-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z"
              /></svg
            >
            Rescan Devices
          </button>
          <button class="btn footer-btn secondary">
            <svg class="btn-icon-svg" viewBox="0 0 24 24"
              ><path
                d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 17h-2v-2h2v2zm2.07-7.75l-.9.92C13.45 12.9 13 13.5 13 15h-2v-.5c0-1.1.45-2.1 1.17-2.83l1.24-1.26c.37-.36.59-.86.59-1.41 0-1.1-.9-2-2-2s-2 .9-2 2H8c0-2.21 1.79-4 4-4s4 1.79 4 4c0 .88-.36 1.68-.93 2.25z"
              /></svg
            >
            Full Troubleshooting Guide
          </button>
        </div>
      </div>
    {:else}
      <!-- Connected State -->
      <div class="connected-view">
        <div class="card device-panel">
          <div class="panel-header">
            <div class="device-badge">
              <div class="status-dot connected"></div>
              <span>{deviceInfo.model || selectedDevice} Connected</span>
            </div>
            <div class="device-actions">
              <select
                bind:value={selectedDevice}
                on:change={() => loadDeviceData(selectedDevice)}
                class="device-select"
              >
                {#each devices as dev}
                  <option value={dev}>{dev}</option>
                {/each}
              </select>
              <button
                class="btn danger small"
                on:click={() => (selectedDevice = "")}>Disconnect</button
              >
            </div>
          </div>

          <div class="device-info-grid">
            <div class="info-item">
              <span class="label">Brand</span>
              <span class="value">{deviceInfo.manufacturer}</span>
            </div>
            <div class="info-item">
              <span class="label">Model</span>
              <span class="value">{deviceInfo.model}</span>
            </div>
            <div class="info-item">
              <span class="label">Version</span>
              <span class="value">Android {deviceInfo.version}</span>
            </div>
          </div>
        </div>

        <div class="filter-pills-container">
          <div class="filter-pills">
            <button
              class="pill"
              class:active={activeFilter === "All"}
              on:click={() => (activeFilter = "All")}>All</button
            >
            <button
              class="pill"
              class:active={activeFilter === "Safe to Remove"}
              on:click={() => (activeFilter = "Safe to Remove")}
              >Safe to Remove</button
            >
            <button
              class="pill"
              class:active={activeFilter === "System Core"}
              on:click={() => (activeFilter = "System Core")}
              >System Core</button
            >
            <button
              class="pill"
              class:active={activeFilter === "Third Party"}
              on:click={() => (activeFilter = "Third Party")}
              >Third Party</button
            >
          </div>
        </div>

        <div class="toolbar card">
          <div class="actions-left">
            <div class="checkbox-container select-all-wrapper">
              <input
                type="checkbox"
                on:change={toggleSelectAll}
                checked={displayPackages.length > 0 &&
                  displayPackages.every((p) =>
                    selectedPackages.includes(p.packageName),
                  )}
              />
            </div>
            <div class="search-bar">
              <svg class="search-icon" viewBox="0 0 24 24"
                ><path
                  d="M15.5 14h-.79l-.28-.27A6.471 6.471 0 0 0 16 9.5 6.5 6.5 0 1 0 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"
                /></svg
              >
              <input
                type="text"
                placeholder="Search apps..."
                bind:value={searchQuery}
              />
            </div>
            <button class="btn secondary" on:click={autoSelectBloatware}>
              <svg class="btn-icon-svg" viewBox="0 0 24 24"
                ><path
                  d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"
                /></svg
              >
              Auto-Select
            </button>
          </div>
          <div class="actions-right">
            <button
              class="btn danger"
              on:click={promptDeleteSelected}
              disabled={selectedPackages.length === 0}
            >
              <svg class="btn-icon-svg" viewBox="0 0 24 24"
                ><path
                  d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"
                /></svg
              >
              Delete ({selectedPackages.length})
            </button>
          </div>
        </div>

        <div class="package-grid">
          {#each displayPackages as pkg (pkg.packageName)}
            <div
              class="package-card card"
              class:is-bloat={pkg.isBloat}
              class:selected={selectedPackages.includes(pkg.packageName)}
            >
              <div class="checkbox-container">
                <input
                  type="checkbox"
                  checked={selectedPackages.includes(pkg.packageName)}
                  on:change={(e) => {
                    const target = /** @type {any} */ (e.target);
                    if (target) {
                      handleCheckboxChange(pkg.packageName, target.checked);
                    }
                  }}
                />
              </div>

              <div
                class="app-avatar"
                style="background-color: {getAvatarColor(pkg.appName)}"
              >
                {#await getIcon(pkg.packageName)}
                  {pkg.appName.charAt(0).toUpperCase()}
                {:then iconUrl}
                  {#if iconUrl}
                    <img src={iconUrl} alt={pkg.appName} class="app-icon-img" />
                  {:else}
                    {pkg.appName.charAt(0).toUpperCase()}
                  {/if}
                {/await}
              </div>

              <div class="app-details">
                <div class="app-name-row">
                  <span class="app-name">{pkg.appName}</span>
                  <button
                    class="info-btn"
                    on:click={() => promptDeletePackage(pkg)}
                    title="Delete Package"
                  >
                    <svg viewBox="0 0 24 24"
                      ><path
                        d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"
                      /></svg
                    >
                  </button>
                </div>
                <div class="pkg-name">{pkg.packageName}</div>
                <div class="tag-row">
                  {#if pkg.isBloat}
                    <span class="bloat-tag">Safe To Remove</span>
                  {:else if pkg.isSystem}
                    <span class="bloat-tag safe">System Core</span>
                  {:else}
                    <span class="bloat-tag user">Third Party</span>
                  {/if}
                </div>
              </div>
            </div>
          {/each}
        </div>
      </div>
    {/if}
  </div>

  <footer class="status-footer">
    <div class="status-left">
      <div class="status-dot" class:connected={devices.length > 0}></div>
      <span>{devices.length > 0 ? "CONNECTED" : "DISCONNECTED"}</span>
      <span class="divider">|</span>
      <span>ADB v1.0.41</span>
    </div>
    <div class="status-right">
      <span>{devices.length > 0 ? deviceInfo.model : "NO DEVICE DETECTED"}</span
      >
    </div>
  </footer>

  {#if showConfirmModal}
    <div class="modal-backdrop">
      <div class="modal card">
        <div class="modal-header">
          <h3 class="modal-title">Confirm Deletion</h3>
          <button class="close-btn" on:click={() => (showConfirmModal = false)}
            >✕</button
          >
        </div>
        <div class="modal-body">
          <p>
            Are you sure you want to delete {confirmPackages.length} package{confirmPackages.length >
            1
              ? "s"
              : ""}?
          </p>
          {#if isSystemWarning}
            <div class="warning-box">
              <strong>⚠️ WARNING: System Core App Detected!</strong>
              <p>
                You are attempting to delete one or more system core apps. This
                can break your phone, cause bootloops, or disable critical
                functionality. Proceed with extreme caution.
              </p>
            </div>
          {/if}
        </div>
        <div class="modal-footer">
          <button
            class="btn secondary"
            on:click={() => (showConfirmModal = false)}>Cancel</button
          >
          <button class="btn danger" on:click={confirmDeletion}
            >Yes, Delete</button
          >
        </div>
      </div>
    </div>
  {/if}

  {#if showAboutModal}
    <!-- svelte:ignore a11y-click-events-have-key-events -->
    <!-- svelte:ignore a11y-no-static-element-interactions -->
    <div class="modal-backdrop" on:click={() => (showAboutModal = false)}>
      <div class="modal card about-modal" on:click|stopPropagation>
        <div class="modal-header">
          <h3 class="modal-title">About ZeroBloat</h3>
          <button class="close-btn" on:click={() => (showAboutModal = false)}
            >✕</button
          >
        </div>
        <div class="modal-body about-body">
          <div class="about-logo-wrapper">
            <img src={logo} class="about-logo" alt="ZeroBloat Logo" />
            <h2 class="about-app-name">ZeroBloat</h2>
            <span class="about-version">v1.0.0</span>
          </div>

          <p class="about-description">
            ZeroBloat is a modern, premium bloatware manager for Android devices
            designed to help you declutter and secure your phone safely and
            silently.
          </p>

          <div class="about-actions">
            <button
              class="btn repo-btn"
              on:click={() =>
                BrowserOpenURL("https://github.com/AdhwaithAS/ZeroBloat")}
            >
              <svg class="btn-icon-svg" viewBox="0 0 24 24">
                <path
                  d="M12 2C6.477 2 2 6.477 2 12c0 4.42 2.865 8.166 6.839 9.489.5.092.682-.217.682-.482 0-.237-.008-.866-.013-1.7-2.782.603-3.369-1.34-3.369-1.34-.454-1.156-1.11-1.464-1.11-1.464-.908-.62.069-.608.069-.608 1.003.07 1.531 1.03 1.531 1.03.892 1.529 2.341 1.087 2.91.831.092-.646.35-1.086.636-1.336-2.22-.253-4.555-1.11-4.555-4.943 0-1.091.39-1.984 1.029-2.683-.103-.253-.446-1.27.098-2.647 0 0 .84-.269 2.75 1.025A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.294 2.747-1.025 2.747-1.025.546 1.377.203 2.394.1 2.647.64.699 1.028 1.592 1.028 2.683 0 3.842-2.339 4.687-4.566 4.935.359.309.678.919.678 1.852 0 1.336-.012 2.415-.012 2.743 0 .267.18.579.688.481C19.137 20.162 22 16.418 22 12c0-5.523-4.477-10-10-10z"
                />
              </svg>
              GitHub
            </button>

            <button
              class="btn coffee-btn"
              on:click={() =>
                BrowserOpenURL("https://buymeacoffee.com/adhwaithas")}
            >
              ☕ Buy Coffee
            </button>
          </div>
        </div>
      </div>
    </div>
  {/if}
</main>

<style>
  @import url("https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap");

  :root {
    --bg-color: #f8fafc;
    --text-color: #0f172a;
    --text-secondary: #475569;
    --card-bg: #ffffff;
    --border-color: #e2e8f0;
    --primary: #2563eb;
    --primary-hover: #1d4ed8;
    --secondary: #f1f5f9;
    --secondary-hover: #e2e8f0;
    --danger: #dc2626;
    --danger-hover: #b91c1c;
    --success: #16a34a;
    --accent-blue: #0ea5e9;
    --icon-color: #475569;
    --font-family: "Inter", -apple-system, BlinkMacSystemFont, "Segoe UI",
      Roboto, sans-serif;
    --navbar-height: 64px;
    --footer-height: 28px;
    --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);
    --shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.1),
      0 2px 4px -2px rgb(0 0 0 / 0.1);
    --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.1),
      0 4px 6px -4px rgb(0 0 0 / 0.1);
  }

  :global(body) {
    background-color: var(--bg-color);
    color: var(--text-color);
    font-family: var(--font-family);
    margin: 0;
    padding: 0;
    overflow-x: hidden;
    transition:
      background-color 0.3s ease,
      color 0.3s ease;
  }

  .container {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
  }

  /* Navbar Styles */
  .navbar {
    height: var(--navbar-height);
    background: var(--card-bg);
    border-bottom: 1px solid var(--border-color);
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 2rem;
    position: sticky;
    top: 0;
    z-index: 100;
    transition:
      background-color 0.3s ease,
      border-color 0.3s ease;
  }

  .nav-left .logo {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .logo-img {
    width: 28px;
    height: 28px;
    object-fit: contain;
  }

  .brand-name {
    font-weight: 700;
    font-size: 1.25rem;
    color: var(--text-color);
    letter-spacing: -0.025em;
  }

  .nav-right {
    display: flex;
    align-items: center;
    gap: 1.25rem;
  }

  .rescan-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background: var(--primary);
    color: white;
    border: none;
    padding: 0.6rem 1.2rem;
    border-radius: 8px;
    font-weight: 600;
    font-size: 0.9rem;
    cursor: pointer;
    transition: all 0.2s;
    box-shadow: var(--shadow-sm);
  }

  .rescan-btn:hover {
    background: var(--primary-hover);
    transform: translateY(-1px);
    box-shadow: var(--shadow-md);
  }

  .icon-btn {
    background: var(--secondary);
    border: none;
    padding: 0.6rem;
    color: var(--icon-color);
    cursor: pointer;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s;
  }

  .icon-btn:hover {
    background: var(--secondary-hover);
    color: var(--text-color);
  }

  .icon-btn svg {
    width: 20px;
    height: 20px;
    fill: currentColor;
  }

  /* Content Area */
  .content-area {
    flex: 1;
    padding: 2.5rem 2rem;
    max-width: 1200px; /* Increased width */
    margin: 0 auto;
    width: 100%;
    box-sizing: border-box;
    padding-bottom: calc(var(--footer-height) + 4rem);
  }

  /* Connection States (Waiting) */
  .connection-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    width: 100%;
  }

  .usb-icon-wrapper {
    width: 80px;
    height: 80px;
    background: var(--card-bg);
    border-radius: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: var(--shadow-md);
    margin: 0 auto 2.5rem auto; /* Explicitly centered */
    border: 1px solid var(--border-color);
    transition: all 0.3s ease;
  }

  .usb-icon-wrapper:hover {
    transform: scale(1.05);
    border-color: var(--primary);
  }

  .usb-icon-wrapper svg {
    width: 40px;
    height: 40px;
    fill: var(--primary);
  }

  .waiting-header h2 {
    font-size: 2.75rem;
    margin: 0 0 1rem 0;
    font-weight: 800;
    color: var(--text-color);
    letter-spacing: -0.04em;
  }

  .subtitle {
    color: var(--text-secondary);
    font-size: 1.125rem;
    max-width: 650px;
    margin: 0 auto 2.5rem auto;
    line-height: 1.6;
  }

  .scanning-status {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.75rem;
    color: var(--primary);
    font-weight: 700;
    font-size: 0.875rem;
    letter-spacing: 0.1em;
    background: var(--secondary);
    padding: 0.6rem 1.25rem;
    border-radius: 24px;
    width: fit-content;
    margin: 0 auto;
    border: 1px solid var(--border-color);
  }

  /* Steps Grid */
  .steps-grid {
    display: grid;
    grid-template-columns: repeat(
      auto-fit,
      minmax(340px, 1fr)
    ); /* Increased card width */
    gap: 2.5rem;
    width: 100%;
    margin-top: 4rem;
  }

  .step-card {
    background: var(--card-bg);
    border: 1px solid var(--border-color);
    border-radius: 20px;
    padding: 2.5rem;
    text-align: left;
    display: flex;
    flex-direction: column;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    box-shadow: var(--shadow-sm);
  }

  .step-card:hover {
    transform: translateY(-8px);
    box-shadow: 0 25px 30px -10px rgb(0 0 0 / 0.1);
    border-color: var(--primary);
  }

  .step-number {
    background: var(--primary);
    color: white;
    width: 36px;
    height: 36px;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 800;
    font-size: 1.125rem;
    margin-bottom: 1.75rem;
  }

  .step-card h3 {
    margin: 0 0 0.875rem 0;
    font-size: 1.35rem;
    font-weight: 800;
    color: var(--text-color);
  }

  .step-card p {
    color: var(--text-secondary);
    font-size: 1.05rem;
    line-height: 1.7;
    margin-bottom: 2.5rem;
    flex-grow: 1;
  }

  .step-visual {
    height: 150px;
    background: var(--secondary);
    border-radius: 16px;
    overflow: hidden;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 1px solid var(--border-color);
  }

  /* Status Badges */
  .status-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.6rem 1.2rem;
    border-radius: 9999px;
    font-weight: 700;
    font-size: 0.9rem;
    width: fit-content;
  }

  .status-badge.success {
    background: rgba(22, 163, 74, 0.1);
    color: var(--success);
    border: 1px solid rgba(22, 163, 74, 0.2);
  }

  .status-badge svg {
    width: 20px;
    height: 20px;
    fill: currentColor;
  }

  /* Terminal Snippet */
  .terminal-snippet {
    background: #0f172a;
    border-radius: 12px;
    padding: 1rem;
    width: 100%;
    box-sizing: border-box;
    box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.2);
  }

  .terminal-header {
    display: flex;
    gap: 6px;
    margin-bottom: 0.75rem;
  }

  .terminal-header .dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: #334155;
  }

  .terminal-snippet code {
    font-family: "JetBrains Mono", monospace;
    font-size: 0.85rem;
    color: #e2e8f0;
    line-height: 1.5;
    display: block;
  }

  .terminal-snippet .prompt {
    color: #10b981;
    font-weight: bold;
    margin-right: 0.5rem;
  }

  .terminal-snippet .warning {
    color: #ef4444;
  }

  /* Icons */
  .btn-icon-svg {
    width: 20px;
    height: 20px;
    fill: currentColor;
    flex-shrink: 0;
  }

  .icon-refresh {
    width: 18px;
    height: 18px;
    fill: currentColor;
    flex-shrink: 0;
  }

  .icon-refresh.small {
    width: 16px;
    height: 16px;
  }

  .error-icon {
    font-size: 3.5rem;
    margin-bottom: 1.5rem;
  }

  /* Footer Actions */
  .footer-actions {
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 2rem;
    margin-top: 5rem;
    flex-wrap: wrap;
    width: 100%;
  }

  .footer-btn {
    padding: 0.875rem 1.75rem;
    border-radius: 14px;
    font-weight: 700;
    font-size: 1rem;
    display: flex;
    align-items: center;
    gap: 0.75rem;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    border: 1px solid transparent;
  }

  .footer-btn.primary.action-main {
    background: var(--primary);
    color: white;
    padding: 1rem 2.5rem;
    border-radius: 16px;
    font-size: 1.125rem;
    box-shadow: 0 10px 20px -10px rgba(37, 99, 235, 0.5);
    z-index: 2;
  }

  .footer-btn.primary.action-main:hover {
    background: var(--primary-hover);
    transform: translateY(-4px) scale(1.05);
    box-shadow: 0 15px 25px -8px rgba(37, 99, 235, 0.6);
  }

  .footer-btn.secondary {
    background: var(--card-bg);
    border: 1px solid var(--border-color);
    color: var(--text-color);
    box-shadow: var(--shadow-sm);
  }

  .footer-btn.secondary:hover {
    background: var(--secondary);
    border-color: var(--primary);
    color: var(--primary);
    transform: translateY(-2px);
    box-shadow: var(--shadow-md);
  }

  /* Connected View */
  .connected-view {
    display: flex;
    flex-direction: column;
    gap: 2rem;
  }

  .device-panel {
    padding: 2rem;
  }

  .panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding-bottom: 2rem;
    border-bottom: 1px solid var(--border-color);
    margin-bottom: 2rem;
  }

  .device-badge {
    display: flex;
    align-items: center;
    gap: 1rem;
    font-weight: 800;
    font-size: 1.25rem;
    color: var(--text-color);
  }

  .status-dot {
    width: 12px;
    height: 12px;
    border-radius: 50%;
    background: var(--text-secondary);
  }

  .status-dot.connected {
    background: #10b981;
    box-shadow: 0 0 12px rgba(16, 185, 129, 0.5);
    animation: pulse 2s infinite;
  }

  @keyframes pulse {
    0% {
      opacity: 1;
    }
    50% {
      opacity: 0.6;
    }
    100% {
      opacity: 1;
    }
  }

  .device-actions {
    display: flex;
    gap: 1rem;
  }

  .device-select {
    padding: 0.6rem 1rem;
    border-radius: 8px;
    border: 1px solid var(--border-color);
    background: var(--secondary);
    color: var(--text-color);
    font-weight: 600;
    cursor: pointer;
  }

  .device-info-grid {
    display: flex;
    gap: 1rem;
    margin-top: 1.5rem;
    flex-wrap: wrap;
  }

  .info-item {
    display: flex;
    flex-direction: column;
    justify-content: center;
    padding: 1rem 1.5rem;
    background: var(--card-bg);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    flex: 1;
    min-width: 200px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.02);
  }

  .info-item .label {
    color: var(--text-secondary);
    font-size: 0.75rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-bottom: 0.25rem;
  }

  .info-item .value {
    color: var(--text-color);
    font-size: 1.1rem;
    font-weight: 800;
  }

  .filter-pills-container {
    margin-bottom: 1.5rem;
    display: flex;
  }

  .filter-pills {
    display: flex;
    gap: 0.5rem;
    background: var(--card-bg);
    padding: 0.5rem;
    border-radius: 9999px;
    border: 1px solid var(--border-color);
  }

  .pill {
    background: transparent;
    border: none;
    padding: 0.5rem 1.25rem;
    border-radius: 9999px;
    font-size: 0.9rem;
    font-weight: 700;
    color: var(--text-secondary);
    cursor: pointer;
    transition: all 0.2s;
  }

  .pill:hover {
    color: var(--text-color);
    background: var(--secondary);
  }

  .pill.active {
    background: var(--primary);
    color: white;
  }

  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.25rem 1.75rem;
  }

  .actions-left {
    display: flex;
    align-items: center;
    gap: 1.5rem;
    flex: 1;
  }

  .search-bar {
    position: relative;
    max-width: 400px;
    width: 100%;
  }

  .search-bar input {
    width: 100%;
    box-sizing: border-box;
    padding: 0.75rem 1rem 0.75rem 2.75rem;
    background: var(--secondary);
    border: 1px solid var(--border-color);
    border-radius: 10px;
    color: var(--text-color);
    font-size: 0.95rem;
    transition: all 0.2s;
  }

  .actions-left .btn {
    flex-shrink: 0;
  }

  .search-bar .search-icon {
    position: absolute;
    left: 14px;
    top: 50%;
    transform: translateY(-50%);
    width: 20px;
    height: 20px;
    fill: var(--text-secondary);
  }

  .package-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 1.5rem;
    margin-bottom: 3rem;
  }

  .package-card {
    display: grid;
    grid-template-columns: auto auto 1fr;
    align-items: center;
    gap: 1rem;
    padding: 1.5rem;
    background: var(--card-bg);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    transition: all 0.2s ease;
  }

  .package-card:hover {
    border-color: var(--primary);
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  }

  .package-card.selected {
    border-color: var(--primary);
    background: rgba(37, 99, 235, 0.03);
  }

  .package-card.is-bloat {
    background-color: rgba(220, 38, 38, 0.015);
  }

  .select-all-wrapper {
    margin-right: 0.5rem;
  }

  .checkbox-container {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .checkbox-container input[type="checkbox"] {
    width: 1.25rem;
    height: 1.25rem;
    accent-color: var(--primary);
    cursor: pointer;
  }

  .app-avatar {
    width: 48px;
    height: 48px;
    border-radius: 12px;
    font-size: 1.25rem;
    box-shadow: var(--shadow-sm);
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    font-weight: bold;
    overflow: hidden;
    flex-shrink: 0;
  }

  .app-icon-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .app-details {
    display: flex;
    flex-direction: column;
    justify-content: center;
  }

  .app-name-row {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    width: 100%;
  }

  .app-name {
    font-size: 1.05rem;
    font-weight: 700;
    color: var(--text-color);
    margin-bottom: 0.25rem;
  }

  .info-btn {
    background: none;
    border: none;
    padding: 4px;
    cursor: pointer;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background 0.2s;
    margin-left: 0.5rem;
  }

  .info-btn:hover {
    background: var(--secondary);
  }

  .info-btn svg {
    width: 18px;
    height: 18px;
    fill: var(--text-secondary);
  }

  .tag-row {
    margin-top: 0.5rem;
  }

  .pkg-name {
    font-size: 0.85rem;
    font-family: "JetBrains Mono", monospace;
    color: var(--text-secondary);
    word-break: break-all;
  }

  .bloat-tag {
    font-size: 0.7rem;
    font-weight: 800;
    text-transform: uppercase;
    background: #fee2e2;
    color: #dc2626;
    padding: 0.2rem 0.5rem;
    border-radius: 6px;
    letter-spacing: 0.05em;
  }

  .bloat-tag.safe {
    background: rgba(16, 185, 129, 0.1);
    color: #10b981;
  }

  .bloat-tag.user {
    background: rgba(37, 99, 235, 0.1);
    color: #3b82f6;
  }

  /* Footer */
  .status-footer {
    position: fixed;
    bottom: 0;
    left: 0;
    right: 0;
    height: var(--footer-height);
    background: var(--card-bg);
    border-top: 1px solid var(--border-color);
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0 1.5rem;
    font-family: "JetBrains Mono", monospace;
    font-size: 10px;
    color: var(--text-secondary);
    z-index: 1000;
    transition: background-color 0.3s ease;
  }

  .divider {
    margin: 0 4px;
    opacity: 0.3;
  }

  /* Buttons */
  .btn {
    padding: 0.75rem 1.5rem;
    border-radius: 10px;
    font-weight: 700;
    font-size: 0.95rem;
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
    box-shadow: var(--shadow-sm);
    cursor: pointer;
    border: none;
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .btn:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: var(--shadow-md);
  }

  .btn:active:not(:disabled) {
    transform: translateY(0);
  }

  .btn.danger {
    background: var(--danger);
    color: white;
  }

  .btn.danger:hover {
    background: var(--danger-hover);
  }

  /* Animations */
  .spinning {
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }

  .error-card {
    background: var(--card-bg);
    padding: 4rem;
    border-radius: 32px;
    border: 1px solid var(--border-color);
    max-width: 600px;
    width: 100%;
  }

  .mock-ui {
    width: 80%;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .mock-line {
    height: 12px;
    background: var(--border-color);
    border-radius: 4px;
    opacity: 0.5;
  }

  .mock-line.highlight {
    background: var(--primary);
    color: white;
    font-size: 9px;
    font-weight: 700;
    display: flex;
    align-items: center;
    padding: 0 8px;
    opacity: 1;
  }

  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .modal {
    background: var(--card-bg);
    width: 90%;
    max-width: 500px;
    border-radius: 16px;
    padding: 1.5rem;
    box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .modal-title {
    margin: 0;
    font-size: 1.25rem;
    color: var(--text-color);
  }

  .close-btn {
    background: transparent;
    border: none;
    font-size: 1.25rem;
    color: var(--text-secondary);
    cursor: pointer;
  }

  .modal-body p {
    color: var(--text-color);
    margin-bottom: 1rem;
    font-size: 1rem;
  }

  .warning-box {
    background: rgba(220, 38, 38, 0.1);
    border: 1px solid rgba(220, 38, 38, 0.3);
    padding: 1rem;
    border-radius: 8px;
    margin-bottom: 1.5rem;
  }

  .warning-box strong {
    color: var(--danger);
    display: block;
    margin-bottom: 0.5rem;
  }

  .warning-box p {
    margin: 0;
    color: var(--danger);
    font-size: 0.9rem;
    line-height: 1.4;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 1rem;
    margin-top: 1.5rem;
  }

  /* About Modal Styles */
  .about-modal {
    max-width: 440px !important;
    padding: 2.25rem 2rem !important;
    border-radius: 24px;
    background: var(--card-bg);
    border: 1px solid var(--border-color);
  }

  .about-body {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1.5rem;
    text-align: center;
  }

  .about-logo-wrapper {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
  }

  .about-logo {
    width: 72px;
    height: 72px;
    object-fit: contain;
    border-radius: 18px;
    box-shadow: var(--shadow-md);
    margin-bottom: 0.25rem;
  }

  .about-app-name {
    font-size: 1.75rem;
    font-weight: 800;
    margin: 0.25rem 0 0.125rem 0;
    color: var(--text-color);
    letter-spacing: -0.025em;
  }

  .about-version {
    font-size: 0.85rem;
    font-weight: 700;
    color: var(--primary);
    background: rgba(37, 99, 235, 0.08);
    padding: 0.25rem 0.75rem;
    border-radius: 9999px;
  }

  .about-description {
    color: var(--text-secondary);
    font-size: 0.95rem;
    line-height: 1.6;
    margin: 0;
  }

  .about-actions {
    display: flex;
    gap: 0.75rem;
    width: 100%;
    margin-top: 0.5rem;
  }

  .about-actions .btn {
    flex: 1;
    justify-content: center;
    padding: 0.8rem 1rem;
    border-radius: 12px;
    font-size: 0.9rem;
    font-weight: 700;
  }

  .repo-btn {
    background: #24292e !important;
    color: #ffffff !important;
    border: none !important;
    flex: 1;
    box-shadow: 0 4px 10px rgba(36, 41, 46, 0.25) !important;
  }

  .repo-btn:hover {
    background: #1b1f23 !important;
    transform: translateY(-2px);
    box-shadow: 0 6px 14px rgba(36, 41, 46, 0.35) !important;
  }

  .coffee-btn {
    background: #ffdd00 !important;
    color: #000000 !important;
    box-shadow: 0 4px 10px rgba(255, 221, 0, 0.3) !important;
    flex: 1;
  }

  .coffee-btn:hover {
    background: #f4d000 !important;
    transform: translateY(-2px);
    box-shadow: 0 6px 14px rgba(255, 221, 0, 0.4) !important;
  }
</style>
