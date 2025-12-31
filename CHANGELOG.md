# Changelog

All notable changes to this project will be documented in this file.

## [0.1.0] - 2025-12-31

### ✨ New Features

#### Virtual Workspace System
- **Temporary Workspace Management**: Automatic creation and cleanup of temporary workspace
  - All generated files stored in system temp directory
  - Automatic cleanup when application closes
  - Warning dialog when closing with unsaved changes
- **Multiple File Support**: Create and manage multiple Go files in workspace
  - Each new file initialized with default template
  - Easy switching between files via file tree sidebar
  - File modification status tracking (● indicator for unsaved changes)
- **File Operations**:
  - Create new Go files with custom names
  - Delete files (with confirmation)
  - Auto-save every 2 seconds after editing
  - Active file highlighted in blue
- **Workspace Export**: One-click export entire workspace to persistent location
  - Exports to `Documents/idensyra-exports/` with timestamp
  - All files exported with complete `package main` structure
  - Marks workspace as saved after successful export
- **File Tree Sidebar**: Visual file management interface
  - Icons and names for all workspace files
  - Modification indicators
  - Quick file switching
  - Hover-to-delete functionality

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
- **Backend Workspace Manager** (`workspace.go`):
  - `InitWorkspace()`: Creates temporary workspace on app start
  - `GetWorkspaceFiles()`: Returns all files in workspace
  - `CreateNewFile()`, `DeleteFile()`, `RenameFile()`: File operations
  - `UpdateFileContent()`: Saves file content to temp directory
  - `ExportWorkspace()`: Exports to persistent location
  - `IsWorkspaceModified()`: Tracks unsaved changes
  - `CleanupWorkspace()`: Automatic cleanup on shutdown
  - `beforeClose()`: Prevents closing with unsaved changes
- **Frontend Workspace UI**:
  - File tree sidebar component with icons and indicators
  - `loadWorkspaceFiles()`: Syncs file list from backend
  - `switchToFile()`: Changes active file and loads content
  - `createNewFile()`: Prompts for filename and creates file
  - `deleteFileConfirm()`: Confirms and deletes file
  - `exportWorkspace()`: Triggers workspace export
  - `saveCurrentFile()`: Manual save with auto-save timer
  - `beforeunload` event handler for unsaved changes warning
- Added system theme detection using `window.matchMedia('prefers-color-scheme: dark')`
- Improved Monaco Editor initialization to accept theme parameter
- Added notification queue management to prevent message stacking
- Fixed autocompletion to insert full symbol names (changed `insertText` from `funcName` to `symbol`)
- Fixed autocompletion range to use `position.column` instead of `word.endColumn`
- All frontend assets (Monaco Editor, Bootstrap, Font Awesome) fully localized without CDN dependencies
- Application lifecycle hooks: `OnDomReady`, `OnBeforeClose`, `OnShutdown`

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
