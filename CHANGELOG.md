# Changelog

All notable changes to this project will be documented in this file.

## [NEXT]

## [0.2.0] - 2026-01-04

### New Features

#### igonb Notebook

- New `.igonb` file format: Jupyter-like interactive notebooks for Go
- Multi-language cells: Go, Python, and Markdown support
- Flexible execution modes:
  - Run single cell
  - Run cell with all cells above
  - Run from cell downward
- Cell management: drag-and-drop reordering, add, delete, collapse
- Markdown live preview in notebook cells
- Execution control: stop execution, reset environment
- Output modes: Full (complete output) / Compact (condensed output)
- Auto-save notebook content

#### Python Support

- Execute `.py` files directly from the workspace
- Built-in Python package manager (pip list/install/uninstall)
- Reinstall Python environment option
- Go-Python interoperability: share variables between Go and Python cells in igonb

#### IPython Notebook Support

- Open and preview `.ipynb` files
- One-click conversion from `.ipynb` to `.igonb` format

#### Workspace Enhancements

- File drag-and-drop for moving and reordering
- Auto-save temporary workspace content

### Improvements

- Go range over integers syntax support (Go 1.22+)
- REPL-style execution: auto-display last expression value
- Updated Go requirement to 1.25
- Updated Insyra to v0.2.12

### Technical Changes

- Added `igonb/` module for notebook core functionality
  - `igonb.go`: Notebook structure and parsing
  - `runner.go`: Executor management with multi-key support
  - `execute.go`: Cell execution logic with callback support
  - `python_bridge.go`: Go-Python variable interoperability
- Added `igonb_exec.go` for notebook execution bindings
- Added `python_exec.go` for Python file execution
- Added `python_packages.go` for pip management via Insyra py module
- Frontend: Added notebook UI components and cell editors
- Frontend: Added Python package manager modal

---

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
