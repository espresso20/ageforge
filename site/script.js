// AgeForge landing page: the scroll-through-time map and the try-it prompt.
// No libraries. The map frames come from frames.js, written by
// TestWriteSiteFrames (ui/mapstyle/all/site_frames_test.go) from real runs.

(function () {
  "use strict";

  const root = document.documentElement;
  root.classList.add("js");
  const reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)");

  // ── The map ────────────────────────────────────────────────────────────

  const F = window.AGEFORGE_FRAMES;
  const timeline = document.getElementById("timeline");
  const stage = timeline && timeline.querySelector(".stage");
  const pane = document.getElementById("pane");
  const screen = document.getElementById("screen");
  const map = document.getElementById("map");
  const hudAge = document.getElementById("hud-age");
  const hudN = document.getElementById("hud-n");
  const ticks = document.getElementById("ticks");
  const hero = document.getElementById("hero");
  const live = document.getElementById("age-live");

  // What each frame shows, for the pane's text alternative. Only the ages
  // the page shows are described: the rest are left for players to find.
  const ALT = {
    primitive_age: "a few huts and a campfire in a forest by a river",
    victorian_age: "a railway running through a dense walled city by the water",
    cyberpunk_age: "a neon megacity of megablocks and sky rails",
    galactic_age: "a starbase among the civilizations of the galaxy",
  };

  if (F && map && timeline) {
    initMap();
  } else {
    // No frames: the static first frame stays, and the scroll stays short.
    root.classList.remove("js");
  }

  function initMap() {
    const ages = F.ages;
    const N = ages.length;
    const W = F.w;
    const H = F.h;
    const code = map.querySelector("code") || map;

    // One rule per style, so a frame is only text and class names.
    const css = F.styles
      .map(function (s, i) {
        return (
          ".map .s" + i + "{color:" + F.palette[s[0]] + ";background:" + F.palette[s[1]] +
          (s[2] ? ";font-weight:700" : "") + "}"
        );
      })
      .join("");
    const style = document.createElement("style");
    style.textContent = css;
    document.head.appendChild(style);

    // How far the scroll runs, in steps (--step in style.css). The first
    // frame gives way after half a step, so the scroll soon shows what it
    // does; each frame after it holds a whole step; the last holds a little
    // less before the page moves on.
    const FIRST_HOLD = 0.5;
    const LAST_HOLD = 0.7;
    const STEPS = FIRST_HOLD + Math.max(0, N - 2) + LAST_HOLD;
    timeline.style.setProperty("--steps", String(STEPS));

    function esc(c) {
      if (c === "&") return "&amp;";
      if (c === "<") return "&lt;";
      if (c === ">") return "&gt;";
      const o = c.codePointAt(0);
      if (o > 127 && (o < 0x2500 || o > 0x259f)) return "<i>" + c + "</i>";
      return c;
    }

    // A frame as cells: per row, its characters and each cell's style. A
    // later frame's unchanged rows point at the first frame's.
    const cellCache = [];
    function frameCells(a, f) {
      if (!cellCache[a]) cellCache[a] = [];
      if (!cellCache[a][f]) {
        cellCache[a][f] = ages[a].frames[f].map(function (r, y) {
          if (!r) return frameCells(a, 0)[y];
          const c = Array.from(r.t);
          const st = new Array(c.length);
          let x = 0;
          for (let k = 0; k + 1 < r.r.length; k += 2) {
            for (let j = 0; j < r.r[k + 1]; j++) st[x++] = r.r[k];
          }
          return { c: c, s: st };
        });
      }
      return cellCache[a][f];
    }

    // One row of cells as HTML: a span per run of one style (none for the
    // plain style).
    function cellsHTML(c, st) {
      let out = "";
      let run = "";
      let cur = -1;
      for (let x = 0; x < c.length; x++) {
        if (st[x] !== cur) {
          if (run) out += cur === 0 ? run : '<span class="s' + cur + '">' + run + "</span>";
          run = "";
          cur = st[x];
        }
        run += esc(c[x]);
      }
      if (run) out += cur === 0 ? run : '<span class="s' + cur + '">' + run + "</span>";
      return out;
    }

    // Rows of every frame as HTML, rendered once on first use.
    const cache = [];
    function frameRows(a, f) {
      if (!cache[a]) cache[a] = [];
      if (!cache[a][f]) {
        cache[a][f] = frameCells(a, f).map(function (row) {
          return cellsHTML(row.c, row.s);
        });
      }
      return cache[a][f];
    }

    // The rows on screen.
    code.innerHTML = "";
    const rows = [];
    const shown = [];
    for (let y = 0; y < H; y++) {
      const el = document.createElement("span");
      el.className = "row";
      code.appendChild(el);
      rows.push(el);
      shown.push(null);
    }
    function setRow(y, html) {
      if (shown[y] !== html) {
        rows[y].innerHTML = html;
        shown[y] = html;
      }
    }
    function paint(a, f) {
      const r = frameRows(a, f);
      for (let y = 0; y < H; y++) setRow(y, r[y]);
    }

    // The progress strip: 22 cells, one per age.
    const total = 22;
    const cells = [];
    for (let i = 1; i <= total; i++) {
      const c = document.createElement("span");
      if (ages.some(function (a) { return a.n === i; })) c.className = "stop";
      ticks.appendChild(c);
      cells.push(c);
    }

    // ── Sizing: fit the whole frame, or crop to the town on a phone ──
    let ch = 0.6;
    let fs = 12;
    function measure() {
      const probe = document.createElement("span");
      probe.style.cssText = "position:absolute;visibility:hidden;font-size:100px;white-space:pre";
      probe.textContent = "0000000000";
      map.appendChild(probe);
      ch = probe.getBoundingClientRect().width / 1000 || 0.6;
      map.removeChild(probe);
    }
    function fit() {
      const mobile = window.innerWidth <= 720;
      const padX = mobile ? 32 : 40;
      const availW = stage.clientWidth - padX - 2;
      let availH = stage.clientHeight - 56 - 44 - 36 - 2;
      if (mobile && hero) availH -= hero.offsetHeight + 70;
      let cols = W;
      fs = Math.min(availW / (cols * ch), availH / (H * 1.2));
      const minFs = 8.5;
      if (fs < minFs) {
        cols = Math.max(56, Math.min(W, Math.floor(availW / (minFs * ch))));
        fs = Math.min(availW / (cols * ch), availH / (H * 1.2));
      }
      fs = Math.max(4, Math.floor(fs * 10) / 10);
      const cw = ch * fs;
      // The town sits left of the legend; center a crop on it.
      const center = 49;
      const off = Math.max(0, Math.min(W - cols, Math.round(center - cols / 2)));
      root.style.setProperty("--fs", fs + "px");
      screen.style.width = Math.ceil(cols * cw) + "px";
      screen.style.height = Math.ceil(H * 1.2 * fs) + "px";
      map.style.transform = off ? "translateX(" + -off * cw + "px)" : "";
      // The headline scales with the pane it sits on.
      const paneW = cols * cw;
      const mobileHero = window.innerWidth <= 720;
      const heroFs = mobileHero ? Math.min(44, Math.max(30, window.innerWidth * 0.1)) : Math.min(104, Math.max(30, paneW * 0.075));
      root.style.setProperty("--hero-fs", heroFs.toFixed(1) + "px");
      root.style.setProperty("--hero-tag", Math.min(16, Math.max(11, heroFs * 0.19)).toFixed(1) + "px");
    }
    measure();
    fit();
    let resizeT = 0;
    window.addEventListener("resize", function () {
      clearTimeout(resizeT);
      resizeT = setTimeout(fit, 80);
    });
    if (document.fonts && document.fonts.ready) {
      document.fonts.ready.then(function () {
        measure();
        fit();
      });
    }

    // ── Which age the scroll is on ──
    let cur = -1;
    let frame = 0;
    function label(a) {
      const age = ages[a];
      hudAge.textContent = age.name;
      hudN.textContent = String(age.n);
      cells.forEach(function (c, i) {
        c.classList.toggle("on", i + 1 <= age.n);
        c.classList.toggle("now", i + 1 === age.n);
      });
      const what = ALT[age.key] ? ": " + ALT[age.key] : "";
      pane.setAttribute(
        "aria-label",
        "The AgeForge map in the " + age.name + ", age " + age.n + " of 22" + what + ", drawn in colored terminal characters."
      );
    }

    let liveT = 0;
    function announce(a) {
      clearTimeout(liveT);
      liveT = setTimeout(function () {
        if (live) live.textContent = "Map: " + ages[a].name + ", age " + ages[a].n + " of 22.";
      }, 900);
    }

    // ── The transition: a quiet dissolve. Cells change one at a time, in a
    // scattered order, from what is on screen to the next age's frame. ──
    const DISSOLVE_MS = 480;
    let anim = null;
    function stop() {
      if (anim) cancelAnimationFrame(anim.raf);
      anim = null;
    }
    function scattered(n) {
      const order = new Uint16Array(n);
      for (let i = 0; i < n; i++) order[i] = i;
      for (let i = n - 1; i > 0; i--) {
        const j = Math.floor(Math.random() * (i + 1));
        const t = order[i];
        order[i] = order[j];
        order[j] = t;
      }
      return order;
    }
    function go(a) {
      if (a === cur) return;
      const from = cur;
      const fromFrame = frame;
      cur = a;
      frame = 0;
      label(a);
      announce(a);
      if (from < 0 || reduceMotion.matches) {
        stop();
        paint(a, 0);
        return;
      }
      // Start from what is on screen: a whole frame, or a dissolve already
      // under way, which then heads for the new frame instead.
      const mix = anim
        ? anim.mix
        : frameCells(from, fromFrame).map(function (row) {
            return { c: row.c.slice(), s: row.s.slice() };
          });
      stop();
      const to = frameCells(a, 0);
      const order = scattered(W * H);
      const t0 = performance.now();
      let done = 0;
      let drawn = 0;
      anim = { mix: mix, raf: 0 };
      const step = function (now) {
        const p = Math.min(1, Math.max(0, (now - t0) / DISSOLVE_MS));
        if (p < 1 && now - drawn < 30) {
          anim.raf = requestAnimationFrame(step);
          return;
        }
        drawn = now;
        // Ease in and out, so the change starts and settles gently.
        const upto = p >= 1 ? order.length : Math.floor(p * p * (3 - 2 * p) * order.length);
        const dirty = [];
        for (; done < upto; done++) {
          const y = Math.floor(order[done] / W);
          const x = order[done] - y * W;
          const row = mix[y];
          if (row.c[x] !== to[y].c[x] || row.s[x] !== to[y].s[x]) {
            row.c[x] = to[y].c[x];
            row.s[x] = to[y].s[x];
            dirty[y] = true;
          }
        }
        for (let y = 0; y < H; y++) {
          if (dirty[y]) setRow(y, cellsHTML(mix[y].c, mix[y].s));
        }
        if (p < 1) {
          anim.raf = requestAnimationFrame(step);
        } else {
          anim = null;
          paint(a, 0);
        }
      };
      anim.raf = requestAnimationFrame(step);
    }

    let ticking = false;
    let atTop = true;
    function onScroll() {
      ticking = false;
      const rect = timeline.getBoundingClientRect();
      const span = timeline.offsetHeight - window.innerHeight;
      const p = span > 0 ? Math.min(1, Math.max(0, -rect.top / span)) : 0;
      const top = -rect.top < 40;
      if (top !== atTop) {
        atTop = top;
        document.body.classList.toggle("scrolled", !top);
      }
      const at = p * STEPS;
      go(at < FIRST_HOLD ? 0 : Math.min(N - 1, 1 + Math.floor(at - FIRST_HOLD)));
    }
    window.addEventListener(
      "scroll",
      function () {
        if (!ticking) {
          ticking = true;
          requestAnimationFrame(onScroll);
        }
      },
      { passive: true }
    );
    onScroll();

    // ── Ambient motion: each age's few frames, while the map is in view ──
    let visible = true;
    if ("IntersectionObserver" in window) {
      new IntersectionObserver(function (entries) {
        visible = entries[0].isIntersecting;
      }).observe(pane);
    }
    setInterval(function () {
      if (anim || !visible || document.hidden || reduceMotion.matches || cur < 0) return;
      const n = ages[cur].frames.length;
      if (n < 2) return;
      frame = (frame + 1) % n;
      paint(cur, frame);
    }, 260);
  }

  // ── Install command ────────────────────────────────────────────────────

  const BASE = "https://github.com/espresso20/ageforge/releases/latest/download/";
  const OS = {
    mac: {
      cmd: "curl -fLo ageforge " + BASE + "ageforge-macos-arm64 && chmod +x ageforge && ./ageforge",
      note: "Apple Silicon. On an Intel Mac, swap in <code>ageforge-macos-amd64</code>.",
    },
    linux: {
      cmd: "curl -fLo ageforge " + BASE + "ageforge-linux-amd64 && chmod +x ageforge && ./ageforge",
      note: "x86_64. On ARM, swap in <code>ageforge-linux-arm64</code>.",
    },
    win: {
      cmd: "iwr " + BASE + "ageforge-windows-amd64.exe -OutFile ageforge.exe; .\\ageforge.exe",
      note: "PowerShell, in Windows Terminal.",
    },
  };
  const cmdEl = document.getElementById("cmd");
  const noteEl = document.getElementById("cmd-note");
  const copyBtn = document.getElementById("copy");
  const tabs = Array.prototype.slice.call(document.querySelectorAll(".os-tab"));

  function pickOS(key) {
    if (!OS[key] || !cmdEl) return;
    // Break the long URL after slashes, not mid-word; copying still gets
    // the plain text.
    cmdEl.innerHTML = OS[key].cmd
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/\//g, "/<wbr>");
    noteEl.innerHTML = OS[key].note;
    tabs.forEach(function (t) {
      t.setAttribute("aria-selected", t.dataset.os === key ? "true" : "false");
    });
  }
  tabs.forEach(function (t) {
    t.addEventListener("click", function () {
      pickOS(t.dataset.os);
    });
  });
  const ua = (navigator.userAgentData && navigator.userAgentData.platform) || navigator.userAgent || "";
  let os = "mac";
  if (/win/i.test(ua) && !/darwin|mac/i.test(ua)) os = "win";
  else if (/linux|x11|cros/i.test(ua) && !/android/i.test(ua)) os = "linux";
  pickOS(os);

  function copyInstall(btn) {
    const text = cmdEl ? cmdEl.textContent : "";
    const done = function () {
      if (!btn) return;
      const old = btn.textContent;
      btn.textContent = "Copied";
      btn.classList.add("done");
      setTimeout(function () {
        btn.textContent = old;
        btn.classList.remove("done");
      }, 1600);
    };
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text).then(done, function () {
        selectText(cmdEl);
      });
    } else {
      selectText(cmdEl);
    }
  }
  function selectText(el) {
    const range = document.createRange();
    range.selectNodeContents(el);
    const sel = window.getSelection();
    sel.removeAllRanges();
    sel.addRange(range);
  }
  if (copyBtn) {
    copyBtn.addEventListener("click", function () {
      copyInstall(copyBtn);
    });
  }

  // ── The try-it prompt: a handful of commands, in the game's voice ──────

  const input = document.getElementById("repl-input");
  const out = document.getElementById("repl-out");
  const ghost = document.getElementById("ghost");
  if (!input || !out) return;

  const frameCount = F && F.ages ? F.ages.length : 0;
  const WORDS = ["zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten", "eleven", "twelve", "thirteen", "fourteen"];
  const REPLIES = {
    help: [
      ["ok", "Commands that work here: help, play, build hut, advance, map."],
      ["ok", "The game has dozens more, a Help panel, and completion as you type."],
    ],
    play: [
      ["ok", "A new run doesn't start in a browser tab. AgeForge opens full screen in your terminal."],
      ["ok", "Paste the command above and your first campfire is a few seconds off."],
    ],
    "build hut": [
      ["err", "Cannot afford Hut: need 14 wood (have 0)."],
      ["ok", "There is no wood on a web page. In the game you start with 50."],
    ],
    advance: [
      ["err", "Not ready to advance. The Next Age bar lists what is missing."],
      ["ok", "Top of the list: a running copy of AgeForge."],
    ],
    map: [
      ["ok", "You just scrolled through it: " + (WORDS[frameCount] || "a dozen") + " ages, drawn by the game from a real run."],
      ["ok", "In the game it is live, built from your own buildings, in two styles."],
    ],
    quit: [["ok", "In the game, Esc saves and returns to the main menu. Here, closing the tab works too."]],
  };
  const ALIASES = { h: "help", "?": "help", "b hut": "build hut", exit: "quit" };
  const NAMES = ["help", "play", "build hut", "advance", "map"];

  function line(cls, text) {
    const p = document.createElement("p");
    p.className = "r-" + cls;
    p.textContent = text;
    out.appendChild(p);
    return p;
  }
  function hint() {
    const p = document.createElement("p");
    p.className = "r-hint";
    p.appendChild(document.createTextNode("→ The real thing is one paste away: "));
    const b = document.createElement("button");
    b.type = "button";
    b.textContent = "copy the install command";
    b.addEventListener("click", function () {
      copyInstall(b);
    });
    p.appendChild(b);
    out.appendChild(p);
  }
  function distance(a, b) {
    const d = [];
    for (let i = 0; i <= a.length; i++) d[i] = [i];
    for (let j = 0; j <= b.length; j++) d[0][j] = j;
    for (let i = 1; i <= a.length; i++) {
      for (let j = 1; j <= b.length; j++) {
        d[i][j] = Math.min(d[i - 1][j] + 1, d[i][j - 1] + 1, d[i - 1][j - 1] + (a[i - 1] === b[j - 1] ? 0 : 1));
      }
    }
    return d[a.length][b.length];
  }
  function run(raw) {
    const typed = raw.trim();
    if (!typed) return;
    line("echo", typed);
    let key = typed.toLowerCase().replace(/\s+/g, " ");
    key = ALIASES[key] || key;
    if (REPLIES[key]) {
      REPLIES[key].forEach(function (l) {
        line(l[0], l[1]);
      });
    } else {
      const word = typed.split(/\s+/)[0];
      let best = "";
      let bestD = 3;
      NAMES.forEach(function (n) {
        const d = distance(key, n);
        if (d < bestD) {
          bestD = d;
          best = n;
        }
      });
      line("err", "Unknown command '" + word + "'." + (best ? " Did you mean '" + best + "'?" : "") + " Type help for all commands.");
    }
    hint();
    while (out.children.length > 14) out.removeChild(out.firstChild);
    out.scrollTop = out.scrollHeight;
  }

  // Ghost text, as in the game: the best completion shows dim after the
  // cursor; Tab or the right arrow takes it.
  function completion() {
    const v = input.value.toLowerCase();
    if (!v) return "";
    for (let i = 0; i < NAMES.length; i++) {
      if (NAMES[i].indexOf(v) === 0 && NAMES[i] !== v) return NAMES[i];
    }
    return "";
  }
  function drawGhost() {
    const c = completion();
    ghost.innerHTML = "";
    if (!c) return;
    const typed = document.createElement("b");
    typed.textContent = input.value;
    ghost.appendChild(typed);
    ghost.appendChild(document.createTextNode(c.slice(input.value.length)));
  }
  input.addEventListener("input", drawGhost);
  input.addEventListener("keydown", function (e) {
    const atEnd = input.selectionStart === input.value.length;
    if ((e.key === "Tab" || (e.key === "ArrowRight" && atEnd)) && completion()) {
      e.preventDefault();
      input.value = completion();
      drawGhost();
    } else if (e.key === "Enter") {
      e.preventDefault();
      run(input.value);
      input.value = "";
      drawGhost();
    }
  });
})();
