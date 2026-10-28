# Architecture Decision Record: Fluent UI Settings, Single-Instance Lock & Modern Icon

## Context
User feedback highlighted three critical issues with the initial prototype:
1. The raw Win32 GDI window suffered from unstyled 1990s controls and missing/clipped buttons on high-DPI scaling.
2. Launching `RectangleWin.exe` multiple times opened multiple running instances in the taskbar/tray competing for hotkeys.
3. The icon was an outdated `#` crosshair grid with poor contrast and low resolution.

## Decisions

1. **Modern Windows 11 Fluent Settings GUI (via Pure-Go WebView2)**:
   - Evaluated pure-Go `go-webview2` which leverages the evergreen Windows 10/11 system WebView2 runtime without Cgo.
   - Built an embedded HTML/CSS/JS interface in `ui/settings.html` featuring:
     - Automatic Windows Dark/Light mode theme switching (`prefers-color-scheme`).
     - Miniature SVG window snap position preview diagrams for every single action.
     - Live search/filtering of shortcuts.
     - Interactive keyboard shortcut recording directly from the card.
     - Flexbox/grid layouts ensuring all controls and buttons are 100% visible across all DPI scales (100% to 200%).
     - Zero GDI clipping or hidden button issues.

2. **Single-Instance Mutex Enforcement**:
   - Implemented a named Win32 mutex (`Local\RectangleWin_SingleInstance_Mutex`).
   - If a second instance is started, it posts a registered broadcast message (`RectangleWin_ShowSettings_Msg`) to bring the existing instance's settings window to the foreground and immediately terminates with `os.Exit(0)`.
   - Guarantees exactly one instance is ever active in the system tray and taskbar.

3. **High-Resolution Modern Icon**:
   - Designed a modern macOS Rectangle-style icon featuring a dark slate translucent window frame with a vibrant Windows Fluent Blue snapped half-pane.
   - Generated multi-resolution layers (16x16, 24x24, 32x32, 48x48, 64x64, 128x128, 256x256) embedded in `assets/tray_icon.ico`, `assets/icon.ico`, and `assets/tray_icon.png`.

## Verification
- Unit test suite passes with 0 failures (`go test -v ./...`).
- Single-instance lock verified: starting secondary process results in 1 active process with automatic exit of the duplicate.
- Binary compiled cleanly as a native Windows GUI app (~7.1 MB).
