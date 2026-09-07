(function () {
  "use strict";

  var presets = {
    "horarios_trabajadores.html": [
      ["vision", "Visión"],
      ["resumen", "Resumen"],
      ["programacion", "Programación"],
      ["lista", "Lista"]
    ],
    "asistencia_empleados.html": [
      ["resumen", "Resumen"],
      ["configuracion", "Configuración"],
      ["cierre", "Cierre"],
      ["reportes", "Reportes"],
      ["registro", "Registro"],
      ["consulta", "Consulta"]
    ],
    "carnets.html": [
      ["resumen", "Resumen"],
      ["consulta", "Consulta"],
      ["emision", "Emisión"],
      ["plantilla", "Plantilla"],
      ["preview", "Vista previa"],
      ["bitacora", "Bitácora"]
    ],
    "vehiculos_registro.html": [
      ["configuracion", "Configuración"],
      ["registro", "Registro"],
      ["consulta", "Consulta"],
      ["reportes", "Reportes"]
    ],
    "hoja_vida_operativa.html": [
      ["resumen", "Resumen"],
      ["consulta", "Consulta"],
      ["registro", "Hojas de vida"],
      ["eventos", "Eventos"],
      ["alertas", "Alertas"]
    ],
    "ubicacion_gps.html": [
      ["dispositivos", "Dispositivos"],
      ["mapa", "Mapa"]
    ],
    "auditoria.html": [
      ["resumen", "Resumen"],
      ["filtros", "Filtros"],
      ["retencion", "Retencion"],
      ["eventos", "Eventos"],
      ["detalle", "Detalle"]
    ],
    "backups.html": [
      ["configuracion", "Configuracion"],
      ["snapshot", "Snapshot"],
      ["reinicio", "Reiniciar datos"],
      ["historial", "Historial"],
      ["detalle", "Detalle"]
    ],
    "mi_horario.html": [
      ["resumen", "Resumen"],
      ["turnos", "Turnos"]
    ],
    "documentos_onlyoffice.html": [
      ["crear", "Crear"],
      ["subir", "Subir"],
      ["archivos", "Archivos"],
      ["editor", "Editor"]
    ]
  };

  function currentFileName() {
    var path = String(window.location.pathname || "").split("/").pop();
    return path || "";
  }

  function findSections(items) {
    return items.map(function (item) {
      var key = item[0];
      return {
        key: key,
        label: item[1],
        nodes: Array.prototype.slice.call(document.querySelectorAll('[data-pcs-submenu-section="' + key + '"]'))
      };
    }).filter(function (item) {
      return item.nodes.length > 0;
    });
  }

  function setActive(sections, key, nav) {
    sections.forEach(function (section) {
      var active = section.key === key;
      section.nodes.forEach(function (node) {
        node.hidden = !active;
        node.classList.toggle("pcs-submenu-section-active", active);
      });
    });
    Array.prototype.forEach.call(document.querySelectorAll("[data-pcs-submenu-group]"), function (group) {
      var hasVisibleSection = Array.prototype.some.call(group.querySelectorAll("[data-pcs-submenu-section]"), function (node) {
        return !node.hidden;
      });
      group.hidden = !hasVisibleSection;
    });
    if (nav) {
      Array.prototype.forEach.call(nav.querySelectorAll("button[data-pcs-submenu-target]"), function (button) {
        var active = button.getAttribute("data-pcs-submenu-target") === key;
        button.classList.toggle("active", active);
        button.setAttribute("aria-selected", active ? "true" : "false");
        button.setAttribute("tabindex", active ? "0" : "-1");
      });
    }
    try {
      window.history.replaceState(null, "", "#tab-" + key);
    } catch (error) {}
  }

  function initialKey(sections) {
    var hash = String(window.location.hash || "").replace(/^#tab-/, "");
    if (hash) {
      for (var i = 0; i < sections.length; i += 1) {
        if (sections[i].key === hash) return hash;
      }
    }
    return sections[0] ? sections[0].key : "";
  }

  function findSectionKey(sections, key) {
    if (!key) return "";
    for (var i = 0; i < sections.length; i += 1) {
      if (sections[i].key === key) return key;
    }
    return "";
  }

  function mountSubmenu(sections) {
    if (sections.length < 2) return;
    var container = document.querySelector(".container") || document.querySelector(".mi-horario-shell") || document.querySelector(".oo-shell") || document.querySelector(".nc-shell") || document.body;
    var anchor = container.querySelector(".page-header") || container.querySelector(".mi-horario-head") || container.querySelector(".empresa-section-header") || container.querySelector(".nc-hero") || container.firstElementChild;
    var nav = document.createElement("nav");
    nav.className = "pcs-section-submenu";
    nav.setAttribute("aria-label", "Secciones del módulo");
    nav.setAttribute("role", "tablist");
    nav.innerHTML = sections.map(function (section) {
      return '<button type="button" role="tab" data-pcs-submenu-target="' + section.key + '">' + section.label + '</button>';
    }).join("");
    if (anchor && anchor.parentNode === container) {
      anchor.insertAdjacentElement("afterend", nav);
    } else {
      container.insertBefore(nav, container.firstChild);
    }
    nav.addEventListener("click", function (event) {
      var button = event.target && event.target.closest ? event.target.closest("button[data-pcs-submenu-target]") : null;
      if (!button) return;
      setActive(sections, button.getAttribute("data-pcs-submenu-target"), nav);
    });
    nav.addEventListener("keydown", function (event) {
      if (event.key !== "ArrowRight" && event.key !== "ArrowLeft" && event.key !== "Home" && event.key !== "End") return;
      var buttons = Array.prototype.slice.call(nav.querySelectorAll("button[data-pcs-submenu-target]"));
      var current = Math.max(0, buttons.indexOf(document.activeElement));
      var next = current;
      if (event.key === "ArrowRight") next = (current + 1) % buttons.length;
      if (event.key === "ArrowLeft") next = (current - 1 + buttons.length) % buttons.length;
      if (event.key === "Home") next = 0;
      if (event.key === "End") next = buttons.length - 1;
      event.preventDefault();
      buttons[next].focus();
      setActive(sections, buttons[next].getAttribute("data-pcs-submenu-target"), nav);
    });
    setActive(sections, initialKey(sections), nav);
    window.addEventListener("hashchange", function () {
      var key = findSectionKey(sections, String(window.location.hash || "").replace(/^#tab-/, ""));
      if (key) setActive(sections, key, nav);
    });
    window.addEventListener("message", function (event) {
      var data = event && event.data;
      if (!data || data.type !== "pcs-submenu-select") return;
      var key = findSectionKey(sections, String(data.key || ""));
      if (key) setActive(sections, key, nav);
    });
  }

  function init() {
    var preset = presets[currentFileName()];
    if (!preset) return;
    mountSubmenu(findSections(preset));
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
