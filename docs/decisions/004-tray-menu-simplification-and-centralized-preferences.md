# 004: Tray Menu Simplification and Centralized Preferences Architecture

## Context & Problem
Originally, RectangleWin did not have a graphical settings interface. As features were added, developer utilities and settings were scattered into the right-click system tray context menu:
- `Open Config File (JSON)`: Exposed raw JSON storage directly in the primary user menu.
- `Reload Configuration`: Manual hotkey reload needed when editing raw files.
- `View Active Shortcuts`: Popped up an obsolete Win32 message box dumping hotkey text.
- `Run on startup`: Placed as a stray checkbox in the tray menu, but completely absent from the Settings window.

Once a modern Windows 11 Fluent preferences modal UI was introduced, having these items in the right-click menu caused:
1. **Severe UX redundancy**: Users had two completely different places to configure the app and view shortcuts (one in modern Fluent UI, one in a raw Win32 message box).
2. **Cognitive clutter**: Low-level developer tools (raw JSON, manual reloads) were placed front-and-center in the everyday user context menu.
3. **Inconsistent state**: "Run on startup" was invisible to users configuring preferences in the Settings modal.

## Decision (Product Owner & UX Perspective)

### 1. Minimalist System Tray Gateway
The system tray menu is streamlined into a clean, modern utility menu:
* **`Settings...`**: Opens or focuses the centralized Preferences modal.
* **`Documentation`**: Opens the online documentation and shortcut guide in the user's default browser.
* **`Quit RectangleWin`**: Gracefully terminates the application.

All redundant, developer-facing, or duplicated items (`View Active Shortcuts`, `Reload Configuration`, `Open Config File (JSON)`, and tray-level `Run on startup`) have been removed from the tray context menu.

### 2. Centralized Preferences Control Center (P2: Zero Duplication)
All application preferences, behavior settings, and advanced tools are consolidated into the Preferences modal:
* **Tab 1: Halves & Corners**: 8 primary half and corner quadrant snap actions with visual diagrams.
* **Tab 2: Thirds & Sizing**: 12 third-splits, maximize, center, restore, and step resize actions.
* **Tab 3: Displays**: Multi-monitor navigation (`Next Monitor`, `Prev Monitor`, `Always on Top`).
* **Tab 4: General Preferences**:
  - **System Integration**: "Launch on Windows Startup" switch toggle directly linked to the Windows registry `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`.
  - **Window Snapping Behavior**: Cycle subdivisions toggle, almost-maximize margin %, step resize %.
  - **Configuration & Storage**: Clean display of the local config file path, plus "Open Config File", "Open Folder", and "Reload from Disk" buttons for power users.
  - **About RectangleWin**: Version display and direct links to Documentation and GitHub.

### 3. Cross-Tab Smart Search
The header search bar now operates globally across all shortcut tabs: if a user types a query while on the General or Displays tab, the UI automatically switches to the tab containing matching shortcuts.

## Consequences & Governance
* Complies with **P1 (Code Quality)**, **P2 (Zero Duplication)**, and **P5 (Separation of Concerns)**.
* Removes dead Win32 message box code (`showActiveShortcutsDialog`).
* Preserves raw configuration access for power users within the General tab without polluting the tray interface.
