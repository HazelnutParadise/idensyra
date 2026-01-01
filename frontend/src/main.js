import "./style.css";
import "bootstrap/dist/css/bootstrap.min.css";
import "@fortawesome/fontawesome-free/css/all.min.css";
import * as monaco from "monaco-editor";
import { marked } from "marked";

import {
  ExecuteCode,
  GetVersion,
  GetDefaultCode,
  GetSymbols,
  SaveCode,
  LoadCode,
  SaveResult,
  OpenGitHub,
  OpenOfficialSite,
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
import { EventsOn } from "../wailsjs/runtime/runtime";

const RenameFile = (...args) => window.go.main.App.RenameFile(...args);
const SaveResultToWorkspace = (...args) =>
  window.go.main.App.SaveResultToWorkspace(...args);
const CreateFolder = (...args) => window.go.main.App.CreateFolder(...args);
const DeleteFolder = (...args) => window.go.main.App.DeleteFolder(...args);
const RenameFolder = (...args) => window.go.main.App.RenameFolder(...args);
const ImportFileToWorkspaceAt = (...args) =>
  window.go.main.App.ImportFileToWorkspaceAt(...args);

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
let currentActionMenu = null;
let workspaceFiles = [];
let activeFileName = "";
let isWorkspaceInitialized = false;
let isImagePreview = false; // Track if current file is an image
let importProgressHideTimer = null;
let isLargeFilePreview = false;
let isBinaryPreview = false;
const expandedDirs = new Set();
let selectedFolderPath = "";
let isRootFolderSelected = false;
let lastExecutionOutput =
  '<div style="color: #888;">Run your code to see output here...</div>';
let previewMode = null;
let previewUpdateTimer = null;
const excelSheetSelections = new Map();
let excelPreviewToken = 0;

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

let cachedGoParse = {
  versionId: null,
  structs: new Map(),
  varTypes: new Map(),
};

function parseGoStructs(source) {
  const structs = new Map();
  const structRegex = /type\s+([A-Za-z_]\w*)\s+struct\s*\{([\s\S]*?)\n\}/g;
  let match;

  while ((match = structRegex.exec(source))) {
    const name = match[1];
    const body = match[2];
    const fields = new Set();

    body.split(/\r?\n/).forEach((line) => {
      const cleaned = line.split("//")[0].trim();
      if (!cleaned) return;
      if (cleaned.startsWith("}")) return;
      const fieldMatch = cleaned.match(/^([A-Za-z_]\w*)\b/);
      if (fieldMatch) {
        fields.add(fieldMatch[1]);
      }
    });

    structs.set(name, { fields: Array.from(fields), methods: [] });
  }

  const methodRegex =
    /func\s*\(\s*\w+\s*(?:\*\s*)?([A-Za-z_]\w*)\s*\)\s*([A-Za-z_]\w*)\s*\(/g;
  while ((match = methodRegex.exec(source))) {
    const typeName = match[1];
    const methodName = match[2];
    const info = structs.get(typeName) || { fields: [], methods: [] };
    if (!info.methods.includes(methodName)) {
      info.methods.push(methodName);
    }
    structs.set(typeName, info);
  }

  return structs;
}

function parseGoVarTypes(source, structNames) {
  const varTypes = new Map();
  const lines = source.split(/\r?\n/);

  lines.forEach((line) => {
    const cleaned = line.split("//")[0];
    let match = cleaned.match(
      /\bvar\s+([A-Za-z_]\w*)\s+(?:\*\s*)?([A-Za-z_]\w*)\b/,
    );
    if (match && structNames.has(match[2])) {
      varTypes.set(match[1], match[2]);
    }

    match = cleaned.match(/\b([A-Za-z_]\w*)\s*:=\s*&?\s*([A-Za-z_]\w*)\s*\{/);
    if (match && structNames.has(match[2])) {
      varTypes.set(match[1], match[2]);
    }

    match = cleaned.match(
      /\b([A-Za-z_]\w*)\s*:=\s*new\(\s*([A-Za-z_]\w*)\s*\)/,
    );
    if (match && structNames.has(match[2])) {
      varTypes.set(match[1], match[2]);
    }
  });

  return varTypes;
}

function getGoParse(model) {
  const versionId = model.getVersionId();
  if (cachedGoParse.versionId === versionId) {
    return cachedGoParse;
  }

  const source = model.getValue();
  const structs = parseGoStructs(source);
  const structNames = new Set(structs.keys());
  const varTypes = parseGoVarTypes(source, structNames);

  cachedGoParse = { versionId, structs, varTypes };
  return cachedGoParse;
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

function showBinaryPreview(filename, label) {
  hideImagePreview();
  const editorContainer = document.getElementById("code-editor");
  editorContainer.style.display = "none";

  let binaryContainer = document.getElementById("binary-file-container");
  if (!binaryContainer) {
    binaryContainer = document.createElement("div");
    binaryContainer.id = "binary-file-container";
    binaryContainer.style.cssText = `
      width: 100%;
      height: 100%;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      background: var(--panel-background-color);
      padding: 24px;
      box-sizing: border-box;
      text-align: center;
      color: var(--text-color);
    `;
    editorContainer.parentElement.appendChild(binaryContainer);
  }

  binaryContainer.innerHTML = `
    <div style="font-size: 18px; margin-bottom: 8px;">
      <i class="fas fa-file"></i> ${filename}
    </div>
    <div style="opacity: 0.7; font-size: 13px;">
      ${label} preview only.
    </div>
  `;

  binaryContainer.style.display = "flex";
  isBinaryPreview = true;
}

function hideBinaryPreview() {
  const binaryContainer = document.getElementById("binary-file-container");
  if (binaryContainer) {
    binaryContainer.style.display = "none";
  }
  const editorContainer = document.getElementById("code-editor");
  editorContainer.style.display = "block";
  isBinaryPreview = false;
}

function showLargeFilePreview(filename, size) {
  hideImagePreview();
  hideBinaryPreview();
  const editorContainer = document.getElementById("code-editor");
  editorContainer.style.display = "none";

  let largeContainer = document.getElementById("large-file-container");
  if (!largeContainer) {
    largeContainer = document.createElement("div");
    largeContainer.id = "large-file-container";
    largeContainer.style.cssText = `
      width: 100%;
      height: 100%;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      background: var(--panel-background-color);
      padding: 24px;
      box-sizing: border-box;
      text-align: center;
      color: var(--text-color);
    `;
    editorContainer.parentElement.appendChild(largeContainer);
  }

  largeContainer.innerHTML = `
    <div style="font-size: 18px; margin-bottom: 8px;">
      <i class="fas fa-file-archive"></i> ${filename}
    </div>
    <div style="opacity: 0.7; margin-bottom: 6px;">
      File size: ${formatBytes(size)}
    </div>
    <div style="opacity: 0.7; font-size: 13px;">
      File is too large to preview or edit in the editor.
    </div>
  `;

  largeContainer.style.display = "flex";
  isLargeFilePreview = true;
}

function hideLargeFilePreview() {
  const largeContainer = document.getElementById("large-file-container");
  if (largeContainer) {
    largeContainer.style.display = "none";
  }
  const editorContainer = document.getElementById("code-editor");
  editorContainer.style.display = "block";
  isLargeFilePreview = false;
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
    triggerCharacters: ["."],
    provideCompletionItems: (model, position) => {
      const word = model.getWordUntilPosition(position);
      const range = {
        startLineNumber: position.lineNumber,
        endLineNumber: position.lineNumber,
        startColumn: word.startColumn,
        endColumn: position.column,
      };

      const linePrefix = model
        .getLineContent(position.lineNumber)
        .slice(0, position.column - 1);
      const memberMatch = linePrefix.match(/([A-Za-z_]\w*)\.$/);

      if (memberMatch) {
        const target = memberMatch[1];
        const suggestions = [];
        const { structs, varTypes } = getGoParse(model);
        const typeName = varTypes.get(target) || target;
        const typeInfo = structs.get(typeName);

        if (typeInfo) {
          typeInfo.fields.forEach((field) => {
            suggestions.push({
              label: field,
              kind: monaco.languages.CompletionItemKind.Field,
              detail: `${typeName} field`,
              insertText: field,
              range: range,
            });
          });

          typeInfo.methods.forEach((method) => {
            suggestions.push({
              label: method,
              kind: monaco.languages.CompletionItemKind.Method,
              detail: `${typeName} method`,
              insertText: method,
              range: range,
            });
          });
        }

        const packageSuggestions = goSymbols
          .filter((symbol) => symbol.startsWith(`${target}.`))
          .map((symbol) => {
            const memberName = symbol.slice(target.length + 1);
            return {
              label: memberName,
              kind: monaco.languages.CompletionItemKind.Function,
              detail: `${target} package`,
              documentation: `Member from ${target}`,
              insertText: memberName,
              range: range,
            };
          });

        suggestions.push(...packageSuggestions);
        return { suggestions: suggestions };
      }

      const suggestions = goSymbols.map((symbol) => {
        const parts = symbol.split(".");
        const packageName = parts[0];

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

      const { structs } = getGoParse(model);
      for (const [structName] of structs) {
        suggestions.push({
          label: structName,
          kind: monaco.languages.CompletionItemKind.Struct,
          detail: "struct type",
          insertText: structName,
          range: range,
        });
      }

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
  if (!isRunnableActiveFile()) {
    showMessage("Run is only available for .go files", "warning");
    return;
  }

  isExecuting = true;
  const runButton = document.getElementById("run-btn");
  const resultOutput = document.getElementById("result-output");
  const resultLabel = document.querySelector(".result-label");

  clearPreviewIfNeeded();
  if (resultLabel) {
    resultLabel.textContent = "Output";
  }

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

    setResultOutput(result);
  } catch (error) {
    setResultOutput(`<div class="error-message">Error: ${error}</div>`);
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

function closeActionMenu() {
  if (!currentActionMenu) return;
  currentActionMenu.classList.remove("active");
  const actions = currentActionMenu.closest(".file-actions");
  if (actions) {
    actions.classList.remove("open");
  }
  currentActionMenu = null;
}

function formatBytes(bytes) {
  if (!Number.isFinite(bytes) || bytes <= 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  let value = bytes;
  let unitIndex = 0;
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex += 1;
  }
  return `${value.toFixed(value >= 10 || unitIndex === 0 ? 0 : 1)} ${
    units[unitIndex]
  }`;
}

function getParentPath(path) {
  const parts = path.split("/");
  if (parts.length <= 1) return "";
  return parts.slice(0, -1).join("/");
}

function getTargetFolder() {
  if (isRootFolderSelected) return "";
  if (selectedFolderPath) return selectedFolderPath;
  if (activeFileName) return getParentPath(activeFileName);
  return "";
}

function getFileExtension(filename) {
  const parts = filename.split(".");
  if (parts.length <= 1) return "";
  return parts.pop().toLowerCase();
}

function getMediaKind(filename) {
  const ext = getFileExtension(filename);
  const imageExts = ["jpg", "jpeg", "png", "gif", "bmp", "webp", "svg", "ico"];
  const videoExts = ["mp4", "webm", "mov", "avi", "mkv", "m4v", "mpg", "mpeg"];
  const audioExts = ["mp3", "wav", "flac", "ogg", "aac", "m4a"];
  const documentExts = ["pdf"];

  if (imageExts.includes(ext)) return "image";
  if (videoExts.includes(ext)) return "video";
  if (audioExts.includes(ext)) return "audio";
  if (documentExts.includes(ext)) return "pdf";
  return "";
}

function getMimeType(filename) {
  const ext = getFileExtension(filename);
  const map = {
    jpg: "image/jpeg",
    jpeg: "image/jpeg",
    png: "image/png",
    gif: "image/gif",
    bmp: "image/bmp",
    webp: "image/webp",
    svg: "image/svg+xml",
    ico: "image/x-icon",
    mp4: "video/mp4",
    webm: "video/webm",
    mov: "video/quicktime",
    avi: "video/x-msvideo",
    mkv: "video/x-matroska",
    m4v: "video/x-m4v",
    mpg: "video/mpeg",
    mpeg: "video/mpeg",
    mp3: "audio/mpeg",
    wav: "audio/wav",
    flac: "audio/flac",
    ogg: "audio/ogg",
    aac: "audio/aac",
    m4a: "audio/mp4",
    pdf: "application/pdf",
  };
  return map[ext] || "application/octet-stream";
}

function updateRunButtonState() {
  const runButton = document.getElementById("run-btn");
  if (!runButton) return;

  const runnable =
    activeFileName &&
    activeFileName.endsWith(".go") &&
    !isImagePreview &&
    !isLargeFilePreview &&
    !isBinaryPreview;

  runButton.disabled = !runnable;
  runButton.title = runnable ? "Run" : "Run is only available for .go files";
}

function isRunnableActiveFile() {
  return (
    activeFileName &&
    activeFileName.endsWith(".go") &&
    !isImagePreview &&
    !isLargeFilePreview &&
    !isBinaryPreview
  );
}

function setResultOutput(html) {
  const resultOutput = document.getElementById("result-output");
  if (!resultOutput) return;
  resultOutput.innerHTML = html;
  lastExecutionOutput = html;
}

function escapeHtml(text) {
  const div = document.createElement("div");
  div.textContent = text;
  return div.innerHTML;
}

marked.setOptions({
  gfm: true,
  breaks: true,
});

function renderMarkdown(markdownText) {
  return marked.parse(markdownText || "");
}

function showPreview(content, type) {
  const resultOutput = document.getElementById("result-output");
  const resultContainer = document.querySelector(".result-container");
  const resultLabel = document.querySelector(".result-label");
  if (!resultOutput || !resultLabel || !resultContainer) return;

  previewMode = type;
  resultLabel.textContent = "Preview";
  setResultPreviewState(true);

  if (type === "html") {
    resultOutput.innerHTML = `<div class="preview-frame-wrap"><iframe class="preview-frame" sandbox=""></iframe></div>`;
    const iframe = resultOutput.querySelector("iframe");
    if (iframe) {
      iframe.srcdoc = content;
    }
    return;
  }

  if (type === "markdown") {
    resultOutput.innerHTML = `<div class="markdown-preview">${renderMarkdown(
      content,
    )}</div>`;
  }
}

function setResultPreviewState(isPreview) {
  const resultSection = document.querySelector(".result-section");
  const resultContainer = document.querySelector(".result-container");
  if (!resultSection || !resultContainer) return;

  if (isPreview) {
    resultSection.classList.add("preview-mode");
    resultContainer.classList.add("preview-mode");
  } else {
    resultSection.classList.remove("preview-mode");
    resultContainer.classList.remove("preview-mode");
  }
}

function showMediaPreview(content, filename) {
  const resultOutput = document.getElementById("result-output");
  const resultContainer = document.querySelector(".result-container");
  const resultLabel = document.querySelector(".result-label");
  if (!resultOutput || !resultLabel || !resultContainer) return;

  const mediaKind = getMediaKind(filename);
  if (!mediaKind) return;

  previewMode = "media";
  resultLabel.textContent = "Preview";
  setResultPreviewState(true);

  const mimeType = getMimeType(filename);
  const dataUrl = `data:${mimeType};base64,${content || ""}`;

  if (mediaKind === "image") {
    resultOutput.innerHTML = `
      <div class="media-preview">
        <img src="${dataUrl}" alt="${filename}" />
      </div>
    `;
  } else if (mediaKind === "video") {
    resultOutput.innerHTML = `
      <div class="media-preview">
        <video controls playsinline src="${dataUrl}"></video>
      </div>
    `;
  } else if (mediaKind === "audio") {
    resultOutput.innerHTML = `
      <div class="media-preview">
        <audio controls src="${dataUrl}"></audio>
      </div>
    `;
  } else if (mediaKind === "pdf") {
    resultOutput.innerHTML = `
      <div class="media-preview">
        <iframe class="pdf-preview" src="${dataUrl}"></iframe>
      </div>
    `;
  }
}

function parseDelimited(text, delimiter) {
  const rows = [];
  let row = [];
  let current = "";
  let inQuotes = false;

  for (let i = 0; i < text.length; i++) {
    const char = text[i];
    const next = text[i + 1];

    if (char === '"' && inQuotes && next === '"') {
      current += '"';
      i += 1;
      continue;
    }

    if (char === '"') {
      inQuotes = !inQuotes;
      continue;
    }

    if (!inQuotes && char === delimiter) {
      row.push(current);
      current = "";
      continue;
    }

    if (!inQuotes && (char === "\n" || char === "\r")) {
      if (char === "\r" && next === "\n") {
        i += 1;
      }
      row.push(current);
      rows.push(row);
      row = [];
      current = "";
      continue;
    }

    current += char;
  }

  row.push(current);
  rows.push(row);
  return rows;
}

function renderTable(rows, maxRows = 200, maxCols = 50) {
  const limitedRows = rows.slice(0, maxRows);
  const htmlRows = limitedRows
    .map((row) => {
      const cells = row
        .slice(0, maxCols)
        .map((cell) => `<td>${escapeHtml(cell || "")}</td>`)
        .join("");
      return `<tr>${cells}</tr>`;
    })
    .join("");
  return `<table><tbody>${htmlRows}</tbody></table>`;
}

function showTablePreviewFromText(content, delimiter) {
  const resultOutput = document.getElementById("result-output");
  const resultContainer = document.querySelector(".result-container");
  const resultLabel = document.querySelector(".result-label");
  if (!resultOutput || !resultLabel || !resultContainer) return;

  previewMode = "table";
  resultLabel.textContent = "Preview";
  setResultPreviewState(true);

  const rows = parseDelimited(content || "", delimiter);
  resultOutput.innerHTML = `<div class="table-preview">${renderTable(rows)}</div>`;
}

async function showExcelPreview(filename) {
  const resultOutput = document.getElementById("result-output");
  const resultContainer = document.querySelector(".result-container");
  const resultLabel = document.querySelector(".result-label");
  if (!resultOutput || !resultLabel || !resultContainer) return;

  const previewToken = ++excelPreviewToken;
  previewMode = "table";
  resultLabel.textContent = "Preview";
  setResultPreviewState(true);
  resultOutput.innerHTML = `
    <div class="excel-preview">
      <div class="excel-preview-toolbar">
        <span class="excel-preview-label">Sheet</span>
        <select id="excel-sheet-select" class="excel-sheet-select" disabled></select>
        <span id="excel-preview-meta" class="excel-preview-meta"></span>
      </div>
      <div id="excel-preview-table" class="table-preview">Loading preview...</div>
    </div>
  `;

  const sheetSelect = document.getElementById("excel-sheet-select");
  const tableContainer = document.getElementById("excel-preview-table");
  const metaEl = document.getElementById("excel-preview-meta");

  if (!sheetSelect || !tableContainer) return;

  try {
    const sheets = await window.go.main.App.GetExcelSheets(filename);
    if (previewToken !== excelPreviewToken) return;

    if (!Array.isArray(sheets) || sheets.length === 0) {
      tableContainer.innerHTML =
        '<div class="error-message">No sheets found</div>';
      return;
    }

    sheetSelect.innerHTML = sheets
      .map(
        (sheet) =>
          `<option value="${escapeHtml(sheet)}">${escapeHtml(sheet)}</option>`,
      )
      .join("");

    let selectedSheet = excelSheetSelections.get(filename) || sheets[0];
    if (!sheets.includes(selectedSheet)) {
      selectedSheet = sheets[0];
    }
    sheetSelect.value = selectedSheet;
    sheetSelect.disabled = false;

    const updateMeta = (sheetName) => {
      const index = sheets.indexOf(sheetName);
      if (metaEl) {
        metaEl.textContent = `Sheet ${index + 1} / ${sheets.length}`;
      }
    };

    const loadSheetPreview = async (sheetName) => {
      excelSheetSelections.set(filename, sheetName);
      updateMeta(sheetName);
      tableContainer.innerHTML = "Loading preview...";

      try {
        const html = await window.go.main.App.GetExcelSheetPreview(
          filename,
          sheetName,
          100,
          30,
        );
        if (previewToken !== excelPreviewToken) return;
        tableContainer.innerHTML = html;
      } catch (error) {
        if (previewToken !== excelPreviewToken) return;
        tableContainer.innerHTML = `<div class="error-message">Preview failed: ${escapeHtml(
          String(error),
        )}</div>`;
      }
    };

    sheetSelect.addEventListener("change", (event) => {
      const value = event.target.value;
      loadSheetPreview(value);
    });

    await loadSheetPreview(selectedSheet);
  } catch (error) {
    if (previewToken !== excelPreviewToken) return;
    tableContainer.innerHTML = `<div class="error-message">Preview failed: ${escapeHtml(
      String(error),
    )}</div>`;
  }
}

function clearPreviewIfNeeded() {
  if (!previewMode) return;
  const resultLabel = document.querySelector(".result-label");
  const resultOutput = document.getElementById("result-output");
  const resultContainer = document.querySelector(".result-container");
  if (!resultLabel || !resultOutput || !resultContainer) return;

  previewMode = null;
  resultLabel.textContent = "Output";
  resultOutput.innerHTML = lastExecutionOutput;
  setResultPreviewState(false);
}

function schedulePreviewUpdate(content, type) {
  if (previewUpdateTimer) {
    clearTimeout(previewUpdateTimer);
  }
  previewUpdateTimer = setTimeout(() => {
    if (type === "csv") {
      showTablePreviewFromText(content, ",");
      return;
    }
    if (type === "tsv") {
      showTablePreviewFromText(content, "\t");
      return;
    }
    showPreview(content, type);
  }, 250);
}

function showImportProgress(
  filename,
  title = "Importing file",
  detailText = "",
) {
  const overlay = document.getElementById("import-progress-overlay");
  if (!overlay) return;

  if (importProgressHideTimer) {
    clearTimeout(importProgressHideTimer);
    importProgressHideTimer = null;
  }

  overlay.classList.add("active");
  const titleEl = document.getElementById("import-progress-title");
  if (titleEl) {
    titleEl.textContent = title;
  }
  document.getElementById("import-progress-filename").textContent =
    filename || "Importing file...";
  document.getElementById("import-progress-percent").textContent = "0%";
  document.getElementById("import-progress-bytes").textContent = detailText;
  document.getElementById("import-progress-fill").style.width = "0%";
}

function updateImportProgress(
  filename,
  percent,
  bytesRead,
  totalBytes,
  title = "Importing file",
  detailText = "",
) {
  const overlay = document.getElementById("import-progress-overlay");
  if (!overlay) return;

  if (!overlay.classList.contains("active")) {
    showImportProgress(filename, title, detailText);
  }

  const titleEl = document.getElementById("import-progress-title");
  if (titleEl) {
    titleEl.textContent = title;
  }
  document.getElementById("import-progress-filename").textContent =
    filename || "Importing file...";
  document.getElementById("import-progress-percent").textContent =
    `${percent}%`;

  if (detailText) {
    document.getElementById("import-progress-bytes").textContent = detailText;
  } else if (totalBytes > 0) {
    document.getElementById("import-progress-bytes").textContent =
      `${formatBytes(bytesRead)} / ${formatBytes(totalBytes)}`;
  } else {
    document.getElementById("import-progress-bytes").textContent = "";
  }

  document.getElementById("import-progress-fill").style.width = `${percent}%`;
}

function hideImportProgress(delayMs = 0) {
  const overlay = document.getElementById("import-progress-overlay");
  if (!overlay) return;

  if (importProgressHideTimer) {
    clearTimeout(importProgressHideTimer);
  }

  importProgressHideTimer = setTimeout(() => {
    overlay.classList.remove("active");
  }, delayMs);
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

  closeActionMenu();
  fileTree.innerHTML = "";

  const treeRoot = buildFileTree(workspaceFiles);
  const initialized = expandedDirs.size > 0;

  renderTreeNodes(treeRoot, fileTree, 0, initialized);
}

function buildFileTree(files) {
  const root = {
    name: "",
    path: "",
    isDir: true,
    meta: null,
    children: new Map(),
  };

  files.forEach((file) => {
    const parts = file.name.split("/");
    let node = root;

    parts.forEach((part, index) => {
      const path = parts.slice(0, index + 1).join("/");
      let child = node.children.get(part);
      const isLeaf = index === parts.length - 1;

      if (!child) {
        child = {
          name: part,
          path,
          isDir: !isLeaf || file.isDir,
          meta: null,
          children: new Map(),
        };
        node.children.set(part, child);
      }

      if (isLeaf) {
        child.isDir = file.isDir;
        child.meta = file;
      }

      node = child;
    });
  });

  return root;
}

function renderTreeNodes(node, container, depth, initialized) {
  const entries = Array.from(node.children.values()).sort((a, b) => {
    if (a.isDir !== b.isDir) {
      return a.isDir ? -1 : 1;
    }
    return a.name.localeCompare(b.name);
  });

  entries.forEach((entry) => {
    if (entry.isDir && !initialized) {
      expandedDirs.add(entry.path);
    }

    const fileItem = document.createElement("div");
    fileItem.className = "file-item";
    fileItem.style.paddingLeft = `${8 + depth * 14}px`;
    fileItem.title = entry.name;

    if (entry.meta && entry.meta.tooLarge) {
      fileItem.classList.add("file-large");
    }
    if (!entry.isDir && entry.path === activeFileName) {
      fileItem.classList.add("active");
    }
    if (entry.isDir && entry.path === selectedFolderPath) {
      fileItem.classList.add("selected");
    }
    if (entry.meta && entry.meta.modified) {
      fileItem.classList.add("modified");
    }

    let iconClass = "fa-file-code";
    if (entry.isDir) {
      iconClass = expandedDirs.has(entry.path) ? "fa-folder-open" : "fa-folder";
    } else if (isImageFile(entry.path)) {
      iconClass = "fa-file-image";
    } else if (entry.path.endsWith(".md")) {
      iconClass = "fa-file-lines";
    } else if (entry.path.endsWith(".json")) {
      iconClass = "fa-file-code";
    } else if (entry.path.endsWith(".txt")) {
      iconClass = "fa-file-lines";
    }

    fileItem.innerHTML = `
      <i class="fas ${iconClass}"></i>
      <span class="file-name">${entry.name}</span>
      ${
        entry.meta && entry.meta.tooLarge
          ? '<span class="large-indicator" title="Large file">L</span>'
          : ""
      }
      ${
        entry.meta && entry.meta.modified
          ? '<span class="modified-indicator">*</span>'
          : ""
      }
      <div class="file-actions">
        <button class="file-action-btn" title="Actions" type="button">
          <i class="fas fa-ellipsis-h"></i>
        </button>
        <div class="file-action-menu">
          <button class="file-action-item file-action-rename" type="button">
            Rename
          </button>
          <button class="file-action-item file-action-delete" type="button">
            Delete
          </button>
        </div>
      </div>
    `;

    fileItem.addEventListener("click", (e) => {
      if (e.target.closest(".file-actions")) {
        return;
      }

      if (entry.isDir) {
        selectedFolderPath = entry.path;
        isRootFolderSelected = false;
        if (expandedDirs.has(entry.path)) {
          expandedDirs.delete(entry.path);
        } else {
          expandedDirs.add(entry.path);
        }
        renderFileTree();
        return;
      }

      switchToFile(entry.path);
    });

    const fileActions = fileItem.querySelector(".file-actions");
    const actionBtn = fileItem.querySelector(".file-action-btn");
    const actionMenu = fileItem.querySelector(".file-action-menu");
    const renameBtn = fileItem.querySelector(".file-action-rename");
    const deleteBtn = fileItem.querySelector(".file-action-delete");

    actionBtn.addEventListener("click", (e) => {
      e.stopPropagation();
      const wasOpen = actionMenu.classList.contains("active");
      closeActionMenu();
      if (!wasOpen) {
        actionMenu.classList.add("active");
        fileActions.classList.add("open");
        currentActionMenu = actionMenu;
      }
    });

    renameBtn.addEventListener("click", (e) => {
      e.stopPropagation();
      closeActionMenu();
      if (entry.isDir) {
        renameFolderPrompt(entry.path);
      } else {
        renameFilePrompt(entry.path);
      }
    });

    deleteBtn.addEventListener("click", (e) => {
      e.stopPropagation();
      closeActionMenu();
      if (entry.isDir) {
        deleteFolderConfirm(entry.path);
      } else {
        deleteFileConfirm(entry.path);
      }
    });

    container.appendChild(fileItem);

    if (entry.isDir && expandedDirs.has(entry.path)) {
      renderTreeNodes(entry, container, depth + 1, true);
    }
  });
}

async function switchToFile(filename, force = false) {
  if (!force && filename === activeFileName) return;

  try {
    // Save current file content (only if not in image preview mode)
    if (
      activeFileName &&
      !isImagePreview &&
      !isLargeFilePreview &&
      !isBinaryPreview
    ) {
      const isKnownFile = workspaceFiles.some(
        (file) => file.name === activeFileName,
      );
      if (isKnownFile) {
        const currentContent = editor.getValue();
        await UpdateFileContent(activeFileName, currentContent);
      } else {
        activeFileName = "";
        hideImagePreview();
        hideLargeFilePreview();
        hideBinaryPreview();
      }
    }

    // Switch to new file
    await SetActiveFile(filename);
    const selectedFile = workspaceFiles.find((file) => file.name === filename);
    activeFileName = filename;
    selectedFolderPath = getParentPath(filename);
    isRootFolderSelected = false;
    document.getElementById("active-file-label").textContent = filename;

    if (selectedFile && selectedFile.tooLarge) {
      showLargeFilePreview(filename, selectedFile.size || 0);
      clearPreviewIfNeeded();
      updateRunButtonState();
      await loadWorkspaceFiles();
      return;
    }

    const content = await GetFileContent(filename);

    // Check if this is an image file
    if (isImageFile(filename)) {
      hideBinaryPreview();
      showImagePreview(filename, content);
      clearPreviewIfNeeded();
      showMediaPreview(content, filename);
    } else if (getMediaKind(filename)) {
      const mediaKind = getMediaKind(filename);
      hideImagePreview();
      hideLargeFilePreview();
      hideBinaryPreview();
      showBinaryPreview(filename, mediaKind === "pdf" ? "PDF" : "Media");
      showMediaPreview(content, filename);
    } else if (
      filename.endsWith(".xlsx") ||
      filename.endsWith(".xlsm") ||
      filename.endsWith(".xltx") ||
      filename.endsWith(".xltm")
    ) {
      hideImagePreview();
      hideLargeFilePreview();
      hideBinaryPreview();
      showBinaryPreview(filename, "Spreadsheet");
      await showExcelPreview(filename);
    } else if (selectedFile && selectedFile.isBinary) {
      hideImagePreview();
      hideLargeFilePreview();
      hideBinaryPreview();
      showBinaryPreview(filename, "Binary");
      clearPreviewIfNeeded();
    } else {
      // Hide binary/image preview if it was showing
      hideImagePreview();
      hideLargeFilePreview();
      hideBinaryPreview();

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

      if (filename.endsWith(".html") || filename.endsWith(".htm")) {
        showPreview(content, "html");
      } else if (filename.endsWith(".md")) {
        showPreview(content, "markdown");
      } else if (filename.endsWith(".csv")) {
        showTablePreviewFromText(content, ",");
      } else if (filename.endsWith(".tsv")) {
        showTablePreviewFromText(content, "\t");
      } else {
        clearPreviewIfNeeded();
      }
    }

    updateRunButtonState();
    // Refresh file tree
    await loadWorkspaceFiles();
  } catch (error) {
    console.error("Failed to switch file:", error);
    showMessage("Failed to switch file: " + error, "error");
  }
}

async function createNewFile() {
  const targetFolder = getTargetFolder();
  const filename = prompt(
    "Enter new file name (e.g., test.go, notes.txt, config.json):",
  );
  if (!filename) return;

  const finalName = targetFolder ? `${targetFolder}/${filename}` : filename;

  try {
    await CreateNewFile(finalName);
    const parentPath = getParentPath(finalName);
    if (parentPath) {
      expandedDirs.add(parentPath);
    }
    await loadWorkspaceFiles();
    showMessage(`File "${finalName}" created successfully`, "success");
    await switchToFile(finalName);
  } catch (error) {
    console.error("Failed to create file:", error);
    showMessage(`Failed to create file: ${error}`, "error");
  }
}

async function createNewFolder() {
  const targetFolder = getTargetFolder();
  const folderName = prompt(
    "Enter new folder name (e.g., docs, assets/icons):",
  );
  if (!folderName) return;

  const finalName = targetFolder ? `${targetFolder}/${folderName}` : folderName;

  try {
    await CreateFolder(finalName);
    const parentPath = getParentPath(finalName);
    if (parentPath) {
      expandedDirs.add(parentPath);
    }
    expandedDirs.add(finalName);
    await loadWorkspaceFiles();
    showMessage(`Folder "${finalName}" created successfully`, "success");
  } catch (error) {
    console.error("Failed to create folder:", error);
    showMessage(`Failed to create folder: ${error}`, "error");
  }
}

async function deleteFileConfirm(filename) {
  const fileCount = workspaceFiles.filter((file) => !file.isDir).length;
  if (fileCount <= 1) {
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
      hideBinaryPreview();
      await switchToFile(workspaceFiles[0].name, true);
    }

    showMessage(`File "${filename}" deleted`, "success");
  } catch (error) {
    console.error("Failed to delete file:", error);
    showMessage("Failed to delete file: " + error, "error");
  }
}

async function deleteFolderConfirm(folderPath) {
  if (!confirm(`Delete folder "${folderPath}" and all its contents?`)) return;

  try {
    await DeleteFolder(folderPath);
    Array.from(expandedDirs).forEach((path) => {
      if (path === folderPath || path.startsWith(`${folderPath}/`)) {
        expandedDirs.delete(path);
      }
    });
    if (
      selectedFolderPath === folderPath ||
      selectedFolderPath.startsWith(`${folderPath}/`)
    ) {
      selectedFolderPath = "";
      isRootFolderSelected = false;
    }
    if (activeFileName && activeFileName.startsWith(`${folderPath}/`)) {
      activeFileName = "";
      hideImagePreview();
      hideLargeFilePreview();
      hideBinaryPreview();
    }
    await loadWorkspaceFiles();
    showMessage(`Folder "${folderPath}" deleted`, "success");
  } catch (error) {
    console.error("Failed to delete folder:", error);
    showMessage("Failed to delete folder: " + error, "error");
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
      hideBinaryPreview();
      await switchToFile(trimmedName, true);
    }

    showMessage(`File renamed to "${trimmedName}"`, "success");
  } catch (error) {
    console.error("Failed to rename file:", error);
    showMessage("Failed to rename file: " + error, "error");
  }
}

async function renameFolderPrompt(folderPath) {
  const newName = prompt("Enter new folder name:", folderPath);
  if (!newName) return;

  const trimmedName = newName.trim();
  if (!trimmedName) {
    showMessage("Folder name cannot be empty", "error");
    return;
  }
  if (trimmedName === folderPath) {
    showMessage("Folder name is unchanged", "warning");
    return;
  }

  try {
    await RenameFolder(folderPath, trimmedName);
    if (activeFileName && activeFileName.startsWith(`${folderPath}/`)) {
      activeFileName = activeFileName.replace(
        `${folderPath}/`,
        `${trimmedName}/`,
      );
    }
    if (selectedFolderPath === folderPath) {
      selectedFolderPath = trimmedName;
      isRootFolderSelected = false;
    } else if (selectedFolderPath.startsWith(`${folderPath}/`)) {
      selectedFolderPath = selectedFolderPath.replace(
        `${folderPath}/`,
        `${trimmedName}/`,
      );
      isRootFolderSelected = false;
    }
    if (expandedDirs.has(folderPath)) {
      expandedDirs.delete(folderPath);
      expandedDirs.add(trimmedName);
    }
    await loadWorkspaceFiles();
    showMessage(`Folder renamed to "${trimmedName}"`, "success");
  } catch (error) {
    console.error("Failed to rename folder:", error);
    showMessage("Failed to rename folder: " + error, "error");
  }
}

async function saveCurrentFile() {
  if (!activeFileName) return;

  // Cannot save image files
  if (isImagePreview) {
    showMessage("Image files cannot be edited", "warning");
    return;
  }
  if (isBinaryPreview) {
    showMessage("This file type cannot be edited in the editor", "warning");
    return;
  }
  if (isLargeFilePreview) {
    showMessage("Large files cannot be edited in the editor", "warning");
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
    if (
      activeFileName &&
      !isImagePreview &&
      !isLargeFilePreview &&
      !isBinaryPreview
    ) {
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
    if (
      activeFileName &&
      !isImagePreview &&
      !isLargeFilePreview &&
      !isBinaryPreview
    ) {
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
      activeFileName = activeFile;
      document.getElementById("active-file-label").textContent = activeFile;
      selectedFolderPath = getParentPath(activeFile);
      isRootFolderSelected = false;

      const activeMeta = workspaceFiles.find(
        (file) => file.name === activeFile,
      );
      if (activeMeta && activeMeta.tooLarge) {
        showLargeFilePreview(activeFile, activeMeta.size || 0);
        updateRunButtonState();
        return;
      }

      const content = await GetFileContent(activeFile);
      if (isImageFile(activeFile)) {
        hideBinaryPreview();
        showImagePreview(activeFile, content);
        showMediaPreview(content, activeFile);
      } else if (getMediaKind(activeFile)) {
        const mediaKind = getMediaKind(activeFile);
        hideImagePreview();
        hideLargeFilePreview();
        hideBinaryPreview();
        showBinaryPreview(activeFile, mediaKind === "pdf" ? "PDF" : "Media");
        showMediaPreview(content, activeFile);
      } else if (
        activeFile.endsWith(".xlsx") ||
        activeFile.endsWith(".xlsm") ||
        activeFile.endsWith(".xltx") ||
        activeFile.endsWith(".xltm")
      ) {
        hideImagePreview();
        hideLargeFilePreview();
        hideBinaryPreview();
        showBinaryPreview(activeFile, "Spreadsheet");
        await showExcelPreview(activeFile);
      } else if (activeMeta && activeMeta.isBinary) {
        hideImagePreview();
        hideLargeFilePreview();
        hideBinaryPreview();
        showBinaryPreview(activeFile, "Binary");
        clearPreviewIfNeeded();
      } else {
        hideImagePreview();
        hideLargeFilePreview();
        hideBinaryPreview();
        editor.setValue(content);

        if (activeFile.endsWith(".html") || activeFile.endsWith(".htm")) {
          showPreview(content, "html");
        } else if (activeFile.endsWith(".md")) {
          showPreview(content, "markdown");
        } else if (activeFile.endsWith(".csv")) {
          showTablePreviewFromText(content, ",");
        } else if (activeFile.endsWith(".tsv")) {
          showTablePreviewFromText(content, "\t");
        } else {
          clearPreviewIfNeeded();
        }
      }
      updateRunButtonState();
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
    const existingNames = new Set(workspaceFiles.map((file) => file.name));
    const targetFolder = getTargetFolder();
    if (targetFolder) {
      await ImportFileToWorkspaceAt(targetFolder);
    } else {
      await ImportFileToWorkspace();
    }
    await loadWorkspaceFiles();

    // Get the newly added file (last modified file in the list)
    if (workspaceFiles.length > 0) {
      const newFile = workspaceFiles.find(
        (file) => !existingNames.has(file.name),
      );
      const fallbackFile = workspaceFiles[workspaceFiles.length - 1];
      const importedFile = newFile || fallbackFile;
      const largeNote = importedFile.tooLarge ? " (large file)" : "";
      showMessage(
        `File "${importedFile.name}" imported successfully${largeNote}`,
        "success",
      );
      // Switch to the newly imported file
      await switchToFile(importedFile.name);
    }
  } catch (error) {
    console.error("Failed to import file:", error);
    showMessage("Failed to import file: " + error, "error");
  } finally {
    hideImportProgress(200);
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
                <button class="secondary" id="hazelnut-btn" title="Official Website">
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
                        <button class="secondary icon-only" id="new-folder-btn" title="New Folder">
                            <i class="fas fa-folder-plus"></i>
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
        <div id="import-progress-overlay" class="import-progress-overlay">
            <div class="import-progress-card">
                <div id="import-progress-title" class="import-progress-title">Importing file</div>
                <div id="import-progress-filename" class="import-progress-filename"></div>
                <div class="import-progress-bar">
                    <div id="import-progress-fill" class="import-progress-fill"></div>
                </div>
                <div class="import-progress-details">
                    <span id="import-progress-percent" class="import-progress-percent">0%</span>
                    <span id="import-progress-bytes" class="import-progress-bytes"></span>
                </div>
            </div>
        </div>
    `;

  document.addEventListener("click", (event) => {
    if (!event.target.closest(".file-actions")) {
      closeActionMenu();
    }
  });

  const resultOutput = document.getElementById("result-output");
  const fileTree = document.getElementById("file-tree");
  if (fileTree) {
    fileTree.addEventListener("click", (event) => {
      if (event.target.closest(".file-item")) {
        return;
      }
      selectedFolderPath = "";
      isRootFolderSelected = true;
      renderFileTree();
    });
  }
  if (resultOutput) {
    lastExecutionOutput = resultOutput.innerHTML;
  }

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
    .getElementById("new-folder-btn")
    .addEventListener("click", createNewFolder);
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
    .addEventListener("click", () => OpenOfficialSite());

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

  updateRunButtonState();

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
        const activeMeta = workspaceFiles.find(
          (file) => file.name === activeFileName,
        );

        if (activeMeta && activeMeta.tooLarge) {
          showLargeFilePreview(activeFileName, activeMeta.size || 0);
          document.getElementById("active-file-label").textContent =
            activeFileName;
          selectedFolderPath = getParentPath(activeFileName);
          isRootFolderSelected = false;
          updateRunButtonState();
          return;
        }

        const content = await GetFileContent(activeFileName);

        // Check if this is an image file
        if (isImageFile(activeFileName)) {
          hideBinaryPreview();
          showImagePreview(activeFileName, content);
          showMediaPreview(content, activeFileName);
        } else if (getMediaKind(activeFileName)) {
          const mediaKind = getMediaKind(activeFileName);
          hideImagePreview();
          hideLargeFilePreview();
          hideBinaryPreview();
          showBinaryPreview(
            activeFileName,
            mediaKind === "pdf" ? "PDF" : "Media",
          );
          showMediaPreview(content, activeFileName);
        } else if (
          activeFileName.endsWith(".xlsx") ||
          activeFileName.endsWith(".xlsm") ||
          activeFileName.endsWith(".xltx") ||
          activeFileName.endsWith(".xltm")
        ) {
          hideImagePreview();
          hideLargeFilePreview();
          hideBinaryPreview();
          showBinaryPreview(activeFileName, "Spreadsheet");
          await showExcelPreview(activeFileName);
        } else if (activeMeta && activeMeta.isBinary) {
          hideImagePreview();
          hideLargeFilePreview();
          hideBinaryPreview();
          showBinaryPreview(activeFileName, "Binary");
          clearPreviewIfNeeded();
        } else {
          // Set the language based on file extension
          const language = getLanguageFromFilename(activeFileName);
          const model = editor.getModel();
          monaco.editor.setModelLanguage(model, language);

          hideImagePreview();
          hideLargeFilePreview();
          hideBinaryPreview();
          editor.setValue(content);
          document.getElementById("active-file-label").textContent =
            activeFileName;
          selectedFolderPath = getParentPath(activeFileName);
          isRootFolderSelected = false;

          if (
            activeFileName.endsWith(".html") ||
            activeFileName.endsWith(".htm")
          ) {
            showPreview(content, "html");
          } else if (activeFileName.endsWith(".md")) {
            showPreview(content, "markdown");
          } else if (activeFileName.endsWith(".csv")) {
            showTablePreviewFromText(content, ",");
          } else if (activeFileName.endsWith(".tsv")) {
            showTablePreviewFromText(content, "\t");
          } else {
            clearPreviewIfNeeded();
          }
        }
        updateRunButtonState();
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
      if (!isImagePreview && !isLargeFilePreview && !isBinaryPreview) {
        if (
          activeFileName.endsWith(".html") ||
          activeFileName.endsWith(".htm")
        ) {
          schedulePreviewUpdate(editor.getValue(), "html");
        } else if (activeFileName.endsWith(".md")) {
          schedulePreviewUpdate(editor.getValue(), "markdown");
        } else if (activeFileName.endsWith(".csv")) {
          schedulePreviewUpdate(editor.getValue(), "csv");
        } else if (activeFileName.endsWith(".tsv")) {
          schedulePreviewUpdate(editor.getValue(), "tsv");
        }
        // Update content in memory
        clearTimeout(window.autoUpdateTimer);
        window.autoUpdateTimer = setTimeout(async () => {
          const currentContent = editor.getValue();
          await UpdateFileContent(activeFileName, currentContent);
          await loadWorkspaceFiles(); // Refresh to show modified indicator
        }, 1000);
      }
    }
  });

  EventsOn("import:file-progress", (payload) => {
    const data = Array.isArray(payload) ? payload[0] : payload;
    if (!data) return;

    const title = "Importing file";
    const fileName = data.fileName || "Importing file...";
    const bytesRead = Number(data.bytesRead || 0);
    const totalBytes = Number(data.totalBytes || 0);
    const percent =
      totalBytes > 0
        ? Math.min(100, Math.round((bytesRead / totalBytes) * 100))
        : 100;

    if (data.phase === "start") {
      showImportProgress(fileName, title);
      updateImportProgress(fileName, 0, bytesRead, totalBytes, title);
      return;
    }

    if (data.phase === "progress") {
      updateImportProgress(fileName, percent, bytesRead, totalBytes, title);
      return;
    }

    if (data.phase === "done") {
      updateImportProgress(fileName, 100, totalBytes, totalBytes, title);
      hideImportProgress(400);
      return;
    }

    if (data.phase === "error") {
      if (data.message) {
        showMessage(data.message, "error");
      }
      hideImportProgress(0);
    }
  });

  EventsOn("workspace:open-progress", (payload) => {
    const data = Array.isArray(payload) ? payload[0] : payload;
    if (!data) return;

    const title = "Opening workspace";
    const fileName = data.fileName || "Opening workspace...";
    const bytesRead = Number(data.bytesRead || 0);
    const totalBytes = Number(data.totalBytes || 0);
    const processedFiles = Number(data.processedFiles || 0);
    const totalFiles = Number(data.totalFiles || 0);
    const progressTotal = totalBytes > 0 ? totalBytes : totalFiles;
    const progressValue = totalBytes > 0 ? bytesRead : processedFiles;
    const percent =
      progressTotal > 0
        ? Math.min(100, Math.round((progressValue / progressTotal) * 100))
        : 100;
    const detailText =
      totalBytes > 0
        ? `${formatBytes(bytesRead)} / ${formatBytes(totalBytes)}`
        : totalFiles > 0
          ? `${processedFiles} / ${totalFiles} files`
          : "";

    if (data.phase === "start") {
      showImportProgress(fileName, title, detailText);
      updateImportProgress(
        fileName,
        0,
        bytesRead,
        totalBytes,
        title,
        detailText,
      );
      return;
    }

    if (data.phase === "progress") {
      updateImportProgress(
        fileName,
        percent,
        bytesRead,
        totalBytes,
        title,
        detailText,
      );
      return;
    }

    if (data.phase === "done") {
      updateImportProgress(
        fileName,
        100,
        bytesRead,
        totalBytes,
        title,
        detailText,
      );
      hideImportProgress(400);
      return;
    }

    if (data.phase === "error") {
      if (data.message) {
        showMessage(data.message, "error");
      }
      hideImportProgress(0);
    }
  });
}

// Start the app when DOM is ready
if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initApp);
} else {
  initApp();
}
