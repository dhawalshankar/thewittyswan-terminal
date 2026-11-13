const win = document.getElementById("terminalWindow");
const titlebar = document.getElementById("titlebar");
const terminalEl = document.getElementById("terminal");
const resizeHandle = document.getElementById("resizeHandle");
const reopenPill = document.getElementById("reopenPill");
const closeBtn = document.getElementById("closeBtn");
const minimizeBtn = document.getElementById("minimizeBtn");
const maximizeBtn = document.getElementById("maximizeBtn");
const scrollControls = document.getElementById("scrollControls");
const safeAreaProbe = document.getElementById("safeAreaProbe");
const scrollUpBtn = document.getElementById("scrollUpBtn");
const scrollDownBtn = document.getElementById("scrollDownBtn");

function isMobile() {
  return window.innerWidth <= 640;
}

function getSafeArea() {
  const cs = getComputedStyle(safeAreaProbe);
  return {
    top: parseFloat(cs.paddingTop) || 0,
    right: parseFloat(cs.paddingRight) || 0,
    bottom: parseFloat(cs.paddingBottom) || 0,
    left: parseFloat(cs.paddingLeft) || 0,
  };
}

function getViewportRect() {
  const vv = window.visualViewport;
  const safe = getSafeArea();
  const left = (vv ? vv.offsetLeft : 0) + safe.left;
  const top = (vv ? vv.offsetTop : 0) + safe.top;
  const width = (vv ? vv.width : window.innerWidth) - safe.left - safe.right;
  const height = (vv ? vv.height : window.innerHeight) - safe.top - safe.bottom;
  return { left, top, width, height };
}

const term = new Terminal({
  cursorBlink: true,
  fontFamily: "Courier New, monospace",
  fontSize: isMobile() ? 13 : 15,
  theme: {
    background: "#111214",
    foreground: "#d8d8dc",
    cursor: "#e0c9a6"
  },
  scrollback: 5000,
});

const webLinksAddon = new WebLinksAddon.WebLinksAddon((event, uri) => {
  if (/^https?:\/\//i.test(uri)) {
    window.open(uri, "_blank", "noopener,noreferrer");
  }
});
term.loadAddon(webLinksAddon);

const fitAddon = new FitAddon.FitAddon();
term.loadAddon(fitAddon);
term.open(terminalEl);

function applyInitialLayout() {
  const vp = getViewportRect();
  const w = Math.min(vp.width * 0.94, 900);
  const h = Math.min(vp.height * 0.82, 650);
  win.style.width = w + "px";
  win.style.height = h + "px";
  win.style.left = (vp.left + (vp.width - w) / 2) + "px";
  win.style.top = (vp.top + (vp.height - h) / 2) + "px";
}
applyInitialLayout();
fitAddon.fit();

const resizeObserver = new ResizeObserver(() => {
  try { fitAddon.fit(); } catch (e) {}
});
resizeObserver.observe(win);

function clampToViewport() {
  const vp = getViewportRect();
  const rect = win.getBoundingClientRect();

  let w = Math.min(rect.width, vp.width);
  let h = Math.min(rect.height, vp.height);
  let left = Math.min(Math.max(rect.left, vp.left), vp.left + vp.width - w);
  let top = Math.min(Math.max(rect.top, vp.top), vp.top + vp.height - h);

  win.style.width = w + "px";
  win.style.height = h + "px";
  win.style.left = left + "px";
  win.style.top = top + "px";
}

// Runs the actual layout update. Called instantly on keyboard show/hide
// (visualViewport resize/scroll) so the input line never gets stuck
// behind the keyboard, and on a short debounce for window resize /
// orientation change (where instant reflow can look janky).
function applyViewportChange() {
  term.options.fontSize = isMobile() ? 13 : 15;
  clampToViewport();
  try { fitAddon.fit(); } catch (e) {}
  term.scrollToBottom();
}

let resizeTimeout;
function handleWindowResize() {
  clearTimeout(resizeTimeout);
  resizeTimeout = setTimeout(applyViewportChange, 100);
}

window.addEventListener("resize", handleWindowResize);
window.addEventListener("orientationchange", handleWindowResize);

if (window.visualViewport) {
  // Keyboard open/close must react immediately — no debounce here,
  // otherwise the prompt sits behind the keyboard for ~100ms+ and
  // on some mobile browsers never gets nudged back into view at all.
  window.visualViewport.addEventListener("resize", applyViewportChange);
  window.visualViewport.addEventListener("scroll", applyViewportChange);
}

// -----------------------------------------------------------------
// Dragging the window by its titlebar (mouse + touch, all screens)
// -----------------------------------------------------------------
let dragOffsetX = 0, dragOffsetY = 0, dragging = false;

function startDrag(clientX, clientY) {
  const rect = win.getBoundingClientRect();
  dragOffsetX = clientX - rect.left;
  dragOffsetY = clientY - rect.top;
  dragging = true;
  win.classList.add("dragging");
  titlebar.classList.add("grabbing");
}

function doDrag(clientX, clientY) {
  if (!dragging) return;
  const vp = getViewportRect();
  const rect = win.getBoundingClientRect();
  let left = clientX - dragOffsetX;
  let top = clientY - dragOffsetY;
  left = Math.min(Math.max(left, vp.left - rect.width + 80), vp.left + vp.width - 80);
  top = Math.min(Math.max(top, vp.top), vp.top + vp.height - 40);
  win.style.left = left + "px";
  win.style.top = top + "px";
}

function endDrag() {
  dragging = false;
  win.classList.remove("dragging");
  titlebar.classList.remove("grabbing");
}

titlebar.addEventListener("mousedown", (e) => {
  if (e.target.closest(".dot")) return;
  startDrag(e.clientX, e.clientY);
  e.preventDefault();
});
window.addEventListener("mousemove", (e) => doDrag(e.clientX, e.clientY));
window.addEventListener("mouseup", endDrag);

titlebar.addEventListener("touchstart", (e) => {
  if (e.target.closest(".dot")) return;
  const t = e.touches[0];
  startDrag(t.clientX, t.clientY);
}, { passive: true });
window.addEventListener("touchmove", (e) => {
  if (!dragging) return;
  const t = e.touches[0];
  doDrag(t.clientX, t.clientY);
  e.preventDefault();
}, { passive: false });
window.addEventListener("touchend", endDrag);
window.addEventListener("touchcancel", endDrag);

// -----------------------------------------------------------------
// Resizing via the bottom-right handle (mouse + touch, all screens)
// -----------------------------------------------------------------
let resizing = false, resizeStartX = 0, resizeStartY = 0, startW = 0, startH = 0;

function startResize(clientX, clientY) {
  const rect = win.getBoundingClientRect();
  resizing = true;
  resizeStartX = clientX;
  resizeStartY = clientY;
  startW = rect.width;
  startH = rect.height;
  win.classList.add("resizing");
}

function doResize(clientX, clientY) {
  if (!resizing) return;
  const rect = win.getBoundingClientRect();
  const vp = getViewportRect();
  const maxW = vp.left + vp.width - rect.left;
  const maxH = vp.top + vp.height - rect.top;
  const minW = Math.min(280, maxW);
  const minH = Math.min(200, maxH);
  let w = startW + (clientX - resizeStartX);
  let h = startH + (clientY - resizeStartY);
  w = Math.min(Math.max(w, minW), maxW);
  h = Math.min(Math.max(h, minH), maxH);
  win.style.width = w + "px";
  win.style.height = h + "px";
}

function endResize() {
  resizing = false;
  win.classList.remove("resizing");
}

resizeHandle.addEventListener("mousedown", (e) => {
  startResize(e.clientX, e.clientY);
  e.preventDefault();
  e.stopPropagation();
});
resizeHandle.addEventListener("keydown", (e) => {
  const directions = {
    ArrowLeft: [-1, 0],
    ArrowRight: [1, 0],
    ArrowUp: [0, -1],
    ArrowDown: [0, 1],
  };
  const direction = directions[e.key];
  if (!direction) return;

  const rect = win.getBoundingClientRect();
  const vp = getViewportRect();
  const maxW = vp.left + vp.width - rect.left;
  const maxH = vp.top + vp.height - rect.top;
  const step = e.shiftKey ? 40 : 10;
  const width = Math.min(maxW, Math.max(Math.min(280, maxW), rect.width + direction[0] * step));
  const height = Math.min(maxH, Math.max(Math.min(200, maxH), rect.height + direction[1] * step));
  win.style.width = width + "px";
  win.style.height = height + "px";
  fitAddon.fit();
  e.preventDefault();
});
window.addEventListener("mousemove", (e) => doResize(e.clientX, e.clientY));
window.addEventListener("mouseup", endResize);

resizeHandle.addEventListener("touchstart", (e) => {
  const t = e.touches[0];
  startResize(t.clientX, t.clientY);
  e.stopPropagation();
}, { passive: true });
window.addEventListener("touchmove", (e) => {
  if (!resizing) return;
  const t = e.touches[0];
  doResize(t.clientX, t.clientY);
  e.preventDefault();
}, { passive: false });
window.addEventListener("touchend", endResize);
window.addEventListener("touchcancel", endResize);

// -----------------------------------------------------------------
// Traffic-light controls: close / minimize / maximize-restore
// -----------------------------------------------------------------
let savedRect = null;
let minimized = false;
let preMinimizeHeight = null;

closeBtn.addEventListener("click", () => {
  win.style.display = "none";
  reopenPill.classList.add("visible");
  reopenPill.focus();
});
reopenPill.addEventListener("click", () => {
  win.style.display = "flex";
  reopenPill.classList.remove("visible");
  if (minimized) setMinimized(false);
  requestAnimationFrame(() => {
    applyViewportChange();
    term.focus();
  });
});

function setMinimized(next) {
  minimized = next;
  if (minimized) {
    preMinimizeHeight = win.style.height || win.getBoundingClientRect().height + "px";
    terminalEl.style.display = "none";
    scrollControls.hidden = true;
    resizeHandle.style.display = "none";
    win.style.height = "auto";
  } else {
    terminalEl.style.display = "block";
    scrollControls.hidden = false;
    resizeHandle.style.display = "block";
    win.style.height = preMinimizeHeight;
  }
  minimizeBtn.setAttribute("aria-label", minimized ? "Restore terminal" : "Minimize terminal");
  minimizeBtn.title = minimized ? "Restore" : "Minimize";
}
minimizeBtn.addEventListener("click", () => {
  setMinimized(!minimized);
  if (!minimized) {
    requestAnimationFrame(() => {
      applyViewportChange();
      term.focus();
    });
  }
});

maximizeBtn.addEventListener("click", () => {
  const vp = getViewportRect();
  if (savedRect) {
    win.style.left = savedRect.left;
    win.style.top = savedRect.top;
    win.style.width = savedRect.width;
    win.style.height = savedRect.height;
    savedRect = null;
  } else {
    const rect = win.getBoundingClientRect();
    savedRect = {
      left: rect.left + "px",
      top: rect.top + "px",
      width: rect.width + "px",
      height: rect.height + "px"
    };
    win.style.left = vp.left + "px";
    win.style.top = vp.top + "px";
    win.style.width = vp.width + "px";
    win.style.height = vp.height + "px";
  }
  maximizeBtn.setAttribute("aria-label", savedRect ? "Restore terminal" : "Maximize terminal");
  maximizeBtn.title = savedRect ? "Restore" : "Maximize";
});

// -----------------------------------------------------------------
// Custom scroll buttons
// -----------------------------------------------------------------
function wireScrollButton(btn, lines) {
  let holdTimer = null, holdInterval = null;
  const doScroll = () => term.scrollLines(lines);
  const start = (e) => {
    e.preventDefault();
    e.stopPropagation();
    doScroll();
    holdTimer = setTimeout(() => {
      holdInterval = setInterval(doScroll, 80);
    }, 350);
  };
  const stop = () => {
    clearTimeout(holdTimer);
    clearInterval(holdInterval);
  };
  btn.addEventListener("mousedown", start);
  btn.addEventListener("touchstart", start, { passive: false });
  ["mouseup", "mouseleave", "touchend", "touchcancel"].forEach(ev =>
    btn.addEventListener(ev, stop)
  );
}
wireScrollButton(scrollUpBtn, -3);
wireScrollButton(scrollDownBtn, 3);

function updateScrollControls() {
  const buffer = term.buffer.active;
  scrollUpBtn.disabled = buffer.viewportY <= 0;
  scrollDownBtn.disabled = buffer.viewportY >= buffer.baseY;
}
term.onScroll(updateScrollControls);
term.onRender(updateScrollControls);

terminalEl.addEventListener("touchend", () => {
  if (!dragging && !resizing) term.focus();
});

// vfsPwdString, vfsMkdir, vfsTouch, vfsLs, vfsCd, vfsRm, vfsCatIfExists,
// vfsWriteFile are all defined in vfs.js, loaded before this file.

const prompt = () => `swan@thewittyswan:${vfsPwdString()}$ `;
let currentLine = "";
let busy = false;
let currentAbortController = null; // tracks the in-flight /run fetch, so Ctrl+C can cancel it

// Command history
const history = [];
let historyIndex = -1;

// Line-editor state — non-null while `edit <file>` is active. When set,
// onData routes all keystrokes into the line editor instead of the
// normal command pipeline (see top of term.onData).
let editMode = null; // null, or { filename, lines }

function reportClipboardFailure(action) {
  term.write(`\r\n[clipboard] Unable to ${action}; check your browser's clipboard permission.\r\n`);
  if (!busy) {
    term.write(editMode ? `> ${currentLine}` : prompt() + currentLine);
  }
  term.scrollToBottom();
}

term.attachCustomKeyEventHandler((event) => {
  if (event.type !== "keydown" || !(event.ctrlKey || event.metaKey)) return true;

  const key = event.key.toLowerCase();
  if (key === "c" && term.hasSelection() && navigator.clipboard?.writeText) {
    event.preventDefault();
    navigator.clipboard.writeText(term.getSelection())
      .catch(() => reportClipboardFailure("copy"));
    return false;
  }

  if (key === "v" && navigator.clipboard?.readText) {
    event.preventDefault();
    navigator.clipboard.readText().then((text) => {
      if (!text) return;
      const pastedText = editMode
        ? text.replace(/\r\n?|\n/g, "\r")
        : text.replace(/\r\n?|\n/g, " ");
      term.paste(pastedText);
    }).catch(() => reportClipboardFailure("paste"));
    return false;
  }

  return true;
});

// writePrompt always autoscrolls to the bottom, like a normal terminal,
// so the newest prompt is guaranteed fully visible. There's no attempt
// to pin long output to the top anymore — for output taller than the
// visible window, scroll up manually (or use the ▲ button) to read it
// from the start.
function writePrompt() {
  term.write("\r\n" + prompt());
  term.scrollToBottom();
}
function showBanner() {
  term.writeln("");
  term.writeln("\x1b[1;38;2;224;201;166mTheWittySwan\x1b[0m");
  term.writeln("\x1b[1mInteractive Portfolio Terminal\x1b[0m");
  term.writeln("");
  term.writeln("Dhawal Shukla");
  term.writeln("");
  term.writeln("Backend Engineering");
  term.writeln("Network Security • Cloud");
  term.writeln("");
  term.writeln("Type 'help' to explore.");
  term.writeln("Ctrl+V paste");
  term.writeln("Select text; Ctrl+C copies");
  term.writeln("");
}
function clearCurrentLine() {
  term.write("\b \b".repeat(currentLine.length));
}

function navigateHistory(direction) {
  if (history.length === 0) return;

  if (direction === -1) {
    if (historyIndex === -1) {
      historyIndex = history.length - 1;
    } else if (historyIndex > 0) {
      historyIndex--;
    } else {
      return;
    }
  } else {
    if (historyIndex === -1) return;
    if (historyIndex < history.length - 1) {
      historyIndex++;
    } else {
      historyIndex = -1;
    }
  }

  clearCurrentLine();
  currentLine = historyIndex === -1 ? "" : history[historyIndex];
  term.write(currentLine);
}

// -----------------------------------------------------------------
// Known-typo normalization — fixes common slips before the command
// ever reaches the backend. Deliberately a fixed lookup table (not
// fuzzy matching) so it never "corrects" something you didn't mean.
// -----------------------------------------------------------------
const COMMON_TYPOS = {
  "localhots": "localhost",
  "helo": "help",
  "hlep": "help",
  "claer": "clear",
  "whoiam": "whoami",
  "prjects": "projects",
  "experiance": "experience",
  "achievments": "achievements",
  "resuem": "resume"
};

function normalizeInput(command) {
  const parts = command.split(" ");
  const corrected = parts.map(word => COMMON_TYPOS[word.toLowerCase()] || word);
  return corrected.join(" ");
}

function isResumeOpenCommand(command) {
  return /^resume\s+(pdf|open|download)$/.test(command.trim());
}

// -----------------------------------------------------------------
// Redirect parsing for `echo <text> > file` / `echo <text> >> file`.
// Returns null if the command isn't an echo-with-redirect, so plain
// `echo hi` (no redirect) falls through to the plain-echo branch below.
// -----------------------------------------------------------------
function parseRedirect(command) {
  const match = command.match(/^echo\s+(.*?)\s*(>>|>)\s*(\S+)\s*$/);
  if (!match) return null;
  let [, rawText, op, filename] = match;
  rawText = rawText.trim();
  if (
    (rawText.startsWith('"') && rawText.endsWith('"')) ||
    (rawText.startsWith("'") && rawText.endsWith("'"))
  ) {
    rawText = rawText.slice(1, -1);
  }
  return { text: rawText, append: op === ">>", filename };
}

// Commits or discards the in-progress `edit` session and returns to the
// normal prompt.
function finishEdit(save) {
  if (save) {
    const content = editMode.lines.join("\n");
    vfsWriteFile(editMode.filename, content, false);
    term.writeln(`\r\n"${editMode.filename}" saved (${editMode.lines.length} line${editMode.lines.length === 1 ? "" : "s"})`);
  } else {
    term.writeln("\r\nedit aborted");
  }
  editMode = null;
  currentLine = "";
  writePrompt();
}

showBanner();
writePrompt();
term.focus();

// onData (not onKey) fires after the browser/OS has resolved the
// actual character — this correctly handles mobile autocorrect,
// predictive text, IME composition, and pasted text.
term.onData(async (data) => {
  // Ctrl+C (\x03) — must work even while a command is in flight or an
  // edit session is open, so this check sits above every other guard.
  if (data === "\x03") {
    if (editMode) {
      clearCurrentLine();
      term.write("^C");
      finishEdit(false);
      return;
    }
    if (busy) currentAbortController?.abort();
    else clearCurrentLine();
    term.write("^C");
    currentLine = "";
    historyIndex = -1;
    if (!busy) writePrompt();
    return;
  }

  // -------------------------------------------------------------
  // Line-editor mode — `edit <file>` routes all keystrokes here
  // instead of the normal command pipeline below, so history
  // navigation, typo-fixing, and /run never see editor input.
  // -------------------------------------------------------------
  if (editMode) {
    let i = 0;
    while (i < data.length) {
      const char = data[i];
      const code = char.charCodeAt(0);

      if (code === 4) { // Ctrl+D
        finishEdit(true);
        return;
      }
      if (char === "\r") {
        term.write("\r\n");
        if (currentLine === ".") {
          finishEdit(true);
          return;
        }
        editMode.lines.push(currentLine);
        currentLine = "";
        term.write("> ");
      } else if (char === "\b" || code === 127) {
        if (currentLine.length > 0) {
          currentLine = currentLine.slice(0, -1);
          term.write("\b \b");
        }
      } else if (code < 32) {
        // ignore other control chars
      } else {
        currentLine += char;
        term.write(char);
      }
      i += 1;
    }
    return;
  }

  if (busy) return; // ignore other input while a command is in flight

  let i = 0;
  while (i < data.length) {
    const char = data[i];
    const code = char.charCodeAt(0);

    if (code === 27 && data[i + 1] === "[") {
      const seq = data[i + 2];
      if (seq === "A") navigateHistory(-1);
      else if (seq === "B") navigateHistory(1);
      i += 3;
      continue;
    }

    if (char === "\r") { // Enter
      let command = currentLine.trim();
      command = normalizeInput(command);

      term.write("\r\n");

      currentLine = "";
      historyIndex = -1;

      if (command.length > 0 && history[history.length - 1] !== command) {
        history.push(command);
      }

      if (command === "clear") {
        term.clear();
        showBanner();
        writePrompt();
        return;
      }

      if (command === "history") {
        if (history.length === 0) {
          term.writeln("No commands in history.");
        } else {
          history.forEach((cmd, idx) => {
            term.writeln(`${idx + 1}  ${cmd}`);
          });
        }
        writePrompt();
        return;
      }

      if (command.length === 0) {
        writePrompt();
        return;
      }

      // -------------------------------------------------------------
      // Virtual filesystem commands — all client-side, no /run call.
      //
      // Naming-conflict resolution:
      // - "ls" with NO args   -> vfs directory listing
      // - "ls blog" / "ls netsec" (has args) -> falls through to the
      //   backend, unchanged, since that's your existing blog listing
      // - "cat <name>" -> checked against the vfs first; only if no
      //   matching file exists there does it fall through to the
      //   backend's blog-post `cat <slug>`
      // - "pwd" -> now vfs-only (backend's old cosmetic pwd is
      //   shadowed and can be removed from commands.go)
      // - "echo <text> > file" / ">> file" -> vfs write/append;
      //   plain "echo <text>" (no redirect) just prints the text
      // - "edit <file>" -> opens the line-mode editor (see editMode
      //   handling at the top of this listener)
      // -------------------------------------------------------------
      const parts = command.split(" ");
      const cmdName = parts[0];
      const cmdArgs = parts.slice(1);

      if (cmdName === "mkdir") {
        const out = vfsMkdir(cmdArgs.join(" "));
        if (out) term.writeln(out);
        writePrompt();
        return;
      }

      if (cmdName === "touch") {
        const out = vfsTouch(cmdArgs.join(" "));
        if (out) term.writeln(out);
        writePrompt();
        return;
      }

      if (cmdName === "cd") {
        const out = vfsCd(cmdArgs.join(" "));
        if (out) term.writeln(out);
        writePrompt();
        return;
      }

      if (cmdName === "rm") {
        const out = vfsRm(cmdArgs.join(" "));
        if (out) term.writeln(out);
        writePrompt();
        return;
      }

      if (cmdName === "ls" && cmdArgs.length === 0) {
        term.writeln(vfsLs());
        writePrompt();
        return;
      }

      if (cmdName === "pwd") {
        term.writeln(vfsPwdString());
        writePrompt();
        return;
      }

      if (cmdName === "echo") {
        const redirect = parseRedirect(command);
        if (redirect) {
          const out = vfsWriteFile(redirect.filename, redirect.text, redirect.append);
          if (out) term.writeln(out);
          writePrompt();
          return;
        }
        // plain echo, no redirect — just print the text back
        term.writeln(cmdArgs.join(" ").replace(/^["']|["']$/g, ""));
        writePrompt();
        return;
      }

      if (cmdName === "edit") {
        const filename = cmdArgs.join(" ");
        if (!filename) {
          term.writeln("usage: edit <file>");
          writePrompt();
          return;
        }
        const dir = resolveDir(vfs, cwd);
        if (dir.children[filename] && dir.children[filename].type === "dir") {
          term.writeln(`edit: ${filename}: is a directory`);
          writePrompt();
          return;
        }
        const existing = dir.children[filename];
        editMode = {
          filename,
          lines: existing && existing.content ? existing.content.split("\n") : [],
        };
        term.writeln(`-- editing ${filename} --`);
        term.writeln("type text one line at a time");
        term.writeln("finish with '.' on its own line");
        term.writeln("or press Ctrl+D");
        if (editMode.lines.length) {
          term.writeln("--- existing content (new lines will be appended) ---");
          editMode.lines.forEach(l => term.writeln(l));
          term.writeln("---");
        }
        term.write("\r\n> ");
        return;
      }

      if (cmdName === "cat" && cmdArgs.length > 0) {
        const vfsResult = vfsCatIfExists(cmdArgs.join(" "));
        if (vfsResult !== null) {
          term.writeln(vfsResult);
          writePrompt();
          return;
        }
        // no matching vfs file — fall through to backend blog cat below
      }

      busy = true;
      currentAbortController = new AbortController();
      const resumeWindow = isResumeOpenCommand(command)
        ? window.open("about:blank", "_blank")
        : null;

      try {
        const res = await fetch("/run", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ command }),
          signal: currentAbortController.signal
        });

        const text = await res.text();
        term.write(text.replace(/\r?\n/g, "\r\n"));

        const resumeURL = text.match(/^opening resume:\s*(https?:\/\/\S+)\s*$/i)?.[1];
        if (resumeWindow && resumeURL) {
          resumeWindow.location.href = resumeURL;
        } else if (isResumeOpenCommand(command) && resumeURL) {
          window.location.href = resumeURL;
        } else if (resumeWindow) {
          resumeWindow.close();
        }
      } catch (err) {
        if (resumeWindow) resumeWindow.close();
        if (err.name === "AbortError") {
          // cancelled via Ctrl+C — already handled above, nothing to print
        } else {
          term.write("error: could not reach server");
        }
      }

      busy = false;
      currentAbortController = null;
      writePrompt();
    } else if (char === "\b" || code === 127) {
      if (currentLine.length > 0) {
        currentLine = currentLine.slice(0, -1);
        term.write("\b \b");
      }
    } else if (code < 32) {
      // ignore other control characters (Tab, etc.)
    } else {
      currentLine += char;
      term.write(char);
    }

    i += 1;
  }
});