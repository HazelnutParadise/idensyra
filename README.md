# Idensyra

`Idensyra` is a mini Go IDE, with [Insyra](https://github.com/HazelnutParadise/insyra) pre-installed. Run Go code without installing Go environment!

**Get Idensyra executable file [HERE](https://github.com/HazelnutParadise/idensyra/releases).**

## [Insyra](https://github.com/HazelnutParadise/insyra) Inside: Simple? It's quite!

With `Idensyra`, you can write Go code, and run it immediately, no need to install Go environment. Also, you can use [Insyra](https://github.com/HazelnutParadise/insyra)'s functions directly in `Idensyra`.

## User Interface

Idensyra has two modes:

- **GUI mode**: Build with [Fyne](https://github.com/fyne-io/fyne), supports live run, but no code highlighting.
- **WebUI mode**: WebUI with [monaco-editor](https://github.com/microsoft/monaco-editor), provides code highlighting, but no live run.

### GUI

![GUI example](./gui_example.png)

### WebUI

![WebUI example](./webui_example.png)

## Limitation

- No third party package support (except Insyra and its dependencies).
- Does not support `py` and `lp` package of Insyra.
- Currently no code highlighting.
