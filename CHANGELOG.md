# Changelog

All notable changes to this project will be documented in this file.

## [0.2.0] - 2025-12-31

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

### 🔧 Technical Changes
- Added system theme detection using `window.matchMedia('prefers-color-scheme: dark')`
- Improved Monaco Editor initialization to accept theme parameter
- Added notification queue management to prevent message stacking
- All frontend assets (Monaco Editor, Bootstrap, Font Awesome) fully localized without CDN dependencies

---

## [0.1.0] - 2024-12-31

### 🎉 Major Changes

#### Migrated from Fyne UI to Wails v2

The entire project has been migrated from Fyne (native Go UI) to Wails v2 (WebView + Go backend), providing a modern web-based interface with better performance and smaller binary size.

### ✨ New Features

#### Frontend
- **Monaco Editor Integration**: Professional code editor with VS Code-like features
  - Syntax highlighting for Go
  - IntelliSense and auto-completion
  - Minimap navigation
  - Line numbers and code folding
  - Bracket pair colorization
  
- **Modern UI/UX**:
  - Responsive design with Bootstrap 5
  - Dark/Light theme toggle with persistent preference
  - Font Awesome icons
  - Split view with resizable panels
  - Smooth animations and transitions

- **Local Assets**: All dependencies are now bundled locally
  - Monaco Editor (no CDN)
  - Bootstrap 5 (no CDN)
  - Font Awesome (no CDN)
  - Faster loading and offline support

#### Backend
- **Wails Integration**: 
  - Native file dialogs for save/load operations
  - System browser integration
  - Smaller memory footprint
  - Faster startup time

- **Enhanced API**:
  ```go
  - ExecuteCode(code string) string
  - ExecuteCodeWithColorBG(code string, colorBG string) string
  - GetVersion() map[string]string
  - GetDefaultCode() string
  - GetSymbols() []string
  - SaveCode(code string) error
  - SaveResult(result string) error
  - OpenGitHub()
  - OpenHazelnutParadise()
  ```

#### Features
- **Live Run Mode**: Automatically execute code on edit with debouncing
- **Keyboard Shortcuts**:
  - `Ctrl/Cmd + Enter`: Run code
  - `Ctrl/Cmd + S`: Save code
- **Copy to Clipboard**: One-click result copying
- **File Operations**: Native dialogs for saving and loading
- **Theme Persistence**: Remembers user's theme preference

### 🗑️ Removed

#### Dependencies
- `fyne.io/fyne/v2` - Replaced by Wails
- `github.com/gorilla/websocket` - No longer needed

#### Features
- Web UI Mode - No longer needed (Wails itself is web-based)
- WebSocket connection - Replaced by Wails bindings
- Go template engine - Replaced by native HTML/JS

### 🔧 Technical Changes

#### Project Structure
```
idensyra/
├── app.go                 # Wails application backend
├── main.go                # Application entry point
├── wails.json             # Wails configuration
├── internal/              # Yaegi symbols
├── frontend/              # Frontend code
│   ├── src/
│   │   ├── main.js       # Main JavaScript
│   │   └── style.css     # Styles
│   ├── vite.config.js    # Vite configuration
│   ├── index.html
│   └── package.json
└── build/                 # Build output
```

#### Build System
- **Vite**: Modern frontend build tool
- **Monaco Editor Plugin**: Proper worker handling
- **Manual Chunks**: Optimized code splitting
- **Hot Module Replacement**: Fast development iteration

#### Dependencies Updated
- `github.com/HazelnutParadise/insyra`: v0.2.10
- `github.com/traefik/yaegi`: v0.16.1
- `gorm.io/gorm`: v1.31.1

#### New Dependencies
- `github.com/wailsapp/wails/v2`: v2.11.0
- `monaco-editor`: v0.55.1
- `bootstrap`: v5.3.8
- `@fortawesome/fontawesome-free`: v7.1.0
- `vite-plugin-monaco-editor`: Latest

### 📦 Build & Deployment

#### Development
```bash
wails dev
```
- Hot reload support
- Browser debugging available at http://localhost:34115

#### Production Build
```bash
wails build
```
- Optimized bundles
- Minified assets
- Single executable output

#### Cross-Platform
```bash
wails build -platform darwin/amd64   # macOS
wails build -platform linux/amd64    # Linux
wails build -platform windows/amd64  # Windows
```

### 🐛 Bug Fixes
- Fixed ANSI color rendering in output
- Improved error messages
- Better memory management

### 📚 Documentation
- New comprehensive README.md
- Added WAILS_MIGRATION.md guide
- Updated CONTRIBUTING.md
- Added inline code comments

### ⚡ Performance Improvements
- Faster startup time (native WebView vs Fyne)
- Reduced memory usage
- Smaller binary size
- Optimized asset loading

### 🔐 Security
- No external CDN dependencies
- Local asset bundling
- Secure file operations with native dialogs

### 🎨 UI/UX Improvements
- Professional code editor experience
- Better color schemes
- Responsive layout
- Intuitive controls
- Status messages and notifications

### 📝 Notes

#### Breaking Changes
- Command line flags removed (not applicable with Wails)
- Window size and position handling changed
- Configuration format changed (wails.json instead of Fyne preferences)

#### Migration Guide
See [WAILS_MIGRATION.md](WAILS_MIGRATION.md) for detailed migration information.

#### Known Issues
- First run on Windows requires WebView2 Runtime
- Monaco Editor CSS warnings during build (cosmetic only)

#### Future Plans
- [ ] Code snippets library
- [ ] Multi-file project support
- [ ] Debugging tools
- [ ] Custom themes
- [ ] Auto-save functionality
- [ ] Code formatting
- [ ] Search and replace
- [ ] Settings panel

---

## [0.0.6] - 2024 (Previous Version)

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