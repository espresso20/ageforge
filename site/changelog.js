const REPO      = "espresso20/ageforge";
const RAW_URL   = `https://raw.githubusercontent.com/${REPO}/master/CHANGELOG.md`;
const GH_RELEASES = `https://github.com/${REPO}/releases`;

function escHtml(s) {
    return s.replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;");
}

// Parse CHANGELOG.md into an array of {version, date, body} objects.
// Handles entries like:  ## [v3.3.0] — 2026-03-18
function parseChangelog(text) {
    const entries = [];
    const sections = text.split(/^## /m).slice(1); // drop header before first ##

    for (const section of sections) {
        const lines = section.split("\n");
        const heading = lines[0].trim();

        // Skip [Unreleased]
        if (/unreleased/i.test(heading)) continue;

        // Extract version + date from:  [v3.3.0] — 2026-03-18
        const m = heading.match(/\[([^\]]+)\](?:\s*[—-]+\s*(\S+))?/);
        if (!m) continue;

        const version = m[1];
        const date    = m[2] || "";
        const body    = lines.slice(1).join("\n").replace(/^---\s*$/m, "").trim();

        entries.push({ version, date, body });
    }
    return entries;
}

// Convert the body of a changelog entry (### sections + lists) to HTML.
function renderBody(md) {
    if (!md) return "<em style='color:var(--muted)'>No notes for this release.</em>";
    const lines = md.split("\n");
    let html = "";
    let inList = false;

    for (const raw of lines) {
        const line = raw.trimEnd();
        if (/^### (.+)/.test(line)) {
            if (inList) { html += "</ul>"; inList = false; }
            const text = line.replace(/^### /, "");
            let cls = "sec-other";
            if (/added/i.test(text))   cls = "sec-added";
            if (/fixed/i.test(text))   cls = "sec-fixed";
            if (/changed/i.test(text)) cls = "sec-changed";
            html += `<h3 class="${cls}">${escHtml(text)}</h3>`;
        } else if (/^[-*] (.+)/.test(line)) {
            if (!inList) { html += "<ul>"; inList = true; }
            html += `<li>${escHtml(line.replace(/^[-*] /, ""))}</li>`;
        } else if (line === "") {
            if (inList) { html += "</ul>"; inList = false; }
        } else {
            if (inList) { html += "</ul>"; inList = false; }
            html += `<p>${escHtml(line)}</p>`;
        }
    }
    if (inList) html += "</ul>";
    return html;
}

function fmtDate(iso) {
    if (!iso) return "";
    const d = new Date(iso);
    return isNaN(d) ? iso : d.toLocaleDateString("en-US", { year:"numeric", month:"long", day:"numeric" });
}

async function loadChangelog() {
    const feed = document.getElementById("cl-feed");
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 8000);

    try {
        const res = await fetch(RAW_URL, { signal: controller.signal });
        clearTimeout(timeout);
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const text = await res.text();
        const entries = parseChangelog(text);

        if (!entries.length) {
            feed.innerHTML = `<div class="cl-empty">No releases yet.</div>`;
            return;
        }

        feed.innerHTML = entries.map(e => `
            <article class="cl-entry">
                <div>
                    <span class="cl-tag">${escHtml(e.version)}</span>
                    ${e.date ? `<span class="cl-date">${escHtml(fmtDate(e.date))}</span>` : ""}
                </div>
                <div class="cl-body">${renderBody(e.body)}</div>
                <a class="cl-gh-link"
                   href="${GH_RELEASES}/tag/${encodeURIComponent(e.version)}"
                   target="_blank" rel="noopener">
                    View on GitHub ↗
                </a>
            </article>
        `).join("");
    } catch (err) {
        clearTimeout(timeout);
        const msg = err.name === "AbortError" ? "Request timed out." : escHtml(err.message);
        feed.innerHTML = `
            <div class="cl-error">
                Could not load changelog: ${msg}<br><br>
                <a href="${GH_RELEASES}" target="_blank" rel="noopener"
                   style="color:var(--muted);text-decoration:underline">
                    View releases on GitHub ↗
                </a>
            </div>`;
    }
}

loadChangelog();
