(function () {
  "use strict";

  var params = new URLSearchParams(window.location.search || "");
  var empresaID = params.get("empresa_id") || localStorage.getItem("empresa_id") || localStorage.getItem("empresaId") || "";
  var state = {
    image: null,
    stream: null,
    analysisID: 0,
    reportURL: "",
    lastResult: null,
    crop: null,
    zoom: 1
  };

  function $(id) {
    return document.getElementById(id);
  }

  function api(action) {
    var url = "/api/empresa/grafologia?empresa_id=" + encodeURIComponent(empresaID || "0");
    if (action) url += "&action=" + encodeURIComponent(action);
    return url;
  }

  function init() {
    bindUpload();
    bindButtons();
    paintBlankCanvas();
    loadDashboard();
  }

  function bindUpload() {
    var file = $("grafologiaFile");
    var drop = $("grafologiaDrop");
    if (!file || !drop) return;
    drop.addEventListener("click", function () { file.click(); });
    drop.addEventListener("keydown", function (ev) {
      if (ev.key === "Enter" || ev.key === " ") {
        ev.preventDefault();
        file.click();
      }
    });
    file.addEventListener("change", function () {
      if (file.files && file.files[0]) loadImageFile(file.files[0]);
    });
    ["dragenter", "dragover"].forEach(function (name) {
      drop.addEventListener(name, function (ev) {
        ev.preventDefault();
        drop.classList.add("drag");
      });
    });
    ["dragleave", "drop"].forEach(function (name) {
      drop.addEventListener(name, function (ev) {
        ev.preventDefault();
        drop.classList.remove("drag");
      });
    });
    drop.addEventListener("drop", function (ev) {
      var dt = ev.dataTransfer;
      if (dt && dt.files && dt.files[0]) loadImageFile(dt.files[0]);
    });
  }

  function bindButtons() {
    $("btnGrafologiaRefresh")?.addEventListener("click", loadDashboard);
    $("btnGrafologiaAnalyze")?.addEventListener("click", analyzeCurrentCanvas);
    $("btnGrafologiaCamera")?.addEventListener("click", startCamera);
    $("btnGrafologiaCapture")?.addEventListener("click", captureCamera);
    $("btnGrafologiaCrop")?.addEventListener("click", cropCenter);
    $("btnGrafologiaPerspective")?.addEventListener("click", autoPerspective);
    $("grafologiaBrightness")?.addEventListener("input", renderCanvas);
    $("grafologiaContrast")?.addEventListener("input", renderCanvas);
    $("grafologiaZoom")?.addEventListener("input", function () { setZoom(numberValue("grafologiaZoom") / 100); });
    $("btnGrafologiaZoomOut")?.addEventListener("click", function () { setZoom(state.zoom - 0.1); });
    $("btnGrafologiaZoomIn")?.addEventListener("click", function () { setZoom(state.zoom + 0.1); });
    $("btnGrafologiaZoomReset")?.addEventListener("click", function () { setZoom(1); });
    $("btnGrafologiaHTML")?.addEventListener("click", function () { openReport("html"); });
    $("btnGrafologiaDOC")?.addEventListener("click", function () { openReport("doc"); });
    $("btnGrafologiaJSON")?.addEventListener("click", function () { openReport("json"); });
    $("btnGrafologiaCSV")?.addEventListener("click", function () { openReport("csv"); });
    $("btnGrafologiaTXT")?.addEventListener("click", function () { openReport("txt"); });
    $("btnGrafologiaPDF")?.addEventListener("click", function () { openReport("pdf"); });
  }

  function loadImageFile(file) {
    var reader = new FileReader();
    reader.onload = function () {
      var img = new Image();
      img.onload = function () {
        state.image = img;
        state.crop = null;
        setZoom(1, true);
        renderCanvas();
      };
      img.src = reader.result;
    };
    reader.readAsDataURL(file);
  }

  function paintBlankCanvas() {
    var canvas = $("grafologiaCanvas");
    if (!canvas) return;
    canvas.width = 900;
    canvas.height = 520;
    var ctx = canvas.getContext("2d");
    ctx.fillStyle = "#ffffff";
    ctx.fillRect(0, 0, canvas.width, canvas.height);
    ctx.fillStyle = "#64748b";
    ctx.font = "700 24px Arial";
    ctx.textAlign = "center";
    ctx.fillText("Vista previa del manuscrito", canvas.width / 2, canvas.height / 2);
  }

  function renderCanvas() {
    var img = state.image;
    var canvas = $("grafologiaCanvas");
    if (!img || !canvas) return;
    var maxW = 1200;
    var scale = Math.min(1, maxW / img.width);
    var source = state.crop || { x: 0, y: 0, w: img.width, h: img.height };
    scale = Math.min(1, maxW / source.w);
    var zoomedSource = zoomSource(source, state.zoom || 1);
    canvas.width = Math.max(320, Math.round(zoomedSource.w * scale));
    canvas.height = Math.max(220, Math.round(zoomedSource.h * scale));
    var ctx = canvas.getContext("2d");
    ctx.filter = "brightness(" + (100 + numberValue("grafologiaBrightness")) + "%) contrast(" + (100 + numberValue("grafologiaContrast")) + "%)";
    ctx.drawImage(img, zoomedSource.x, zoomedSource.y, zoomedSource.w, zoomedSource.h, 0, 0, canvas.width, canvas.height);
    ctx.filter = "none";
  }

  function zoomSource(source, zoom) {
    zoom = Math.max(0.5, Math.min(3, Number(zoom) || 1));
    if (Math.abs(zoom - 1) < 0.01) return source;
    var img = state.image;
    var nextW = Math.max(1, Math.round(source.w / zoom));
    var nextH = Math.max(1, Math.round(source.h / zoom));
    var centerX = source.x + source.w / 2;
    var centerY = source.y + source.h / 2;
    var x = Math.round(centerX - nextW / 2);
    var y = Math.round(centerY - nextH / 2);
    x = Math.max(0, Math.min(x, img.width - nextW));
    y = Math.max(0, Math.min(y, img.height - nextH));
    return { x: x, y: y, w: Math.min(nextW, img.width), h: Math.min(nextH, img.height) };
  }

  function setZoom(value, skipRender) {
    state.zoom = Math.max(0.5, Math.min(3, Number(value) || 1));
    var pct = Math.round(state.zoom * 100);
    var slider = $("grafologiaZoom");
    var label = $("grafologiaZoomValue");
    if (slider) slider.value = String(pct);
    if (label) label.textContent = pct + "%";
    if (!skipRender) renderCanvas();
  }

  function cropCenter() {
    if (!state.image) return;
    var img = state.image;
    state.crop = {
      x: Math.round(img.width * 0.05),
      y: Math.round(img.height * 0.05),
      w: Math.round(img.width * 0.90),
      h: Math.round(img.height * 0.90)
    };
    renderCanvas();
  }

  function autoPerspective() {
    if (!state.image) return;
    var previousBrightness = numberValue("grafologiaBrightness");
    var previousContrast = numberValue("grafologiaContrast");
    $("grafologiaBrightness").value = Math.max(previousBrightness, 8);
    $("grafologiaContrast").value = Math.max(previousContrast, 24);
    cropInkBoundingBox();
    renderCanvas();
  }

  function cropInkBoundingBox() {
    var canvas = document.createElement("canvas");
    var img = state.image;
    canvas.width = img.width;
    canvas.height = img.height;
    var ctx = canvas.getContext("2d", { willReadFrequently: true });
    ctx.drawImage(img, 0, 0);
    var data = ctx.getImageData(0, 0, canvas.width, canvas.height).data;
    var minX = canvas.width, minY = canvas.height, maxX = 0, maxY = 0, found = false;
    for (var y = 0; y < canvas.height; y += 2) {
      for (var x = 0; x < canvas.width; x += 2) {
        var i = (y * canvas.width + x) * 4;
        var lum = 0.299 * data[i] + 0.587 * data[i + 1] + 0.114 * data[i + 2];
        if (lum < 205) {
          found = true;
          if (x < minX) minX = x;
          if (y < minY) minY = y;
          if (x > maxX) maxX = x;
          if (y > maxY) maxY = y;
        }
      }
    }
    if (!found) return;
    var padX = Math.round(canvas.width * 0.04);
    var padY = Math.round(canvas.height * 0.04);
    state.crop = {
      x: Math.max(0, minX - padX),
      y: Math.max(0, minY - padY),
      w: Math.min(canvas.width, maxX - minX + padX * 2),
      h: Math.min(canvas.height, maxY - minY + padY * 2)
    };
  }

  function numberValue(id) {
    return Number($(id)?.value || 0) || 0;
  }

  async function startCamera() {
    var video = $("grafologiaVideo");
    if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia || !video) {
      alert("La cámara no está disponible en este navegador.");
      return;
    }
    try {
      state.stream = await navigator.mediaDevices.getUserMedia({ video: { facingMode: "environment" }, audio: false });
      video.srcObject = state.stream;
      video.style.display = "block";
      await video.play();
      $("btnGrafologiaCapture").disabled = false;
    } catch (err) {
      alert("No se pudo abrir la cámara: " + (err && err.message ? err.message : err));
    }
  }

  function captureCamera() {
    var video = $("grafologiaVideo");
    if (!video || !video.videoWidth) return;
    var canvas = document.createElement("canvas");
    canvas.width = video.videoWidth;
    canvas.height = video.videoHeight;
    canvas.getContext("2d").drawImage(video, 0, 0);
    var img = new Image();
    img.onload = function () {
      state.image = img;
      state.crop = null;
      setZoom(1, true);
      renderCanvas();
    };
    img.src = canvas.toDataURL("image/png");
  }

  async function analyzeCurrentCanvas() {
    if (!empresaID || empresaID === "0") {
      alert("No se detectó empresa_id para guardar el análisis.");
      return;
    }
    var canvas = $("grafologiaCanvas");
    if (!state.image || !canvas) {
      alert("Carga una imagen manuscrita antes de analizar.");
      return;
    }
    setLoading(true);
    try {
      var blob = await new Promise(function (resolve) { canvas.toBlob(resolve, "image/png", 0.95); });
      var form = new FormData();
      form.append("imagen", blob, "manuscrito.png");
      form.append("titulo", $("grafologiaTitle")?.value || "Informe grafológico GRAFOLOGIX");
      form.append("ocr_texto", $("grafologiaOCRText")?.value || "");
      var res = await fetch(api("analizar"), { method: "POST", body: form, credentials: "include" });
      if (res.status === 401 || res.status === 403) {
        throw new Error("Tu sesión no está activa o no tienes permiso para analizar manuscritos en esta empresa. Entra desde Administrar empresa y vuelve a abrir GRAFOLOGIX.");
      }
      if (!res.ok) throw new Error(cleanErrorMessage(await res.text()));
      var data = await res.json();
      state.analysisID = data.id || 0;
      state.reportURL = data.reporte_url || "";
      state.lastResult = data.resultado || null;
      renderResult(state.lastResult);
      await loadDashboard();
    } catch (err) {
      alert("No se pudo analizar el manuscrito: " + (err && err.message ? err.message : err));
    } finally {
      setLoading(false);
    }
  }

  async function loadDashboard() {
    if (!empresaID) return;
    try {
      var res = await fetch(api("dashboard") + "&limit=12", { credentials: "include" });
      if (!res.ok) throw new Error(await res.text());
      var data = await res.json();
      var items = data.items || [];
      $("grafKpiAnalisis").textContent = String(items.length || 0);
      if (items[0]) {
        $("grafKpiConfianza").textContent = fmtPct(items[0].confianza_global || 0);
      }
      renderHistory(items);
    } catch (err) {
      renderHistory([]);
    }
  }

  function renderResult(result) {
    if (!result) return;
    $("grafologiaEmpty").style.display = "none";
    $("grafologiaResult").style.display = "block";
    $("grafologiaSummary").textContent = result.summary || "";
    $("grafKpiConfianza").textContent = fmtPct(result.global_trust || 0);
    $("grafKpiLineas").textContent = String((result.image && result.image.lines_detected) || 0);
    var metrics = $("grafologiaMetrics");
    metrics.innerHTML = "";
    (result.metrics || []).forEach(function (m) {
      var tr = document.createElement("tr");
      tr.innerHTML = "<td><strong></strong><br><small></small></td><td></td><td></td>";
      tr.querySelector("strong").textContent = m.name || m.key || "";
      tr.querySelector("small").textContent = m.explanation || "";
      tr.children[1].textContent = m.value || m.category || "";
      tr.children[2].textContent = fmtPct(m.confidence || 0);
      metrics.appendChild(tr);
    });
    renderPreprocess(result.preprocess || null);
    var traits = $("grafologiaTraits");
    traits.innerHTML = "";
    (result.traits || []).forEach(function (t) {
      var div = document.createElement("div");
      div.className = "grafologia-bar";
      div.innerHTML = '<div class="grafologia-bar-head"><span></span><strong></strong></div><div class="grafologia-track"><div class="grafologia-fill"></div></div><p class="grafologia-warning"></p>';
      div.querySelector("span").textContent = t.name || t.key || "";
      div.querySelector("strong").textContent = fmtPct(t.percent || 0) + " · " + (t.level || "");
      div.querySelector(".grafologia-fill").style.width = Math.max(0, Math.min(100, Number(t.percent || 0))) + "%";
      div.querySelector("p").textContent = t.explanation || "";
      traits.appendChild(div);
    });
  }

  function renderPreprocess(preprocess) {
    var box = $("grafologiaPreprocess");
    renderQuality(preprocess);
    if (!box) return;
    box.innerHTML = "";
    var urls = preprocess && preprocess.image_urls ? preprocess.image_urls : {};
    var labels = {
      grayscale: "Escala de grises",
      binary: "Binarización",
      edges: "Bordes",
      lines: "Líneas y márgenes"
    };
    var keys = ["grayscale", "binary", "edges", "lines"];
    var hasAny = keys.some(function (key) { return !!urls[key]; });
    if (!hasAny) {
      box.innerHTML = '<p class="grafologia-warning">El preprocesamiento visual se mostrará en los nuevos análisis.</p>';
      return;
    }
    keys.forEach(function (key) {
      if (!urls[key]) return;
      var item = document.createElement("div");
      item.className = "grafologia-preprocess-item";
      item.innerHTML = "<strong></strong><img alt=\"\">";
      item.querySelector("strong").textContent = labels[key] || key;
      item.querySelector("img").src = urls[key];
      item.querySelector("img").alt = labels[key] || key;
      box.appendChild(item);
    });
  }

  function renderQuality(preprocess) {
    var box = $("grafologiaQuality");
    if (!box) return;
    box.innerHTML = "";
    if (!preprocess || !preprocess.quality) return;
    var q = preprocess.quality;
    var items = [
      ["Contraste", fmtPct(q.contrast || 0), q.lighting_warning ? "Revisar iluminación" : "Correcto"],
      ["Densidad de tinta", fmtPct((q.ink_density || 0) * 100), "Tinta detectada"],
      ["Nitidez estimada", fmtPct(q.sharpness || 0), q.resolution_warning ? "Mejorar resolución" : "Aceptable"],
      ["Umbral Otsu", String(preprocess.threshold || 0), q.crop_suggested ? "Recorte recomendado" : "Encuadre correcto"]
    ];
    items.forEach(function (row) {
      var div = document.createElement("div");
      div.className = "grafologia-bar";
      div.innerHTML = '<div class="grafologia-bar-head"><span></span><strong></strong></div><p class="grafologia-warning"></p>';
      div.querySelector("span").textContent = row[0];
      div.querySelector("strong").textContent = row[1];
      div.querySelector("p").textContent = row[2];
      box.appendChild(div);
    });
  }

  function renderHistory(items) {
    var box = $("grafologiaHistory");
    if (!box) return;
    box.innerHTML = "";
    if (!items.length) {
      box.innerHTML = '<p class="grafologia-warning">Aún no hay informes grafológicos guardados para esta empresa.</p>';
      return;
    }
    items.forEach(function (item) {
      var div = document.createElement("div");
      div.className = "grafologia-history-item";
      div.innerHTML = '<strong></strong><p></p><div class="grafologia-export"><button class="btn secondary" type="button">HTML</button><button class="btn secondary" type="button">Word</button><button class="btn secondary" type="button">JSON</button><button class="btn secondary" type="button">CSV</button><button class="btn secondary" type="button">TXT</button><button class="btn primary" type="button">PDF</button></div>';
      div.querySelector("strong").textContent = item.titulo || "Informe GRAFOLOGIX";
      div.querySelector("p").textContent = (item.fecha_creacion || "") + " · Confianza " + fmtPct(item.confianza_global || 0);
      var buttons = div.querySelectorAll("button");
      buttons[0].addEventListener("click", function () { window.open(reportUrl(item.id, "html"), "_blank", "noopener"); });
      buttons[1].addEventListener("click", function () { window.open(reportUrl(item.id, "doc"), "_blank", "noopener"); });
      buttons[2].addEventListener("click", function () { window.open(reportUrl(item.id, "json"), "_blank", "noopener"); });
      buttons[3].addEventListener("click", function () { window.open(reportUrl(item.id, "csv"), "_blank", "noopener"); });
      buttons[4].addEventListener("click", function () { window.open(reportUrl(item.id, "txt"), "_blank", "noopener"); });
      buttons[5].addEventListener("click", function () { window.open(reportUrl(item.id, "pdf"), "_blank", "noopener"); });
      box.appendChild(div);
    });
  }

  function openReport(format) {
    if (!state.analysisID) {
      alert("Primero genera un análisis.");
      return;
    }
    window.open(reportUrl(state.analysisID, format), "_blank", "noopener");
  }

  function reportUrl(id, format) {
    return api("reporte") + "&id=" + encodeURIComponent(id) + "&format=" + encodeURIComponent(format || "html");
  }

  function setLoading(active) {
    var loader = $("grafologiaLoader");
    if (loader) loader.classList.toggle("show", !!active);
    var btn = $("btnGrafologiaAnalyze");
    if (btn) btn.disabled = !!active;
  }

  function cleanErrorMessage(raw) {
    raw = String(raw || "").trim();
    if (!raw) return "No se pudo completar la operación.";
    try {
      var data = JSON.parse(raw);
      return data.error || data.message || raw;
    } catch (_) {
      return raw;
    }
  }

  function fmtPct(value) {
    return (Math.round(Number(value || 0) * 10) / 10).toLocaleString("es-CO") + "%";
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
