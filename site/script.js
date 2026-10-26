// The front page: the light and dark switch, the install box and the plate. No
// libraries. The map frames come from frames.js, written by TestWriteSiteFrames.
(() => {
"use strict";
const $ = id => document.getElementById(id);
const reduce = !!(window.matchMedia && matchMedia("(prefers-reduced-motion: reduce)").matches);

// ---- light and dark: the system's choice until the reader makes one ----------
// theme.js has already put a saved choice on <html> before the first paint.
const root = document.documentElement;
const dark = window.matchMedia ? matchMedia("(prefers-color-scheme: dark)") : null;
const metas = Array.from(document.querySelectorAll('meta[name="theme-color"]'));
const current = () => {
  const t = root.getAttribute("data-theme");
  return t === "dark" || t === "light" ? t : dark && dark.matches ? "dark" : "light";
};
function syncTheme() {
  const night = current() === "dark", b = $("theme");
  b.textContent = night ? "Day" : "Night";
  b.setAttribute("aria-pressed", String(night));
  b.setAttribute("aria-label", night ? "Switch to light mode" : "Switch to dark mode");
  metas.forEach(m => m.setAttribute("content", night ? "#121822" : "#e8dcc0"));
}
$("theme").hidden = false;
$("theme").addEventListener("click", () => {
  const next = current() === "dark" ? "light" : "dark";
  root.setAttribute("data-theme", next);
  try { localStorage.setItem("ageforge-theme", next); } catch (e) { /* private window: the choice lasts until the page closes */ }
  syncTheme();
});
if (dark && dark.addEventListener) dark.addEventListener("change", syncTheme);
syncTheme();

// ---- the install box ---------------------------------------------------------
const BASE = "https://github.com/espresso20/ageforge/releases/latest/download/";
const OS = {
  mac: ["curl -fLo ageforge " + BASE + "ageforge-macos-arm64 && chmod +x ageforge && ./ageforge", "For Apple Silicon. On an Intel Mac, use ageforge-macos-amd64 instead."],
  linux: ["curl -fLo ageforge " + BASE + "ageforge-linux-amd64 && chmod +x ageforge && ./ageforge", "For x86_64. On ARM, use ageforge-linux-arm64 instead."],
  win: ["iwr " + BASE + "ageforge-windows-amd64.exe -OutFile ageforge.exe; .\\ageforge.exe", "Paste it into PowerShell in Windows Terminal."],
};
let os = /Win/.test(navigator.platform) ? "win" : /Linux/.test(navigator.platform) ? "linux" : "mac";
function setOS(k) {
  os = k;
  $("cmd").textContent = OS[k][0];
  $("note").textContent = OS[k][1];
  document.querySelectorAll(".inst-os button").forEach(b => b.setAttribute("aria-pressed", String(b.dataset.os === k)));
}
document.querySelectorAll(".inst-os button").forEach(b => b.addEventListener("click", () => setOS(b.dataset.os)));
setOS(os);
$("copy").addEventListener("click", () => {
  const done = t => { $("copy").textContent = t; $("status").textContent = t; setTimeout(() => { $("copy").textContent = "Copy"; $("status").textContent = ""; }, 1500); };
  const pick = () => { const r = document.createRange(); r.selectNodeContents($("cmd")); const s = getSelection(); s.removeAllRanges(); s.addRange(r); done("Selected"); };
  try { navigator.clipboard.writeText(OS[os][0]).then(() => done("Copied"), pick); } catch (e) { pick(); }
});

// ---- the plate: one ink as engraved, the game's own colors as played ----------
const plate = $("plate"), btn = $("toggle");
const AGE = Number(plate.dataset.age);
const CROP = [0, 1, 94, 32];   // columns 0-93, rows 1-31: the map without its legend and status lines
let F = null;
const cache = [];
const esc = c => c === "&" ? "&amp;" : c === "<" ? "&lt;" : c === ">" ? "&gt;" : c;
function lum(hex) {
  const n = parseInt(hex.slice(1), 16), r = (n >> 16) & 255, g = (n >> 8) & 255, b = n & 255;
  return (0.299 * r + 0.587 * g + 0.114 * b) / 255;
}
function styles() {
  // a brighter color on the terminal prints as heavier ink
  const css = F.styles.map((s, i) => {
    const o = Math.max(0.34, Math.min(1, 0.2 + lum(F.palette[s[0]]) * 0.95));
    return ".s" + i + "{--o:" + o.toFixed(2) + ";--c:" + F.palette[s[0]] + (F.palette[s[1]] !== F.bg ? ";--g:" + F.palette[s[1]] : "") + "}";
  }).join("");
  const el = document.createElement("style"); el.textContent = css; document.head.appendChild(el);
}
function rowHTML(row) {
  const chars = Array.from(row.t), r = row.r;
  let out = "", x = 0;
  for (let i = 0; i + 1 < r.length; i += 2) {
    const s = r[i], n = r[i + 1], a = Math.max(x, CROP[0]), b = Math.min(x + n, CROP[2]);
    if (b > a) {
      let run = "";
      for (let k = a; k < b; k++) { const ch = chars[k] === undefined ? " " : chars[k]; run += ch.codePointAt(0) > 127 ? "<i>" + ch + "</i>" : esc(ch); }
      out += '<span class="s' + s + (F.styles[s][2] & 1 ? " bold" : "") + '">' + run + "</span>";
    }
    x += n;
  }
  return '<span class="row">' + out + "</span>";
}
function frameHTML(f) {
  const frames = F.ages[AGE].frames, n = f % frames.length;
  if (cache[n] === undefined) {
    let h = "";
    for (let y = CROP[1]; y < CROP[3]; y++) h += rowHTML(frames[n][y] || frames[0][y]);
    cache[n] = h;
  }
  return cache[n];
}
let frame = 0;
const paint = () => { plate.querySelector(".map code").innerHTML = frameHTML(frame); };
Promise.resolve(window.AGEFORGE_FRAMES).then(d => {
  if (!d) throw new Error("no frames");
  F = d;
  styles();
  const old = plate.querySelector(".pic style"); if (old) old.remove();   // the static frame's own colors, no longer needed
  paint();
  btn.hidden = false;
  btn.addEventListener("click", () => {
    const on = btn.getAttribute("aria-pressed") !== "true";
    btn.setAttribute("aria-pressed", String(on));
    btn.textContent = on ? "Show in ink" : "Show in color";
    plate.classList.toggle("ink", !on);
  });
  // the map moves, as it does in the game. Reduced motion holds it still.
  let seen = true;
  if ("IntersectionObserver" in window) new IntersectionObserver(es => { seen = es[es.length - 1].isIntersecting; }).observe(plate);
  if (!reduce) setInterval(() => { if (!document.hidden && seen) { frame++; paint(); } }, 520);
}).catch(() => { /* the static first frame stays */ });
})();
