(function () {
  "use strict";

  function escapeHtml(value) {
    return String(value || "").replace(/[&<>"']/g, function (character) {
      return ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[character];
    });
  }

  function injectStyles() {
    var style = document.createElement("style");
    style.textContent = [
      "body.pcs-erd-page{margin:0;background:var(--bg,#f4f7fb);color:var(--text,#111827);font-family:system-ui,-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif}",
      ".erd-shell{min-height:100vh;padding:18px;box-sizing:border-box}",
      ".erd-header{display:flex;justify-content:space-between;gap:18px;align-items:flex-start;margin-bottom:16px}",
      ".erd-header h1{margin:0;font-size:clamp(1.35rem,2.4vw,2rem)}.erd-header p{margin:8px 0 0;max-width:900px;line-height:1.45;color:var(--muted,#64748b)}",
      ".erd-metrics{display:flex;gap:8px;flex-wrap:wrap}.erd-metric{border:1px solid var(--border,#cbd5e1);border-radius:7px;background:var(--surface,#fff);padding:7px 10px;font-size:.84rem;font-weight:800;white-space:nowrap}",
      ".erd-layout{display:grid;grid-template-columns:minmax(260px,330px) minmax(0,1fr);gap:16px;align-items:start}",
      ".erd-panel{background:var(--surface,#fff);border:1px solid var(--border,#dbe3ef);border-radius:8px;box-shadow:0 8px 24px rgba(15,23,42,.07)}",
      ".erd-sidebar{position:sticky;top:12px;padding:12px}.erd-controls{display:grid;gap:8px;margin-bottom:10px}.erd-controls input,.erd-controls select{width:100%;box-sizing:border-box;padding:8px 9px;border:1px solid var(--border,#cbd5e1);border-radius:6px;background:var(--surface,#fff);color:var(--text,#111827)}",
      ".erd-table-list{max-height:calc(100vh - 210px);overflow:auto;border-top:1px solid var(--border,#dbe3ef)}.erd-table-button{display:block;width:100%;border:0;border-bottom:1px solid var(--border,#edf1f6);background:transparent;color:inherit;text-align:left;padding:9px 4px;cursor:pointer}.erd-table-button:hover,.erd-table-button.is-selected{background:#eff6ff}.erd-table-button strong{display:block;font-family:Consolas,'Liberation Mono',monospace;font-size:.86rem}.erd-table-button span{display:block;color:var(--muted,#64748b);font-size:.75rem;margin-top:2px}",
      ".erd-detail{padding:16px;min-width:0}.erd-detail h2{margin:0;font:800 1.18rem Consolas,'Liberation Mono',monospace}.erd-source{margin:6px 0 16px;color:var(--muted,#64748b);font-size:.85rem;word-break:break-word}.erd-section{margin-top:16px}.erd-section h3{margin:0 0 8px;font-size:1rem}.erd-columns{width:100%;border-collapse:collapse;font-size:.86rem}.erd-columns th,.erd-columns td{padding:8px;text-align:left;border-bottom:1px solid var(--border,#e2e8f0);vertical-align:top}.erd-columns th{background:#f8fafc}.erd-column-name{font-family:Consolas,'Liberation Mono',monospace;font-weight:800}.erd-tag{display:inline-block;margin:1px 3px 1px 0;padding:2px 5px;border-radius:4px;background:#e2e8f0;font-size:.72rem;font-weight:800}.erd-tag.pk{background:#dbeafe;color:#1d4ed8}.erd-tag.fk{background:#dcfce7;color:#166534}.erd-tag.logical{background:#fef3c7;color:#92400e}",
      ".erd-relations{display:grid;gap:7px}.erd-relation{padding:9px;border:1px solid var(--border,#e2e8f0);border-radius:6px;background:#fafcff;font-family:Consolas,'Liberation Mono',monospace;font-size:.82rem;overflow-wrap:anywhere}.erd-empty{padding:18px;color:var(--muted,#64748b);text-align:center}",
      "@media(max-width:860px){.erd-shell{padding:12px}.erd-header{display:block}.erd-metrics{margin-top:12px}.erd-layout{grid-template-columns:1fr}.erd-sidebar{position:static}.erd-table-list{max-height:270px}.erd-columns{font-size:.78rem}.erd-columns th,.erd-columns td{padding:6px}}"
    ].join("\n");
    document.head.appendChild(style);
  }

  function relationHtml(relation) {
    var kind = relation.physical ? '<span class="erd-tag fk">FK fisica</span>' : '<span class="erd-tag logical">Relacion logica</span>';
    return '<div class="erd-relation">' + kind + " " + escapeHtml(relation.table + "." + relation.column + " -> " + relation.refTable + "." + relation.refColumn) + "</div>";
  }

  function boot(manifest) {
    var state = { manifest: manifest, selected: manifest.tables[0].name, query: "", category: "" };
    var tableByName = {};
    manifest.tables.forEach(function (table) { tableByName[table.name] = table; });
    var categories = Array.from(new Set(manifest.tables.map(function (table) { return table.category; }))).sort();

    document.body.classList.add("pcs-erd-page");
    document.title = "ERD PostgreSQL completo - Super Administrador";
    document.body.innerHTML = [
      '<main class="erd-shell">',
      '<header class="erd-header"><div><h1>ERD PostgreSQL completo</h1><p>Catalogo navegable de la base de datos: tablas, atributos, claves y relaciones. Las FKs fisicas se distinguen de las relaciones logicas mantenidas por la aplicacion.</p></div>',
      '<div class="erd-metrics"><span class="erd-metric">' + manifest.table_count + ' tablas</span><span class="erd-metric">' + manifest.physical_fk_count + ' FKs fisicas</span><span class="erd-metric">' + manifest.logical_relation_count + ' relaciones logicas</span></div></header>',
      '<div class="erd-layout"><aside class="erd-panel erd-sidebar"><div class="erd-controls"><input id="erdSearch" type="search" placeholder="Buscar tabla o atributo"><select id="erdCategory"><option value="">Todos los dominios</option>' + categories.map(function (category) { return '<option value="' + escapeHtml(category) + '">' + escapeHtml(category) + '</option>'; }).join("") + '</select></div><div id="erdTableList" class="erd-table-list"></div></aside><section id="erdDetail" class="erd-panel erd-detail"></section></div>',
      '</main>'
    ].join("");

    var list = document.getElementById("erdTableList");
    var detail = document.getElementById("erdDetail");
    var search = document.getElementById("erdSearch");
    var category = document.getElementById("erdCategory");

    function filteredTables() {
      var query = state.query.toLowerCase();
      return manifest.tables.filter(function (table) {
        var matchesCategory = !state.category || table.category === state.category;
        var haystack = table.name + " " + table.category + " " + table.columns.map(function (column) { return column.name; }).join(" ");
        return matchesCategory && (!query || haystack.toLowerCase().indexOf(query) !== -1);
      });
    }

    function renderList() {
      var tables = filteredTables();
      if (!tables.some(function (table) { return table.name === state.selected; })) state.selected = tables.length ? tables[0].name : "";
      list.innerHTML = tables.length ? tables.map(function (table) {
        return '<button type="button" class="erd-table-button' + (table.name === state.selected ? ' is-selected' : '') + '" data-table="' + escapeHtml(table.name) + '"><strong>' + escapeHtml(table.name) + '</strong><span>' + escapeHtml(table.category) + ' · ' + table.columns.length + ' atributos</span></button>';
      }).join("") : '<div class="erd-empty">No hay tablas que coincidan.</div>';
      Array.prototype.forEach.call(list.querySelectorAll("[data-table]"), function (button) {
        button.addEventListener("click", function () { state.selected = button.getAttribute("data-table"); renderList(); renderDetail(); });
      });
    }

    function renderDetail() {
      var table = tableByName[state.selected];
      if (!table) { detail.innerHTML = '<div class="erd-empty">Seleccione una tabla.</div>'; return; }
      var physical = manifest.physical_foreign_keys.filter(function (relation) { return relation.table === table.name || relation.refTable === table.name; });
      var logical = manifest.logical_relations.filter(function (relation) { return relation.table === table.name || relation.refTable === table.name; });
      detail.innerHTML = [
        '<h2>' + escapeHtml(table.name) + '</h2><p class="erd-source">Dominio: ' + escapeHtml(table.category) + ' · Fuente: ' + escapeHtml((table.sources || []).join(", ")) + '</p>',
        '<section class="erd-section"><h3>Atributos (' + table.columns.length + ')</h3><div style="overflow:auto"><table class="erd-columns"><thead><tr><th>Columna</th><th>Tipo PostgreSQL</th><th>Restricciones</th></tr></thead><tbody>',
        table.columns.map(function (column) {
          var tags = [];
          if (column.pk) tags.push('<span class="erd-tag pk">PK</span>');
          if (column.notNull) tags.push('<span class="erd-tag">NOT NULL</span>');
          if (column.unique) tags.push('<span class="erd-tag">UNIQUE</span>');
          if (column.default) tags.push('<span class="erd-tag">DEFAULT</span>');
          return '<tr><td class="erd-column-name">' + escapeHtml(column.name) + '</td><td>' + escapeHtml(column.type) + '</td><td>' + (tags.join("") || "-") + '</td></tr>';
        }).join(""),
        '</tbody></table></div></section>',
        '<section class="erd-section"><h3>FKs fisicas (' + physical.length + ')</h3><div class="erd-relations">' + (physical.map(relationHtml).join("") || '<div class="erd-empty">No hay constraints FK fisicos declarados en el SQL extraido.</div>') + '</div></section>',
        '<section class="erd-section"><h3>Relaciones logicas (' + logical.length + ')</h3><div class="erd-relations">' + (logical.map(relationHtml).join("") || '<div class="erd-empty">No se detectaron relaciones logicas para esta tabla.</div>') + '</div></section>'
      ].join("");
    }

    search.addEventListener("input", function () { state.query = search.value; renderList(); renderDetail(); });
    category.addEventListener("change", function () { state.category = category.value; renderList(); renderDetail(); });
    renderList();
    renderDetail();
  }

  injectStyles();
  fetch("/data/documentacion_tecnica_completa_manifest.json?v=20260710").then(function (response) {
    if (!response.ok) throw new Error("No se pudo cargar el catalogo ERD.");
    return response.json();
  }).then(boot).catch(function () {
    document.body.innerHTML = '<main class="erd-shell"><div class="erd-panel erd-detail"><h1>ERD PostgreSQL completo</h1><p>No se pudo cargar el catalogo estructurado.</p></div></main>';
  });
})();
