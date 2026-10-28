# Architecture Decision Record: Native Win32 In-App Settings UI & Shortcut Recorder

## Context
While the JSON configuration system enabled customizable keybindings, editing raw JSON files is not user-friendly for everyday users. Users expected an in-app Preferences / Settings GUI where they could visually inspect actions, adjust bindings, and click "Record" to capture keystrokes naturally, similar to macOS Rectangle.

## Key Constraints
- The Windows development and runtime environment lacks `gcc`/Cgo compilers, preventing the use of Cgo-based UI toolkits like Fyne, Qt, or Walk.
- WebView2 / Electron adds multi-megabyte footprints, external runtime dependencies, or latency.
- The utility must remain lightweight, fast, and maintain a tiny memory and binary footprint (Principles P6 Dependency Discipline and P9 Minimalism).

## Decisions

1. **Pure Go Win32 Native Dialog**:
   - Built directly using `user32.dll` and `gonutz/w32/v2`.
   - Native Windows controls with crisp "Segoe UI" typography.
   - Tabbed layout dividing 23 actions across 3 logical groups:
     - Halves & Corners
     - Thirds & Sizing
     - Screens & Behavior

2. **Interactive Shortcut Capture Modal**:
   - Implemented an interactive recorder window that intercepts `WM_KEYDOWN` and `WM_SYSKEYDOWN`.
   - Automatically detects modifiers (`Win`, `Ctrl`, `Alt`, `Shift`) via `GetAsyncKeyState` and formats canonical key combinations (e.g. `Win+Alt+Left`).
   - Supports `Esc` to cancel and `Backspace`/`Delete` to clear.

3. **Inter-Thread Reload Synchronization**:
   - Win32 requires `RegisterHotKey` to execute on the thread that runs the message pump (`GetMessage`).
   - The Settings UI operates in its own goroutine with its own message loop.
   - When "Save & Apply" is clicked, settings are saved to disk, and a custom message (`WM_RELOAD_HOTKEYS = WM_USER + 100`) is posted to `mainThreadID` via `PostThreadMessageW`.
   - The main loop safely unbinds and rebinds all hotkeys without race conditions or thread affinity issues.

4. **Validation Engine**:
   - Built `ValidateConfig` to catch duplicate shortcuts and invalid key names before saving, presenting clear warnings to the user.

## Consequences & Verification
- 0 new external dependencies.
- Binary size remains compact (~6.7 MB).
- Comprehensive unit tests in `settings_validator_test.go` and `config_test.go`.
