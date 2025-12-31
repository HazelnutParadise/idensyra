# Changelog

All notable changes to this project will be documented in this file.

## [0.1.0] - 2025-12-31

### ✨ New Features

#### Theme System
- **Follow System Theme**: Automatically detects and follows OS dark/light theme on startup
  - Windows: Follows "Settings > Personalization > Colors"
  - macOS: Follows "System Preferences > General > Appearance"
- **Monaco Theme Consistency**: Editor theme now syncs with application theme (no longer defaults to dark)
- **Theme Persistence**: Manual theme changes are saved and restored on next launch

#### Editor Features
- **Undo/Redo Support**: Full undo/redo functionality with keyboard shortcuts
  - Undo: `Ctrl + Z` (Windows/Linux) or `Cmd + Z` (macOS)
  - Redo: `Ctrl + Shift + Z` or `Ctrl + Y` (Windows/Linux), `Cmd + Shift + Z` or `Cmd + Y` (macOS)
- **Minimap Toggle**: Toggle code minimap on/off with toolbar button
  - Defaults to off (saves screen space)
  - Active state shown with blue highlight
  - Setting persisted across sessions
- **Word Wrap Toggle**: Choose between line wrapping or horizontal scrolling
  - Defaults to off (horizontal scroll)
  - Active state shown with blue highlight
  - Setting persisted across sessions

#### User Experience
- **Smart Notifications**: New notifications automatically hide old ones to avoid stacking
- **Visual Feedback**: Active feature buttons show blue highlight for clear status indication
- **Complete Keyboard Support**: All major features support keyboard shortcuts

### 🔄 Improvements
- **Settings Persistence**: All user preferences automatically saved to localStorage
  - Theme selection
  - Minimap enabled/disabled
  - Word wrap enabled/disabled
  - Editor font size
  - Output font size
  - Panel split ratio

### 🗑️ Removed
- **Live Run Confirmation**: Removed confirmation dialog when disabling Live Run (now shows notification only)

### 🐛 Bug Fixes
- **Autocompletion**: Fixed completion suggestions inserting only partial text (e.g., typing `in` and selecting `insyra.Config` now correctly inserts `insyra.Config` instead of just `Config`)
- **Autocompletion Range**: Fixed completion range to prevent overwriting previous text

### 🔧 Technical Changes
- Added system theme detection using `window.matchMedia('prefers-color-scheme: dark')`
- Improved Monaco Editor initialization to accept theme parameter
- Added notification queue management to prevent message stacking
- Fixed autocompletion to insert full symbol names (changed `insertText` from `funcName` to `symbol`)
- Fixed autocompletion range to use `position.column` instead of `word.endColumn`
- All frontend assets (Monaco Editor, Bootstrap, Font Awesome) fully localized without CDN dependencies

---

## [0.0.6] - (Previous Version)

### Features
- Fyne UI based desktop application
- Basic code editor
- Code execution with Yaegi
- Web UI mode
- WebSocket communication
- File save/load operations

### Dependencies
- Fyne v2.5.1
- Insyra (older version)
- Yaegi v0.16.1
- Gorilla WebSocket

---

## Credits

- **Author**: TimLai666
- **Email**: tim930102@icloud.com
- **Organization**: [HazelnutParadise](https://hazelnut-paradise.com)

## Links

- [GitHub Repository](https://github.com/HazelnutParadise/idensyra)
- [Insyra Library](https://insyra.hazelnut-paradise.com)
- [Wails Framework](https://wails.io)

---

*For more information about how to use this application, see [README.md](README.md)*
