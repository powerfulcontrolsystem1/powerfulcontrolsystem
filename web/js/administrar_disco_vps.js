(function () {
    var endpoint = "/super/api/administrar_disco_vps";
    var candidatesEl = document.getElementById("candidates");
    var emptyEl = document.getElementById("emptyState");
    var confirmationEl = document.getElementById("confirmation");
    var cleanupBtn = document.getElementById("cleanupBtn");
    var refreshBtn = document.getElementById("refreshBtn");
    var selectAllBtn = document.getElementById("selectAllBtn");
    var statusEl = document.getElementById("status");
    var state = { categories: [] };

    function bytes(value) { value = Number(value || 0); if (!value) return "0 B"; var units=["B","KB","MB","GB","TB"], index=0; while(value>=1024 && index<units.length-1){ value/=1024; index++; } return (value>=10 || index===0 ? value.toFixed(0) : value.toFixed(1)) + " " + units[index]; }
    function selected() { return Array.prototype.slice.call(candidatesEl.querySelectorAll("input[type=checkbox]:checked")).map(function (item) { return item.value; }); }
    function updateAction() { cleanupBtn.disabled = confirmationEl.value.trim() !== "LIBERAR ESPACIO" || selected().length === 0; }
    function setStatus(message) { statusEl.textContent = message || ""; }
    function render(data) {
      state.categories = Array.isArray(data.categories) ? data.categories : [];
      var disk = data.disk || {};
      document.getElementById("totalValue").textContent = bytes(disk.total_bytes);
      document.getElementById("usedValue").textContent = bytes(disk.used_bytes);
      document.getElementById("freeValue").textContent = bytes(disk.free_bytes);
      candidatesEl.innerHTML = "";
      state.categories.forEach(function (item) {
        var row = document.createElement("label"); row.className = "candidate";
        var count = Number(item.candidate_count || 0);
        var estimate = item.estimate_known ? bytes(item.estimated_bytes) + " estimados" : (count ? "tamaño se calculará al limpiar" : "sin elementos");
        row.innerHTML = '<input type="checkbox" value="' + escapeHTML(item.id) + '"' + (count ? '' : ' disabled') + '><span><strong></strong><small></small></span><b></b>';
        row.querySelector("strong").textContent = item.title || "Categoría segura";
        row.querySelector("small").textContent = item.description || "";
        row.querySelector("b").textContent = count + " elemento(s) · " + estimate;
        candidatesEl.appendChild(row);
      });
      emptyEl.classList.toggle("hidden", state.categories.some(function (item) { return Number(item.candidate_count || 0) > 0; }));
      updateAction();
    }
    function escapeHTML(value) { var node=document.createElement("span"); node.textContent=String(value || ""); return node.innerHTML; }
    function request(method, body) { return fetch(endpoint, { method:method, credentials:"same-origin", headers: body ? {"Content-Type":"application/json"} : {}, body: body ? JSON.stringify(body) : undefined }).then(function(response){ return response.json().catch(function(){ return {}; }).then(function(data){ if(!response.ok || !data.ok){ throw new Error(data.error || "No se pudo completar la operación"); } return data; }); }); }
    function load(message) { refreshBtn.disabled = true; setStatus("Consultando recursos recuperables..."); request("GET").then(function(data){ render(data); setStatus(message || "Actualizado."); }).catch(function(error){ setStatus(error.message); }).finally(function(){ refreshBtn.disabled = false; }); }
    refreshBtn.addEventListener("click", load);
    confirmationEl.addEventListener("input", updateAction);
    candidatesEl.addEventListener("change", updateAction);
    selectAllBtn.addEventListener("click", function(){ Array.prototype.slice.call(candidatesEl.querySelectorAll("input[type=checkbox]:not(:disabled)")).forEach(function(item){ item.checked=true; }); updateAction(); });
    cleanupBtn.addEventListener("click", function(){ var categories=selected(); if(!categories.length || confirmationEl.value.trim()!=="LIBERAR ESPACIO") return; cleanupBtn.disabled=true; refreshBtn.disabled=true; setStatus("Liberando únicamente recursos recuperables..."); request("POST", { action:"cleanup", categories:categories, confirmation:"LIBERAR ESPACIO" }).then(function(data){ confirmationEl.value=""; load("Limpieza terminada. Espacio liberado: " + bytes(data.freed_bytes) + "."); }).catch(function(error){ setStatus(error.message); }).finally(function(){ refreshBtn.disabled=false; updateAction(); }); });
    load();
  })();
