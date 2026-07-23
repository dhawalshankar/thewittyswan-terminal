// -----------------------------------------------------------------
// Virtual filesystem — persisted in localStorage, purely client-side.
// No round-trip to the Go backend for any of these commands.
// -----------------------------------------------------------------
const FS_KEY = "swan_vfs";

function loadFS() {
  const raw = localStorage.getItem(FS_KEY);
  if (raw) {
    try { return JSON.parse(raw); } catch (e) { /* fall through to fresh fs */ }
  }
  return { type: "dir", name: "/", children: {} };
}
function saveFS(fs) {
  localStorage.setItem(FS_KEY, JSON.stringify(fs));
}

let vfs = loadFS();
let cwd = ["/"]; // path stack, e.g. ["/", "notes", "drafts"]

function resolveDir(fs, pathParts) {
  let node = fs;
  for (const part of pathParts.slice(1)) {
    if (!node.children[part] || node.children[part].type !== "dir") return null;
    node = node.children[part];
  }
  return node;
}

function vfsPwdString() {
  if (cwd.length === 1) return "/";
  return "/" + cwd.slice(1).join("/");
}

function vfsMkdir(name) {
  if (!name) return "usage: mkdir <name>";
  const dir = resolveDir(vfs, cwd);
  if (dir.children[name]) return `mkdir: cannot create directory '${name}': already exists`;
  dir.children[name] = { type: "dir", name, children: {} };
  saveFS(vfs);
  return "";
}

function vfsTouch(name) {
  if (!name) return "usage: touch <name>";
  const dir = resolveDir(vfs, cwd);
  if (!dir.children[name]) {
    dir.children[name] = { type: "file", name, content: "" };
    saveFS(vfs);
  }
  return "";
}

function vfsLs() {
  const dir = resolveDir(vfs, cwd);
  const entries = Object.values(dir.children);
  if (entries.length === 0) return "(empty)";
  return entries
    .map(e => (e.type === "dir" ? e.name + "/" : e.name))
    .join("  ");
}

function vfsCd(name) {
  if (!name || name === "~") { cwd = ["/"]; return ""; }
  if (name === "/") { cwd = ["/"]; return ""; }
  if (name === "..") {
    if (cwd.length > 1) cwd.pop();
    return "";
  }
  const dir = resolveDir(vfs, cwd);
  if (!dir.children[name] || dir.children[name].type !== "dir") {
    return `cd: no such directory: ${name}`;
  }
  cwd.push(name);
  return "";
}

function vfsRm(name) {
  if (!name) return "usage: rm <name>";
  const dir = resolveDir(vfs, cwd);
  if (!dir.children[name]) return `rm: no such file or directory: ${name}`;
  delete dir.children[name];
  saveFS(vfs);
  return "";
}

// Returns null if no matching file exists in the current vfs dir —
// caller falls back to the backend's blog `cat <slug>` in that case.
function vfsCatIfExists(name) {
  const dir = resolveDir(vfs, cwd);
  const f = dir.children[name];
  if (!f) return null;
  if (f.type === "dir") return `cat: ${name}: is a directory`;
  return f.content.length > 0 ? f.content : "(empty file)";
}

// Overwrites (or appends to) a file's content. Creates the file if it
// doesn't exist yet, same as touch would.
function vfsWriteFile(name, content, append) {
  if (!name) return "usage: echo <text> > <file>";
  const dir = resolveDir(vfs, cwd);
  if (dir.children[name] && dir.children[name].type === "dir") {
    return `${name}: is a directory`;
  }
  if (!dir.children[name]) {
    dir.children[name] = { type: "file", name, content: "" };
  }
  const file = dir.children[name];
  if (append && file.content) {
    file.content += "\n" + content;
  } else {
    file.content = content;
  }
  saveFS(vfs);
  return "";
}