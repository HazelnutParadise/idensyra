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
  GetWorkspaceFiles,
  GetActiveFile,
  SetActiveFile,
  GetFileContent,
  UpdateFileContent,
  CreateNewFile,
  DeleteFile,
  SaveFile,
  SaveAllFiles,
  OpenWorkspace,
  CreateWorkspace,
  ImportFileToWorkspace,
  ExportCurrentFile,
  IsWorkspaceModified,
  GetWorkspaceInfo,
} from "../wailsjs/go/main/App";

const RenameFile = (...args) => window.go.main.App.RenameFile(...args);
const SaveResultToWorkspace = (...args) =>
  window.go.main.App.SaveResultToWorkspace(...args);

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
let workspaceFiles = [];
let activeFileName = "";
let isWorkspaceInitialized = false;
let isImagePreview = false; // Track if current file is an image

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

// Get Monaco language from file extension
function getLanguageFromFilename(filename) {
  const ext = filename.split(".").pop().toLowerCase();
  const languageMap = {
    go: "go",
    js: "javascript",
    ts: "typescript",
    jsx: "javascript",
    tsx: "typescript",
    json: "json",
    html: "html",
    htm: "html",
    css: "css",
    scss: "scss",
    sass: "sass",
    less: "less",
    md: "markdown",
    txt: "plaintext",
    xml: "xml",
    yaml: "yaml",
    yml: "yaml",
    py: "python",
    rb: "ruby",
    java: "java",
    c: "c",
    cpp: "cpp",
    cs: "csharp",
    php: "php",
    sh: "shell",
    bash: "shell",
    sql: "sql",
    r: "r",
    swift: "swift",
    kt: "kotlin",
    rs: "rust",
    dockerfile: "dockerfile",
  };
  return languageMap[ext] || "plaintext";
}

// Check if file is an image
function isImageFile(filename) {
  const ext = filename.split(".").pop().toLowerCase();
  return ["jpg", "jpeg", "png", "gif", "bmp", "webp", "svg", "ico"].includes(
    ext,
  );
}

// Show image preview
function showImagePreview(filename, base64Data) {
  const editorContainer = document.getElementById("code-editor");
  editorContainer.style.display = "none";

  let imageContainer = document.getElementById("image-preview-container");
  if (!imageContainer) {
    imageContainer = document.createElement("div");
    imageContainer.id = "image-preview-container";
    imageContainer.style.cssText = `
      width: 100%;
      height: 100%;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      background: var(--panel-background-color);
      padding: 20px;
      box-sizing: border-box;
      overflow: auto;
    `;
    editorContainer.parentElement.appendChild(imageContainer);
  }

  const ext = filename.split(".").pop().toLowerCase();
  const mimeType = ext === "svg" ? "image/svg+xml" : `image/${ext}`;

  imageContainer.innerHTML = `
    <div style="text-align: center; width: 100%;">
      <div style="margin-bottom: 15px; color: var(--text-color); font-size: 14px;">
        <i class="fas fa-image"></i> ${filename}
      </div>
      <img src="data:${mimeType};base64,${base64Data}"
           alt="${filename}"
           style="max-width: 100%; max-height: calc(100% - 60px); object-fit: contain;
                  border: 1px solid var(--border-color); border-radius: 4px; background: white;" />
      <div style="margin-top: 15px; color: var(--text-color); opacity: 0.7; font-size: 12px;">
        Image preview mode - this file cannot be edited
      </div>
    </div>
  `;
  imageContainer.style.display = "flex";
  isImagePreview = true;
}

// Hide image preview and show editor
function hideImagePreview() {
  const imageContainer = document.getElementById("image-preview-container");
  if (imageContainer) {
    imageContainer.style.display = "none";
  }
  const editorContainer = document.getElementById("code-editor");
  editorContainer.style.display = "block";
  isImagePreview = false;
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
    insertSpaces: true, // Use spaces for non-Go files
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
    runButton.innerHTML = '<i class="fas fa-play"></i> Run';
    await loadWorkspaceFiles();
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

    const savedPath = await SaveResultToWorkspace(text);
    await loadWorkspaceFiles();
    showMessage(`Result saved to workspace: ${savedPath}`, "success");
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

// Workspace functions
async function loadWorkspaceFiles() {
  try {
    workspaceFiles = await GetWorkspaceFiles();
    activeFileName = await GetActiveFile();
    renderFileTree();
  } catch (error) {
    console.error("Failed to load workspace files:", error);
  }
}

// Load workspace with retry logic to ensure backend is ready
async function loadWorkspaceWithRetry(maxRetries = 5, delayMs = 100) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      workspaceFiles = await GetWorkspaceFiles();
      activeFileName = await GetActiveFile();

      // Check if we got any files
      if (workspaceFiles && workspaceFiles.length > 0) {
        renderFileTree();
        console.log(
          "Workspace loaded successfully with",
          workspaceFiles.length,
          "files",
        );
        return true;
      }

      // If no files yet, wait and retry
      console.log(`Workspace not ready, retrying (${i + 1}/${maxRetries})...`);
      await new Promise((resolve) => setTimeout(resolve, delayMs));
    } catch (error) {
      console.error(`Failed to load workspace (attempt ${i + 1}):`, error);
      await new Promise((resolve) => setTimeout(resolve, delayMs));
    }
  }

  console.error("Failed to load workspace after", maxRetries, "attempts");
  return false;
}

function renderFileTree() {
  const fileTree = document.getElementById("file-tree");
  if (!fileTree) return;

  fileTree.innerHTML = "";

  workspaceFiles.forEach((file) => {
    const fileItem = document.createElement("div");
    fileItem.className = "file-item";
    if (file.name === activeFileName) {
      fileItem.classList.add("active");
    }
    if (file.modified) {
      fileItem.classList.add("modified");
    }

    // Choose icon based on file type
    let iconClass = "fa-file-code";
    if (isImageFile(file.name)) {
      iconClass = "fa-file-image";
    } else if (file.name.endsWith(".md")) {
      iconClass = "fa-file-lines";
    } else if (file.name.endsWith(".json")) {
      iconClass = "fa-file-code";
    } else if (file.name.endsWith(".txt")) {
      iconClass = "fa-file-lines";
    }

    fileItem.innerHTML = `
      <i class="fas ${iconClass}"></i>
      <span class="file-name">${file.name}</span>
      ${file.modified ? '<span class="modified-indicator">*</span>' : ""}
      <button class="file-rename-btn" title="Rename file">
        <i class="fas fa-pen"></i>
      </button>
      <button class="file-delete-btn" title="Delete file">
        <i class="fas fa-times"></i>
      </button>
    `;

    fileItem.addEventListener("click", (e) => {
      if (
        !e.target.closest(".file-delete-btn") &&
        !e.target.closest(".file-rename-btn")
      ) {
        switchToFile(file.name);
      }
    });

    const renameBtn = fileItem.querySelector(".file-rename-btn");
    renameBtn.addEventListener("click", (e) => {
      e.stopPropagation();
      renameFilePrompt(file.name);
    });

    const deleteBtn = fileItem.querySelector(".file-delete-btn");
    deleteBtn.addEventListener("click", (e) => {
      e.stopPropagation();
      deleteFileConfirm(file.name);
    });

    fileTree.appendChild(fileItem);
  });
}

async function switchToFile(filename, force = false) {
  if (!force && filename === activeFileName) return;

  try {
    // Save current file content (only if not in image preview mode)
    if (activeFileName && !isImagePreview) {
      const isKnownFile = workspaceFiles.some(
        (file) => file.name === activeFileName,
      );
      if (isKnownFile) {
        const currentContent = editor.getValue();
        await UpdateFileContent(activeFileName, currentContent);
      } else {
        activeFileName = "";
        hideImagePreview();
      }
    }

    // Switch to new file
    await SetActiveFile(filename);
    const content = await GetFileContent(filename);
    activeFileName = filename;

    // Check if this is an image file
    if (isImageFile(filename)) {
      showImagePreview(filename, content);
    } else {
      // Hide image preview if it was showing
      hideImagePreview();

      // Set the language based on file extension
      const language = getLanguageFromFilename(filename);
      const model = editor.getModel();
      monaco.editor.setModelLanguage(model, language);

      // Set editor content
      editor.setValue(content);

      // Adjust editor options based on file type
      if (language === "go") {
        editor.updateOptions({ insertSpaces: false, tabSize: 4 });
      } else {
        editor.updateOptions({ insertSpaces: true, tabSize: 2 });
      }
    }

    // Refresh file tree
    await loadWorkspaceFiles();
  } catch (error) {
    console.error("Failed to switch file:", error);
    showMessage("Failed to switch file: " + error, "error");
  }
}

async function createNewFile() {
  const filename = prompt(
    "Enter new file name (e.g., test.go, notes.txt, config.json):",
  );
  if (!filename) return;

  try {
    await CreateNewFile(filename);
    await loadWorkspaceFiles();
    showMessage(`File "${filename}" created successfully`, "success");
    await switchToFile(filename);
  } catch (error) {
    console.error("Failed to create file:", error);
    showMessage(`Failed to create file: ${error}`, "error");
  }
}

async function deleteFileConfirm(filename) {
  if (workspaceFiles.length <= 1) {
    showMessage("Cannot delete the last file in workspace", "error");
    return;
  }

  if (!confirm(`Delete file "${filename}"?`)) return;

  try {
    await DeleteFile(filename);
    activeFileName = "";
    await loadWorkspaceFiles();

    // If deleted file was active, switch to first available
    if (workspaceFiles.length > 0) {
      hideImagePreview();
      await switchToFile(workspaceFiles[0].name, true);
    }

    showMessage(`File "${filename}" deleted`, "success");
  } catch (error) {
    console.error("Failed to delete file:", error);
    showMessage("Failed to delete file: " + error, "error");
  }
}

async function renameFilePrompt(filename) {
  const newName = prompt("Enter new file name:", filename);
  if (!newName) return;

  const trimmedName = newName.trim();
  if (!trimmedName) {
    showMessage("File name cannot be empty", "error");
    return;
  }
  if (trimmedName === filename) {
    return;
  }
  if (workspaceFiles.some((file) => file.name === trimmedName)) {
    showMessage(`File "${trimmedName}" already exists`, "error");
    return;
  }

  try {
    if (filename === activeFileName && !isImagePreview) {
      const currentContent = editor.getValue();
      await UpdateFileContent(filename, currentContent);
    }

    await RenameFile(filename, trimmedName);
    const wasActive = filename === activeFileName;

    activeFileName = "";
    await loadWorkspaceFiles();

    if (wasActive) {
      hideImagePreview();
      await switchToFile(trimmedName, true);
    }

    showMessage(`File renamed to "${trimmedName}"`, "success");
  } catch (error) {
    console.error("Failed to rename file:", error);
    showMessage("Failed to rename file: " + error, "error");
  }
}

async function saveCurrentFile() {
  if (!activeFileName) return;

  // Cannot save image files
  if (isImagePreview) {
    showMessage("Image files cannot be edited", "warning");
    return;
  }

  try {
    const currentContent = editor.getValue();
    await UpdateFileContent(activeFileName, currentContent);

    // Try to save to disk
    await SaveFile(activeFileName);
    await loadWorkspaceFiles();
    showMessage(`Saved ${activeFileName}`, "success");
  } catch (error) {
    console.error("Failed to save file:", error);
    if (error && error.toString().includes("temporary workspace")) {
      // Prompt to create workspace
      if (
        confirm(
          "You are in a temporary workspace. Would you like to create a workspace folder to save your files?",
        )
      ) {
        await createWorkspace();
      }
    } else {
      showMessage("Failed to save file: " + error, "error");
    }
  }
}

async function saveAllFiles() {
  try {
    // Update current file content first (only if not in image preview mode)
    if (activeFileName && !isImagePreview) {
      const currentContent = editor.getValue();
      await UpdateFileContent(activeFileName, currentContent);
    }

    await SaveAllFiles();
    await loadWorkspaceFiles();
    showMessage("All files saved", "success");
  } catch (error) {
    console.error("Failed to save files:", error);
    if (error && error.toString().includes("temporary workspace")) {
      if (
        confirm(
          "You are in a temporary workspace. Would you like to create a workspace folder to save your files?",
        )
      ) {
        await createWorkspace();
      }
    } else {
      showMessage("Failed to save files: " + error, "error");
    }
  }
}

async function createWorkspace() {
  try {
    const workspacePath = await CreateWorkspace();
    if (!workspacePath) {
      return; // User cancelled
    }

    await loadWorkspaceFiles();
    showMessage(`Workspace created at: ${workspacePath}`, "success");
  } catch (error) {
    console.error("Failed to create workspace:", error);
    showMessage("Failed to create workspace: " + error, "error");
  }
}

async function exportCurrentFile() {
  try {
    // Update current file content first
    if (activeFileName) {
      const currentContent = editor.getValue();
      await UpdateFileContent(activeFileName, currentContent);
    }

    await ExportCurrentFile();
    showMessage(`File exported successfully`, "success");
  } catch (error) {
    console.error("Failed to export file:", error);
    if (error && error.toString().includes("User cancelled")) {
      return;
    }
    showMessage("Failed to export file: " + error, "error");
  }
}

async function openWorkspace() {
  try {
    // Check for unsaved changes
    const modified = await IsWorkspaceModified();
    if (modified) {
      if (
        !confirm(
          "Opening a workspace will discard all unsaved changes. Continue?",
        )
      ) {
        return;
      }
    }

    const workspacePath = await OpenWorkspace();
    if (!workspacePath) {
      return; // User cancelled
    }

    await loadWorkspaceFiles();

    // Load the active file
    if (workspaceFiles.length > 0) {
      const activeFile = await GetActiveFile();
      const content = await GetFileContent(activeFile);
      editor.setValue(content);
      activeFileName = activeFile;
      document.getElementById("active-file-label").textContent = activeFile;
    }

    showMessage(`Workspace opened: ${workspacePath}`, "success");
  } catch (error) {
    console.error("Failed to open workspace:", error);
    if (error && error.toString().includes("User cancelled")) {
      return;
    }
    showMessage("Failed to open workspace: " + error, "error");
  }
}

async function importFileToWorkspace() {
  try {
    await ImportFileToWorkspace();
    await loadWorkspaceFiles();

    // Get the newly added file (last modified file in the list)
    if (workspaceFiles.length > 0) {
      const lastFile = workspaceFiles[workspaceFiles.length - 1];
      showMessage(`File "${lastFile.name}" imported successfully`, "success");
      // Switch to the newly imported file
      await switchToFile(lastFile.name);
    }
  } catch (error) {
    console.error("Failed to import file:", error);
    showMessage("Failed to import file: " + error, "error");
  }
}

// Note: beforeunload removed to allow proper window closing in Wails

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

  // Setup UI with workspace sidebar
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
            <div class="workspace-sidebar">
                <div class="workspace-header">
                    <span class="workspace-label">Workspace</span>
                    <div class="workspace-buttons">
                        <button class="secondary icon-only" id="new-file-btn" title="New File (Ctrl+N)">
                            <i class="fas fa-file-circle-plus"></i>
                        </button>
                        <button class="secondary icon-only" id="import-file-btn" title="Import File to Workspace">
                            <i class="fas fa-file-import"></i>
                        </button>
                        <button class="secondary icon-only" id="open-workspace-btn" title="Open Workspace Folder">
                            <i class="fas fa-folder-open"></i>
                        </button>
                        <button class="secondary icon-only" id="save-workspace-btn" title="Save All Files (Ctrl+Shift+S)">
                            <i class="fas fa-save"></i>
                        </button>
                    </div>
                </div>
                <div id="file-tree" class="file-tree"></div>
            </div>
            <div class="editor-section">
                <div class="editor-header">
                    <span class="editor-label" id="active-file-label">Code Input</span>
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
                        <button class="secondary" id="save-code-btn" title="Save Current File (Ctrl+S)">
                            <i class="fas fa-save"></i> Save
                        </button>
                        <button class="secondary" id="export-file-btn" title="Export Current File">
                            <i class="fas fa-file-export"></i> Export
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
                            <i class="fas fa-play"></i> Run
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
  document
    .getElementById("save-code-btn")
    .addEventListener("click", saveCurrentFile);
  document
    .getElementById("export-file-btn")
    .addEventListener("click", exportCurrentFile);
  document
    .getElementById("new-file-btn")
    .addEventListener("click", createNewFile);
  document
    .getElementById("import-file-btn")
    .addEventListener("click", importFileToWorkspace);
  document
    .getElementById("open-workspace-btn")
    .addEventListener("click", openWorkspace);
  document
    .getElementById("save-workspace-btn")
    .addEventListener("click", saveAllFiles);
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
    // Ctrl/Cmd + S to save current file
    if ((e.ctrlKey || e.metaKey) && e.key === "s" && !e.shiftKey) {
      e.preventDefault();
      saveCurrentFile();
    }
    // Ctrl/Cmd + Shift + S to save all files
    if ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key === "S") {
      e.preventDefault();
      saveAllFiles();
    }
    // Ctrl/Cmd + N to create new file
    if ((e.ctrlKey || e.metaKey) && e.key === "n") {
      e.preventDefault();
      createNewFile();
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

  // Load workspace files with retry to ensure backend is initialized
  const loaded = await loadWorkspaceWithRetry();

  if (loaded) {
    // Load initial file content
    if (activeFileName) {
      try {
        const content = await GetFileContent(activeFileName);

        // Check if this is an image file
        if (isImageFile(activeFileName)) {
          showImagePreview(activeFileName, content);
        } else {
          // Set the language based on file extension
          const language = getLanguageFromFilename(activeFileName);
          const model = editor.getModel();
          monaco.editor.setModelLanguage(model, language);

          editor.setValue(content);
          document.getElementById("active-file-label").textContent =
            activeFileName;
        }
      } catch (error) {
        console.error("Failed to load initial file:", error);
      }
    }
  } else {
    console.error("Workspace failed to initialize properly");
    showMessage("Failed to initialize workspace", "error");
  }

  // Mark file as modified on content change
  editor.onDidChangeModelContent(() => {
    if (activeFileName) {
      // Update content in memory
      clearTimeout(window.autoUpdateTimer);
      window.autoUpdateTimer = setTimeout(async () => {
        const currentContent = editor.getValue();
        await UpdateFileContent(activeFileName, currentContent);
        await loadWorkspaceFiles(); // Refresh to show modified indicator
      }, 1000);
    }
  });
}

// Start the app when DOM is ready
if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initApp);
} else {
  initApp();
}
