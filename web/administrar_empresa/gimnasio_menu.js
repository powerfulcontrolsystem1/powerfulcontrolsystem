(function () {
      var VERSION = "20260508-gimnasio-submenu";
      var activeLink = null;

      function parsePositiveInt(raw) {
        var n = Number(String(raw || "").trim());
        if (!Number.isFinite(n)) return 0;
        n = Math.trunc(n);
        return n > 0 ? n : 0;
      }

      window.__resolveEmpresaIdContext = function () {
        try {
          var params = new URLSearchParams(window.location.search || "");
          var own = parsePositiveInt(params.get("empresa_id") || params.get("id"));
          if (own > 0) return own;
        } catch (_) {}
        try {
          if (window.parent && window.parent !== window && typeof window.parent.__resolveEmpresaIdContext === "function") {
            var parentResolved = parsePositiveInt(window.parent.__resolveEmpresaIdContext());
            if (parentResolved > 0) return parentResolved;
          }
        } catch (_) {}
        try {
          var candidates = [
            sessionStorage.getItem("active_empresa_id"),
            sessionStorage.getItem("empresa_id"),
            localStorage.getItem("active_empresa_id"),
            localStorage.getItem("empresa_id")
          ];
          for (var i = 0; i < candidates.length; i += 1) {
            var parsed = parsePositiveInt(candidates[i]);
            if (parsed > 0) return parsed;
          }
        } catch (_) {}
        return 0;
      };

      function withEmpresaAndVersion(rawUrl) {
        try {
          var url = new URL(rawUrl, window.location.origin);
          var empresaId = window.__resolveEmpresaIdContext();
          if (empresaId > 0) url.searchParams.set("empresa_id", String(empresaId));
          url.searchParams.set("submenu", "1");
          if (!url.searchParams.get("v")) url.searchParams.set("v", VERSION);
          return url.pathname + url.search + url.hash;
        } catch (_) {
          return rawUrl;
        }
      }

      function markActive(link) {
        if (activeLink) activeLink.classList.remove("active");
        activeLink = link || null;
        if (activeLink) {
          activeLink.classList.add("active");
          openGroup(activeLink.closest(".admin-nav-group"));
        }
      }

      function openGroup(group) {
        if (!group || !group.parentElement) return;
        Array.prototype.slice.call(group.parentElement.querySelectorAll(".admin-nav-group")).forEach(function (item) {
          var open = item === group;
          item.classList.toggle("is-open", open);
          var title = item.querySelector(".admin-nav-group-title");
          if (title) title.setAttribute("aria-expanded", open ? "true" : "false");
        });
      }

      function setupGroups() {
        Array.prototype.slice.call(document.querySelectorAll("#gimnasioSidebarNav .admin-nav-group")).forEach(function (group, index) {
          var title = group.querySelector(".admin-nav-group-title");
          if (!title) return;
          title.setAttribute("role", "button");
          title.setAttribute("tabindex", "0");
          if (index === 0) openGroup(group);
          var toggle = function () {
            var willOpen = !group.classList.contains("is-open");
            if (willOpen) openGroup(group);
          };
          title.addEventListener("click", toggle);
          title.addEventListener("keydown", function (event) {
            if (event.key === "Enter" || event.key === " ") {
              event.preventDefault();
              toggle();
            }
          });
        });
      }

      var links = Array.prototype.slice.call(document.querySelectorAll("#gimnasioSidebarNav a[href]"));
      setupGroups();
      links.forEach(function (link, index) {
        link.setAttribute("href", withEmpresaAndVersion(link.getAttribute("href")));
        link.addEventListener("click", function () { markActive(link); });
        if (index === 0) markActive(link);
      });

      var frame = document.getElementById("gimnasioContentFrame");
      if (frame) {
        frame.setAttribute("src", withEmpresaAndVersion(frame.getAttribute("src") || "/administrar_empresa/gimnasio.html?tab=dashboard&submenu=1"));
      }
    })();
