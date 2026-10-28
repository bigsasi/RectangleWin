# Architecture Decision Record: Configurable Keybindings & Modern Windows Integration

## Context
RectangleWin was originally created with hardcoded Win32 hotkeys and legacy Vista-era DPI awareness (`SetProcessDPIAware`). Users needed a simple way to customize keyboard shortcuts to match their personal muscle memory, disable conflicting keys, support multi-monitor setups, and adjust snapping behaviors without recompilation.

## Decisions

1. **Human-Readable JSON Configuration with Default Fallback**:
   - Store configuration at `%APPDATA%\RectangleWin\config.json` (falling back to `./config.json` for portable use).
   - Use natural key descriptor strings (e.g. `"Win+Alt+Left"`, `"Ctrl+Shift+F"`, `"Win+Alt+Space"`, `"none"`).
   - Automatically generate formatted default configuration on first launch if missing.

2. **Hot-Reloadable Win32 HotKey Engine**:
   - Provide `UnregisterAllHotKeys` and dynamic rebinding via Win32 `RegisterHotKey` / `UnregisterHotKey`.
   - Add tray menu actions ("Open Config File" and "Reload Configuration") allowing instant editing and reloading without process restart.

3. **Per-Monitor V2 DPI Awareness**:
   - Utilize `SetProcessDpiAwarenessContext(DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2)` with fallback to legacy APIs.
   - Ensures exact window frame boundaries and monitor coordinate calculations on high-DPI and mixed-scaling multi-monitor systems.

4. **Multi-Monitor Window Migration**:
   - Implement `MoveWindowToMonitor` calculating proportional scale and work area offsets when shifting windows across displays.

## Consequences & Verification
- Unit tests in `keymap_test.go`, `config_test.go`, and `snap_test.go` validate string parsing, config serialization, and snap geometry.
- Full compatibility with Windows 10 & 11 without external runtime dependencies.
