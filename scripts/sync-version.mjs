import { readFileSync, writeFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const root = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const versionPath = resolve(root, "VERSION.txt");
const version = readFileSync(versionPath, "utf8").trim();

if (!version) {
  throw new Error("VERSION.txt is empty");
}

const updateJson = (path, update) => {
  const data = JSON.parse(readFileSync(path, "utf8"));
  update(data);
  writeFileSync(path, JSON.stringify(data, null, 2) + "\n");
};

updateJson(resolve(root, "wails.json"), (data) => {
  data.info = data.info || {};
  data.info.productVersion = version;
});

updateJson(resolve(root, "frontend", "package.json"), (data) => {
  data.version = version;
});

console.log(`Synced version ${version}`);
