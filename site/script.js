// ── AgeForge site JS ─────────────────────────────────────────────────────────

// ── Starfield canvas (multi-layer parallax + nebulae) ────────────────────────
(function () {
  const canvas = document.getElementById("starfield");
  if (!canvas) return;
  const ctx = canvas.getContext("2d");

  // Track scroll for parallax
  let scrollY = 0;
  window.addEventListener("scroll", () => { scrollY = window.scrollY; }, { passive: true });

  // Three depth layers: [background, mid, foreground]
  const LAYER_DEFS = [
    { count: 160, rMin: 0.15, rMax: 0.65, sMin: 0.03, sMax: 0.12, parallax: 0.00, aMin: 0.15, aMax: 0.50 },
    { count:  70, rMin: 0.65, rMax: 1.35, sMin: 0.06, sMax: 0.25, parallax: 0.04, aMin: 0.30, aMax: 0.78 },
    { count:  22, rMin: 1.35, rMax: 2.40, sMin: 0.10, sMax: 0.38, parallax: 0.11, aMin: 0.55, aMax: 1.00 },
  ];

  // Nebula blobs — fixed viewport fractions, very subtle colour clouds
  const NEBULAE = [
    { fx: 0.12, fy: 0.22, fr: 0.32, r: 70, g: 35, b: 155 },   // indigo upper-left
    { fx: 0.82, fy: 0.10, fr: 0.26, r:  0, g: 75, b: 160 },   // cobalt upper-right
    { fx: 0.55, fy: 0.68, fr: 0.38, r: 150, g: 25, b: 80 },   // ruby lower-center
    { fx: 0.22, fy: 0.82, fr: 0.22, r: 15, g: 80, b: 120 },   // teal lower-left
  ];

  let layers = [];

  function mkStars(def) {
    return Array.from({ length: def.count }, () => ({
      x: Math.random() * canvas.width,
      y: Math.random() * canvas.height,
      r: def.rMin + Math.random() * (def.rMax - def.rMin),
      speed: def.sMin + Math.random() * (def.sMax - def.sMin),
      phase: Math.random() * Math.PI * 2,
    }));
  }

  function resize() {
    canvas.width  = window.innerWidth;
    canvas.height = window.innerHeight;
    layers = LAYER_DEFS.map(def => ({ ...def, stars: mkStars(def) }));
  }

  function draw() {
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    const t = Date.now() / 1000;

    // Nebulae (no parallax — ambient colour wash behind everything)
    for (const n of NEBULAE) {
      const cx = n.fx * canvas.width;
      const cy = n.fy * canvas.height;
      const r  = n.fr * Math.max(canvas.width, canvas.height);
      const g  = ctx.createRadialGradient(cx, cy, 0, cx, cy, r);
      g.addColorStop(0,   `rgba(${n.r},${n.g},${n.b},0.048)`);
      g.addColorStop(0.5, `rgba(${n.r},${n.g},${n.b},0.020)`);
      g.addColorStop(1,   `rgba(${n.r},${n.g},${n.b},0)`);
      ctx.fillStyle = g;
      ctx.beginPath();
      ctx.arc(cx, cy, r, 0, Math.PI * 2);
      ctx.fill();
    }

    // Star layers — foreground layers shift more with scroll (parallax depth)
    for (const layer of layers) {
      for (const s of layer.stars) {
        const sy  = ((s.y + scrollY * layer.parallax) % canvas.height + canvas.height) % canvas.height;
        const alpha = layer.aMin + (layer.aMax - layer.aMin) *
          Math.abs(Math.sin(t * s.speed + s.phase));

        // Halo glow for the largest (foreground) stars
        if (s.r > 1.3) {
          const halo = ctx.createRadialGradient(s.x, sy, 0, s.x, sy, s.r * 5);
          halo.addColorStop(0, `rgba(201,209,217,${(alpha * 0.35).toFixed(3)})`);
          halo.addColorStop(1, "rgba(201,209,217,0)");
          ctx.fillStyle = halo;
          ctx.beginPath();
          ctx.arc(s.x, sy, s.r * 5, 0, Math.PI * 2);
          ctx.fill();
        }

        ctx.beginPath();
        ctx.arc(s.x, sy, s.r, 0, Math.PI * 2);
        ctx.fillStyle = `rgba(201,209,217,${alpha.toFixed(3)})`;
        ctx.fill();
      }
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
    // Stagger the float animation so adjacent cards aren't in sync
    node.style.animationDelay = `${((i * 0.55) % 7).toFixed(2)}s`;
    node.innerHTML = `
      <div class="age-dot">${age.icon}</div>
      <div class="age-order">AGE ${String(i).padStart(2, '0')}</div>
      <div class="age-name">${age.name}</div>
      <div class="age-desc">${age.desc}</div>
    `;
    track.appendChild(node);
  });

  // ── Arrow navigation ────────────────────────────────────────────────────
  const agesScroll = document.getElementById('ages-scroll');
  const prevBtn    = document.getElementById('ages-prev');
  const nextBtn    = document.getElementById('ages-next');

  if (agesScroll && prevBtn && nextBtn) {
    let agesIdx = 0;

    function scrollToAge(idx) {
      const cards = track.querySelectorAll('.age-node');
      if (!cards.length) return;
      agesIdx = Math.max(0, Math.min(idx, cards.length - 1));
      agesScroll.scrollTo({ left: cards[agesIdx].offsetLeft, behavior: 'smooth' });
    }

    prevBtn.addEventListener('click', () => scrollToAge(agesIdx - 1));
    nextBtn.addEventListener('click', () => scrollToAge(agesIdx + 1));

    // Keep agesIdx in sync when user swipes/scrolls manually
    agesScroll.addEventListener('scrollend', () => {
      const cards = [...track.querySelectorAll('.age-node')];
      const mid = agesScroll.scrollLeft + agesScroll.offsetWidth / 2;
      let best = 0, bestDist = Infinity;
      cards.forEach((c, i) => {
        const d = Math.abs(c.offsetLeft + c.offsetWidth / 2 - mid);
        if (d < bestDist) { bestDist = d; best = i; }
      });
      agesIdx = best;
    }, { passive: true });
  }
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
