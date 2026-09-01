(function () {
      var modules = (window.PCS_NUEVAS_PLANTILLAS || []).map(function (item) {
        return [item.id, item.module, item.title, item.summary || item.lead, item.icon || "/img/company-briefcase-color.svg"];
      });

      function esc(value) {
        return String(value == null ? "" : value).replace(/[&<>"']/g, function (ch) {
          return {"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#39;"}[ch];
        });
      }

      function href(module) {
        return "/administrar_empresa/modulo_menu.html?module=" + encodeURIComponent(module);
      }

      function renderModules(items) {
        var grid = document.getElementById("plantillasLaunchGrid");
        if (!items.length) {
          grid.innerHTML = '<div class="plantillas-launch-empty">No hay plantillas habilitados para esta empresa o rol. Revisa la licencia y los permisos por rol en el panel de super administrador.</div>';
          return;
        }
        grid.innerHTML = items.map(function (item, idx) {
        return '<a id="' + esc(item[0]) + '" class="plantillas-launch-card tone-' + ((idx % 6) + 1) + '" href="' + esc(href(item[1])) + '">' +
          '<img class="icon" src="' + esc(item[4]) + '" alt="">' +
          '<strong>' + esc(item[2]) + '</strong>' +
          '<span>' + esc(item[3]) + '</span>' +
          '</a>';
        }).join("");
      }

      function resolveEmpresaId() {
        try {
          if (window.__empresaModuleGuard && typeof window.__empresaModuleGuard.resolveEmpresaId === "function") {
            return Number(window.__empresaModuleGuard.resolveEmpresaId() || 0);
          }
        } catch (_) {}
        try {
          return Number((new URLSearchParams(window.location.search || "")).get("empresa_id") || 0);
        } catch (_) {
          return 0;
        }
      }

      function moduleRowAllowsCreate(row) {
        if (!row || typeof row !== "object") return false;
        if (typeof row.create !== "undefined") return !!row.create;
        if (row.acciones && typeof row.acciones === "object") return !!row.acciones.C;
        return false;
      }

      function moduleAllowedByContext(item, context) {
        if (!context || typeof context !== "object") return true;
        var pages = context.paginas || {};
        if (Object.prototype.hasOwnProperty.call(pages, item[0])) return !!pages[item[0]];
        var rows = Array.isArray(context.modulos) ? context.modulos : [];
        for (var i = 0; i < rows.length; i += 1) {
          if (String(rows[i].modulo || "").trim().toLowerCase() === item[1]) {
            return moduleRowAllowsCreate(rows[i]);
          }
        }
        return false;
      }

      function fetchPermissionContext(empresaId) {
        if (!empresaId) return Promise.resolve(null);
        return fetch("/api/empresa/permisos_contexto?empresa_id=" + encodeURIComponent(empresaId), { credentials: "same-origin" })
          .then(function (resp) { return resp.ok ? resp.json() : null; })
          .catch(function () { return null; });
      }

      function fetchBackendCatalog(empresaId) {
        if (!empresaId) return Promise.resolve(null);
        return fetch("/api/empresa/plantillas_nuevas/catalogo?empresa_id=" + encodeURIComponent(empresaId), { credentials: "same-origin" })
          .then(function (resp) { return resp.ok ? resp.json() : null; })
          .catch(function () { return null; });
      }

      function mergeBackendCatalog(staticModules, catalog) {
        var rows = catalog && Array.isArray(catalog.items) ? catalog.items : [];
        if (!rows.length) return staticModules;
        var staticByModule = {};
        staticModules.forEach(function (item) { staticByModule[item[1]] = item; });
        return rows.map(function (row) {
          var module = String(row.module || row.modulo || "").trim();
          var local = staticByModule[module] || [];
          return [
            row.id || row.page || local[0] || "",
            module,
            row.title || row.titulo || local[2] || module,
            row.summary || row.resumen || local[3] || "",
            local[4] || "/img/company-briefcase-color.svg"
          ];
        }).filter(function (item) { return item[0] && item[1]; });
      }

      var empresaId = resolveEmpresaId();
      Promise.all([fetchPermissionContext(empresaId), fetchBackendCatalog(empresaId)]).then(function (results) {
        var context = results[0];
        var catalog = results[1];
        var effectiveModules = mergeBackendCatalog(modules, catalog);
        renderModules(effectiveModules.filter(function (item) { return moduleAllowedByContext(item, context); }));
      });
    })();
