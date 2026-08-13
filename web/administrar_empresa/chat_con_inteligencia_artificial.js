(function(){
      function setMsg(text, isError) {
        var el = document.getElementById('msg');
        if (!el) return;
        el.textContent = text || '';
        el.classList.toggle('value-negative', !!isError);
      }
      function openCentral() {
        try {
          if (window.parent && window.parent !== window) {
            window.parent.postMessage({
              type: 'pcs-ai-drawer-open',
              mode: 'operativo',
              prompt: 'Continúa tu consulta desde el asistente IA central de la empresa.'
            }, '*');
            setMsg('Solicitud enviada al asistente IA central.', false);
            return;
          }
        } catch (error) {}
        setMsg('No se pudo abrir el asistente central desde esta vista.', true);
      }
      document.getElementById('openCentralAI').addEventListener('click', openCentral);
      window.setTimeout(openCentral, 250);
    })();
