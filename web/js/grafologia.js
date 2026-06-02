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
    clientes: [],
    selectedCliente: null,
    clienteSearchTimer: 0,
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
    window.addEventListener("resize", function () {
      if (state.image) renderCanvas();
      else paintBlankCanvas();
    });
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
    $("btnGrafologiaAnalyzeAI")?.addEventListener("click", analyzeCurrentCanvasWithGPT55);
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
    $("btnGrafologiaBuscarCliente")?.addEventListener("click", searchClientes);
    $("btnGrafologiaNuevoCliente")?.addEventListener("click", function () { toggleClienteForm(true); });
    $("btnGrafologiaCancelarCliente")?.addEventListener("click", function () { toggleClienteForm(false); });
    $("btnGrafologiaCrearCliente")?.addEventListener("click", createClienteCentral);
    $("grafologiaClienteSearch")?.addEventListener("input", function () {
      clearTimeout(state.clienteSearchTimer);
      state.clienteSearchTimer = setTimeout(searchClientes, 280);
    });
  }

  function clientesApi(q) {
    var url = "/api/empresa/clientes?empresa_id=" + encodeURIComponent(empresaID || "0");
    if (q) url += "&q=" + encodeURIComponent(q);
    return url;
  }

  async function searchClientes() {
    var box = $("grafologiaClienteSugerencias");
    var input = $("grafologiaClienteSearch");
    if (!box || !input || !empresaID) return;
    var q = (input.value || "").trim();
    if (q.length < 2) {
      box.innerHTML = '<p class="grafologia-warning">Escribe al menos 2 caracteres para buscar clientes.</p>';
      return;
    }
    box.innerHTML = '<p class="grafologia-warning">Buscando clientes...</p>';
    try {
      var res = await fetch(clientesApi(q), { credentials: "include" });
      if (!res.ok) throw new Error(cleanErrorMessage(await res.text()));
      state.clientes = await res.json();
      renderClienteSuggestions(state.clientes || []);
    } catch (err) {
      box.innerHTML = '<p class="grafologia-warning">No se pudieron buscar clientes: ' + escapeHTML(err && err.message ? err.message : err) + '</p>';
    }
  }

  function renderClienteSuggestions(items) {
    var box = $("grafologiaClienteSugerencias");
    if (!box) return;
    box.innerHTML = "";
    if (!items.length) {
      box.innerHTML = '<p class="grafologia-warning">No se encontro el cliente. Puedes crearlo con el boton Nuevo cliente.</p>';
      return;
    }
    items.slice(0, 12).forEach(function (cliente) {
      var btn = document.createElement("button");
      btn.type = "button";
      btn.className = "grafologia-client-option";
      btn.innerHTML = '<span><strong></strong><br><small></small></span><span class="btn secondary" aria-hidden="true">Asociar</span>';
      btn.querySelector("strong").textContent = cliente.nombre_razon_social || "Cliente sin nombre";
      btn.querySelector("small").textContent = clienteLabel(cliente);
      btn.addEventListener("click", function () { selectCliente(cliente); });
      box.appendChild(btn);
    });
  }

  function selectCliente(cliente) {
    state.selectedCliente = cliente || null;
    renderSelectedCliente();
    var input = $("grafologiaClienteSearch");
    if (input && cliente) input.value = cliente.nombre_razon_social || "";
    var box = $("grafologiaClienteSugerencias");
    if (box) box.innerHTML = "";
    toggleClienteForm(false);
  }

  function renderSelectedCliente() {
    var box = $("grafologiaClienteSeleccionado");
    if (!box) return;
    if (!state.selectedCliente) {
      box.classList.remove("show");
      box.innerHTML = "";
      return;
    }
    var cliente = state.selectedCliente;
    box.classList.add("show");
    box.innerHTML = '<strong></strong><p class="grafologia-warning"></p><div class="grafologia-actions"><button class="btn secondary" type="button">Quitar cliente</button></div>';
    box.querySelector("strong").textContent = cliente.nombre_razon_social || "Cliente asociado";
    box.querySelector("p").textContent = clienteLabel(cliente);
    box.querySelector("button").addEventListener("click", function () {
      state.selectedCliente = null;
      var input = $("grafologiaClienteSearch");
      if (input) input.value = "";
      renderSelectedCliente();
    });
  }

  function toggleClienteForm(show) {
    var form = $("grafologiaClienteForm");
    if (form) form.classList.toggle("show", !!show);
  }

  async function createClienteCentral() {
    if (!empresaID || empresaID === "0") {
      alert("No se detecto empresa_id para crear el cliente.");
      return;
    }
    var nombre = ($("grafologiaNuevoClienteNombre")?.value || "").trim();
    var numero = ($("grafologiaNuevoClienteDocumento")?.value || "").trim();
    if (!nombre || !numero) {
      alert("Nombre y numero de documento son obligatorios para crear el cliente central.");
      return;
    }
    var payload = {
      empresa_id: Number(empresaID),
      tipo_documento: ($("grafologiaNuevoClienteTipo")?.value || "CC").trim(),
      numero_documento: numero,
      tipo_persona: "natural",
      nombre_razon_social: nombre,
      email: ($("grafologiaNuevoClienteEmail")?.value || "").trim(),
      telefono: ($("grafologiaNuevoClienteTelefono")?.value || "").trim(),
      pais: ($("grafologiaNuevoClientePais")?.value || "CO").trim() || "CO",
      observaciones: ($("grafologiaPersonaDescripcion")?.value || "").trim()
    };
    try {
      var res = await fetch("/api/empresa/clientes", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify(payload)
      });
      if (!res.ok) throw new Error(cleanErrorMessage(await res.text()));
      var data = await res.json();
      var cliente = Object.assign({}, payload, { id: data.id || data.cliente_id || 0, estado: "activo" });
      selectCliente(cliente);
      alert("Cliente creado y asociado al manuscrito.");
    } catch (err) {
      alert("No se pudo crear el cliente: " + (err && err.message ? err.message : err));
    }
  }

  function clienteLabel(cliente) {
    if (!cliente) return "";
    var doc = [cliente.tipo_documento, cliente.numero_documento].filter(Boolean).join(" ");
    var parts = [doc, cliente.telefono, cliente.email].filter(Boolean);
    return parts.join(" · ");
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
    var viewport = prepareCanvasViewport(canvas);
    var ctx = viewport.ctx;
    ctx.fillStyle = "#ffffff";
    ctx.fillRect(0, 0, viewport.width, viewport.height);
    ctx.fillStyle = "#64748b";
    ctx.font = "700 24px Arial";
    ctx.textAlign = "center";
    ctx.fillText("Vista previa del manuscrito", viewport.width / 2, viewport.height / 2);
    ctx.setTransform(1, 0, 0, 1, 0, 0);
  }

  function renderCanvas() {
    var img = state.image;
    var canvas = $("grafologiaCanvas");
    if (!img || !canvas) return;
    var source = state.crop || { x: 0, y: 0, w: img.width, h: img.height };
    var viewport = prepareCanvasViewport(canvas);
    var zoomedSource = zoomSourceForViewport(source, state.zoom || 1, viewport.width / viewport.height);
    var ctx = viewport.ctx;
    ctx.fillStyle = "#ffffff";
    ctx.fillRect(0, 0, viewport.width, viewport.height);
    ctx.filter = "brightness(" + (100 + numberValue("grafologiaBrightness")) + "%) contrast(" + (100 + numberValue("grafologiaContrast")) + "%)";
    ctx.drawImage(img, zoomedSource.x, zoomedSource.y, zoomedSource.w, zoomedSource.h, 0, 0, viewport.width, viewport.height);
    ctx.filter = "none";
    ctx.setTransform(1, 0, 0, 1, 0, 0);
  }

  function prepareCanvasViewport(canvas) {
    var wrap = canvas.parentElement;
    var width = Math.max(320, Math.round((wrap && wrap.clientWidth) || canvas.clientWidth || 900));
    var height = Math.max(260, Math.round((wrap && wrap.clientHeight) || canvas.clientHeight || 520));
    var dpr = Math.max(1, Math.min(2, window.devicePixelRatio || 1));
    canvas.width = Math.round(width * dpr);
    canvas.height = Math.round(height * dpr);
    canvas.style.width = width + "px";
    canvas.style.height = height + "px";
    var ctx = canvas.getContext("2d");
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    return { ctx: ctx, width: width, height: height };
  }

  function zoomSourceForViewport(source, zoom, viewportRatio) {
    var cropped = zoomSource(source, zoom);
    viewportRatio = viewportRatio > 0 ? viewportRatio : cropped.w / Math.max(1, cropped.h);
    var next = { x: cropped.x, y: cropped.y, w: cropped.w, h: cropped.h };
    var currentRatio = next.w / Math.max(1, next.h);
    if (currentRatio > viewportRatio) {
      var nextW = Math.max(1, Math.round(next.h * viewportRatio));
      next.x = Math.round(next.x + (next.w - nextW) / 2);
      next.w = nextW;
    } else if (currentRatio < viewportRatio) {
      var nextH = Math.max(1, Math.round(next.w / viewportRatio));
      next.y = Math.round(next.y + (next.h - nextH) / 2);
      next.h = nextH;
    }
    return clampZoomSource(next);
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

  function clampZoomSource(source) {
    var img = state.image;
    if (!img) return source;
    var w = Math.max(1, Math.min(Math.round(source.w), img.width));
    var h = Math.max(1, Math.min(Math.round(source.h), img.height));
    var x = Math.max(0, Math.min(Math.round(source.x), img.width - w));
    var y = Math.max(0, Math.min(Math.round(source.y), img.height - h));
    return { x: x, y: y, w: w, h: h };
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
      var blob = await canvasToOptimizedBlob(canvas);
      var form = new FormData();
      form.append("imagen", blob, "manuscrito.jpg");
      form.append("titulo", $("grafologiaTitle")?.value || "Informe grafológico GRAFOLOGIX");
      form.append("ocr_texto", $("grafologiaOCRText")?.value || "");
      form.append("cliente_id", state.selectedCliente && state.selectedCliente.id ? String(state.selectedCliente.id) : "");
      form.append("persona_nombre", state.selectedCliente ? (state.selectedCliente.nombre_razon_social || "") : "");
      form.append("cliente_documento", state.selectedCliente ? clienteLabel(state.selectedCliente) : "");
      form.append("persona_descripcion", $("grafologiaPersonaDescripcion")?.value || "");
      form.append("persona_caracteristicas", $("grafologiaPersonaCaracteristicas")?.value || "");
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

  async function analyzeCurrentCanvasWithGPT55() {
    if (!empresaID || empresaID === "0") {
      alert("No se detecto empresa_id para analizar con GPT-5.5.");
      return;
    }
    var canvas = $("grafologiaCanvas");
    if (!state.image || !canvas) {
      alert("Carga una imagen manuscrita antes de analizar con GPT-5.5.");
      return;
    }
    renderAIResult(null, true);
    setLoading(true);
    try {
      var blob = await canvasToOptimizedBlob(canvas);
      var form = new FormData();
      form.append("imagen", blob, "manuscrito-gpt55.jpg");
      form.append("titulo", $("grafologiaTitle")?.value || "Informe grafologico GRAFOLOGIX");
      form.append("ocr_texto", $("grafologiaOCRText")?.value || "");
      form.append("cliente_id", state.selectedCliente && state.selectedCliente.id ? String(state.selectedCliente.id) : "");
      form.append("persona_nombre", state.selectedCliente ? (state.selectedCliente.nombre_razon_social || "") : "");
      form.append("cliente_documento", state.selectedCliente ? clienteLabel(state.selectedCliente) : "");
      form.append("persona_descripcion", $("grafologiaPersonaDescripcion")?.value || "");
      form.append("persona_caracteristicas", $("grafologiaPersonaCaracteristicas")?.value || "");
      var res = await fetch(api("analizar_ia"), { method: "POST", body: form, credentials: "include" });
      if (res.status === 401 || res.status === 403) {
        throw new Error("Tu sesion no esta activa o no tienes permiso para usar GPT-5.5 en GRAFOLOGIX.");
      }
      if (!res.ok) throw new Error(cleanErrorMessage(await res.text()));
      var data = await res.json();
      renderAIResult(data, false);
    } catch (err) {
      renderAIResult({
        error: err && err.message ? err.message : String(err || "No se pudo completar el analisis GPT-5.5.")
      }, false);
      alert("No se pudo analizar con GPT-5.5: " + (err && err.message ? err.message : err));
    } finally {
      setLoading(false);
    }
  }

  function renderAIResult(data, loading) {
    var box = $("grafologiaAIResult");
    if (!box) return;
    box.classList.add("show");
    $("grafologiaEmpty").style.display = "none";
    if (loading) {
      box.innerHTML = '<div class="grafologia-ai-meta"><span>GPT-5.5</span><span>Analizando manuscrito...</span></div><div class="grafologia-ai-response">Procesando imagen con el modelo de vision configurado en el sistema.</div>';
      return;
    }
    if (!data || data.error) {
      box.innerHTML = '<div class="grafologia-ai-meta"><span>GPT-5.5</span><span>No completado</span></div><div class="grafologia-ai-response"></div>';
      box.querySelector(".grafologia-ai-response").textContent = data && data.error ? data.error : "No se recibio respuesta del analisis GPT-5.5.";
      return;
    }
    var usage = data.usage || {};
    var meta = [
      data.display_name || data.model_id || "GPT-5.5",
      data.upstream_model ? "Modelo " + data.upstream_model : "",
      usage.daily_limit ? "Uso diario " + (usage.daily_used || 0) + "/" + usage.daily_limit : "",
      usage.daily_remaining !== undefined ? "Restantes " + usage.daily_remaining : ""
    ].filter(Boolean);
    box.innerHTML = '<div class="grafologia-ai-meta"></div><div class="grafologia-ai-response"></div>';
    box.querySelector(".grafologia-ai-meta").textContent = meta.join(" · ");
    box.querySelector(".grafologia-ai-response").textContent = data.respuesta || "GPT-5.5 no devolvio contenido.";
  }

  async function canvasToOptimizedBlob(canvas) {
    var maxDim = 1600;
    var ratio = Math.min(1, maxDim / Math.max(canvas.width || 1, canvas.height || 1));
    var target = canvas;
    if (ratio < 1) {
      target = document.createElement("canvas");
      target.width = Math.max(1, Math.round(canvas.width * ratio));
      target.height = Math.max(1, Math.round(canvas.height * ratio));
      var tctx = target.getContext("2d");
      tctx.imageSmoothingEnabled = true;
      tctx.imageSmoothingQuality = "high";
      tctx.fillStyle = "#ffffff";
      tctx.fillRect(0, 0, target.width, target.height);
      tctx.drawImage(canvas, 0, 0, target.width, target.height);
    }
    return new Promise(function (resolve, reject) {
      target.toBlob(function (blob) {
        if (blob && blob.size > 0) {
          resolve(blob);
        } else {
          reject(new Error("No se pudo preparar la imagen optimizada para enviar."));
        }
      }, "image/jpeg", 0.82);
    });
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
    renderResultSubject(result.subject || null);
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

  function renderResultSubject(subject) {
    var box = $("grafologiaSubject");
    if (!box) return;
    if (!subject || (!subject.cliente_nombre && !subject.persona_descripcion && !subject.persona_caracteristicas)) {
      box.classList.remove("show");
      box.innerHTML = "";
      return;
    }
    box.classList.add("show");
    box.innerHTML = '<strong></strong><p class="grafologia-warning"></p>';
    box.querySelector("strong").textContent = subject.cliente_nombre || "Persona asociada";
    var lines = [];
    if (subject.cliente_documento) lines.push(subject.cliente_documento);
    if (subject.persona_descripcion) lines.push(subject.persona_descripcion);
    if (subject.persona_caracteristicas) lines.push(subject.persona_caracteristicas);
    box.querySelector("p").textContent = lines.join(" · ");
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
      div.querySelector("p").textContent = [item.fecha_creacion || "", item.cliente_nombre ? "Cliente " + item.cliente_nombre : "", "Confianza " + fmtPct(item.confianza_global || 0)].filter(Boolean).join(" · ");
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
    var btnAI = $("btnGrafologiaAnalyzeAI");
    if (btnAI) btnAI.disabled = !!active;
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

  function escapeHTML(value) {
    return String(value || "").replace(/[&<>"']/g, function (ch) {
      return ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[ch] || ch;
    });
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
