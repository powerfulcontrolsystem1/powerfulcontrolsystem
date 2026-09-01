const state = { limit: 100, offset: 0 };
    const listContainer = document.getElementById('historialList');

    function getEmpresaId() {
      try {
        const params = new URLSearchParams(window.location.search || '');
        const own = Number(params.get('empresa_id') || params.get('id') || 0);
        if (own > 0) return Math.trunc(own);
      } catch (_) {}
      try {
        if (window.parent && window.parent !== window && typeof window.parent.__resolveEmpresaIdContext === 'function') {
          const parentID = Number(window.parent.__resolveEmpresaIdContext() || 0);
          if (parentID > 0) return Math.trunc(parentID);
        }
      } catch (_) {}
      try {
        const stored = Number(sessionStorage.getItem('active_empresa_id') || sessionStorage.getItem('empresa_id') || localStorage.getItem('active_empresa_id') || localStorage.getItem('empresa_id') || 0);
        if (stored > 0) return Math.trunc(stored);
      } catch (_) {}
      return 0;
    }

    function buildEmpresaURL(endpoint) {
      const url = new URL(endpoint, window.location.origin);
      const empresaID = getEmpresaId();
      if (empresaID > 0 && !url.searchParams.has('empresa_id')) {
        url.searchParams.set('empresa_id', String(empresaID));
      }
      return url;
    }

    async function loadHistorial() {
      try {
        listContainer.innerHTML = '<p class="empty-state">Cargando...</p>';
        const url = buildEmpresaURL('/api/empresa/productos/precios_historial');
        url.searchParams.set('limit', String(state.limit));
        url.searchParams.set('offset', String(state.offset));
        const res = await fetch(url.pathname + url.search, { credentials: 'same-origin' });
        if (!res.ok) throw new Error(await res.text());
        const data = await res.json();
        renderHistorial(data || []);
      } catch (err) {
        listContainer.innerHTML = `<p class="error-msg">Error cargando historial: ${err.message}</p>`;
      }
    }

    function exportHistorial(formato) {
      const url = buildEmpresaURL('/api/empresa/productos/precios_historial');
      url.searchParams.set('limit', '5000');
      url.searchParams.set('offset', '0');
      url.searchParams.set('formato', formato);
      const a = document.createElement('a');
      a.href = url.pathname + url.search;
      a.target = '_blank';
      a.rel = 'noopener';
      document.body.appendChild(a);
      a.click();
      a.remove();
    }

    function renderHistorial(items) {
      if (!items.length) {
        listContainer.innerHTML = '<p class="empty-state">No hay registros en el historial de precios.</p>';
        return;
      }

      listContainer.innerHTML = '';
      items.forEach(h => {
        const itemDiv = document.createElement('div');
        itemDiv.className = 'historial-item';

        const formatMoney = v => `$${Number(v).toFixed(2)}`;
        let changesHTML = '';

        if (h.precio_anterior !== h.precio_nuevo) {
          changesHTML += `<div><span class="amount-label">Precio:</span> <span class="amount-old">${formatMoney(h.precio_anterior)}</span> → <span class="amount-new">${formatMoney(h.precio_nuevo)}</span></div>`;
        }
        if (h.costo_anterior !== h.costo_nuevo) {
            changesHTML += `<div><span class="amount-label">Costo:</span> <span class="amount-old">${formatMoney(h.costo_anterior)}</span> → <span class="amount-new">${formatMoney(h.costo_nuevo)}</span></div>`;
        }
        if (h.impuesto_anterior !== h.impuesto_nuevo) {
            changesHTML += `<div><span class="amount-label">Impuesto:</span> <span class="amount-old">${h.impuesto_anterior}%</span> → <span class="amount-new">${h.impuesto_nuevo}%</span></div>`;
        }
        if (!changesHTML) {
            changesHTML = `<div class="muted empty-note-italic">Registro inicial o sin cambios directos</div>`;
        }

        const dateStr = h.fecha_cambio ? new Date(h.fecha_cambio).toLocaleString() : 'Fecha desc.';

        itemDiv.innerHTML = `
          <div class="historial-date">${dateStr}</div>
          <div>
            <div class="historial-product-name">${h.producto_nombre || 'Producto ID: ' + h.producto_id}</div>
            <div class="historial-reason">${h.motivo || 'Motivo desconocido'}</div>
          </div>
          <div class="historial-changes">${changesHTML}</div>
          <div class="user-info">
            <strong>Usuario:</strong> ${h.usuario_creador || 'Sistema'}<br>
            <strong>Ref:</strong> ${h.referencia || 'N/A'}<br>
            <strong>Estado:</strong> ${h.estado || 'activo'}
          </div>
        `;
        listContainer.appendChild(itemDiv);
      });
    }

    document.getElementById('btnRefresh').addEventListener('click', loadHistorial);
    document.querySelectorAll('[data-export-format]').forEach(btn => {
      btn.addEventListener('click', () => exportHistorial(btn.dataset.exportFormat || 'csv'));
    });
    document.addEventListener('DOMContentLoaded', loadHistorial);
