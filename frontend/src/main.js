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
const ExecutePythonFile = (...args) =>
  window.go.main.App.ExecutePythonFile(...args);
const ExecuteIgonbCells = (...args) =>
  window.go.main.App.ExecuteIgonbCells(...args);

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
let isIgonbView = false;
let igonbState = null;
let igonbSaveTimer = null;
let igonbRenderToken = 0;
let igonbIdCounter = 0;
const igonbEditors = new Map();
let igonbOutputMode = "full";
let igonbSelectedId = null;
let igonbDragId = null;
const expandedDirs = new Set();
let selectedFolderPath = "";
let isRootFolderSelected = false;
let workspaceLoadToken = 0;
let fileTreeRenderToken = 0;
const fileTreeChunkSize = 120;
const fileModelCache = new Map();
const fileModelSizes = new Map();
let fileModelBytes = 0;
const maxOpenModels = 12;
const maxOpenBytes = 12 * 1024 * 1024;
let workspaceRefreshTimer = null;
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
    igonb: "json",
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
  aliases: new Map(),
  modelUri: null,
};

function parseGoStructs(source) {
  const structs = new Map();
  const aliases = new Map();
  const structRegex = /type\s+([A-Za-z_]\w*)\s+struct\s*\{([\s\S]*?)\n\}/g;
  let match;

  while ((match = structRegex.exec(source))) {
    const name = match[1];
    const body = match[2];
    const fields = [];
    const fieldTypes = new Map();

    body.split(/\r?\n/).forEach((line) => {
      const cleaned = line.split("//")[0].trim();
      if (!cleaned) return;
      if (cleaned.startsWith("}")) return;
      const noTag = cleaned.split("`")[0].trim();
      if (!noTag) return;

      let fieldMatch = noTag.match(
        /^([A-Za-z_]\w*(?:\s*,\s*[A-Za-z_]\w*)*)\s+(.+)$/,
      );
      if (fieldMatch) {
        const names = fieldMatch[1]
          .split(",")
          .map((namePart) => namePart.trim())
          .filter(Boolean);
        const fieldType = fieldMatch[2].trim();

        names.forEach((fieldName) => {
          fields.push(fieldName);
          fieldTypes.set(fieldName, fieldType);
        });
        return;
      }

      fieldMatch = noTag.match(/^(\*?\s*[A-Za-z_]\w*(?:\.[A-Za-z_]\w*)*)$/);
      if (fieldMatch) {
        const fieldType = fieldMatch[1].replace(/\s+/g, "");
        const fieldName = fieldType.replace(/^\*/, "").split(".").pop();
        fields.push(fieldName);
        fieldTypes.set(fieldName, fieldType);
      }
    });

    structs.set(name, { fields, fieldTypes, methods: [] });
  }

  const methodRegex =
    /func\s*\(\s*\w+\s*(?:\*\s*)?([A-Za-z_]\w*)\s*\)\s*([A-Za-z_]\w*)\s*\(/g;
  while ((match = methodRegex.exec(source))) {
    const typeName = match[1];
    const methodName = match[2];
    const info = structs.get(typeName) || {
      fields: [],
      fieldTypes: new Map(),
      methods: [],
    };
    if (!info.methods.includes(methodName)) {
      info.methods.push(methodName);
    }
    structs.set(typeName, info);
  }

  source.split(/\r?\n/).forEach((line) => {
    const cleaned = line.split("//")[0].trim();
    if (!cleaned) return;
    if (cleaned.includes(" struct")) return;
    if (cleaned.includes(" interface")) return;

    let aliasMatch = cleaned.match(
      /^type\s+([A-Za-z_]\w*)\s*=\s*([A-Za-z_]\w*)\b/,
    );
    if (!aliasMatch) {
      aliasMatch = cleaned.match(/^type\s+([A-Za-z_]\w*)\s+([A-Za-z_]\w*)\b/);
    }

    if (aliasMatch) {
      const aliasName = aliasMatch[1];
      const targetName = aliasMatch[2];
      if (structs.has(targetName)) {
        const targetInfo = structs.get(targetName);
        structs.set(aliasName, {
          fields: [...targetInfo.fields],
          fieldTypes: new Map(targetInfo.fieldTypes),
          methods: [...targetInfo.methods],
        });
        aliases.set(aliasName, targetName);
      }
    }
  });

  return { structs, aliases };
}

function normalizeTypeName(typeName) {
  let name = (typeName || "").trim();
  if (!name) return "";
  name = name.replace(/\(.*\)$/, "").trim();
  name = name.replace(/^\*+/, "");

  const mapMatch = name.match(/^map\s*\[[^\]]*\]\s*(.+)$/);
  if (mapMatch) {
    name = mapMatch[1].trim();
  }

  const chanMatch = name.match(/^(?:<-)?\s*chan\s+(.+)$/);
  if (chanMatch) {
    name = chanMatch[1].trim();
  }

  while (name.startsWith("[")) {
    const bracketMatch = name.match(/^\[\s*\d*\s*\]\s*(.+)$/);
    if (!bracketMatch) break;
    name = bracketMatch[1].trim();
  }

  name = name.replace(/^\*+/, "");

  if (name.includes(".")) {
    name = name.split(".").pop();
  }

  return name.trim();
}

function resolveStructType(typeName, structs, aliases) {
  let resolved = normalizeTypeName(typeName);
  let guard = 0;
  while (aliases.has(resolved) && guard < 10) {
    resolved = aliases.get(resolved);
    guard += 1;
  }
  if (structs.has(resolved)) {
    return resolved;
  }
  return resolved;
}

function parseGoVarTypes(source, structs, aliases) {
  const varTypes = new Map();
  const structNames = new Set(structs.keys());
  const lines = source.split(/\r?\n/);

  lines.forEach((line) => {
    const cleaned = line.split("//")[0];
    let match = cleaned.match(
      /\bvar\s+([A-Za-z_]\w*)\s+(?:\*\s*)?([A-Za-z_]\w*)\b/,
    );
    if (match && structNames.has(match[2])) {
      varTypes.set(match[1], resolveStructType(match[2], structs, aliases));
    }

    match = cleaned.match(
      /\bvar\s+([A-Za-z_]\w*)\s*=\s*&?\s*([A-Za-z_]\w*)\s*\{/,
    );
    if (match && structNames.has(match[2])) {
      varTypes.set(match[1], resolveStructType(match[2], structs, aliases));
    }

    match = cleaned.match(/\b([A-Za-z_]\w*)\s*:=\s*&?\s*([A-Za-z_]\w*)\s*\{/);
    if (match && structNames.has(match[2])) {
      varTypes.set(match[1], resolveStructType(match[2], structs, aliases));
    }

    match = cleaned.match(
      /\b([A-Za-z_]\w*)\s*:=\s*new\(\s*([A-Za-z_]\w*)\s*\)/,
    );
    if (match && structNames.has(match[2])) {
      varTypes.set(match[1], resolveStructType(match[2], structs, aliases));
    }

    match = cleaned.match(/\b([A-Za-z_]\w*)\s*:=\s*&?\s*([A-Za-z_]\w*)\s*\(/);
    if (match && structNames.has(match[2])) {
      varTypes.set(match[1], resolveStructType(match[2], structs, aliases));
    }

    match = cleaned.match(
      /\bvar\s+([A-Za-z_]\w*)\s*=\s*new\(\s*([A-Za-z_]\w*)\s*\)/,
    );
    if (match && structNames.has(match[2])) {
      varTypes.set(match[1], resolveStructType(match[2], structs, aliases));
    }
  });

  return varTypes;
}

function getGoParse(model) {
  const versionId = model.getVersionId();
  const modelUri = model.uri ? model.uri.toString() : "";
  if (
    cachedGoParse.versionId === versionId &&
    cachedGoParse.modelUri === modelUri
  ) {
    return cachedGoParse;
  }

  const source = model.getValue();
  const { structs, aliases } = parseGoStructs(source);
  const varTypes = parseGoVarTypes(source, structs, aliases);

  cachedGoParse = { versionId, structs, varTypes, aliases, modelUri };
  return cachedGoParse;
}

function resolveExpressionType(expression, structs, varTypes, aliases) {
  const parts = expression.split(".").filter(Boolean);
  if (parts.length === 0) return "";

  const base = parts[0];
  let currentType = null;

  if (varTypes.has(base)) {
    currentType = varTypes.get(base);
  } else if (structs.has(base)) {
    currentType = base;
  } else {
    return "";
  }

  let resolved = resolveStructType(currentType, structs, aliases);

  for (let i = 1; i < parts.length; i += 1) {
    const fieldName = parts[i];
    const typeInfo = structs.get(resolved);
    if (!typeInfo || !typeInfo.fieldTypes.has(fieldName)) {
      return "";
    }
    const fieldType = typeInfo.fieldTypes.get(fieldName);
    resolved = resolveStructType(fieldType, structs, aliases);
  }

  return resolved;
}

function touchFileModel(filename) {
  const model = fileModelCache.get(filename);
  if (!model) return;
  fileModelCache.delete(filename);
  fileModelCache.set(filename, model);
}

function updateFileModelSize(filename, model) {
  if (!model) return;
  const newSize = model.getValueLength();
  const oldSize = fileModelSizes.get(filename) || 0;
  fileModelSizes.set(filename, newSize);
  fileModelBytes += newSize - oldSize;
}

function evictFileModels() {
  let guard = 0;
  while (
    (fileModelCache.size > maxOpenModels || fileModelBytes > maxOpenBytes) &&
    guard < 1000
  ) {
    const oldestKey = fileModelCache.keys().next().value;
    if (!oldestKey) break;
    if (oldestKey === activeFileName && fileModelCache.size > 1) {
      touchFileModel(oldestKey);
      guard += 1;
      continue;
    }

    const model = fileModelCache.get(oldestKey);
    fileModelCache.delete(oldestKey);
    const size = fileModelSizes.get(oldestKey) || 0;
    fileModelSizes.delete(oldestKey);
    fileModelBytes -= size;
    if (model && !model.isDisposed()) {
      model.dispose();
    }
    guard += 1;
  }
}

function getCachedFileModel(filename) {
  const model = fileModelCache.get(filename);
  if (model && !model.isDisposed()) {
    touchFileModel(filename);
    return model;
  }
  if (model && model.isDisposed()) {
    fileModelCache.delete(filename);
    fileModelSizes.delete(filename);
  }
  return null;
}

function getOrCreateFileModel(filename, content) {
  const cached = getCachedFileModel(filename);
  if (cached) {
    return cached;
  }

  const uri = monaco.Uri.parse(
    `inmemory://model/${encodeURIComponent(filename)}`,
  );
  let model = monaco.editor.getModel(uri);
  if (model && !model.isDisposed()) {
    if (!model.getValue()) {
      model.setValue(content);
    }
  } else {
    const language = getLanguageFromFilename(filename);
    model = monaco.editor.createModel(content, language, uri);
  }

  fileModelCache.set(filename, model);
  updateFileModelSize(filename, model);
  touchFileModel(filename);
  evictFileModels();
  return model;
}

function updateActiveFileModelSize() {
  if (!activeFileName) return;
  const model = fileModelCache.get(activeFileName);
  if (!model || model.isDisposed()) return;
  updateFileModelSize(activeFileName, model);
  touchFileModel(activeFileName);
  evictFileModels();
}

function removeFileModel(filename) {
  const model = fileModelCache.get(filename);
  if (model) {
    fileModelCache.delete(filename);
    const size = fileModelSizes.get(filename) || 0;
    fileModelSizes.delete(filename);
    fileModelBytes -= size;
    if (editor && editor.getModel() === model) {
      return;
    }
    if (!model.isDisposed()) {
      model.dispose();
    }
  }
}

function removeFileModelsInFolder(folderPath) {
  Array.from(fileModelCache.keys()).forEach((key) => {
    if (key === folderPath || key.startsWith(`${folderPath}/`)) {
      removeFileModel(key);
    }
  });
}

function disposeOrphanModel(filename, model) {
  if (!model || model.isDisposed()) return;
  if (!filename) return;
  if (fileModelCache.has(filename)) return;
  if (editor && editor.getModel() === model) return;
  model.dispose();
}

function renameFileModel(oldName, newName) {
  const model = fileModelCache.get(oldName);
  if (!model) return;
  const size = fileModelSizes.get(oldName) || 0;
  fileModelCache.delete(oldName);
  fileModelSizes.delete(oldName);

  fileModelCache.set(newName, model);
  fileModelSizes.set(newName, size);

  const newLanguage = getLanguageFromFilename(newName);
  if (model.getLanguageId() !== newLanguage) {
    monaco.editor.setModelLanguage(model, newLanguage);
  }
  touchFileModel(newName);
}

function renameFileModelsInFolder(oldPath, newPath) {
  Array.from(fileModelCache.keys()).forEach((key) => {
    if (key === oldPath || key.startsWith(`${oldPath}/`)) {
      const newKey = key.replace(`${oldPath}/`, `${newPath}/`);
      if (newKey === key) return;
      renameFileModel(key, newKey);
    }
  });
}

function applyTextFileContent(filename, content, fromCache = false) {
  if (filename.endsWith(".igonb")) {
    showIgonbNotebook(content, filename);
    return;
  }

  hideIgonbNotebook();

  const model = fromCache
    ? getCachedFileModel(filename)
    : getOrCreateFileModel(filename, content);
  if (!model) return;

  editor.setModel(model);

  const language = getLanguageFromFilename(filename);
  if (model.getLanguageId() !== language) {
    monaco.editor.setModelLanguage(model, language);
  }

  if (!fromCache && content !== model.getValue()) {
    model.setValue(content);
  }

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

function ensureIgonbContainer() {
  let container = document.getElementById("igonb-container");
  if (container) {
    return container;
  }

  const editorContainer = document.getElementById("code-editor");
  container = document.createElement("div");
  container.id = "igonb-container";
  container.className = "igonb-container";
  editorContainer.parentElement.appendChild(container);

  container.innerHTML = `
    <div class="igonb-toolbar">
      <div class="igonb-toolbar-left">
        <button class="secondary" id="igonb-add-go"><i class="fas fa-plus"></i> Go</button>
        <button class="secondary" id="igonb-add-py"><i class="fas fa-plus"></i> Python</button>
        <button class="secondary" id="igonb-add-md"><i class="fas fa-plus"></i> Markdown</button>
      </div>
      <div class="igonb-toolbar-right">
        <button class="secondary" id="igonb-toggle-output" title="Toggle output height">
          <i class="fas fa-align-left"></i> Scroll Output
        </button>
        <button class="success" id="igonb-run-all"><i class="fas fa-play"></i> Run All</button>
      </div>
    </div>
    <div id="igonb-cells" class="igonb-cells"></div>
  `;

  container
    .querySelector("#igonb-add-go")
    .addEventListener("click", () => addIgonbCell("go"));
  container
    .querySelector("#igonb-add-py")
    .addEventListener("click", () => addIgonbCell("python"));
  container
    .querySelector("#igonb-add-md")
    .addEventListener("click", () => addIgonbCell("markdown"));
  container
    .querySelector("#igonb-run-all")
    .addEventListener("click", () => runIgonbAll());
  container
    .querySelector("#igonb-toggle-output")
    .addEventListener("click", () => toggleIgonbOutputMode());

  return container;
}

function showIgonbNotebook(content, filename) {
  const parsed = parseIgonbContent(content);
  if (!parsed) {
    showMessage("Invalid igonb content, showing raw JSON", "error");
    hideIgonbNotebook();
    const model = getOrCreateFileModel(filename, content);
    editor.setModel(model);
    monaco.editor.setModelLanguage(model, getLanguageFromFilename(filename));
    return;
  }

  igonbState = parsed;
  ensureIgonbSelection();
  disposeIgonbEditors();
  ensureIgonbContainer();

  clearPreviewIfNeeded();
  hideImagePreview();
  hideLargeFilePreview();
  hideBinaryPreview();

  const editorContainer = document.getElementById("code-editor");
  editorContainer.style.display = "none";

  const container = document.getElementById("igonb-container");
  container.style.display = "flex";
  isIgonbView = true;
  document.body.classList.add("igonb-mode");
  applyIgonbOutputMode();
  applyIgonbFontSizes();

  renderIgonbCells();
  setResultOutput(
    '<div style="color: #888;">Notebook output is shown inline.</div>',
  );
}

function hideIgonbNotebook() {
  const container = document.getElementById("igonb-container");
  if (container) {
    container.style.display = "none";
  }
  const editorContainer = document.getElementById("code-editor");
  editorContainer.style.display = "block";
  isIgonbView = false;
  igonbState = null;
  igonbSelectedId = null;
  igonbDragId = null;
  document.body.classList.remove("igonb-mode", "igonb-dragging");
  if (igonbSaveTimer) {
    clearTimeout(igonbSaveTimer);
    igonbSaveTimer = null;
  }
  disposeIgonbEditors();
}

function toggleIgonbOutputMode() {
  igonbOutputMode = igonbOutputMode === "scroll" ? "full" : "scroll";
  applyIgonbOutputMode();
}

function applyIgonbOutputMode() {
  const container = document.getElementById("igonb-container");
  if (!container) return;
  const toggleButton = container.querySelector("#igonb-toggle-output");
  if (igonbOutputMode === "scroll") {
    container.classList.add("igonb-output-scroll");
    if (toggleButton) {
      toggleButton.innerHTML = '<i class="fas fa-align-left"></i> Full Output';
      toggleButton.title = "Show full output";
    }
  } else {
    container.classList.remove("igonb-output-scroll");
    if (toggleButton) {
      toggleButton.innerHTML =
        '<i class="fas fa-align-left"></i> Scroll Output';
      toggleButton.title = "Limit output height with scrolling";
    }
  }
}

function parseIgonbContent(content) {
  try {
    const data = JSON.parse(content || "{}");
    const rawCells = Array.isArray(data.cells) ? data.cells : [];
    const cells =
      rawCells.length > 0 ? rawCells : [{ language: "go", source: "" }];
    return {
      version: data.version || 1,
      cells: cells.map((cell) => ({
        id: cell.id || nextIgonbId(),
        language: normalizeIgonbLanguage(cell.language),
        source: cell.source || "",
        output: "",
        error: "",
        running: false,
        waiting: false,
        editing: false,
      })),
    };
  } catch (err) {
    console.error("Failed to parse igonb:", err);
    return null;
  }
}

function nextIgonbId() {
  igonbIdCounter += 1;
  return `igonb-${igonbIdCounter}`;
}

function getIgonbIndexById(id) {
  if (!igonbState || !id) return -1;
  return igonbState.cells.findIndex((cell) => cell.id === id);
}

function updateIgonbCellSelection(id, selected) {
  const container = document.querySelector(`.igonb-cell[data-cell-id="${id}"]`);
  if (!container) return;
  container.classList.toggle("selected", selected);
}

function setIgonbSelectedId(id) {
  if (igonbSelectedId === id) {
    return;
  }
  const prevId = igonbSelectedId;
  igonbSelectedId = id;
  if (prevId) {
    updateIgonbCellSelection(prevId, false);
  }
  if (igonbSelectedId) {
    updateIgonbCellSelection(igonbSelectedId, true);
  }
}

function ensureIgonbSelection() {
  if (!igonbState || igonbState.cells.length === 0) {
    igonbSelectedId = null;
    return;
  }
  const selectedIndex = getIgonbIndexById(igonbSelectedId);
  if (selectedIndex === -1) {
    igonbSelectedId = null;
  }
}

function getIgonbSelectedIndex() {
  const index = getIgonbIndexById(igonbSelectedId);
  return index === -1 ? null : index;
}

function getIgonbSelectedLanguage() {
  if (!igonbState) return "go";
  const index = getIgonbSelectedIndex();
  if (index === null) return "go";
  return igonbState.cells[index].language || "go";
}

function normalizeIgonbLanguage(language) {
  const normalized = String(language || "")
    .toLowerCase()
    .trim();
  if (normalized === "py") return "python";
  if (normalized === "md") return "markdown";
  if (
    normalized === "go" ||
    normalized === "python" ||
    normalized === "markdown"
  ) {
    return normalized;
  }
  return "go";
}

function renderIgonbCells() {
  const renderToken = ++igonbRenderToken;
  const cellList = document.getElementById("igonb-cells");
  if (!cellList || !igonbState) return;

  ensureIgonbSelection();
  disposeIgonbEditors();
  cellList.innerHTML = "";
  igonbState.cells.forEach((cell, index) => {
    if (renderToken !== igonbRenderToken) return;
    cellList.appendChild(createIgonbCellElement(cell, index));
  });

  requestAnimationFrame(() => {
    if (renderToken !== igonbRenderToken) return;
    initIgonbEditors();
  });
}

function createIgonbCellElement(cell, index) {
  const container = document.createElement("div");
  container.className = "igonb-cell";
  container.dataset.index = index;
  container.dataset.cellId = cell.id;
  container.dataset.language = cell.language;
  if (cell.language === "markdown") {
    container.classList.add("markdown");
  }
  if (cell.editing) {
    container.classList.add("editing");
  }
  if (cell.running) {
    container.classList.add("running");
  }
  if (cell.waiting) {
    container.classList.add("waiting");
  }
  if (cell.id === igonbSelectedId) {
    container.classList.add("selected");
  }

  const toolbar = document.createElement("div");
  toolbar.className = "igonb-cell-toolbar";

  const dragHandle = document.createElement("button");
  dragHandle.type = "button";
  dragHandle.className = "secondary igonb-drag-handle";
  dragHandle.title = "Drag to reorder";
  dragHandle.innerHTML = '<i class="fas fa-grip-vertical"></i>';
  dragHandle.draggable = true;
  dragHandle.addEventListener("dragstart", (event) => {
    startIgonbDrag(event, cell.id);
  });
  dragHandle.addEventListener("dragend", () => {
    clearIgonbDrag();
  });

  const title = document.createElement("div");
  title.className = "igonb-cell-title";
  title.textContent = `Cell ${index + 1} - ${cell.language.toUpperCase()}`;

  const status = document.createElement("div");
  status.className = "igonb-cell-status";
  status.textContent = cell.running
    ? "Running..."
    : cell.waiting
      ? "Waiting..."
      : "";

  const actionGroup = document.createElement("div");
  actionGroup.className = "igonb-cell-actions";

  const languageSelect = document.createElement("select");
  languageSelect.className = "igonb-cell-language";
  ["go", "python", "markdown"].forEach((lang) => {
    const option = document.createElement("option");
    option.value = lang;
    option.textContent = lang.toUpperCase();
    if (lang === cell.language) {
      option.selected = true;
    }
    languageSelect.appendChild(option);
  });

  languageSelect.addEventListener("change", () => {
    cell.language = languageSelect.value;
    cell.output = "";
    cell.error = "";
    cell.running = false;
    cell.waiting = false;
    cell.editing = false;
    scheduleIgonbSave();
    renderIgonbCells();
  });

  actionGroup.appendChild(languageSelect);

  if (cell.language === "markdown") {
    const toggleBtn = document.createElement("button");
    toggleBtn.className = "secondary igonb-md-toggle";
    toggleBtn.innerHTML = cell.editing
      ? '<i class="fas fa-eye"></i> Preview'
      : '<i class="fas fa-pen"></i> Edit';
    toggleBtn.addEventListener("click", () => {
      cell.editing = !cell.editing;
      renderIgonbCells();
    });
    actionGroup.appendChild(toggleBtn);
  } else {
    const runBtn = document.createElement("button");
    runBtn.className = "secondary igonb-cell-run";
    runBtn.disabled = cell.running;
    runBtn.innerHTML = '<i class="fas fa-play"></i> Run';
    runBtn.addEventListener("click", () => runIgonbCell(index));
    actionGroup.appendChild(runBtn);
  }

  const deleteBtn = document.createElement("button");
  deleteBtn.className = "secondary";
  deleteBtn.innerHTML = '<i class="fas fa-trash"></i>';
  deleteBtn.addEventListener("click", () => deleteIgonbCell(index));
  actionGroup.appendChild(deleteBtn);

  toolbar.appendChild(dragHandle);
  toolbar.appendChild(title);
  toolbar.appendChild(status);
  toolbar.appendChild(actionGroup);

  container.addEventListener("click", () => {
    setIgonbSelectedId(cell.id);
  });
  container.addEventListener("dragover", (event) => {
    handleIgonbDragOver(event, cell.id);
  });
  container.addEventListener("dragleave", () => {
    clearIgonbDragOver(cell.id);
  });
  container.addEventListener("drop", (event) => {
    handleIgonbDrop(event, cell.id);
  });

  const editorWrap = document.createElement("div");
  editorWrap.className = "igonb-cell-editor";

  container.appendChild(toolbar);
  container.appendChild(editorWrap);

  if (cell.language === "markdown") {
    const preview = document.createElement("div");
    preview.className = "igonb-markdown-preview";
    preview.innerHTML = renderMarkdown(cell.source);
    preview.addEventListener("click", () => {
      if (!cell.editing) {
        cell.editing = true;
        renderIgonbCells();
      }
    });
    container.appendChild(preview);
  } else {
    const output = document.createElement("div");
    output.className = "igonb-cell-output";
    output.innerHTML = cell.output
      ? cell.output
      : '<div class="igonb-empty-output">No output</div>';
    if (cell.error) {
      output.innerHTML += `<div class="igonb-error-output">${escapeHtml(cell.error)}</div>`;
    }
    container.appendChild(output);
  }

  return container;
}

function updateMarkdownPreview(container, source) {
  const preview = container.querySelector(".igonb-markdown-preview");
  if (preview) {
    preview.innerHTML = renderMarkdown(source);
  }
}

function startIgonbDrag(event, cellId) {
  igonbDragId = cellId;
  setIgonbSelectedId(cellId);
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData("text/plain", cellId);
  }
  document.body.classList.add("igonb-dragging");
}

function clearIgonbDrag() {
  igonbDragId = null;
  document.body.classList.remove("igonb-dragging");
  document
    .querySelectorAll(".igonb-cell.drag-over, .igonb-cell.drag-over-bottom")
    .forEach((el) => {
      el.classList.remove("drag-over", "drag-over-bottom");
    });
}

function handleIgonbDragOver(event, targetId) {
  if (!igonbDragId || !igonbState) return;
  if (targetId === igonbDragId) return;
  event.preventDefault();

  const container = event.currentTarget;
  if (!container) return;
  const rect = container.getBoundingClientRect();
  const isBottom = event.clientY > rect.top + rect.height / 2;

  container.classList.add("drag-over");
  container.classList.toggle("drag-over-bottom", isBottom);
}

function clearIgonbDragOver(targetId) {
  const container = document.querySelector(
    `.igonb-cell[data-cell-id="${targetId}"]`,
  );
  if (!container) return;
  container.classList.remove("drag-over", "drag-over-bottom");
}

function handleIgonbDrop(event, targetId) {
  if (!igonbDragId || !igonbState) return;
  event.preventDefault();

  const fromIndex = getIgonbIndexById(igonbDragId);
  const targetIndex = getIgonbIndexById(targetId);
  if (fromIndex === -1 || targetIndex === -1) {
    clearIgonbDrag();
    return;
  }

  const container = event.currentTarget;
  if (!container) {
    clearIgonbDrag();
    return;
  }

  const rect = container.getBoundingClientRect();
  const insertAfter = event.clientY > rect.top + rect.height / 2;
  let insertIndex = insertAfter ? targetIndex + 1 : targetIndex;
  moveIgonbCell(fromIndex, insertIndex);
  clearIgonbDrag();
}

function moveIgonbCell(fromIndex, insertIndex) {
  if (!igonbState) return;
  if (fromIndex === insertIndex || fromIndex + 1 === insertIndex) {
    renderIgonbCells();
    return;
  }
  const cells = igonbState.cells;
  const [moved] = cells.splice(fromIndex, 1);
  let targetIndex = insertIndex;
  if (fromIndex < insertIndex) {
    targetIndex -= 1;
  }
  if (targetIndex < 0) targetIndex = 0;
  if (targetIndex > cells.length) targetIndex = cells.length;
  cells.splice(targetIndex, 0, moved);
  scheduleIgonbSave();
  renderIgonbCells();
}

function addIgonbCell(language) {
  if (!igonbState) return;
  const newCell = {
    id: nextIgonbId(),
    language: language,
    source: "",
    output: "",
    error: "",
    running: false,
    waiting: false,
    editing: language === "markdown",
    focus: true,
  };

  const selectedIndex = getIgonbSelectedIndex();
  if (selectedIndex === null) {
    igonbState.cells.push(newCell);
  } else {
    igonbState.cells.splice(selectedIndex + 1, 0, newCell);
  }

  igonbSelectedId = newCell.id;
  scheduleIgonbSave();
  renderIgonbCells();
}

function deleteIgonbCell(index) {
  if (!igonbState) return;
  if (igonbState.cells.length <= 1) {
    showMessage("Cannot delete the last cell", "warning");
    return;
  }
  if (!confirm("Delete this cell?")) {
    return;
  }
  const removedId = igonbState.cells[index]?.id;
  igonbState.cells.splice(index, 1);
  if (removedId && removedId === igonbSelectedId) {
    const nextCell = igonbState.cells[index] || igonbState.cells[index - 1];
    igonbSelectedId = nextCell ? nextCell.id : null;
  }
  scheduleIgonbSave();
  renderIgonbCells();
}

function scheduleIgonbSave() {
  if (!activeFileName || !activeFileName.endsWith(".igonb")) return;
  if (!igonbState) return;
  clearTimeout(igonbSaveTimer);
  igonbSaveTimer = setTimeout(async () => {
    const payload = getIgonbContent();
    await UpdateFileContent(activeFileName, payload);
    scheduleWorkspaceRefresh();
  }, 600);
}

function getIgonbContent() {
  const payload = {
    version: igonbState ? igonbState.version || 1 : 1,
    cells: (igonbState ? igonbState.cells : []).map((cell) => ({
      language: cell.language,
      source: cell.source,
    })),
  };
  return JSON.stringify(payload, null, 2);
}

async function runIgonbCell(index) {
  if (!igonbState) return;
  try {
    scheduleIgonbSave();
    const target = igonbState.cells[index];
    if (target) {
      setIgonbSelectedId(target.id);
    }
    setIgonbRunning(index);
    const content = getIgonbContent();
    const results = await ExecuteIgonbCells(content, index);
    applyIgonbResults(results);
  } catch (error) {
    clearIgonbRunning();
    showMessage("Failed to execute cell: " + error, "error");
  }
}

async function runIgonbAll() {
  if (!igonbState) return;
  try {
    scheduleIgonbSave();
    setIgonbRunning(-1);
    const content = getIgonbContent();
    const results = await ExecuteIgonbCells(content, -1);
    applyIgonbResults(results);
  } catch (error) {
    clearIgonbRunning();
    showMessage("Failed to execute notebook: " + error, "error");
  }
}

function applyIgonbResults(results) {
  if (!Array.isArray(results) || !igonbState) return;
  let hadError = false;
  results.forEach((result) => {
    const idx = result.index;
    if (idx < 0 || idx >= igonbState.cells.length) return;
    igonbState.cells[idx].output = result.output || "";
    igonbState.cells[idx].error = result.error || "";
    igonbState.cells[idx].running = false;
    if (result.error) {
      hadError = true;
    }
    updateIgonbCellOutput(igonbState.cells[idx]);
  });
  clearIgonbRunning();
  if (hadError) {
    showMessage("Notebook execution stopped due to an error", "error");
  }
}

function disposeIgonbEditors() {
  igonbEditors.forEach((entry) => {
    if (entry.resizeObserver) {
      entry.resizeObserver.disconnect();
    }
    if (entry.editor) {
      entry.editor.dispose();
    }
    if (entry.model && !entry.model.isDisposed()) {
      entry.model.dispose();
    }
  });
  igonbEditors.clear();
}

function getIgonbMonacoLanguage(language) {
  switch (language) {
    case "python":
      return "python";
    case "markdown":
      return "markdown";
    default:
      return "go";
  }
}

function initIgonbEditors() {
  if (!igonbState) return;
  const theme = document.body.getAttribute("data-theme") || "dark";

  igonbState.cells.forEach((cell) => {
    const container = document.querySelector(
      `.igonb-cell[data-cell-id="${cell.id}"]`,
    );
    if (!container) return;
    const editorHost = container.querySelector(".igonb-cell-editor");
    if (!editorHost) return;

    const uri = monaco.Uri.parse(
      `inmemory://igonb/${encodeURIComponent(cell.id)}`,
    );
    let model = monaco.editor.getModel(uri);
    if (model) {
      model.dispose();
    }
    model = monaco.editor.createModel(
      cell.source,
      getIgonbMonacoLanguage(cell.language),
      uri,
    );

    const editorInstance = monaco.editor.create(editorHost, {
      model,
      theme: theme === "light" ? "vs-light" : "vs-dark",
      minimap: { enabled: false },
      scrollBeyondLastLine: false,
      lineNumbers: "on",
      glyphMargin: false,
      folding: true,
      wordWrap: "on",
      fontSize: editorFontSize,
      automaticLayout: true,
      overviewRulerLanes: 0,
    });

    let lastHeight = 0;
    const updateHeight = () => {
      const height = Math.max(120, editorInstance.getContentHeight());
      if (height !== lastHeight) {
        lastHeight = height;
        editorHost.style.height = `${height}px`;
      }
      editorInstance.layout();
    };
    const scheduleHeight = () => requestAnimationFrame(updateHeight);
    updateHeight();

    editorInstance.onDidContentSizeChange(scheduleHeight);
    let resizeObserver = null;
    if (typeof ResizeObserver !== "undefined") {
      resizeObserver = new ResizeObserver(() => {
        scheduleHeight();
      });
      resizeObserver.observe(editorHost);
    }
    editorInstance.onDidChangeModelContent(() => {
      cell.source = model.getValue();
      scheduleIgonbSave();
      if (cell.language === "markdown") {
        updateMarkdownPreview(container, cell.source);
      }
    });

    igonbEditors.set(cell.id, {
      editor: editorInstance,
      model,
      updateHeight,
      resizeObserver,
    });

    if (cell.focus) {
      editorInstance.focus();
      cell.focus = false;
    }

    if (cell.language === "markdown") {
      updateMarkdownPreview(container, cell.source);
    }
  });
}

function updateIgonbCellOutput(cell) {
  const container = document.querySelector(
    `.igonb-cell[data-cell-id="${cell.id}"]`,
  );
  if (!container) return;
  if (cell.language === "markdown") return;

  const output = container.querySelector(".igonb-cell-output");
  if (!output) return;
  output.innerHTML = cell.output
    ? cell.output
    : '<div class="igonb-empty-output">No output</div>';
  if (cell.error) {
    output.innerHTML += `<div class="igonb-error-output">${escapeHtml(cell.error)}</div>`;
  }
}

function setIgonbRunning(upToIndex) {
  if (!igonbState) return;
  const runnableIndices = [];
  igonbState.cells.forEach((cell, idx) => {
    const shouldRun = upToIndex < 0 ? true : idx <= upToIndex;
    if (shouldRun && cell.language !== "markdown") {
      runnableIndices.push(idx);
    }
  });

  const runningIndex = runnableIndices.length > 0 ? runnableIndices[0] : -1;
  const runnableSet = new Set(runnableIndices);

  igonbState.cells.forEach((cell, idx) => {
    const shouldRun = runnableSet.has(idx);
    cell.running = idx === runningIndex;
    cell.waiting = shouldRun && idx !== runningIndex;
    updateIgonbCellRunningUI(cell);
  });
}

function clearIgonbRunning() {
  if (!igonbState) return;
  igonbState.cells.forEach((cell) => {
    cell.running = false;
    cell.waiting = false;
    updateIgonbCellRunningUI(cell);
  });
}

function updateIgonbCellRunningUI(cell) {
  const container = document.querySelector(
    `.igonb-cell[data-cell-id="${cell.id}"]`,
  );
  if (!container) return;
  if (cell.running) {
    container.classList.add("running");
  } else {
    container.classList.remove("running");
  }
  if (cell.waiting) {
    container.classList.add("waiting");
  } else {
    container.classList.remove("waiting");
  }
  const status = container.querySelector(".igonb-cell-status");
  if (status) {
    status.textContent = cell.running
      ? "Running..."
      : cell.waiting
        ? "Waiting..."
        : "";
  }
  const runBtn = container.querySelector(".igonb-cell-run");
  if (runBtn) {
    runBtn.disabled = cell.running || cell.waiting;
  }
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
    fontSize: editorFontSize,
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
      const memberMatch = linePrefix.match(
        /([A-Za-z_]\w*(?:\.[A-Za-z_]\w*)*)\.$/,
      );

      if (memberMatch) {
        const target = memberMatch[1];
        const suggestions = [];
        const { structs, varTypes, aliases } = getGoParse(model);
        const resolvedType = resolveExpressionType(
          target,
          structs,
          varTypes,
          aliases,
        );
        const typeName = resolvedType || varTypes.get(target) || target;
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

        if (!typeInfo) {
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
        }
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
    updateActiveFileModelSize();
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
    showMessage(
      "Run is only available for .go, .py, and .igonb files",
      "warning",
    );
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
    if (activeFileName.endsWith(".igonb")) {
      await runIgonbAll();
      setResultOutput(
        '<div style="color: #888;">Notebook output is shown inline.</div>',
      );
      return;
    }

    const code = editor.getValue();
    let result = "";

    if (activeFileName.endsWith(".go")) {
      result = await ExecuteCode(code);
    } else if (activeFileName.endsWith(".py")) {
      result = await ExecutePythonFile(activeFileName, code);
    } else {
      showMessage(
        "Run is only available for .go, .py, and .igonb files",
        "warning",
      );
      return;
    }

    setResultOutput(result);
  } catch (error) {
    setResultOutput(`<div class="error-message">Error: ${error}</div>`);
  } finally {
    isExecuting = false;
    runButton.disabled = false;
    runButton.innerHTML = '<i class="fas fa-play"></i> Run';
    scheduleWorkspaceRefresh();
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
    scheduleWorkspaceRefresh();
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
    isRunnableActiveFile() &&
    !isImagePreview &&
    !isLargeFilePreview &&
    !isBinaryPreview;

  runButton.disabled = !runnable;
  runButton.title = runnable
    ? "Run"
    : "Run is only available for .go, .py, and .igonb files";
}

function isRunnableFileName(filename) {
  return (
    filename.endsWith(".go") ||
    filename.endsWith(".py") ||
    filename.endsWith(".igonb")
  );
}

function isRunnableActiveFile() {
  if (!activeFileName) return false;
  if (!isRunnableFileName(activeFileName)) return false;
  if (activeFileName.endsWith(".igonb") && !isIgonbView) return false;
  return !isImagePreview && !isLargeFilePreview && !isBinaryPreview;
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
  const loadToken = ++workspaceLoadToken;
  try {
    const [files, activeFile] = await Promise.all([
      GetWorkspaceFiles(),
      GetActiveFile(),
    ]);
    if (loadToken !== workspaceLoadToken) {
      return;
    }
    workspaceFiles = files;
    activeFileName = activeFile;
    renderFileTree();
  } catch (error) {
    console.error("Failed to load workspace files:", error);
  }
}

function scheduleWorkspaceRefresh(delayMs = 120) {
  if (workspaceRefreshTimer) return;
  workspaceRefreshTimer = setTimeout(() => {
    workspaceRefreshTimer = null;
    loadWorkspaceFiles();
  }, delayMs);
}

// Load workspace with retry logic to ensure backend is ready
async function loadWorkspaceWithRetry(maxRetries = 5, delayMs = 100) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      const [files, activeFile] = await Promise.all([
        GetWorkspaceFiles(),
        GetActiveFile(),
      ]);
      workspaceFiles = files;
      activeFileName = activeFile;

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

  const renderToken = ++fileTreeRenderToken;
  closeActionMenu();
  fileTree.innerHTML = "";

  const treeRoot = buildFileTree(workspaceFiles);
  const initialized = expandedDirs.size > 0;
  const renderList = [];

  collectRenderEntries(treeRoot, 0, initialized, renderList);

  let index = 0;
  const renderChunk = () => {
    if (renderToken !== fileTreeRenderToken) return;
    const fragment = document.createDocumentFragment();
    const end = Math.min(index + fileTreeChunkSize, renderList.length);

    for (; index < end; index += 1) {
      const { entry, depth } = renderList[index];
      fragment.appendChild(createFileItem(entry, depth));
    }

    fileTree.appendChild(fragment);

    if (index < renderList.length) {
      requestAnimationFrame(renderChunk);
    }
  };

  requestAnimationFrame(renderChunk);
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

function sortTreeEntries(node) {
  return Array.from(node.children.values()).sort((a, b) => {
    if (a.isDir !== b.isDir) {
      return a.isDir ? -1 : 1;
    }
    return a.name.localeCompare(b.name);
  });
}

function collectRenderEntries(node, depth, initialized, list) {
  const entries = sortTreeEntries(node);

  entries.forEach((entry) => {
    if (entry.isDir && !initialized) {
      expandedDirs.add(entry.path);
    }

    list.push({ entry, depth });

    if (entry.isDir && expandedDirs.has(entry.path)) {
      collectRenderEntries(entry, depth + 1, true, list);
    }
  });
}

function createFileItem(entry, depth) {
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
  } else if (entry.path.endsWith(".igonb")) {
    iconClass = "fa-book";
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

  return fileItem;
}

async function switchToFile(filename, force = false) {
  if (!force && filename === activeFileName) return;

  try {
    const previousFileName = activeFileName;
    const previousModel = editor.getModel();

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
        const currentContent =
          isIgonbView && activeFileName.endsWith(".igonb")
            ? getIgonbContent()
            : editor.getValue();
        await UpdateFileContent(activeFileName, currentContent);
      } else {
        activeFileName = "";
        hideImagePreview();
        hideLargeFilePreview();
        hideBinaryPreview();
        hideIgonbNotebook();
      }
    }

    // Switch to new file
    const selectedFile = workspaceFiles.find((file) => file.name === filename);
    activeFileName = filename;
    selectedFolderPath = getParentPath(filename);
    isRootFolderSelected = false;
    document.getElementById("active-file-label").textContent = filename;
    renderFileTree();
    const setActivePromise = SetActiveFile(filename);

    if (selectedFile && selectedFile.tooLarge) {
      showLargeFilePreview(filename, selectedFile.size || 0);
      clearPreviewIfNeeded();
      updateRunButtonState();
      await setActivePromise;
      scheduleWorkspaceRefresh();
      return;
    }

    const cachedModel = getCachedFileModel(filename);
    if (cachedModel) {
      const cachedContent = cachedModel.getValue();
      hideImagePreview();
      hideLargeFilePreview();
      hideBinaryPreview();
      applyTextFileContent(filename, cachedContent, true);
      disposeOrphanModel(previousFileName, previousModel);
      updateRunButtonState();
      await setActivePromise;
      scheduleWorkspaceRefresh();
      return;
    }

    const content = await GetFileContent(filename);
    hideIgonbNotebook();

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

      applyTextFileContent(filename, content, false);
      disposeOrphanModel(previousFileName, previousModel);
    }

    updateRunButtonState();
    await setActivePromise;
    // Refresh file tree without blocking UI
    scheduleWorkspaceRefresh();
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
    scheduleWorkspaceRefresh();
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
    const wasActive = filename === activeFileName;
    const previousModel = wasActive ? editor.getModel() : null;

    await DeleteFile(filename);
    removeFileModel(filename);
    activeFileName = "";
    await loadWorkspaceFiles();

    // If deleted file was active, switch to first available
    if (workspaceFiles.length > 0) {
      hideImagePreview();
      hideBinaryPreview();
      await switchToFile(workspaceFiles[0].name, true);
    }
    if (wasActive) {
      disposeOrphanModel(filename, previousModel);
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
    const activeFileBefore = activeFileName;
    const wasActive =
      activeFileBefore && activeFileBefore.startsWith(`${folderPath}/`);
    const previousModel = wasActive ? editor.getModel() : null;

    await DeleteFolder(folderPath);
    removeFileModelsInFolder(folderPath);
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
    scheduleWorkspaceRefresh();
    if (wasActive) {
      disposeOrphanModel(activeFileBefore, previousModel);
    }
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
    renameFileModel(filename, trimmedName);

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
    renameFileModelsInFolder(folderPath, trimmedName);
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
    scheduleWorkspaceRefresh();
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
    const currentContent =
      isIgonbView && activeFileName.endsWith(".igonb")
        ? getIgonbContent()
        : editor.getValue();
    await UpdateFileContent(activeFileName, currentContent);

    // Try to save to disk
    await SaveFile(activeFileName);
    scheduleWorkspaceRefresh();
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
      const currentContent =
        isIgonbView && activeFileName.endsWith(".igonb")
          ? getIgonbContent()
          : editor.getValue();
      await UpdateFileContent(activeFileName, currentContent);
    }

    await SaveAllFiles();
    scheduleWorkspaceRefresh();
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

    scheduleWorkspaceRefresh();
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
        applyTextFileContent(activeFile, content, false);
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

function applyIgonbFontSizes() {
  document.documentElement.style.setProperty(
    "--igonb-editor-font-size",
    `${editorFontSize}px`,
  );
  document.documentElement.style.setProperty(
    "--igonb-output-font-size",
    `${outputFontSize}px`,
  );

  if (!igonbEditors.size) {
    return;
  }

  igonbEditors.forEach((entry) => {
    if (!entry.editor) return;
    entry.editor.updateOptions({ fontSize: editorFontSize });
    if (entry.updateHeight) {
      entry.updateHeight();
    } else {
      entry.editor.layout();
    }
  });
}

// Change editor font size
function changeEditorFontSize(delta) {
  editorFontSize = Math.max(8, Math.min(32, editorFontSize + delta));
  editor.updateOptions({ fontSize: editorFontSize });
  document.getElementById("editor-font-size").textContent = editorFontSize;
  applyIgonbFontSizes();
}

// Change output font size
function changeOutputFontSize(delta) {
  outputFontSize = Math.max(8, Math.min(32, outputFontSize + delta));
  const resultContainer = document.querySelector(".result-container");
  resultContainer.style.fontSize = outputFontSize + "px";
  document.getElementById("output-font-size").textContent = outputFontSize;
  applyIgonbFontSizes();
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
  applyIgonbFontSizes();

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
    // Ctrl/Cmd + Shift + Enter to add a new cell in igonb
    if (
      isIgonbView &&
      (e.ctrlKey || e.metaKey) &&
      e.shiftKey &&
      e.key === "Enter"
    ) {
      e.preventDefault();
      addIgonbCell(getIgonbSelectedLanguage());
      return;
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
          hideImagePreview();
          hideLargeFilePreview();
          hideBinaryPreview();
          document.getElementById("active-file-label").textContent =
            activeFileName;
          selectedFolderPath = getParentPath(activeFileName);
          isRootFolderSelected = false;
          applyTextFileContent(activeFileName, content, false);
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
      if (activeFileName.endsWith(".igonb") && isIgonbView) {
        return;
      }
      updateActiveFileModelSize();
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
          scheduleWorkspaceRefresh();
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
