# Changelog

All notable changes to this project will be documented in this file.

## [NEXT]

## [0.1.0] - 2026-01-01

### New Features

#### Workspace & File Management

- Temporary workspace on startup with save prompt on exit
- Create or open workspace folders and save files to disk
- File/folder tree with nested folders
- New file/folder, rename, delete (cannot delete last file)
- Import external files of any type with progress UI
- Export current file to a user-selected location
- Modified (`*`) and large-file (`L`) indicators
- Large/binary files open in preview-only mode

#### Preview & Media

- HTML/Markdown preview in output panel
- CSV/TSV table preview
- Excel preview with sheet selector (`.xlsx`/`.xlsm`/`.xltx`/`.xltm`)
- Image/video/audio/PDF preview
- Output panel switches between Output and Preview modes

#### Editor & Execution

- Multi-language syntax highlighting
- Go completion: stdlib + Insyra + struct fields/methods
- Minimap and Word Wrap toggles
- Run only for `.go` files; Live Run with debounce

#### Output

- Save output directly to workspace (`result.txt` / `result_#.txt`)
- Copy output to clipboard
- ANSI color rendering for terminal-style output

### Improvements

- Theme follows system with manual toggle and Monaco sync
- Progress overlay for import/open workflows
- Safer workspace switch with unsaved-change confirmation

### Bug Fixes

- Autocompletion now inserts full symbol names
- Autocompletion range respects cursor position

### Technical Changes

- Added workspace manager for folder operations and open/create flows
- Added Excel preview backend via `excelize` v2.10.0
- Updated Insyra to v0.2.11
- All frontend assets localized without CDN dependencies
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

_For more information about how to use this application, see [README.md](README.md)_
