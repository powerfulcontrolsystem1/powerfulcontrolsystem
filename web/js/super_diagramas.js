(function () {
  var data = window.PCSSuperDiagramData || { order: [], diagrams: {} };
  var diagramId = document.body ? document.body.getAttribute("data-diagram-id") : "";
  var diagram = data.diagrams[diagramId] || data.diagrams.modulos;

  function injectStyles() {
    var style = document.createElement("style");
    style.textContent = [
      "body.pcs-diagram-page{margin:0;background:var(--bg,#f4f7fb);color:var(--text,#111827);font-family:system-ui,-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif}",
      ".diagram-shell{min-height:100vh;padding:18px;box-sizing:border-box}",
      ".diagram-header{display:flex;align-items:flex-start;justify-content:space-between;gap:16px;margin:0 0 16px}",
      ".diagram-header h1{margin:0;font-size:clamp(1.35rem,2.3vw,2rem);line-height:1.12}",
      ".diagram-header p{margin:8px 0 0;max-width:860px;color:var(--muted,#64748b);line-height:1.45}",
      ".diagram-badge{display:inline-flex;align-items:center;justify-content:center;min-height:28px;padding:4px 10px;border:1px solid var(--border,#cbd5e1);border-radius:999px;background:var(--surface,#fff);font-size:.82rem;font-weight:800;color:var(--accent,#0f6fcb);white-space:nowrap}",
      ".diagram-layout{display:block}",
      ".diagram-panel{background:var(--surface,#fff);border:1px solid var(--border,#dbe3ef);border-radius:8px;box-shadow:0 10px 28px rgba(15,23,42,.08)}",
      ".diagram-panel{padding:14px;margin-bottom:16px;overflow:hidden}",
      ".diagram-canvas{overflow:auto;border:1px solid var(--border,#dbe3ef);border-radius:8px;background:linear-gradient(180deg,#f8fbff,#eef5ff);padding:8px}",
      ".diagram-canvas svg{display:block;min-width:1180px;width:100%;height:auto}",
      ".diagram-source-head{display:flex;align-items:center;justify-content:space-between;gap:12px;margin-bottom:10px}",
      ".diagram-source-head h2{font-size:1rem;margin:0}",
      ".diagram-source{white-space:pre;overflow:auto;max-height:440px;margin:0;padding:14px;border-radius:8px;background:#0f172a;color:#e2e8f0;font:13px/1.5 Consolas,'Liberation Mono',monospace}",
      ".diagram-note{margin:12px 0 0;color:var(--muted,#64748b);font-size:.9rem;line-height:1.45}",
      ".diagram-actions{display:flex;gap:8px;flex-wrap:wrap}",
      ".diagram-actions button{border:1px solid var(--border,#cbd5e1);background:var(--surface,#fff);color:var(--text,#111827);border-radius:7px;padding:7px 10px;font-weight:800;cursor:pointer}",
      "@media(max-width:860px){.diagram-shell{padding:12px}.diagram-header{display:block}.diagram-badge{margin-top:10px}}"
    ].join("\n");
    document.head.appendChild(style);
  }

  function svgText(lines, x, y, cls) {
    return lines.map(function (line, index) {
      var size = index === 0 ? 15 : 12;
      var weight = index === 0 ? 800 : 600;
      return '<text x="' + x + '" y="' + (y + index * 18) + '" text-anchor="middle" font-size="' + size + '" font-weight="' + weight + '" fill="#0f172a">' + escapeHtml(line) + "</text>";
    }).join("");
  }

  function escapeHtml(value) {
    return String(value || "").replace(/[&<>"']/g, function (ch) {
      return ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[ch];
    });
  }

  function nodeFill(kind) {
    if (kind === "core") return "#dbeafe";
    if (kind === "db") return "#dcfce7";
    if (kind === "external") return "#ffedd5";
    if (kind === "infra") return "#e0e7ff";
    return "#ffffff";
  }

  function renderSvg(target, item) {
    var nodes = item.nodes || [];
    var nodeMap = {};
    nodes.forEach(function (node) {
      node.w = node.w || 190;
      node.h = node.h || 90;
      nodeMap[node.id] = node;
    });
    var maxX = 0;
    var maxY = 0;
    nodes.forEach(function (node) {
      maxX = Math.max(maxX, node.x + node.w + 70);
      maxY = Math.max(maxY, node.y + node.h + 70);
    });
    var edges = (item.edges || []).map(function (edge) {
      var from = nodeMap[edge[0]];
      var to = nodeMap[edge[1]];
      if (!from || !to) return "";
      var x1 = from.x + from.w / 2;
      var y1 = from.y + from.h / 2;
      var x2 = to.x + to.w / 2;
      var y2 = to.y + to.h / 2;
      return '<line x1="' + x1 + '" y1="' + y1 + '" x2="' + x2 + '" y2="' + y2 + '" stroke="#64748b" stroke-width="2" marker-end="url(#arrow)" opacity=".82"/>';
    }).join("");
    var nodeSvg = nodes.map(function (node) {
      var lines = String(node.label || "").split("\n");
      var textY = node.y + node.h / 2 - ((lines.length - 1) * 18) / 2 + 5;
      return [
        '<g>',
        '<rect x="' + node.x + '" y="' + node.y + '" width="' + node.w + '" height="' + node.h + '" rx="8" fill="' + nodeFill(node.kind) + '" stroke="#94a3b8" stroke-width="1.5"/>',
        svgText(lines, node.x + node.w / 2, textY, ""),
        "</g>"
      ].join("");
    }).join("");
    target.innerHTML = [
      '<svg viewBox="0 0 ' + Math.max(maxX, 1280) + ' ' + Math.max(maxY, 620) + '" role="img" aria-label="' + escapeHtml(item.title) + '">',
      "<defs>",
      '<marker id="arrow" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto" markerUnits="strokeWidth">',
      '<path d="M0,0 L0,6 L9,3 z" fill="#64748b"/>',
      "</marker>",
      "</defs>",
      edges,
      nodeSvg,
      "</svg>"
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

  function boot() {
    if (!diagram) return;
    document.title = diagram.title + " - Super Administrador";
    injectStyles();
    document.body.classList.add("pcs-diagram-page");
    document.body.innerHTML = [
      '<main class="diagram-shell">',
      '<header class="diagram-header">',
      '<div><h1>' + escapeHtml(diagram.title) + '</h1><p>' + escapeHtml(diagram.summary) + '</p></div>',
      '<span class="diagram-badge">' + escapeHtml(diagram.badge || "Diagrama") + '</span>',
      "</header>",
      '<div class="diagram-layout">',
      '<section>',
      '<article class="diagram-panel"><div class="diagram-canvas" id="diagramCanvas"></div><p class="diagram-note">Vista interna para super administrador. La fuente Mermaid queda abajo para Codex y mantenimiento tecnico.</p></article>',
      '<article class="diagram-panel"><div class="diagram-source-head"><h2>Fuente Mermaid para Codex</h2><div class="diagram-actions"><button id="copyDiagramSource" type="button">Copiar Mermaid</button></div></div><pre class="diagram-source" id="diagramSource"></pre></article>',
      "</section>",
      "</div>",
      "</main>"
    ].join("");
    renderSvg(document.getElementById("diagramCanvas"), diagram);
    var source = document.getElementById("diagramSource");
    source.textContent = diagram.source || "";
    copySource(document.getElementById("copyDiagramSource"), diagram.source || "");
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", boot);
  } else {
    boot();
  }
})();
