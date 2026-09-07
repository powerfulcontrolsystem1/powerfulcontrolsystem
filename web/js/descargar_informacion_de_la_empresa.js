(function () {
  var params = new URLSearchParams(window.location.search || "");
  var empresaID = parseInt(params.get("empresa_id") || params.get("id") || "0", 10);
  var statusEl = document.getElementById("empresaExportStatus");
  var titleEl = document.getElementById("empresaExportTitle");
  var subtitleEl = document.getElementById("empresaExportSubtitle");
  var formEl = document.getElementById("empresaExportForm");
  var formatEl = document.getElementById("empresaExportFormat");
  var downloadBtn = document.getElementById("empresaExportDownload");
  var backBtn = document.getElementById("empresaExportBack");
  var empresaName = "empresa";
  var isDownloading = false;

  if (params.get("embedded") === "1" && document.body) {
    document.body.classList.add("empresa-export-page--embedded");
  }

  function setStatus(message, type) {
    if (!statusEl) return;
    statusEl.textContent = message || "";
    statusEl.classList.toggle("is-error", type === "error");
    statusEl.classList.toggle("is-success", type === "success");
    statusEl.classList.toggle("is-muted", !type || type === "muted");
  }

  function setBusy(active) {
    isDownloading = !!active;
    if (downloadBtn) {
      downloadBtn.disabled = isDownloading || !empresaID;
      downloadBtn.classList.toggle("is-loading", isDownloading);
      downloadBtn.textContent = isDownloading ? "Descargando..." : "Descargar";
    }
    if (formatEl) {
      formatEl.disabled = isDownloading || !empresaID;
    }
  }

  function getFilenameFromHeaders(response, fallbackBaseName, format) {
    var disposition = response && response.headers ? response.headers.get("Content-Disposition") : "";
    var match = disposition && disposition.match(/filename\*=UTF-8''([^;]+)|filename="?([^";]+)"?/i);
    var raw = match ? (match[1] || match[2] || "") : "";
    if (raw) {
      try {
        return decodeURIComponent(raw);
      } catch (err) {
        return raw;
      }
    }
    return fallbackBaseName + "." + format;
  }

  function slugEmpresa() {
    return String(empresaName || "empresa")
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, "_")
      .replace(/^_+|_+$/g, "") || "empresa";
  }

  async function loadEmpresaName() {
    if (!empresaID) {
      if (titleEl) titleEl.textContent = "Falta empresa";
      if (subtitleEl) subtitleEl.textContent = "Abre esta pagina desde el selector de empresas para descargar una empresa concreta.";
      setStatus("No se encontro empresa_id en la URL.", "error");
      setBusy(false);
      return;
    }
    setStatus("Listo para descargar.", "muted");
    setBusy(false);
    try {
      var response = await fetch("/super/api/empresas?id=" + encodeURIComponent(String(empresaID)) + "&action=resumen_descarga", {
        credentials: "same-origin"
      });
      if (!response.ok) return;
      var data = await response.json();
      var empresa = data && data.snapshot && data.snapshot.empresa ? data.snapshot.empresa : {};
      if (empresa.nombre) {
        empresaName = empresa.nombre;
        if (titleEl) titleEl.textContent = "Informacion de " + empresa.nombre;
        if (subtitleEl) {
          subtitleEl.textContent = "Selecciona backup completo o un formato de exportacion para descargar la copia de datos de esta empresa.";
        }
      }
    } catch (err) {
      setStatus("Puedes intentar descargar. Si falla, vuelve al selector de empresas.", "muted");
    }
  }

  async function downloadSelectedFormat() {
    if (!empresaID || isDownloading) return;
    var format = formatEl && formatEl.value ? formatEl.value : "xls";
    setBusy(true);
    var label = format === "backup" ? "BACKUP" : String(format).toUpperCase();
    setStatus("Preparando archivo " + label + "...", "muted");
    try {
      var response = await fetch("/super/api/empresas?id=" + encodeURIComponent(String(empresaID)) + "&action=exportar_informacion&format=" + encodeURIComponent(format), {
        credentials: "same-origin"
      });
      if (!response.ok) {
        var errorText = await response.text();
        throw new Error(errorText || ("HTTP " + response.status));
      }
      var blob = await response.blob();
      var fileName = getFilenameFromHeaders(response, "empresa_" + slugEmpresa(), format);
      var blobURL = window.URL.createObjectURL(blob);
      var anchor = document.createElement("a");
      anchor.href = blobURL;
      anchor.download = fileName;
      document.body.appendChild(anchor);
      anchor.click();
      anchor.remove();
      window.URL.revokeObjectURL(blobURL);
      setStatus("Descarga lista: " + fileName, "success");
    } catch (err) {
      setStatus("No fue posible descargar: " + (err && err.message ? err.message : err), "error");
    } finally {
      setBusy(false);
    }
  }

  if (formEl) {
    formEl.addEventListener("submit", function (event) {
      event.preventDefault();
      downloadSelectedFormat();
    });
  }

  if (backBtn) {
    backBtn.addEventListener("click", function () {
      if (window.parent && window.parent !== window) {
        window.parent.postMessage({ type: "pcs:selector-show-empresas" }, window.location.origin);
        return;
      }
      window.location.href = "/seleccionar_empresa.html";
    });
  }

  loadEmpresaName();
})();
