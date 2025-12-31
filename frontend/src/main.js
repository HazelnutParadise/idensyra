import "./style.css";
import "bootstrap/dist/css/bootstrap.min.css";
import "@fortawesome/fontawesome-free/css/all.min.css";
import * as monaco from "monaco-editor";
import {
  ExecuteCode,
  GetVersion,
  GetDefaultCode,
  GetSymbols,
  SaveCode,
  LoadCode,
  SaveResult,
  OpenGitHub,
  OpenHazelnutParadise,
} from "../wailsjs/go/main/App";

let editor;
let liveRun = false;
let isExecuting = false;
let currentCode = "";
let goSymbols = [];
let editorFontSize = 14;
let outputFontSize = 13;
let isResizing = false;
let editorWidth = 50; // percentage
let minimapEnabled = false;
let wordWrapEnabled = false;
let currentNotification = null; // Track current notification

// Detect system theme preference
function getSystemTheme() {
  if (
    window.matchMedia &&
    window.matchMedia("(prefers-color-scheme: dark)").matches
  ) {
    return "dark";
  }
  return "light";
}

// Initialize Monaco Editor
async function initMonacoEditor(theme = "dark") {
  // Load symbols first
  try {
    goSymbols = await GetSymbols();
    console.log("Loaded symbols:", goSymbols.length);
  } catch (error) {
    console.error("Failed to load symbols:", error);
  }

  editor = monaco.editor.create(document.getElementById("code-editor"), {
    value: "",
    language: "go",
    theme: theme === "light" ? "vs-light" : "vs-dark",
    automaticLayout: true,
    fontSize: 14,
    minimap: { enabled: minimapEnabled },
    scrollBeyondLastLine: false,
    wordWrap: wordWrapEnabled ? "on" : "off",
    tabSize: 4,
    insertSpaces: false,
    lineNumbers: "on",
    renderWhitespace: "selection",
    folding: true,
    bracketPairColorization: {
      enabled: true,
    },
    suggest: {
      showKeywords: true,
      showSnippets: true,
    },
  });

  // Apply initial font size
  editor.updateOptions({ fontSize: editorFontSize });

  // Register completion provider for Go
  monaco.languages.registerCompletionItemProvider("go", {
    provideCompletionItems: (model, position) => {
      const word = model.getWordUntilPosition(position);
      const range = {
        startLineNumber: position.lineNumber,
        endLineNumber: position.lineNumber,
        startColumn: word.startColumn,
        endColumn: position.column,
      };

      const suggestions = goSymbols.map((symbol) => {
        const parts = symbol.split(".");
        const packageName = parts[0];
        const funcName = parts.slice(1).join(".");

        return {
          label: symbol,
          kind: monaco.languages.CompletionItemKind.Function,
          detail: `${packageName} package`,
          documentation: `Function from ${packageName}`,
          insertText: symbol,
          range: range,
        };
      });

      // Add Go keywords
      const keywords = [
        "break",
        "case",
        "chan",
        "const",
        "continue",
        "default",
        "defer",
        "else",
        "fallthrough",
        "for",
        "func",
        "go",
        "goto",
        "if",
        "import",
        "interface",
        "map",
        "package",
        "range",
        "return",
        "select",
        "struct",
        "switch",
        "type",
        "var",
      ];

      keywords.forEach((keyword) => {
        suggestions.push({
          label: keyword,
          kind: monaco.languages.CompletionItemKind.Keyword,
          insertText: keyword,
          range: range,
        });
      });

      // Add common Go types
      const types = [
        "string",
        "int",
        "int8",
        "int16",
        "int32",
        "int64",
        "uint",
        "uint8",
        "uint16",
        "uint32",
        "uint64",
        "float32",
        "float64",
        "bool",
        "byte",
        "rune",
        "error",
      ];

      types.forEach((type) => {
        suggestions.push({
          label: type,
          kind: monaco.languages.CompletionItemKind.TypeParameter,
          insertText: type,
          range: range,
        });
      });

      return { suggestions: suggestions };
    },
  });

  // Load default code
  GetDefaultCode().then((defaultCode) => {
    editor.setValue(defaultCode);
    currentCode = defaultCode;
  });

  // Listen for changes
  editor.onDidChangeModelContent(() => {
    currentCode = editor.getValue();
    if (liveRun && !isExecuting) {
      debounceExecute();
    }
  });
}

// Debounce function for live run
let debounceTimer;
function debounceExecute() {
  clearTimeout(debounceTimer);
  debounceTimer = setTimeout(() => {
    executeCode();
  }, 1000);
}

// Execute code
async function executeCode() {
  if (isExecuting) return;

  isExecuting = true;
  const runButton = document.getElementById("run-btn");
  const resultOutput = document.getElementById("result-output");

  // Update button state
  runButton.disabled = true;
  runButton.innerHTML = '<span class="loading"></span> Running...';

  // Show executing message
  resultOutput.innerHTML =
    '<div style="color: #4ec9b0;">Executing code...</div>';

  try {
    const code = editor.getValue();
    const theme = document.body.getAttribute("data-theme") || "dark";
    const result = await ExecuteCode(code);

    resultOutput.innerHTML = result;
  } catch (error) {
    resultOutput.innerHTML = `<div class="error-message">Error: ${error}</div>`;
  } finally {
    isExecuting = false;
    runButton.disabled = false;
    runButton.innerHTML = '<i class="fas fa-play"></i> Run Code';
  }
}

// Copy result to clipboard
function copyResult() {
  const resultOutput = document.getElementById("result-output");
  const text = resultOutput.innerText;

  if (!text) {
    showMessage("No content to copy", "error");
    return;
  }

  navigator.clipboard
    .writeText(text)
    .then(() => {
      showMessage("Result copied to clipboard!", "success");
    })
    .catch((err) => {
      showMessage("Failed to copy: " + err, "error");
    });
}

// Save code
async function saveCode() {
  try {
    const code = editor.getValue();
    await SaveCode(code);
    showMessage("Code saved successfully!", "success");
  } catch (error) {
    if (error) {
      showMessage("Failed to save code: " + error, "error");
    }
  }
}

// Save result
async function saveResult() {
  try {
    const resultOutput = document.getElementById("result-output");
    const text = resultOutput.innerText;

    if (!text) {
      showMessage("No result to save", "error");
      return;
    }

    await SaveResult(text);
    showMessage("Result saved successfully!", "success");
  } catch (error) {
    if (error) {
      showMessage("Failed to save result: " + error, "error");
    }
  }
}

// Toggle theme
function toggleTheme() {
  const currentTheme = document.body.getAttribute("data-theme");
  const newTheme = currentTheme === "light" ? "dark" : "light";
  document.body.setAttribute("data-theme", newTheme);

  if (editor) {
    monaco.editor.setTheme(newTheme === "light" ? "vs-light" : "vs-dark");
  }

  localStorage.setItem("theme", newTheme);
}

// Toggle minimap
function toggleMinimap() {
  minimapEnabled = !minimapEnabled;
  if (editor) {
    editor.updateOptions({ minimap: { enabled: minimapEnabled } });
  }
  const minimapBtn = document.getElementById("minimap-toggle");
  if (minimapEnabled) {
    minimapBtn.classList.add("active");
  } else {
    minimapBtn.classList.remove("active");
  }
  localStorage.setItem("minimapEnabled", minimapEnabled);
  showMessage(`Minimap ${minimapEnabled ? "enabled" : "disabled"}`, "success");
}

// Toggle word wrap
function toggleWordWrap() {
  wordWrapEnabled = !wordWrapEnabled;
  if (editor) {
    editor.updateOptions({ wordWrap: wordWrapEnabled ? "on" : "off" });
  }
  const wordwrapBtn = document.getElementById("wordwrap-toggle");
  if (wordWrapEnabled) {
    wordwrapBtn.classList.add("active");
  } else {
    wordwrapBtn.classList.remove("active");
  }
  localStorage.setItem("wordWrapEnabled", wordWrapEnabled);
  showMessage(
    `Word wrap ${wordWrapEnabled ? "enabled" : "disabled"}`,
    "success",
  );
}

// Undo
function undo() {
  if (editor) {
    editor.trigger("keyboard", "undo", null);
  }
}

// Redo
function redo() {
  if (editor) {
    editor.trigger("keyboard", "redo", null);
  }
}

// Show message
function showMessage(message, type = "success") {
  // Remove current notification if exists
  if (currentNotification && document.body.contains(currentNotification)) {
    currentNotification.style.animation = "slideOut 0.3s ease-out";
    const oldNotification = currentNotification;
    setTimeout(() => {
      if (document.body.contains(oldNotification)) {
        document.body.removeChild(oldNotification);
      }
    }, 300);
  }

  // Create new notification
  const messageDiv = document.createElement("div");
  messageDiv.className = `notification-message ${type}`;
  messageDiv.textContent = message;
  messageDiv.style.animation = "slideIn 0.3s ease-out";

  document.body.appendChild(messageDiv);
  currentNotification = messageDiv;

  setTimeout(() => {
    messageDiv.style.animation = "slideOut 0.3s ease-out";
    setTimeout(() => {
      if (document.body.contains(messageDiv)) {
        document.body.removeChild(messageDiv);
      }
      if (currentNotification === messageDiv) {
        currentNotification = null;
      }
    }, 300);
  }, 3000);
}

// Load code from file
async function loadCode() {
  try {
    const code = await LoadCode();
    if (code) {
      editor.setValue(code);
      showMessage("Code loaded successfully!", "success");
    }
  } catch (error) {
    if (error) {
      showMessage("Failed to load code: " + error, "error");
    }
  }
}

// Change editor font size
function changeEditorFontSize(delta) {
  editorFontSize = Math.max(8, Math.min(32, editorFontSize + delta));
  editor.updateOptions({ fontSize: editorFontSize });
  document.getElementById("editor-font-size").textContent = editorFontSize;
}

// Change output font size
function changeOutputFontSize(delta) {
  outputFontSize = Math.max(8, Math.min(32, outputFontSize + delta));
  const resultContainer = document.querySelector(".result-container");
  resultContainer.style.fontSize = outputFontSize + "px";
  document.getElementById("output-font-size").textContent = outputFontSize;
}

// Initialize resizer
function initResizer() {
  const resizer = document.getElementById("resizer");
  const editorSection = document.querySelector(".editor-section");
  const resultSection = document.querySelector(".result-section");
  const mainContent = document.querySelector(".main-content");

  resizer.addEventListener("mousedown", (e) => {
    isResizing = true;
    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";
  });

  document.addEventListener("mousemove", (e) => {
    if (!isResizing) return;

    const containerRect = mainContent.getBoundingClientRect();
    const newEditorWidth =
      ((e.clientX - containerRect.left) / containerRect.width) * 100;

    if (newEditorWidth > 20 && newEditorWidth < 80) {
      editorWidth = newEditorWidth;
      editorSection.style.width = editorWidth + "%";
      resultSection.style.width = 100 - editorWidth + "%";
    }
  });

  document.addEventListener("mouseup", () => {
    if (isResizing) {
      isResizing = false;
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
    }
  });
}

// Add CSS animations
const style = document.createElement("style");
style.textContent = `
    @keyframes slideIn {
        from {
            transform: translateX(400px);
            opacity: 0;
        }
        to {
            transform: translateX(0);
            opacity: 1;
        }
    }

    @keyframes slideOut {
        from {
            transform: translateX(0);
            opacity: 1;
        }
        to {
            transform: translateX(400px);
            opacity: 0;
        }
    }
`;
document.head.appendChild(style);

// Initialize app
async function initApp() {
  // Load saved preferences or use system theme
  const systemTheme = getSystemTheme();
  const savedTheme = localStorage.getItem("theme") || systemTheme;
  document.body.setAttribute("data-theme", savedTheme);

  // Load saved editor preferences
  minimapEnabled = localStorage.getItem("minimapEnabled") === "true";
  wordWrapEnabled = localStorage.getItem("wordWrapEnabled") === "true";

  // Setup UI
  document.getElementById("app").innerHTML = `
        <div class="header">
            <div class="header-left">
                <h1 class="header-title">Idensyra</h1>
                <span class="version-info" id="version-info">Loading version...</span>
            </div>
            <div class="header-right">
                <label class="checkbox-container">
                    <input type="checkbox" id="live-run-check">
                    <span>Live Run</span>
                </label>
                <button class="secondary icon-only" id="minimap-toggle" title="Toggle Minimap">
                    <i class="fas fa-map"></i>
                </button>
                <button class="secondary icon-only" id="wordwrap-toggle" title="Toggle Word Wrap">
                    <i class="fas fa-text-width"></i>
                </button>
                <button class="secondary icon-only" id="theme-toggle" title="Toggle Theme">
                    <i class="fas fa-adjust"></i>
                </button>
                <button class="secondary" id="github-btn" title="View on GitHub">
                    <i class="fab fa-github"></i>
                </button>
                <button class="secondary" id="hazelnut-btn" title="HazelnutParadise">
                    <i class="fas fa-link"></i>
                </button>
            </div>
        </div>
        <div class="main-content">
            <div class="editor-section">
                <div class="editor-header">
                    <span class="editor-label">Code Input</span>
                    <div class="editor-actions">
                        <div class="font-size-controls">
                            <span class="font-size-label">Font:</span>
                            <button class="secondary font-size-btn" id="editor-font-decrease">
                                <i class="fas fa-minus"></i>
                            </button>
                            <span class="font-size-display" id="editor-font-size">14</span>
                            <button class="secondary font-size-btn" id="editor-font-increase">
                                <i class="fas fa-plus"></i>
                            </button>
                        </div>
                        <button class="secondary" id="load-code-btn">
                            <i class="fas fa-folder-open"></i> Load
                        </button>
                        <button class="secondary" id="save-code-btn">
                            <i class="fas fa-save"></i> Save
                        </button>
                    </div>
                </div>
                <div class="editor-container">
                    <div id="code-editor"></div>
                </div>
            </div>
            <div id="resizer" class="resizer"></div>
            <div class="result-section">
                <div class="result-header">
                    <span class="result-label">Output</span>
                    <div class="result-actions">
                        <div class="font-size-controls">
                            <span class="font-size-label">Font:</span>
                            <button class="secondary font-size-btn" id="output-font-decrease">
                                <i class="fas fa-minus"></i>
                            </button>
                            <span class="font-size-display" id="output-font-size">13</span>
                            <button class="secondary font-size-btn" id="output-font-increase">
                                <i class="fas fa-plus"></i>
                            </button>
                        </div>
                        <button class="success" id="run-btn">
                            <i class="fas fa-play"></i> Run Code
                        </button>
                        <button class="secondary" id="copy-result-btn">
                            <i class="fas fa-copy"></i> Copy
                        </button>
                        <button class="secondary" id="save-result-btn">
                            <i class="fas fa-save"></i> Save
                        </button>
                    </div>
                </div>
                <div class="result-container">
                    <div id="result-output" class="result-output">
                        <div style="color: #888;">Run your code to see output here...</div>
                    </div>
                </div>
            </div>
        </div>
    `;

  // Load version info after UI is created
  try {
    const versionInfo = await GetVersion();
    document.getElementById("version-info").textContent =
      `v${versionInfo.idensyra} with Insyra v${versionInfo.insyra}`;
  } catch (error) {
    console.error("Failed to load version info:", error);
    document.getElementById("version-info").textContent = "Version unavailable";
  }

  // Initialize Monaco Editor with theme
  initMonacoEditor(savedTheme);

  // Initialize resizer
  initResizer();

  // Set initial panel sizes
  const editorSection = document.querySelector(".editor-section");
  const resultSection = document.querySelector(".result-section");
  editorSection.style.width = editorWidth + "%";
  resultSection.style.width = 100 - editorWidth + "%";

  // Apply initial output font size
  const resultContainer = document.querySelector(".result-container");
  resultContainer.style.fontSize = outputFontSize + "px";

  // Setup event listeners
  document.getElementById("run-btn").addEventListener("click", executeCode);
  document
    .getElementById("copy-result-btn")
    .addEventListener("click", copyResult);
  document.getElementById("save-code-btn").addEventListener("click", saveCode);
  document.getElementById("load-code-btn").addEventListener("click", loadCode);
  document
    .getElementById("save-result-btn")
    .addEventListener("click", saveResult);
  document
    .getElementById("theme-toggle")
    .addEventListener("click", toggleTheme);
  document
    .getElementById("github-btn")
    .addEventListener("click", () => OpenGitHub());
  document
    .getElementById("hazelnut-btn")
    .addEventListener("click", () => OpenHazelnutParadise());

  // Font size controls
  document
    .getElementById("editor-font-decrease")
    .addEventListener("click", () => changeEditorFontSize(-1));
  document
    .getElementById("editor-font-increase")
    .addEventListener("click", () => changeEditorFontSize(1));
  document
    .getElementById("output-font-decrease")
    .addEventListener("click", () => changeOutputFontSize(-1));
  document
    .getElementById("output-font-increase")
    .addEventListener("click", () => changeOutputFontSize(1));

  document.getElementById("live-run-check").addEventListener("change", (e) => {
    liveRun = e.target.checked;
    if (liveRun) {
      showMessage(
        "Live Run enabled - code will execute automatically",
        "success",
      );
    } else {
      showMessage("Live Run disabled", "success");
    }
  });

  // Minimap toggle
  document
    .getElementById("minimap-toggle")
    .addEventListener("click", toggleMinimap);

  // Word wrap toggle
  document
    .getElementById("wordwrap-toggle")
    .addEventListener("click", toggleWordWrap);

  // Set initial button states
  if (minimapEnabled) {
    document.getElementById("minimap-toggle").classList.add("active");
  }
  if (wordWrapEnabled) {
    document.getElementById("wordwrap-toggle").classList.add("active");
  }

  // Keyboard shortcuts
  document.addEventListener("keydown", (e) => {
    // Ctrl/Cmd + Enter to run
    if ((e.ctrlKey || e.metaKey) && e.key === "Enter") {
      e.preventDefault();
      executeCode();
    }
    // Ctrl/Cmd + S to save code
    if ((e.ctrlKey || e.metaKey) && e.key === "s") {
      e.preventDefault();
      saveCode();
    }
    // Ctrl/Cmd + Z to undo
    if ((e.ctrlKey || e.metaKey) && e.key === "z" && !e.shiftKey) {
      e.preventDefault();
      undo();
    }
    // Ctrl/Cmd + Shift + Z or Ctrl/Cmd + Y to redo
    if (
      ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key === "Z") ||
      ((e.ctrlKey || e.metaKey) && e.key === "y")
    ) {
      e.preventDefault();
      redo();
    }
  });
}

// Start the app when DOM is ready
if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initApp);
} else {
  initApp();
}
