(function () {
  var data = window.PCSSuperTechnicalDocs || { diagrams: [] };
  var selectedId = document.body ? document.body.getAttribute("data-tech-diagram-id") : "all";
  var diagrams = selectedId === "all"
    ? data.diagrams
    : data.diagrams.filter(function (item) { return item.id === selectedId; });

  function escapeHtml(value) {
    return String(value || "").replace(/[&<>"']/g, function (ch) {
      return ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[ch];
    });
  }

  function injectStyles() {
    var style = document.createElement("style");
    style.textContent = [
      "body.tech-doc-page{margin:0;background:var(--bg,#f4f7fb);color:var(--text,#111827);font-family:system-ui,-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif}",
      ".tech-doc-shell{min-height:100vh;padding:18px;box-sizing:border-box}",
      ".tech-doc-header{display:flex;align-items:flex-start;justify-content:space-between;gap:16px;margin-bottom:16px}",
      ".tech-doc-header h1{margin:0;font-size:clamp(1.35rem,2.4vw,2rem);line-height:1.12}",
      ".tech-doc-header p{margin:8px 0 0;color:var(--muted,#64748b);max-width:920px;line-height:1.45}",
      ".tech-doc-badge{display:inline-flex;align-items:center;justify-content:center;min-height:28px;padding:4px 10px;border:1px solid var(--border,#cbd5e1);border-radius:999px;background:var(--surface,#fff);font-size:.82rem;font-weight:800;color:var(--accent,#0f6fcb);white-space:nowrap}",
      ".tech-doc-index{display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:10px;margin:0 0 16px}",
      ".tech-doc-index a,.tech-doc-card{background:var(--surface,#fff);border:1px solid var(--border,#dbe3ef);border-radius:8px;box-shadow:0 10px 28px rgba(15,23,42,.08)}",
      ".tech-doc-index a{display:block;padding:10px 12px;color:var(--text,#111827);text-decoration:none;font-weight:800}",
      ".tech-doc-card{padding:14px;margin-bottom:16px;overflow:hidden}",
      ".tech-doc-card h2{margin:0 0 10px;font-size:1.18rem}",
      ".tech-doc-meta{display:flex;gap:8px;flex-wrap:wrap;margin:0 0 12px;color:var(--muted,#64748b);font-size:.86rem}",
      ".tech-doc-meta span{border:1px solid var(--border,#dbe3ef);border-radius:999px;padding:3px 8px;background:rgba(255,255,255,.55)}",
      ".tech-doc-flow{border:1px solid var(--border,#dbe3ef);border-radius:8px;background:linear-gradient(180deg,#f8fbff,#eef5ff);padding:12px;overflow:auto;margin-bottom:12px}",
      ".tech-doc-flow-grid{display:flex;gap:10px;align-items:center;min-width:max-content;flex-wrap:wrap}",
      ".tech-doc-node{border:1px solid #94a3b8;background:#fff;border-radius:8px;padding:8px 10px;min-width:128px;max-width:220px;font-weight:800;text-align:center;box-sizing:border-box}",
      ".tech-doc-arrow{font-weight:900;color:#64748b}",
      ".tech-doc-source-head{display:flex;align-items:center;justify-content:space-between;gap:10px;margin-bottom:8px}",
      ".tech-doc-source-head h3{font-size:1rem;margin:0}",
      ".tech-doc-source{white-space:pre;overflow:auto;max-height:420px;margin:0;padding:14px;border-radius:8px;background:#0f172a;color:#e2e8f0;font:13px/1.5 Consolas,'Liberation Mono',monospace}",
      ".tech-doc-actions{display:flex;gap:8px;flex-wrap:wrap}",
      ".tech-doc-actions button{border:1px solid var(--border,#cbd5e1);background:var(--surface,#fff);color:var(--text,#111827);border-radius:7px;padding:7px 10px;font-weight:800;cursor:pointer}",
      ".tech-doc-empty{padding:18px;border:1px dashed var(--border,#cbd5e1);border-radius:8px;background:var(--surface,#fff)}",
      "@media(max-width:860px){.tech-doc-shell{padding:12px}.tech-doc-header{display:block}.tech-doc-badge{margin-top:10px}}"
    ].join("\n");
    document.head.appendChild(style);
  }

  function extractSteps(source) {
    var lines = String(source || "").split(/\r?\n/);
    var steps = [];
    lines.forEach(function (line) {
      var clean = line.trim();
      if (!clean || /^(flowchart|sequenceDiagram|stateDiagram|classDiagram|erDiagram|participant|actor|classDef|class\s)/.test(clean)) return;
      var parts = clean.split(/-->|->>|-->>|:\s/).map(function (part) {
        return part.replace(/^\w+\s*[\[{("]*/, "").replace(/[\]})"]+$/g, "").trim();
      }).filter(Boolean);
      parts.forEach(function (part) {
        var normalized = part.replace(/\s+/g, " ");
        if (normalized && steps.indexOf(normalized) === -1 && steps.length < 12) steps.push(normalized);
      });
    });
    return steps;
  }

  function renderPreview(source) {
    var steps = extractSteps(source);
    if (!steps.length) {
      return '<div class="tech-doc-flow"><div class="tech-doc-flow-grid"><div class="tech-doc-node">Fuente Mermaid disponible</div></div></div>';
    }
    return [
      '<div class="tech-doc-flow"><div class="tech-doc-flow-grid">',
      steps.map(function (step, index) {
        return (index ? '<span class="tech-doc-arrow">&rarr;</span>' : "") + '<div class="tech-doc-node">' + escapeHtml(step) + '</div>';
      }).join(""),
      "</div></div>"
    ].join("");
  }

  function copySource(button, source) {
    button.addEventListener("click", function () {
      if (!navigator.clipboard) return;
      navigator.clipboard.writeText(source || "").then(function () {
        button.textContent = "Copiado";
        window.setTimeout(function () { button.textContent = "Copiar Mermaid"; }, 1400);
      }).catch(function () {});
    });
  }

  function renderCard(item, index) {
    var id = "techDiagram" + index;
    return [
      '<article class="tech-doc-card" id="' + id + '">',
      "<h2>" + escapeHtml(item.title) + "</h2>",
      '<div class="tech-doc-meta"><span>' + escapeHtml(item.type || "mermaid") + '</span><span>' + escapeHtml(item.id) + "</span></div>",
      renderPreview(item.source),
      '<div class="tech-doc-source-head"><h3>Fuente Mermaid para Codex</h3><div class="tech-doc-actions"><button type="button" data-copy-index="' + index + '">Copiar Mermaid</button></div></div>',
      '<pre class="tech-doc-source">' + escapeHtml(item.source) + "</pre>",
      "</article>"
    ].join("");
  }

  function boot() {
    injectStyles();
    document.body.classList.add("tech-doc-page");
    var title = selectedId === "all" ? "Documentacion tecnica completa" : (diagrams[0] ? diagrams[0].title : "Diagrama tecnico");
    document.title = title + " - Super Administrador";
    var indexHtml = selectedId === "all"
      ? '<nav class="tech-doc-index" aria-label="Indice de diagramas">' + data.diagrams.map(function (item, index) {
          return '<a href="#techDiagram' + index + '">' + escapeHtml(item.title) + "</a>";
        }).join("") + "</nav>"
      : "";
    document.body.innerHTML = [
      '<main class="tech-doc-shell">',
      '<header class="tech-doc-header">',
      '<div><h1>' + escapeHtml(title) + '</h1><p>' + escapeHtml(data.summary || "") + ' Fuente documental: ' + escapeHtml(data.sourceDocument || "") + '.</p></div>',
      '<span class="tech-doc-badge">' + String(diagrams.length) + ' diagramas</span>',
      "</header>",
      indexHtml,
      diagrams.length ? diagrams.map(renderCard).join("") : '<div class="tech-doc-empty">No se encontro el diagrama solicitado.</div>',
      "</main>"
    ].join("");
    Array.prototype.forEach.call(document.querySelectorAll("[data-copy-index]"), function (button) {
      var item = diagrams[Number(button.getAttribute("data-copy-index"))];
      copySource(button, item ? item.source : "");
    });
  }

  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", boot);
  else boot();
})();
