(function(){
      var target = new URL('/administrar_empresa/administrar_productos.html', window.location.origin);
      target.searchParams.set('view', 'bodegas');
      try {
        var params = new URLSearchParams(window.location.search || '');
        params.forEach(function(value, key){
          if (key !== 'view') target.searchParams.set(key, value);
        });
      } catch (_) {}
      window.location.replace(target.pathname + target.search);
    })();
