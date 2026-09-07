(function () {
      var current = new URL(window.location.href);
      var params = current.searchParams;
      var target = new URL('/pagar_licencia.html', window.location.origin);
      var status = params.get('x_response') || params.get('x_transaction_state') || params.get('response') || params.get('status') || '';
      var reference = params.get('x_ref_payco') || params.get('ref_payco') || params.get('reference') || '';
      var invoice = params.get('invoice') || params.get('x_id_invoice') || params.get('reference') || '';
      var transactionId = params.get('x_transaction_id') || params.get('transaction_id') || params.get('id') || '';
      var licenciaId = params.get('licencia_id') || params.get('extra1') || '';
      var empresaId = params.get('empresa_id') || params.get('extra2') || '';

      target.searchParams.set('provider', 'epayco');
      if (status) target.searchParams.set('status', String(status).toLowerCase());
      if (reference) target.searchParams.set('reference', reference);
      if (invoice) target.searchParams.set('invoice', invoice);
      if (transactionId) target.searchParams.set('transaction_id', transactionId);
      if (licenciaId) target.searchParams.set('licencia_id', licenciaId);
      if (empresaId) target.searchParams.set('empresa_id', empresaId);

      var continueLink = document.getElementById('epaycoContinueLink');
      if (continueLink) {
        continueLink.href = target.toString();
      }

      var msg = document.getElementById('epaycoResponseMessage');
      if (msg) {
        if (status) {
          msg.textContent = 'Epayco respondió con estado ' + String(status).toUpperCase() + '. Te llevaremos a validar el resultado real para dejar tu licencia activa.';
        } else {
          msg.textContent = 'Recibimos la respuesta de Epayco. Te llevaremos a validar el resultado real para dejar tu licencia activa.';
        }
      }

      window.setTimeout(function () {
        window.location.replace(target.toString());
      }, 1200);
    })();
