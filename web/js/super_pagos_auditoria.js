(function () {
  "use strict";

  var form = document.getElementById("paymentAuditFilters");
  var message = document.getElementById("auditMessage");
  var offset = 0;

  function text(value) {
    return value === null || value === undefined || value === "" ? "—" : String(value);
  }

  function statusNode(value) {
    var span = document.createElement("span");
    span.className = "audit-status";
    span.textContent = text(value);
    return span;
  }

  function cell(row, value, isStatus) {
    var td = document.createElement("td");
    if (isStatus) td.appendChild(statusNode(value));
    else td.textContent = text(value);
    row.appendChild(td);
  }

  function renderRows(targetID, rows, fields, statusFields) {
    var body = document.getElementById(targetID);
    body.replaceChildren();
    if (!rows.length) {
      var emptyRow = document.createElement("tr");
      var emptyCell = document.createElement("td");
      emptyCell.colSpan = fields.length;
      emptyCell.className = "audit-empty";
      emptyCell.textContent = "Sin registros para estos filtros";
      emptyRow.appendChild(emptyCell);
      body.appendChild(emptyRow);
      return;
    }
    rows.forEach(function (item) {
      var row = document.createElement("tr");
      fields.forEach(function (field) {
        cell(row, item[field], statusFields.indexOf(field) !== -1);
      });
      body.appendChild(row);
    });
  }

  function params() {
    var query = new URLSearchParams();
    new FormData(form).forEach(function (value, key) {
      value = String(value || "").trim();
      if (value) query.set(key, value);
    });
    query.set("offset", String(offset));
    return query;
  }

  async function loadAudit() {
    message.className = "audit-empty";
    message.textContent = "Consultando auditoría…";
    try {
      var response = await fetch("/super/api/pagos/auditoria?" + params().toString(), {
        credentials: "same-origin",
        headers: { "Accept": "application/json" }
      });
      if (!response.ok) throw new Error("HTTP " + response.status);
      var data = await response.json();
      var transactions = Array.isArray(data.transactions) ? data.transactions : [];
      var checkout = Array.isArray(data.checkout_attempts) ? data.checkout_attempts : [];
      var effects = Array.isArray(data.post_effects) ? data.post_effects : [];

      renderRows("auditTransactionsBody", transactions,
        ["updated_at", "provider", "empresa_id", "licencia_id", "reference", "transaction_id", "status", "activation_status", "activation_attempts"],
        ["status", "activation_status"]);
      renderRows("auditCheckoutBody", checkout,
        ["updated_at", "provider", "empresa_id", "reference", "status", "response_code"], ["status"]);
      renderRows("auditEffectsBody", effects,
        ["updated_at", "provider", "empresa_id", "licencia_id", "effect", "status", "attempts"], ["status"]);

      document.getElementById("auditTransactionsCount").textContent = String(transactions.length);
      document.getElementById("auditApprovedCount").textContent = String(transactions.filter(function (item) {
        return String(item.status || "").toUpperCase() === "APPROVED";
      }).length);
      document.getElementById("auditCheckoutCount").textContent = String(checkout.length);
      document.getElementById("auditEffectsAlertCount").textContent = String(effects.filter(function (item) {
        var state = String(item.status || "").toLowerCase();
        return state !== "completado";
      }).length);
      document.getElementById("auditPrevious").disabled = offset === 0;
      document.getElementById("auditNext").disabled = transactions.length < Number(document.getElementById("auditLimit").value || 50);
      message.textContent = "Auditoría actualizada.";
    } catch (error) {
      message.className = "audit-error";
      message.textContent = "No fue posible consultar la auditoría de pagos. Verifica la sesión y el estado del servicio.";
    }
  }

  form.addEventListener("submit", function (event) {
    event.preventDefault();
    offset = 0;
    loadAudit();
  });
  document.getElementById("auditPrevious").addEventListener("click", function () {
    offset = Math.max(0, offset - Number(document.getElementById("auditLimit").value || 50));
    loadAudit();
  });
  document.getElementById("auditNext").addEventListener("click", function () {
    offset += Number(document.getElementById("auditLimit").value || 50);
    loadAudit();
  });
  loadAudit();
})();
