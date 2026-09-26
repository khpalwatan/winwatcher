// winwatcher frontend.

import {
  GetTree,
  SetSelectedPID,
  GetDetails,
  OpenURL,
  KillProcess,
  GetVersion,
} from "../wailsjs/go/main/App";
import { EventsOn } from "../wailsjs/runtime/runtime";

// Alert thresholds. A row lights up red when it crosses either.
const ALERT_CPU_PERCENT = 25;
const ALERT_MEMORY_BYTES = 500 * 1024 * 1024; // 500 MB

const THEMES = ["dark", "amoled", "white"];
const THEME_ICONS = { dark: "◐", amoled: "●", white: "○" };

// Graph colors per theme. Darker tones for white mode so they stay bold.
const GRAPH_COLORS = {
  dark:   { cpu: "#22d3ee", mem: "#d0bcff", grid: "rgba(128,128,128,0.12)" },
  amoled: { cpu: "#22d3ee", mem: "#d0bcff", grid: "rgba(128,128,128,0.10)" },
  white:  { cpu: "#0891b2", mem: "#ca8a04", grid: "rgba(0,0,0,0.10)" },
};

let nodes = [];
let selectedPid = 0;
let filterText = "";
let lastGraphSamples = [];

const treeEl = document.getElementById("tree");
const detailsEl = document.getElementById("details");
const detailsEmptyEl = document.getElementById("detailsEmpty");
const filterEl = document.getElementById("filter");
const procCountEl = document.getElementById("procCount");
const detailNameEl = document.getElementById("detailName");
const detailPidEl = document.getElementById("detailPid");
const statCpuEl = document.getElementById("statCpu");
const statMemEl = document.getElementById("statMem");
const statThreadsEl = document.getElementById("statThreads");
const detailCmdEl = document.getElementById("detailCmd");
const netListEl = document.getElementById("netList");
const fileListEl = document.getElementById("fileList");
const netCountEl = document.getElementById("netCount");
const fileCountEl = document.getElementById("fileCount");
const githubBtnEl = document.getElementById("githubBtn");
const themeBtnEl = document.getElementById("themeBtn");
const themeIconEl = document.getElementById("themeIcon");
const killBtnEl = document.getElementById("killBtn");
const versionTagEl = document.getElementById("versionTag");
const cpuGraphEl = document.getElementById("cpuGraph");
const memGraphEl = document.getElementById("memGraph");
const cpuValueEl = document.getElementById("cpuValue");
const memValueEl = document.getElementById("memValue");

function fmtBytes(n) {
  const kb = 1024, mb = kb * 1024, gb = mb * 1024;
  if (n >= gb) return (n / gb).toFixed(1) + " GB";
  if (n >= mb) return (n / mb).toFixed(1) + " MB";
  if (n >= kb) return Math.round(n / kb) + " KB";
  return n + " B";
}

function fmtCpu(p) {
  return p < 0.05 ? "0.0%" : p.toFixed(1) + "%";
}

function esc(s) {
  return String(s)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;");
}

// ---------- themes ----------

function currentTheme() {
  return localStorage.getItem("theme") || "dark";
}

function applyTheme(name) {
  document.body.classList.remove("theme-amoled", "theme-white");
  if (name === "amoled") document.body.classList.add("theme-amoled");
  else if (name === "white") document.body.classList.add("theme-white");

  themeIconEl.textContent = THEME_ICONS[name] || "◐";
  localStorage.setItem("theme", name);

  // Repaint graphs with the new palette.
  if (lastGraphSamples.length) {
    renderGraph(lastGraphSamples);
  }
}

function cycleTheme() {
  const idx = THEMES.indexOf(currentTheme());
  applyTheme(THEMES[(idx + 1) % THEMES.length]);
}

// ---------- tree rendering ----------

function isAlert(n) {
  return n.cpuPercent >= ALERT_CPU_PERCENT || n.memoryRss >= ALERT_MEMORY_BYTES;
}

function renderTree() {
  const q = filterText.trim().toLowerCase();
  const frag = document.createDocumentFragment();

  for (const n of nodes) {
    if (q && !n.name.toLowerCase().includes(q) && !String(n.pid).includes(q)) {
      continue;
    }

    const classes = ["row"];
    if (n.pid === selectedPid) classes.push("selected");
    if (n.tag === "system") classes.push("system");
    if (isAlert(n)) classes.push("alert");

    const row = document.createElement("div");
    row.className = classes.join(" ");
    row.dataset.pid = n.pid;

    row.innerHTML =
      '<span class="name">' +
        '<span class="indent" style="width:' + (n.depth * 12) + 'px"></span>' +
        esc(n.name) +
      "</span>" +
      '<span class="tag ' + esc(n.tag) + '">' + esc(n.tag) + "</span>" +
      '<span class="pid">' + n.pid + "</span>" +
      '<span class="cpu' + (n.cpuPercent >= 5 ? " hot" : "") + '">' + fmtCpu(n.cpuPercent) + "</span>" +
      '<span class="mem">' + fmtBytes(n.memoryRss) + "</span>";

    frag.appendChild(row);
  }

  treeEl.replaceChildren(frag);
  procCountEl.textContent = String(nodes.length);
}

// ---------- details rendering ----------

function renderDetails(d) {
  if (!d || !d.pid) {
    detailsEl.hidden = true;
    detailsEmptyEl.hidden = false;
    return;
  }

  detailsEl.hidden = false;
  detailsEmptyEl.hidden = true;

  detailNameEl.textContent = nameForPid(d.pid);
  detailPidEl.textContent = "PID " + d.pid;
  detailCmdEl.textContent = d.commandLine || "(unavailable)";
  statThreadsEl.textContent = d.threads > 0 ? String(d.threads) : "-";

  const node = nodes.find((n) => n.pid === d.pid);
  statCpuEl.textContent = node ? fmtCpu(node.cpuPercent) : "-";
  statMemEl.textContent = node ? fmtBytes(node.memoryRss) : "-";

  netCountEl.textContent = String(d.net ? d.net.length : 0);
  if (d.net && d.net.length) {
    const frag = document.createDocumentFragment();
    for (const c of d.net) {
      const stateClass =
        c.state === "ESTABLISHED" ? "established" :
        c.state === "LISTEN" ? "listen" : "";
      const item = document.createElement("div");
      item.className = "listItem";
      item.innerHTML =
        '<span class="state ' + stateClass + '">' + esc(c.state) + "</span>" +
        esc(c.protocol + " " + c.localAddr + " -> " + (c.remoteAddr || "*"));
      frag.appendChild(item);
    }
    netListEl.replaceChildren(frag);
  } else {
    netListEl.replaceChildren();
  }

  fileCountEl.textContent = String(d.files ? d.files.length : 0);
  if (d.filesErr) {
    const note = document.createElement("div");
    note.className = "errorNote";
    note.textContent = d.filesErr;
    fileListEl.replaceChildren(note);
  } else if (d.files && d.files.length) {
    const frag = document.createDocumentFragment();
    for (const f of d.files) {
      const item = document.createElement("div");
      item.className = "listItem";
      item.textContent = f.path;
      frag.appendChild(item);
    }
    fileListEl.replaceChildren(frag);
  } else {
    fileListEl.replaceChildren();
  }
}

function nameForPid(pid) {
  const n = nodes.find((x) => x.pid === pid);
  return n ? n.name : "PID " + pid;
}

// ---------- selection ----------

async function select(pid) {
  if (pid === selectedPid) return;
  selectedPid = pid;
  await SetSelectedPID(pid);
  renderTree();
  const d = await GetDetails(pid);
  renderDetails(d);
}

// ---------- graph ----------

function drawGraph(canvas, samples, getValue, maxValue, color, grid) {
  const dpr = window.devicePixelRatio || 1;
  const rect = canvas.getBoundingClientRect();
  const w = rect.width;
  const h = rect.height;

  if (canvas.width !== w * dpr || canvas.height !== h * dpr) {
    canvas.width = w * dpr;
    canvas.height = h * dpr;
  }

  const ctx = canvas.getContext("2d");
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
  ctx.clearRect(0, 0, w, h);

  ctx.strokeStyle = grid;
  ctx.lineWidth = 1;
  for (let i = 1; i < 4; i++) {
    const y = (h / 4) * i;
    ctx.beginPath();
    ctx.moveTo(0, y);
    ctx.lineTo(w, y);
    ctx.stroke();
  }

  if (!samples || samples.length < 2) return;

  const n = samples.length;
  const step = w / (n - 1);
  const max = maxValue > 0 ? maxValue : 1;

  ctx.beginPath();
  ctx.moveTo(0, h);
  for (let i = 0; i < n; i++) {
    const v = Math.min(getValue(samples[i]), max);
    const y = h - (v / max) * (h - 2) - 1;
    ctx.lineTo(i * step, y);
  }
  ctx.lineTo((n - 1) * step, h);
  ctx.closePath();

  const grad = ctx.createLinearGradient(0, 0, 0, h);
  grad.addColorStop(0, color + "66");
  grad.addColorStop(1, color + "08");
  ctx.fillStyle = grad;
  ctx.fill();

  ctx.beginPath();
  for (let i = 0; i < n; i++) {
    const v = Math.min(getValue(samples[i]), max);
    const y = h - (v / max) * (h - 2) - 1;
    if (i === 0) ctx.moveTo(0, y);
    else ctx.lineTo(i * step, y);
  }
  ctx.strokeStyle = color;
  ctx.lineWidth = 1.8;
  ctx.stroke();
}

function renderGraph(samples) {
  if (!samples || samples.length === 0) return;

  lastGraphSamples = samples;

  const palette = GRAPH_COLORS[currentTheme()] || GRAPH_COLORS.dark;
  const last = samples[samples.length - 1];
  cpuValueEl.textContent = fmtCpu(last.cpu || 0);
  memValueEl.textContent = fmtBytes(last.memory || 0);

  let maxCpu = 100;
  let maxMem = 0;
  for (const s of samples) {
    if (s.cpu > maxCpu) maxCpu = s.cpu;
    if (s.memory > maxMem) maxMem = s.memory;
  }
  maxMem = maxMem * 1.1 || 1;

  drawGraph(cpuGraphEl, samples, (s) => s.cpu, maxCpu, palette.cpu, palette.grid);
  drawGraph(memGraphEl, samples, (s) => s.memory, maxMem, palette.mem, palette.grid);
}

// ---------- events ----------

treeEl.addEventListener("click", (e) => {
  const row = e.target.closest(".row");
  if (!row) return;
  select(Number(row.dataset.pid));
});

filterEl.addEventListener("input", () => {
  filterText = filterEl.value;
  renderTree();
});

githubBtnEl.addEventListener("click", () => {
  OpenURL("https://github.com/khpalwatan");
});

themeBtnEl.addEventListener("click", cycleTheme);

killBtnEl.addEventListener("click", async () => {
  if (!selectedPid) return;
  const node = nodes.find((n) => n.pid === selectedPid);
  const name = node ? node.name : "PID " + selectedPid;
  const ok = window.confirm("Kill " + name + " (PID " + selectedPid + ")?\n\nThis cannot be undone.");
  if (!ok) return;

  const result = await KillProcess(selectedPid);
  if (result && !result.ok) {
    window.alert(result.message || "Failed to kill process.");
  } else {
    selectedPid = 0;
    await SetSelectedPID(0);
    detailsEl.hidden = true;
    detailsEmptyEl.hidden = false;
    renderTree();
  }
});

document.addEventListener("keydown", (e) => {
  if (e.key === "F11") {
    e.preventDefault();
    if (!document.fullscreenElement) {
      document.documentElement.requestFullscreen();
    } else {
      document.exitFullscreen();
    }
    return;
  }
  if (e.key === "Escape" && document.fullscreenElement) {
    document.exitFullscreen();
    return;
  }
  if (e.key === "/" && document.activeElement !== filterEl) {
    e.preventDefault();
    filterEl.focus();
    filterEl.select();
    return;
  }
  if (e.key === "Escape" && document.activeElement === filterEl) {
    filterEl.value = "";
    filterText = "";
    filterEl.blur();
    renderTree();
    return;
  }
  if (e.key === "t" && document.activeElement !== filterEl) {
    cycleTheme();
    return;
  }
});

EventsOn("proc:tree", (payload) => {
  nodes = payload || [];
  renderTree();
});

EventsOn("proc:details", (payload) => {
  if (payload && payload.pid === selectedPid) {
    renderDetails(payload);
  }
});

EventsOn("graph:update", (payload) => {
  renderGraph(payload || []);
});

// ---------- init ----------

(async function init() {
  applyTheme(currentTheme());

  try {
    const v = await GetVersion();
    if (v) versionTagEl.textContent = "v" + v;
  } catch (_) {}

  try {
    const tree = await GetTree();
    nodes = tree || [];
    renderTree();
  } catch (err) {
    console.error("GetTree failed:", err);
  }
})();