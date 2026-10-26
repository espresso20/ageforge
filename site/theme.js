// Runs in the head, ahead of the stylesheet, so a saved light or dark choice is
// in place before the first paint. Everything else about themes is in script.js.
(function () {
  try {
    var t = localStorage.getItem("ageforge-theme");
    if (t === "dark" || t === "light") document.documentElement.setAttribute("data-theme", t);
  } catch (e) {}
})();
