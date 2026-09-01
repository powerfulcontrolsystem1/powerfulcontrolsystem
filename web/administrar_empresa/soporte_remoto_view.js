(function () {
      var q = new URLSearchParams(window.location.search || '');
      var empresaID = Number(q.get('empresa_id') || q.get('id') || 0);
      var codigoSesion = String(q.get('codigo_sesion') || '').trim();
      var token = String(q.get('token') || '').trim();

      var msgEl = document.getElementById('srvMsg');
      var metaEl = document.getElementById('srvMeta');
      var frameEl = document.getElementById('srvFrame');

      function setMsg(text, type) {
        msgEl.textContent = text || '';
        msgEl.classList.remove('error', 'success');
        if (type) msgEl.classList.add(type);
      }

      function escapeHTML(value) {
        return String(value == null ? '' : value)
          .replace(/&/g, '&amp;')
          .replace(/</g, '&lt;')
          .replace(/>/g, '&gt;')
          .replace(/"/g, '&quot;')
          .replace(/'/g, '&#39;');
      }

      function renderMeta(session, embedURL, allowed, reason) {
        var row = session || {};
        var html = '';
        html += '<table class="table">';
        html += '<tbody>';
        html += '<tr><th>Empresa</th><td>' + Number(row.empresa_id || empresaID || 0) + '</td></tr>';
        html += '<tr><th>Código sesión</th><td>' + escapeHTML(row.codigo_sesion || codigoSesion || '-') + '</td></tr>';
        html += '<tr><th>Estado</th><td>' + escapeHTML(row.estado_sesion || '-') + '</td></tr>';
        html += '<tr><th>Dispositivo</th><td>' + escapeHTML(row.dispositivo_nombre || '-') + ' [' + escapeHTML(row.dispositivo_codigo || '-') + ']</td></tr>';
        html += '<tr><th>Operador</th><td>' + escapeHTML(row.operador_nombre || '-') + '</td></tr>';
        html += '<tr><th>Expira</th><td>' + escapeHTML(row.expira_en || '-') + '</td></tr>';
        html += '<tr><th>Acceso permitido</th><td>' + (allowed ? 'Sí' : 'No') + '</td></tr>';
        html += '<tr><th>Motivo bloqueo</th><td>' + escapeHTML(reason || '-') + '</td></tr>';
        html += '<tr><th>URL embebible</th><td>' + escapeHTML(embedURL || '-') + '</td></tr>';
        html += '</tbody>';
        html += '</table>';
        metaEl.innerHTML = html;
      }

      async function resolveAndLoad() {
        if (!empresaID || !codigoSesion || !token) {
          setMsg('Faltan parámetros obligatorios: empresa_id, codigo_sesion, token.', 'error');
          renderMeta(null, '', false, 'parametros faltantes');
          return;
        }
        setMsg('Resolviendo visualización segura de sesión...');

        try {
          var url = new URL('/api/empresa/soporte_remoto', window.location.origin);
          url.searchParams.set('empresa_id', String(empresaID));
          url.searchParams.set('action', 'resolver_visualizacion');
          url.searchParams.set('codigo_sesion', codigoSesion);
          url.searchParams.set('token', token);

          var response = await fetch(url.toString(), { method: 'GET', credentials: 'same-origin' });
          var text = await response.text();
          var data = {};
          if (text) {
            try { data = JSON.parse(text); } catch (_) { data = {}; }
          }
          if (!response.ok) {
            throw new Error((data && (data.error || data.message)) || text || ('HTTP ' + response.status));
          }

          var allowed = !!data.acceso_permitido;
          var embedURL = String(data.embed_url || '').trim();
          var reason = String(data.motivo_bloqueo || '').trim();
          renderMeta(data.session || {}, embedURL, allowed, reason);

          if (!allowed) {
            frameEl.src = 'about:blank';
            setMsg('Sesión válida, pero no habilitada para visualizar: ' + (reason || 'sin motivo reportado'), 'error');
            return;
          }
          if (!embedURL) {
            frameEl.src = 'about:blank';
            setMsg('No hay URL embebible configurada para este dispositivo.', 'error');
            return;
          }

          frameEl.src = embedURL;
          setMsg('Visor remoto cargado correctamente.', 'success');
        } catch (err) {
          frameEl.src = 'about:blank';
          renderMeta(null, '', false, err.message);
          setMsg('No se pudo resolver la sesión remota: ' + err.message, 'error');
        }
      }

      resolveAndLoad();
    })();
