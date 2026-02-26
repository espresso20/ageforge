// ── AgeForge site JS ─────────────────────────────────────────────────────────

// ── Starfield canvas ─────────────────────────────────────────────────────────
(function () {
  const canvas = document.getElementById("starfield");
  if (!canvas) return;
  const ctx = canvas.getContext("2d");
  let stars = [];

  function resize() {
    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;
    stars = Array.from({ length: 180 }, () => ({
      x: Math.random() * canvas.width,
      y: Math.random() * canvas.height,
      r: Math.random() * 1.4 + 0.2,
      speed: Math.random() * 0.3 + 0.05,
      bright: Math.random(),
    }));
  }

  function draw() {
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    const t = Date.now() / 1000;
    for (const s of stars) {
      const alpha =
        0.35 + 0.65 * Math.abs(Math.sin(t * s.speed + s.bright * 6.28));
      ctx.beginPath();
      ctx.arc(s.x, s.y, s.r, 0, Math.PI * 2);
      ctx.fillStyle = `rgba(201,209,217,${alpha.toFixed(2)})`;
      ctx.fill();
    }
    requestAnimationFrame(draw);
  }

  resize();
  window.addEventListener("resize", resize);
  draw();
})();

// ── Nav scroll shadow ─────────────────────────────────────────────────────────
window.addEventListener(
  "scroll",
  () => {
    document
      .getElementById("nav")
      .classList.toggle("scrolled", window.scrollY > 10);
  },
  { passive: true },
);

// ── Intersection observer (fade-in) ──────────────────────────────────────────
const obs = new IntersectionObserver(
  (entries) => {
    for (const e of entries) {
      if (e.isIntersecting) {
        e.target.classList.add("visible");
        obs.unobserve(e.target);
      }
    }
  },
  { threshold: 0.1 },
);
document.querySelectorAll("[data-anim]").forEach((el) => obs.observe(el));

// ── Ages timeline ─────────────────────────────────────────────────────────────
const AGES = [
  { name: 'Primitive\nAge',     icon: '🪨', era: 'primitive',   desc: 'Survival. Nothing but your hands and wits.' },
  { name: 'Stone\nAge',         icon: '🪓', era: 'primitive',   desc: 'Crude tools of stone change everything.' },
  { name: 'Bronze\nAge',        icon: '🛡', era: 'ancient',     desc: 'Metalworking unlocks a new world of possibility.' },
  { name: 'Iron\nAge',          icon: '⚔️',  era: 'ancient',     desc: 'Iron tools and weapons transform society.' },
  { name: 'Classical\nAge',     icon: '🏛',  era: 'classical',   desc: 'Great empires rise. Philosophy and art flourish.' },
  { name: 'Medieval\nAge',      icon: '🏰', era: 'medieval',    desc: 'Kingdoms clash. Feudalism takes hold.' },
  { name: 'Renaissance\nAge',   icon: '🎨', era: 'renaissance', desc: 'Art, science, and exploration bloom.' },
  { name: 'Colonial\nAge',      icon: '⚓', era: 'renaissance', desc: 'Navies and colonies reshape the known world.' },
  { name: 'Industrial\nAge',    icon: '🏭', era: 'industrial',  desc: 'Steam power ignites exponential growth.' },
  { name: 'Victorian\nAge',     icon: '🎩', era: 'industrial',  desc: 'Empires at their absolute peak of confidence.' },
  { name: 'Electric\nAge',      icon: '⚡', era: 'electric',    desc: 'Electricity rewires every corner of civilization.' },
  { name: 'Atomic\nAge',        icon: '☢️',  era: 'atomic',      desc: 'The atom unlocks both limitless power and peril.' },
  { name: 'Modern\nAge',        icon: '🌐', era: 'atomic',      desc: 'Global infrastructure. Mass production. Superpowers.' },
  { name: 'Information\nAge',   icon: '📡', era: 'digital',     desc: 'Data becomes the most valuable resource on Earth.' },
  { name: 'Digital\nAge',       icon: '💻', era: 'digital',     desc: 'Code shapes reality. The physical world goes virtual.' },
  { name: 'Cyberpunk\nAge',     icon: '🤖', era: 'cyber',       desc: 'Megacorps rule. Augmented streets glow neon.' },
  { name: 'Fusion\nAge',        icon: '🔬', era: 'cyber',       desc: 'Clean unlimited energy. The energy crisis: solved.' },
  { name: 'Space\nAge',         icon: '🚀', era: 'cosmic',      desc: 'Humanity escapes the cradle and reaches for the stars.' },
  { name: 'Interstellar\nAge',  icon: '🛸', era: 'cosmic',      desc: 'Colony ships cross the void to distant star systems.' },
  { name: 'Galactic\nAge',      icon: '🌌', era: 'cosmic',      desc: 'An empire that spans hundreds of star systems.' },
  { name: 'Quantum\nAge',       icon: '⚛️',  era: 'cosmic',      desc: 'Reality itself becomes programmable.' },
  { name: 'Transcendent\nAge',  icon: '✨', era: 'cosmic',      desc: 'Beyond physical form. The final age of civilisation.' },
];

const track = document.getElementById('ages-track');
if (track) {
  AGES.forEach((age, i) => {
    const node = document.createElement('div');
    node.className = 'age-node';
    node.dataset.era = age.era;
    node.innerHTML = `
      <div class="age-dot">${age.icon}</div>
      <div class="age-order">AGE ${String(i).padStart(2, '0')}</div>
      <div class="age-name">${age.name}</div>
      <div class="age-desc">${age.desc}</div>
    `;
    track.appendChild(node);
  });
}

// ── Platform tabs ─────────────────────────────────────────────────────────────
document.querySelectorAll(".ptab").forEach((tab) => {
  tab.addEventListener("click", () => {
    const id = tab.dataset.tab;
    document
      .querySelectorAll(".ptab")
      .forEach((t) => t.classList.remove("active"));
    document
      .querySelectorAll(".pane")
      .forEach((p) => p.classList.remove("active"));
    tab.classList.add("active");
    const pane = document.getElementById("pane-" + id);
    if (pane) pane.classList.add("active");
  });
});

// ── Terminal mockup — static HTML lines, typewriter per line ─────────────────
const m = (s) => `<span class="tc-m">${s}</span>`; // muted
const g = (s) => `<span class="tc-g">${s}</span>`; // gold
const gr = (s) => `<span class="tc-gr">${s}</span>`; // green
const rd = (s) => `<span class="tc-rd">${s}</span>`; // red
const cy = (s) => `<span class="tc-cy">${s}</span>`; // cyan
const pu = (s) => `<span class="tc-pu">${s}</span>`; // purple
const yw = (s) => `<span class="tc-yw">${s}</span>`; // yellow

// Fixed-width helpers (no ANSI, just spaces)
const pad = (s, n) => s.padEnd(n);
const lpad = (s, n) => s.padStart(n);
const bar = (n, t) => g("█".repeat(n)) + m("░".repeat(t - n));

const TERM_LINES = [
  // ── row 1: status bar
  g("  🏛 Stone Age") +
    m("  ") +
    m('"Founder"') +
    "                      " +
    m("Tick: ") +
    g("1,247") +
    m("  Pop: ") +
    "18/30" +
    m("  ×1  F1-F7  Esc=Save"),

  // ── row 2: age progress
  m(
    "  ─────────────────────────────────────────────────────────────────────────────────────────────",
  ),

  // ── row 3: next age bar
  g("  Next: Bronze Age") +
    "  " +
    rd("food:3,102/8,000") +
    " " +
    bar(4, 7) +
    "  " +
    rd("stone:890/4,000") +
    " " +
    bar(2, 7) +
    "  " +
    gr("wood:1,240/8,000") +
    " " +
    bar(5, 7),

  // ── row 4: tab bar
  m("  F1:") +
    g("Economy") +
    m("  F2:Research  F3:Military  F4:Trade  F5:Stats  F6:Wonders  F7:Logs"),

  // ── row 5: divider
  m(
    "  ─────────────────────────────────────────────────────────────────────────────────────────────",
  ),

  // ── row 6: column headers
  m("  ──────────────────── Resources ─────────────────────") +
    m("│") +
    m("──────────────── Buildings ──────────────"),

  // ── row 7-10: resources left, buildings right
  "  " +
    cy(pad("food", 10)) +
    lpad("3,102", 6) +
    m("/") +
    pad("8,000", 6) +
    " " +
    bar(4, 6) +
    " " +
    gr(lpad("+8", 4)) +
    m("/t") +
    m("               │") +
    "  " +
    m("── Housing ──"),

  "  " +
    cy(pad("wood", 10)) +
    lpad("1,240", 6) +
    m("/") +
    pad("6,000", 6) +
    " " +
    bar(3, 6) +
    " " +
    gr(lpad("+12", 4)) +
    m("/t") +
    m("               │") +
    "  " +
    gr("✓") +
    " " +
    g(pad("Hut", 16)) +
    m("[12]  wood:8"),

  "  " +
    cy(pad("stone", 10)) +
    lpad("  890", 6) +
    m("/") +
    pad("4,000", 6) +
    " " +
    bar(2, 6) +
    " " +
    gr(lpad("+4", 4)) +
    m("/t") +
    m("               │") +
    "  " +
    m("·") +
    " " +
    m(pad("Stash", 16)) +
    m("[ 5]  wood:45"),

  "  " +
    cy(pad("knowledge", 10)) +
    lpad("  450", 6) +
    m("/") +
    pad("2,000", 6) +
    " " +
    bar(2, 6) +
    " " +
    gr(lpad("+3", 4)) +
    m("/t") +
    m("               │") +
    "  " +
    m("── Production ──"),

  // ── row 11: villager header / building row
  m("  ──────────────────── Villagers ─────────────────────") +
    m("│") +
    "  " +
    gr("✓") +
    " " +
    g(pad("Gathering Camp", 16)) +
    m("[ 3]  wood:200"),

  // ── row 12-13: villagers / buildings
  "  " +
    m("Pop: ") +
    "18" +
    m("/30  Idle: ") +
    g("2") +
    m("  Food: ") +
    rd("-4/t") +
    "    " +
    m("                 │") +
    "  " +
    gr("✓") +
    " " +
    g(pad("Woodcutter Camp", 16)) +
    m("[ 2]  wood:350"),

  "  " +
    pu(pad("worker", 10)) +
    m("×12 ") +
    g("(2 idle)") +
    m("  food:8") +
    "      " +
    m("                │") +
    "  " +
    m("·") +
    " " +
    m(pad("Stone Pit", 16)) +
    m("[ 0]  stone:15"),

  "  " +
    pu(pad("shaman", 10)) +
    m("× 6") +
    m("  knowledge:3") +
    "          " +
    m("                │"),

  // ── row 14: divider + logs
  m(
    "  ─────────────────────────────────────────────────────────────────────────────────────────────",
  ),
  "  " + m("[ 1242] ") + gr("✓ Built: Gathering Camp"),
  "  " + m("[ 1244] ") + yw("★ Tribe reaches 15 members — morale rises!"),
  "  " + m("[ 1247] ") + m("· wood +12/t  |  food +8/t  |  knowledge +3/t"),
  m(
    "  ─────────────────────────────────────────────────────────────────────────────────────────────",
  ),
  // ── last row: prompt
  g("  > "),
];

const termPre = document.getElementById("term-pre");
const termCursor = document.getElementById("term-cursor");

if (termPre) {
  // Add terminal colour classes to the stylesheet dynamically
  const style = document.createElement("style");
  style.textContent = `
    .tc-m  { color: #8b949e; }
    .tc-g  { color: #f0a500; }
    .tc-gr { color: #3fb950; }
    .tc-rd { color: #f85149; }
    .tc-cy { color: #39d0c8; }
    .tc-pu { color: #d2a8ff; }
    .tc-yw { color: #e3c07b; }
  `;
  document.head.appendChild(style);

  // Reveal lines one by one with a short delay between each
  let lineIdx = 0;
  function revealNextLine() {
    if (lineIdx >= TERM_LINES.length) {
      // Show cursor after last line
      if (termCursor) termCursor.style.display = "inline";
      return;
    }
    const div = document.createElement("div");
    div.innerHTML = TERM_LINES[lineIdx];
    termPre.appendChild(div);
    lineIdx++;
    setTimeout(revealNextLine, lineIdx < 5 ? 60 : 30);
  }
  if (termCursor) termCursor.style.display = "none";
  setTimeout(revealNextLine, 500);
}

// ── Scroll animations fallback ────────────────────────────────────────────────
// Stagger feat-cards
document.querySelectorAll(".feat-card").forEach((card, i) => {
  card.style.transitionDelay = i * 80 + "ms";
  obs.observe(card);
});
document.querySelectorAll(".step").forEach((step, i) => {
  step.style.transitionDelay = i * 60 + "ms";
  obs.observe(step);
});
