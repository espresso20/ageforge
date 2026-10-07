// The almanac's plates and its install card. No libraries. The map frames come
// from frames.js, written by TestWriteSiteFrames.
(() => {
"use strict";
const $ = id => document.getElementById(id);
const reduce = !!(window.matchMedia && matchMedia("(prefers-reduced-motion: reduce)").matches);
const BASE = "https://github.com/espresso20/ageforge/releases/latest/download/";
const OS = {
  mac: ["curl -fLo ageforge " + BASE + "ageforge-macos-arm64 && chmod +x ageforge && ./ageforge", "Apple Silicon. On an Intel Mac, put ageforge-macos-amd64 in its place."],
  linux: ["curl -fLo ageforge " + BASE + "ageforge-linux-amd64 && chmod +x ageforge && ./ageforge", "For x86_64. On ARM, put ageforge-linux-arm64 in its place."],
  win: ["iwr " + BASE + "ageforge-windows-amd64.exe -OutFile ageforge.exe; .\\ageforge.exe", "In PowerShell, in Windows Terminal."],
};
function setOS(k) {
  $("cmd").textContent = OS[k][0];
  $("note").textContent = OS[k][1];
  document.querySelectorAll(".os button").forEach(b => b.setAttribute("aria-selected", String(b.dataset.os === k)));
}
document.querySelectorAll(".os button").forEach(b => b.addEventListener("click", () => setOS(b.dataset.os)));
setOS(/Win/.test(navigator.platform) ? "win" : /Linux/.test(navigator.platform) ? "linux" : "mac");
$("copy").addEventListener("click", () => {
  const done = () => { $("copy").textContent = "Copied"; setTimeout(() => { $("copy").textContent = "Copy the line"; }, 1600); };
  const pick = () => { const r = document.createRange(); r.selectNodeContents($("cmd")); const s = getSelection(); s.removeAllRanges(); s.addRange(r); $("copy").textContent = "Selected"; };
  try { navigator.clipboard.writeText($("cmd").textContent).then(done, pick); } catch (e) { pick(); }
});

// the plates: one ink as engraved, the game's own colors as played
const esc = c => c === "&" ? "&amp;" : c === "<" ? "&lt;" : c === ">" ? "&gt;" : c;
const plates = Array.from(document.querySelectorAll(".plate"));
let F = null, html = [];
function lum(hex) {
  const n = parseInt(hex.slice(1), 16), r = (n >> 16) & 255, g = (n >> 8) & 255, b = n & 255;
  return (0.299 * r + 0.587 * g + 0.114 * b) / 255;
}
function build() {
  // a brighter color on the terminal prints as heavier ink
  const css = F.styles.map((s, i) => {
    const o = Math.max(0.34, Math.min(1, 0.2 + lum(F.palette[s[0]]) * 0.95));
    return ".s" + i + "{--o:" + o.toFixed(2) + ";--c:" + F.palette[s[0]] + (F.palette[s[1]] !== F.bg ? ";--g:" + F.palette[s[1]] : "") + "}";
  }).join("");
  const st = document.createElement("style"); st.textContent = css; document.head.appendChild(st);
  html = F.ages.map(a => a.frames.map(fr => fr.map((row, y) => {
    const r = row || a.frames[0][y];
    const chars = Array.from(r.t);
    let out = '<span class="row">', x = 0;
    for (let i = 0; i < r.r.length; i += 2) {
      const s = r.r[i], n = r.r[i + 1];
      let run = "";
      for (const ch of chars.slice(x, x + n)) run += ch.charCodeAt(0) > 127 ? "<b>" + ch + "</b>" : esc(ch);
      out += '<span class="s' + s + (F.styles[s][2] ? " bold" : "") + '">' + run + "</span>";
      x += n;
    }
    return out + "</span>";
  }).join("")));
}
function fit(p) {
  const frame = p.querySelector(".plate-frame"), map = p.querySelector(".map");
  const fs = Math.max(6.2, (frame.clientWidth - 6) / (F.w * 0.6 + 1.4));   // 0.6em a cell, plus the plate's own padding
  map.style.setProperty("--fs", fs.toFixed(2) + "px");
}
function paint(p) { p.querySelector(".map").innerHTML = html[p._age][p._frame % html[p._age].length]; }
Promise.resolve(window.AGEFORGE_FRAMES).then(d => {
  if (!d) throw new Error("no frames");
  F = d;
  build();
  for (const p of plates) {
    p._age = Number(p.dataset.age); p._frame = 0;
    fit(p); paint(p);
    const btn = p.querySelector(".toggle");
    btn.addEventListener("click", () => {
      const on = btn.getAttribute("aria-pressed") !== "true";
      btn.setAttribute("aria-pressed", String(on));
      btn.textContent = on ? "Show as engraved" : "Show as played";
      p.classList.toggle("played", on);
      p._frame = 0; paint(p);
    });
  }
  if ("ResizeObserver" in window) { const ro = new ResizeObserver(() => plates.forEach(fit)); plates.forEach(p => ro.observe(p)); }
  else addEventListener("resize", () => plates.forEach(fit));
  // a played plate moves, as the game does. An engraved one holds still.
  if (!reduce) setInterval(() => {
    if (document.hidden) return;
    for (const p of plates) if (p.classList.contains("played")) { p._frame++; paint(p); }
  }, 520);
}).catch(() => {
  for (const p of plates) { const m = p.querySelector(".map"); if (!m.textContent.trim()) m.textContent = "  The plate was damaged at the printer's."; }
});
})();
