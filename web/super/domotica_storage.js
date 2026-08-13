(function(){
      'use strict';
      var state = { empresas: [] };
      function $(id){ return document.getElementById(id); }
      function esc(v){ return String(v == null ? '' : v).replace(/[&<>"']/g, function(c){ return {'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]; }); }
      function setMsg(text, type){ var box = $('msg'); box.textContent = text || ''; box.className = 'form-help' + (type ? ' ' + type : ''); }
      async function requestJSON(url, options){
        var res = await fetch(url, Object.assign({ credentials: 'same-origin' }, options || {}));
        var ct = res.headers.get('content-type') || '';
        var data = ct.indexOf('application/json') >= 0 ? await res.json() : await res.text();
        if (!res.ok) throw new Error(typeof data === 'string' ? data : (data.error || 'Error'));
        return data;
      }
      function render(){
        var rows = $('rows');
        if (!state.empresas.length) {
          rows.innerHTML = '<tr><td colspan="6">Sin empresas.</td></tr>';
          return;
        }
        rows.innerHTML = state.empresas.map(function(e){
          return '<tr data-empresa-id="' + esc(e.empresa_id) + '">' +
            '<td><strong>' + esc(e.nombre || ('Empresa ' + e.empresa_id)) + '</strong><br><small>ID ' + esc(e.empresa_id) + '</small></td>' +
            '<td><code>' + esc(e.folder || '') + '</code><br><small>' + esc(e.public_path || '') + '</small></td>' +
            '<td>' + esc(e.imagenes || 0) + '</td>' +
            '<td>' + esc(e.usado_mb || '0.00') + '</td>' +
            '<td><input class="form-input max-kb" type="number" min="128" max="20480" step="128" value="' + esc(e.max_image_kb || 2048) + '"></td>' +
            '<td><button class="btn secondary small save-company" type="button">Guardar</button></td>' +
          '</tr>';
        }).join('');
      }
      async function load(){
        setMsg('Cargando almacenamiento...');
        var data = await requestJSON('/super/api/domotica_storage');
        $('defaultMaxKb').value = data.default_max_image_kb || 2048;
        state.empresas = Array.isArray(data.empresas) ? data.empresas : [];
        render();
        setMsg('');
      }
      async function saveDefault(){
        await requestJSON('/super/api/domotica_storage', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ default_max_image_kb: Number($('defaultMaxKb').value || 2048) })
        });
        setMsg('Limite general guardado.', 'success');
        await load();
      }
      async function saveCompany(row){
        await requestJSON('/super/api/domotica_storage', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ empresa_id: Number(row.getAttribute('data-empresa-id') || 0), max_image_kb: Number(row.querySelector('.max-kb').value || 2048) })
        });
        setMsg('Limite de empresa guardado.', 'success');
        await load();
      }
      document.addEventListener('click', function(ev){
        var row = ev.target.closest('tr[data-empresa-id]');
        if (ev.target.id === 'reloadBtn') load().catch(function(err){ setMsg(err.message, 'error'); });
        if (ev.target.id === 'saveDefaultBtn') saveDefault().catch(function(err){ setMsg(err.message, 'error'); });
        if (row && ev.target.classList.contains('save-company')) saveCompany(row).catch(function(err){ setMsg(err.message, 'error'); });
      });
      load().catch(function(err){ setMsg(err.message, 'error'); });
    }());
