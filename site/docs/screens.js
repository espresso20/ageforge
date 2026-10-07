// The wiki's plates: real views of the game's screens, drawn as text.
//
// A page asks for one with a figure:
//
//   <figure class="screen" data-screen="dashboard"><figcaption>The dashboard in the Bronze Age.</figcaption></figure>
//
// This Docsify plugin fetches screens/dashboard.json (written by
// TestWriteSiteScreens, ui/site_screens_test.go) when the figure nears the
// viewport and draws it above the caption: one block per row, one span per
// color run, and every character outside ASCII in a box one cell wide, so a
// fallback font cannot break the grid. The font is sized so the screen fits
// its column; below a readable size it stops shrinking and the plate scrolls
// sideways instead. If the file cannot be had, the caption stays as it is.
// No libraries.
(function () {
  "use strict";

  var MIN_PX = 9;        // a screen is never drawn smaller than this
  var PAIR_MIN_PX = 6.5; // except side by side (.screen-pair), where two must fit
  var MAX_PX = 14;
  var NAME = /^[a-z0-9][a-z0-9-]*$/;

  var script = document.currentScript;
  var folder = (script && script.src ? script.src.replace(/[^/]*$/, "") : "") + "screens/";
  var files = {};        // name -> promise of the parsed file
  var near = null;       // IntersectionObserver for the page now shown
  var sized = null;      // ResizeObserver for the page now shown
  var drawn = [];        // figures drawn on the page now shown

  function load(name) {
    if (!files[name]) {
      files[name] = fetch(folder + name + ".json").then(function (res) {
        if (!res.ok) throw new Error("screen " + name + ": " + res.status);
        return res.json();
      }).then(function (s) {
        if (!s || !(s.w > 0) || !(s.h > 0) || !Array.isArray(s.rows) || !Array.isArray(s.styles) || !Array.isArray(s.palette)) {
          throw new Error("screen " + name + ": not a screen file");
        }
        return s;
      });
      files[name].catch(function () { delete files[name]; });
    }
    return files[name];
  }

  // term builds the <pre> for a screen.
  function term(s) {
    var pre = document.createElement("pre");
    pre.className = "screen-term";
    pre.setAttribute("aria-hidden", "true");
    pre.style.setProperty("--cols", s.w);
    var ground = s.styles[0] ? s.palette[s.styles[0][1]] : "";
    for (var y = 0; y < s.h; y++) {
      var row = s.rows[y] || { t: "", r: [] };
      var line = document.createElement("span");
      line.className = "screen-row";
      var chars = Array.from(row.t || "");
      var x = 0;
      for (var i = 0; i + 1 < row.r.length; i += 2) {
        var st = s.styles[row.r[i]] || [0, 0, 0];
        var n = row.r[i + 1];
        var run = document.createElement("span");
        run.style.color = s.palette[st[0]];
        if (s.palette[st[1]] !== ground) run.style.background = s.palette[st[1]];
        if (st[2] & 1) run.style.fontWeight = "700";
        if (st[2] & 2) run.style.fontStyle = "italic";
        if (st[2] & 4) run.style.textDecoration = "underline";
        var plain = "";
        for (var k = x; k < x + n && k < chars.length; k++) {
          var ch = chars[k];
          if (ch.charCodeAt(0) < 128) { plain += ch; continue; }
          if (plain) { run.appendChild(document.createTextNode(plain)); plain = ""; }
          var cell = document.createElement("i");
          cell.textContent = ch;
          run.appendChild(cell);
        }
        if (plain) run.appendChild(document.createTextNode(plain));
        line.appendChild(run);
        x += n;
      }
      pre.appendChild(line);
    }
    return pre;
  }

  // fit sizes a drawn figure's font so its screen fills the plate, within the
  // limits. A cell's width is measured, not assumed: the font may still be on
  // its way.
  function fit(fig) {
    var d = fig.__screen;
    if (!d || !d.scroll.clientWidth) return;
    var style = getComputedStyle(d.scroll);
    var room = d.scroll.clientWidth - parseFloat(style.paddingLeft) - parseFloat(style.paddingRight);
    var cell = d.probe.getBoundingClientRect().width / 10 / parseFloat(getComputedStyle(d.pre).fontSize);
    if (!(cell > 0) || !(room > 0)) return;
    var min = fig.closest(".screen-pair") ? PAIR_MIN_PX : MIN_PX;
    var px = Math.floor(room / (d.cols * cell) * 100) / 100;
    px = Math.max(min, Math.min(MAX_PX, px));
    d.pre.style.fontSize = px + "px";
    var wide = d.scroll.scrollWidth > d.scroll.clientWidth + 1;
    fig.classList.toggle("screen-scrolls", wide);
    if (wide) d.scroll.setAttribute("tabindex", "0"); else d.scroll.removeAttribute("tabindex");
  }

  function draw(fig, s) {
    if (fig.__screen || !fig.isConnected) return;
    var caption = fig.querySelector("figcaption");
    var plate = document.createElement("div");
    plate.className = "screen-plate";
    if (s.bg) plate.style.background = s.bg; // a screen in a theme of its own keeps that theme's ground
    var scroll = document.createElement("div");
    scroll.className = "screen-scroll";
    var pre = term(s);
    var probe = document.createElement("span");
    probe.className = "screen-probe";
    probe.textContent = "0000000000";
    pre.appendChild(probe);
    scroll.appendChild(pre);
    plate.appendChild(scroll);
    fig.insertBefore(plate, caption || null);
    var label = caption ? caption.textContent.replace(/\s+/g, " ").trim() : "";
    fig.setAttribute("role", "img");
    fig.setAttribute("aria-label", label || "A screen from the game.");
    fig.classList.add("screen-drawn");
    fig.__screen = { scroll: scroll, pre: pre, probe: probe, cols: s.w };
    drawn.push(fig);
    fit(fig);
    if (sized) sized.observe(scroll);
  }

  function show(fig) {
    var name = fig.getAttribute("data-screen") || "";
    if (!NAME.test(name)) return;
    load(name).then(function (s) { draw(fig, s); }, function () { /* the caption stands alone */ });
  }

  // scan runs after every page render: it drops the last page's observers and
  // watches the new page's figures.
  function scan() {
    if (near) near.disconnect();
    if (sized) sized.disconnect();
    drawn = [];
    var figs = document.querySelectorAll("figure.screen[data-screen]");
    if (!figs.length) return;
    sized = "ResizeObserver" in window ? new ResizeObserver(function (entries) {
      for (var i = 0; i < entries.length; i++) {
        var fig = entries[i].target.closest("figure.screen");
        if (fig) fit(fig);
      }
    }) : null;
    if (!("IntersectionObserver" in window)) {
      Array.prototype.forEach.call(figs, show);
      return;
    }
    near = new IntersectionObserver(function (entries, io) {
      entries.forEach(function (e) {
        if (!e.isIntersecting) return;
        io.unobserve(e.target);
        show(e.target);
      });
    }, { rootMargin: "600px 0px" });
    Array.prototype.forEach.call(figs, function (fig) { near.observe(fig); });
  }

  function refit() { drawn.forEach(fit); }
  if (!("ResizeObserver" in window)) window.addEventListener("resize", refit);
  if (document.fonts && document.fonts.ready) document.fonts.ready.then(refit);

  window.$docsify = window.$docsify || {};
  window.$docsify.plugins = [function (hook) { hook.doneEach(scan); }].concat(window.$docsify.plugins || []);
})();
