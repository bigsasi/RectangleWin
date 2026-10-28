# RectangleWin

A lightweight, keyboard-driven window snapping and management tool for Windows, inspired by macOS [Rectangle.app](https://rectangleapp.com) and [Spectacle.app](https://www.spectacleapp.com).

Move and snap windows to edges, corners, halves, and thirds using intuitive hotkeys or customize every action through a modern Windows 11 Fluent interface.

![RectangleWin demo](./assets/RectangleWin-demo.gif)

## Features

- **Edge & Half Snapping**: Left, right, top, and bottom snapping with automatic cycling between ½, ⅔, and ⅓ sizes.
- **Corner Quadrants**: Precise snapping to top-left, top-right, bottom-left, and bottom-right corners.
- **Thirds & Two-Thirds**: Direct shortcuts for vertical one-third and two-thirds splits.
- **Window Sizing**: Maximize, Almost Maximize (with customizable margin padding), Center, Restore, and Step Resize increments.
- **Multi-Monitor Display Moving**: Seamlessly move windows across displays preserving relative coordinates.
- **Always on Top**: Pin any window above all other windows.
- **Modern Fluent Settings UI**:
  - Dark and Light mode theme support (`prefers-color-scheme`).
  - Interactive visual snap diagrams for all 23 actions.
  - Modifier pill toggles (`[Win]`, `[Alt]`, `[Ctrl]`, `[Shift]`), direct key typing, and interactive key recorder.
  - Instant duplicate hotkey validation.
  - Global cross-tab search filter.
  - "Launch on Windows Startup" switch with live registry synchronization.
- **Single-Instance Enforcement**: Named Win32 mutex ensures only one process runs; subsequent launches bring Preferences to the front.
- **Pure Go**: Zero external C runtime or GCC requirements; built entirely with Go and native Windows Win32 APIs.

---

## Default Keyboard Bindings

### Snap to Edges (½, ⅔, ⅓)
* <kbd>Win</kbd> + <kbd>Alt</kbd> + <kbd>&larr;</kbd> / <kbd>&rarr;</kbd> / <kbd>&uarr;</kbd> / <kbd>&darr;</kbd>
* Pressing repeatedly cycles through ½, ⅔, and ⅓ sizes.

### Corner Quadrants
* **Top-Left**: <kbd>Win</kbd> + <kbd>Ctrl</kbd> + <kbd>Alt</kbd> + <kbd>&larr;</kbd>
* **Top-Right**: <kbd>Win</kbd> + <kbd>Ctrl</kbd> + <kbd>Alt</kbd> + <kbd>&uarr;</kbd>
* **Bottom-Left**: <kbd>Win</kbd> + <kbd>Ctrl</kbd> + <kbd>Alt</kbd> + <kbd>&darr;</kbd>
* **Bottom-Right**: <kbd>Win</kbd> + <kbd>Ctrl</kbd> + <kbd>Alt</kbd> + <kbd>&rarr;</kbd>

### Window Management & Monitors
* **Maximize**: <kbd>Win</kbd> + <kbd>Shift</kbd> + <kbd>F</kbd>
* **Almost Maximize**: <kbd>Win</kbd> + <kbd>Alt</kbd> + <kbd>F</kbd>
* **Center Window**: <kbd>Win</kbd> + <kbd>Alt</kbd> + <kbd>C</kbd>
* **Always on Top (Toggle)**: <kbd>Win</kbd> + <kbd>Alt</kbd> + <kbd>A</kbd>
* **Next Monitor**: <kbd>Win</kbd> + <kbd>Ctrl</kbd> + <kbd>Alt</kbd> + <kbd>PageUp</kbd>
* **Previous Monitor**: <kbd>Win</kbd> + <kbd>Ctrl</kbd> + <kbd>Alt</kbd> + <kbd>PageDown</kbd>

---

## Customizing Keybindings & Preferences

You can customize all shortcuts and behavior directly through the in-app **Preferences GUI**:

1. Right-click the **RectangleWin** icon in your system tray and select **"Settings..."** (or simply launch `RectangleWin.exe` while it is already running).
2. Browse through the tabs:
   - **Halves & Corners**: Half-screen and corner quadrant snaps with visual diagrams.
   - **Thirds & Sizing**: Thirds, maximize, almost-maximize margin, and step resizes.
   - **Displays**: Multi-monitor jumping and always-on-top pinning.
   - **Preferences**: System startup toggle and snapping behavior settings (subdivisions, margins, step resize).
   - **About**: App version, feature overview, and project links.
3. To change a keybinding:
   - Click the modifier pills (`[Win]`, `[Alt]`, `[Ctrl]`, `[Shift]`) to toggle keys.
   - Or click **"Record"** and press your combo.
   - Click `✕` to clear or disable any shortcut.
4. Click **"Save & Apply"** to validate and instantly activate your new keybindings without restarting.

---

## Development & Building

Requires Go 1.22+:

```powershell
# Run automated tests
go test -v ./...

# Build standalone Windows GUI binary
go build -ldflags "-H=windowsgui" -o RectangleWin.exe .
```

---

## License & Attribution

This project is licensed under the **Apache License 2.0**. See the [LICENSE](./LICENSE) and [NOTICE](./NOTICE) files for complete terms and third-party notices.

* **Repository**: [github.com/bigsasi/RectangleWin](https://github.com/bigsasi/RectangleWin)
* **Modernized & Maintained by**: [bigsasi](https://github.com/bigsasi)
* **Originally created by**: [Ahmet Alp Balkan](https://github.com/ahmetb) (Copyright 2022, Apache 2.0)
* **Inspired by**: [Rectangle for macOS](https://rectangleapp.com) by Ryan Hanson (MIT License)
