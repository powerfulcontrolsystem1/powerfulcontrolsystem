(function () {
  "use strict";

  var DEFAULT_LABELS = {
    singular: "Estacion",
    plural: "Estaciones"
  };

  function cleanText(value) {
    return String(value == null ? "" : value).trim().replace(/\s+/g, " ");
  }

  function parsePositiveInt(raw) {
    var n = Number(String(raw || "").trim());
    if (!Number.isFinite(n)) return 0;
    n = Math.trunc(n);
    return n > 0 ? n : 0;
  }

  function resolveEmpresaId() {
    try {
      var params = new URLSearchParams(window.location.search || "");
      var own = parsePositiveInt(params.get("empresa_id") || params.get("id"));
      if (own > 0) return own;
    } catch (e) {}
    try {
      if (window.parent && window.parent !== window) {
        if (typeof window.parent.__resolveEmpresaIdContext === "function") {
          var parentResolved = parsePositiveInt(window.parent.__resolveEmpresaIdContext());
          if (parentResolved > 0) return parentResolved;
        }
        var parentParams = new URLSearchParams(window.parent.location.search || "");
        var fromParent = parsePositiveInt(parentParams.get("empresa_id") || parentParams.get("id"));
        if (fromParent > 0) return fromParent;
      }
    } catch (e) {}
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
    } catch (e) {}
    return 0;
  }

  function fallbackPlural(singular) {
    var clean = cleanText(singular) || DEFAULT_LABELS.singular;
    var lower = clean.toLowerCase();
    if (lower.endsWith("s")) return clean;
    if (/[aeiou]$/i.test(clean)) return clean + "s";
    return clean + "es";
  }

  function normalizeLabels(config) {
    var singular = cleanText(
      config && (
        config.singular ||
        config.estacion_nombre_singular ||
        config.nombre_estacion_singular ||
        config.tipo_recurso ||
        config.entidad_estacion_singular
      )
    ) || DEFAULT_LABELS.singular;
    var plural = cleanText(
      config && (
        config.plural ||
        config.estacion_nombre_plural ||
        config.nombre_estacion_plural ||
        config.tipo_recurso_plural ||
        config.entidad_estacion_plural
      )
    ) || fallbackPlural(singular);
    return {
      singular: singular,
      plural: plural,
      singularLower: singular.toLocaleLowerCase("es-CO"),
      pluralLower: plural.toLocaleLowerCase("es-CO")
    };
  }

  function parseConfigFromPrefs(items) {
    if (!Array.isArray(items)) return null;
    for (var i = 0; i < items.length; i += 1) {
      var item = items[i];
      if (!item || Number(item.estacion_id) !== 0 || String(item.clave || "") !== "estaciones_config") {
        continue;
      }
      try {
        var current = item.valor;
        for (var j = 0; j < 8; j += 1) {
          if (typeof current !== "string") break;
          current = JSON.parse(String(current || "{}"));
        }
        return current && typeof current === "object" ? current : null;
      } catch (e) {
        return null;
      }
    }
    return null;
  }

  async function loadLabels(empresaId) {
    empresaId = parsePositiveInt(empresaId || resolveEmpresaId());
    if (!empresaId) return normalizeLabels(null);
    var cacheKey = "pcs_estaciones_labels_" + empresaId;
    try {
      var cached = sessionStorage.getItem(cacheKey);
      if (cached) {
        var parsed = JSON.parse(cached);
        if (parsed && parsed.singular && parsed.plural) {
          return normalizeLabels(parsed);
        }
      }
    } catch (e) {}

    try {
      var resp = await fetch("/api/empresa/estacion_prefs?empresa_id=" + encodeURIComponent(empresaId), {
        credentials: "same-origin"
      });
      if (!resp.ok) return normalizeLabels(null);
      var data = await resp.json();
      var config = parseConfigFromPrefs(Array.isArray(data) ? data : (data && data.prefs ? data.prefs : []));
      var labels = normalizeLabels(config);
      try { sessionStorage.setItem(cacheKey, JSON.stringify(labels)); } catch (e) {}
      return labels;
    } catch (e) {
      return normalizeLabels(null);
    }
  }

  function replaceLabelText(text, labels) {
    if (!text || !labels) return text;
    var out = String(text);
    var replacements = [
      [/\bEstaciones\b/g, labels.plural],
      [/\bestaciones\b/g, labels.pluralLower],
      [/\bEstacion\b/g, labels.singular],
      [/\bestacion\b/g, labels.singularLower],
      [/\bEstación\b/g, labels.singular],
      [/\bestación\b/g, labels.singularLower]
    ];
    replacements.forEach(function (pair) {
      out = out.replace(pair[0], pair[1]);
    });
    return out;
  }

  function applyTemplate(el, labels) {
    var template = el.getAttribute("data-estacion-label-template") || "";
    if (!template) return false;
    var value = template
      .replace(/\{singular\}/g, labels.singular)
      .replace(/\{plural\}/g, labels.plural)
      .replace(/\{singularLower\}/g, labels.singularLower)
      .replace(/\{pluralLower\}/g, labels.pluralLower);
    el.textContent = value;
    return true;
  }

  function applyLabels(root, labels) {
    root = root || document.body;
    labels = normalizeLabels(labels);
    try {
      document.documentElement.setAttribute("data-estacion-singular", labels.singular);
      document.documentElement.setAttribute("data-estacion-plural", labels.plural);
    } catch (e) {}

    Array.prototype.slice.call(root.querySelectorAll("[data-estacion-label-template]")).forEach(function (el) {
      applyTemplate(el, labels);
    });

    if (root.getAttribute && root.getAttribute("data-auto-estacion-labels") !== "true" && document.body.getAttribute("data-auto-estacion-labels") !== "true") {
      return labels;
    }

    var walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT, {
      acceptNode: function (node) {
        if (!node || !node.nodeValue || !node.nodeValue.trim()) return NodeFilter.FILTER_REJECT;
        var parent = node.parentElement;
        if (!parent) return NodeFilter.FILTER_REJECT;
        var tag = parent.tagName ? parent.tagName.toLowerCase() : "";
        if (tag === "script" || tag === "style" || tag === "code" || tag === "pre" || tag === "textarea") {
          return NodeFilter.FILTER_REJECT;
        }
        return NodeFilter.FILTER_ACCEPT;
      }
    });
    var nodes = [];
    while (walker.nextNode()) nodes.push(walker.currentNode);
    nodes.forEach(function (node) {
      var source = node.__pcsEstacionLabelSource || node.nodeValue;
      node.__pcsEstacionLabelSource = source;
      node.nodeValue = replaceLabelText(source, labels);
    });

    Array.prototype.slice.call(root.querySelectorAll("[placeholder],[title],[aria-label]")).forEach(function (el) {
      ["placeholder", "title", "aria-label"].forEach(function (attr) {
        if (el.hasAttribute(attr)) {
          var sourceAttr = "data-pcs-estacion-label-" + attr.replace(/[^a-z0-9_-]/gi, "-") + "-source";
          var source = el.getAttribute(sourceAttr);
          if (!source) {
            source = el.getAttribute(attr);
            el.setAttribute(sourceAttr, source);
          }
          el.setAttribute(attr, replaceLabelText(source, labels));
        }
      });
    });
    return labels;
  }

  async function init(root) {
    var labels = await loadLabels();
    applyLabels(root || document.body, labels);
    return labels;
  }

  window.PCSEstacionLabels = {
    load: loadLabels,
    apply: applyLabels,
    init: init,
    normalize: normalizeLabels,
    replace: replaceLabelText
  };

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", function () { init(); });
  } else {
    init();
  }
})();
