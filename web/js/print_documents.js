(function(global) {
  'use strict';

  function text(value) {
    return String(value == null ? '' : value);
  }

  function escapeHTML(value) {
    return text(value).replace(/[&<>"']/g, function(ch) {
      return {
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;',
        '"': '&quot;',
        "'": '&#39;'
      }[ch];
    });
  }

  function normalizeFormat(value) {
    var v = text(value).trim().toLowerCase();
    return v === 'carta' || v === 'letter' ? 'carta' : 'pos';
  }

  function normalizeKind(value) {
    var v = text(value).trim().toLowerCase();
    if (v === 'factura' || v === 'invoice') return 'factura';
    if (v === 'comprobante' || v === 'voucher') return 'comprobante';
    if (v === 'orden' || v === 'orden_servicio') return 'orden';
    if (v === 'reporte' || v === 'report') return 'reporte';
    return 'recibo';
  }

  function clampFontSize(value, fallback, min, max) {
    var n = Number(value);
    var base = Number(fallback);
    if (!Number.isFinite(base)) base = min;
    base = Math.max(min, Math.min(max, Math.trunc(base)));
    if (!Number.isFinite(n)) return base;
    n = Math.trunc(n);
    if (n < min) return min;
    if (n > max) return max;
    return n;
  }

  function resolvePrintFontSize(format, kind, options) {
    format = normalizeFormat(format);
    kind = normalizeKind(kind);
    options = options || {};
    var cfg = options.printConfig || options.config || options.companyConfig || {};
    if (Object.prototype.hasOwnProperty.call(options, 'printFontSize')) {
      return clampFontSize(options.printFontSize, format === 'pos' ? 11 : 13, format === 'pos' ? 8 : 10, format === 'pos' ? 16 : 22);
    }
    if (kind === 'reporte') {
      return format === 'pos'
        ? clampFontSize(cfg.impresion_reporte_fuente_pos, 11, 8, 16)
        : clampFontSize(cfg.impresion_reporte_fuente_carta, 13, 10, 22);
    }
    return format === 'pos'
      ? clampFontSize(cfg.impresion_factura_fuente_pos, 11, 8, 16)
      : clampFontSize(cfg.impresion_factura_fuente_carta, 13, 10, 22);
  }

  var DEFAULT_PRINT_ITEMS = {
    empresa: true,
    carrito: true,
    codigo: true,
    numero_legal: true,
    fecha: true,
    estado: true,
    cliente: true,
    cliente_email: true,
    cliente_documento: true,
    cajero: true,
    metodo_pago: true,
    total: true,
    moneda: true,
    periodo: true,
    control_documental: true,
    tipo_documento: true,
    codigo_validacion: true,
    pais: true,
    ambiente: true,
    observaciones: true,
    notas_legales: true,
    qr_dian: true,
    total_en_letras: false,
    formato: true,
    impresora: true,
    copias: true,
    campo_personalizado: false,
    campo_personalizado_etiqueta: 'Domicilio',
    campo_personalizado_valor: '',
    campo_personalizado_descripcion_visible: true,
    campo_personalizado_descripcion: ''
  };

  var ELECTRONIC_REQUIRED_PRINT_KEYS = {
    empresa: true,
    codigo: true,
    numero_legal: true,
    fecha: true,
    estado: true,
    cliente: true,
    cliente_documento: true,
    total: true,
    moneda: true,
    control_documental: true,
    tipo_documento: true,
    codigo_validacion: true,
    pais: true,
    ambiente: true,
    qr_dian: true
  };

  function parsePrintItemsJSON(raw, defaults) {
    var base = Object.assign({}, DEFAULT_PRINT_ITEMS, defaults || {});
    if (raw && typeof raw === 'object' && !Array.isArray(raw)) {
      raw = JSON.stringify(raw);
    }
    var body = text(raw).trim();
    if (!body) return base;
    try {
      var parsed = JSON.parse(body);
      if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) return base;
      Object.keys(parsed).forEach(function(key) {
        if (!Object.prototype.hasOwnProperty.call(base, key)) return;
        if (typeof base[key] === 'boolean') {
          base[key] = parsed[key] !== false;
        } else if (typeof base[key] === 'string') {
          base[key] = text(parsed[key]).trim();
        } else {
          base[key] = parsed[key];
        }
      });
    } catch (e) {}
    return base;
  }

  function printItemsFromConfig(config, defaults) {
    return parsePrintItemsJSON(config && config.impresion_recibo_items_json, defaults);
  }

  function isElectronicDocument(item) {
    var tipo = text(item && item.tipo_documento).trim().toLowerCase();
    return tipo === 'factura_electronica' ||
      tipo === 'nota_credito' ||
      tipo === 'nota_debito' ||
      tipo === 'documento_soporte' ||
      tipo === 'nomina_electronica' ||
      tipo === 'documento_equivalente_pos';
  }

  function printItemEnabled(items, key, item) {
    if (isElectronicDocument(item) && ELECTRONIC_REQUIRED_PRINT_KEYS[key]) return true;
    if (!items || typeof items !== 'object') return true;
    return items[key] !== false;
  }

  function numberToSpanishInteger(value) {
    var n = Math.floor(Math.abs(Number(value) || 0));
    var units = ['cero','uno','dos','tres','cuatro','cinco','seis','siete','ocho','nueve','diez','once','doce','trece','catorce','quince','dieciseis','diecisiete','dieciocho','diecinueve'];
    var tens = ['', '', 'veinte', 'treinta', 'cuarenta', 'cincuenta', 'sesenta', 'setenta', 'ochenta', 'noventa'];
    var hundreds = ['', 'ciento', 'doscientos', 'trescientos', 'cuatrocientos', 'quinientos', 'seiscientos', 'setecientos', 'ochocientos', 'novecientos'];
    function belowHundred(x) {
      if (x < 20) return units[x];
      if (x === 20) return 'veinte';
      if (x < 30) return 'veinti' + units[x - 20];
      var t = Math.floor(x / 10);
      var u = x % 10;
      return tens[t] + (u ? ' y ' + units[u] : '');
    }
    function belowThousand(x) {
      if (x < 100) return belowHundred(x);
      if (x === 100) return 'cien';
      var h = Math.floor(x / 100);
      var r = x % 100;
      return hundreds[h] + (r ? ' ' + belowHundred(r) : '');
    }
    function chunk(x) {
      if (x < 1000) return belowThousand(x);
      if (x < 1000000) {
        var th = Math.floor(x / 1000);
        var rem = x % 1000;
        return (th === 1 ? 'mil' : belowThousand(th) + ' mil') + (rem ? ' ' + belowThousand(rem) : '');
      }
      if (x < 1000000000000) {
        var mill = Math.floor(x / 1000000);
        var rest = x % 1000000;
        return (mill === 1 ? 'un millon' : chunk(mill) + ' millones') + (rest ? ' ' + chunk(rest) : '');
      }
      var bill = Math.floor(x / 1000000000000);
      var tail = x % 1000000000000;
      return (bill === 1 ? 'un billon' : chunk(bill) + ' billones') + (tail ? ' ' + chunk(tail) : '');
    }
    return n === 0 ? 'cero' : chunk(n).replace(/\buno mil\b/g, 'un mil');
  }

  function amountToWords(amount, currency) {
    var value = Number(amount || 0);
    var cur = text(currency || 'COP').trim().toUpperCase() || 'COP';
    var sign = value < 0 ? 'menos ' : '';
    var entero = Math.floor(Math.abs(value));
    var cents = Math.round((Math.abs(value) - entero) * 100);
    var currencyName = cur === 'COP' ? (entero === 1 ? 'peso colombiano' : 'pesos colombianos') : cur;
    var out = sign + numberToSpanishInteger(entero) + ' ' + currencyName;
    if (cents > 0) out += ' con ' + numberToSpanishInteger(cents) + ' centavos';
    if (cur === 'COP' && cents === 0) out += ' m/cte';
    return out.toUpperCase();
  }

  function safeLogoURL(value) {
    var raw = text(value).trim();
    var lower = raw.toLowerCase();
    if (!raw) return '';
    if (raw.charAt(0) === '/' || lower.indexOf('https://') === 0 || lower.indexOf('http://') === 0) return raw;
    return '';
  }

  function safeQRImageURL(value) {
    var raw = text(value).trim();
    var lower = raw.toLowerCase();
    if (!raw) return '';
    if (lower.indexOf('data:image/') === 0 || raw.charAt(0) === '/' || lower.indexOf('https://') === 0 || lower.indexOf('http://') === 0) return raw;
    return '';
  }

  function resolveDocumentLogos(config, kind) {
    var cfg = config || {};
    var documentKind = normalizeKind(kind);
    var logos = [];
    var companyLogo = documentKind === 'factura'
      ? safeLogoURL(cfg.logo_factura_url || cfg.logo_url)
      : safeLogoURL(cfg.logo_url);
    var systemLogo = safeLogoURL(cfg.logo_sistema_url || '/img/logo.png');
    var showCompany = documentKind === 'factura'
      ? cfg.mostrar_logo_factura !== false && cfg.mostrar_logo !== false
      : cfg.mostrar_logo_empresa !== false && cfg.mostrar_logo !== false;
    var showSystem = cfg.mostrar_logo_sistema === true;
    if (showCompany && companyLogo) {
      logos.push({ src: companyLogo, alt: 'Logo empresa' });
    }
    if (showSystem && systemLogo) {
      logos.push({ src: systemLogo, alt: 'Logo sistema' });
    }
    return logos;
  }

  function logosHTML(logos) {
    var list = Array.isArray(logos) ? logos : [];
    var html = list.map(function(item) {
      var src = safeLogoURL(item && item.src);
      if (!src) return '';
      return '<img class="pcs-print-logo" src="' + escapeHTML(src) + '" alt="' + escapeHTML((item && item.alt) || 'Logo') + '">';
    }).join('');
    return html ? '<div class="pcs-print-logos">' + html + '</div>' : '';
  }

  function documentCSS(format, kind, options) {
    format = normalizeFormat(format);
    kind = normalizeKind(kind);
    var accent = '#111827';
    var fontSize = resolvePrintFontSize(format, kind, options || {});
    var smallFont = Math.max(7, fontSize - 2);
    var metaFont = Math.max(8, fontSize - 1);
    var titleFont = format === 'pos' ? fontSize + 3 : fontSize + 9;
    var totalFont = format === 'pos' ? fontSize + 2 : fontSize + 3;
    var page = format === 'pos'
      ? '@page{size:80mm auto;margin:3mm;}html,body{width:74mm;max-width:74mm;}'
      : '@page{size:letter;margin:12mm;}html,body{max-width:190mm;}';
    var base = [
      page,
      'html,body{margin:0 auto;background:#fff;color:#111827;font-family:Arial,Helvetica,sans-serif;-webkit-print-color-adjust:exact;print-color-adjust:exact;}',
      'body{box-sizing:border-box;}*{box-sizing:border-box;}',
      '.pcs-print-doc{width:100%;margin:0 auto;background:#fff;color:#111827;overflow-wrap:anywhere;word-break:break-word;}',
      '.pcs-print-head{display:flex;align-items:flex-start;justify-content:space-between;gap:12px;border-bottom:2px solid ' + accent + ';padding-bottom:10px;margin-bottom:12px;}',
      '.pcs-print-logos{display:flex;align-items:center;gap:8px;flex-wrap:wrap;max-width:220px;}',
      '.pcs-print-logo{max-width:100px;max-height:54px;object-fit:contain;display:block;filter:grayscale(1) contrast(1.08);}',
      '.pcs-print-brand h1{margin:0 0 4px;font-size:22px;line-height:1.1;color:#111827;}',
      '.pcs-print-brand p{margin:2px 0;color:#374151;font-size:12px;line-height:1.35;}',
      '.pcs-print-badge{border:1px solid ' + accent + ';color:' + accent + ';border-radius:8px;padding:6px 9px;font-weight:900;font-size:12px;text-transform:uppercase;white-space:nowrap;}',
      '.pcs-print-meta{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:8px;margin:10px 0 12px;}',
      '.pcs-print-box{border:1px solid #d1d5db;border-radius:8px;padding:9px;background:#fff;}',
      '.pcs-print-box span{display:block;color:#374151;font-size:10px;text-transform:uppercase;font-weight:800;letter-spacing:.04em;margin-bottom:3px;}',
      '.pcs-print-box strong{display:block;color:#111827;font-size:13px;line-height:1.25;overflow-wrap:anywhere;word-break:break-word;}',
      '.pcs-print-table{width:100%;border-collapse:collapse;margin:10px 0;font-size:12px;}',
      '.pcs-print-table th{color:#111827;text-align:left;font-size:11px;text-transform:uppercase;border-bottom:1px solid #111827;padding:6px 4px;}',
      '.pcs-print-table td{border-bottom:1px solid #d1d5db;padding:6px 4px;vertical-align:top;}',
      '.pcs-print-number{text-align:right;white-space:nowrap;}',
      '.pcs-print-total{display:flex;justify-content:space-between;gap:12px;border:2px solid ' + accent + ';border-radius:8px;padding:10px 12px;margin-top:10px;font-size:16px;font-weight:900;}',
      '.pcs-print-summary{margin-top:8px;border-top:1px solid #d1d5db;border-bottom:1px solid #d1d5db;}',
      '.pcs-print-summary-row{display:flex;justify-content:space-between;gap:12px;padding:5px 0;border-bottom:1px dashed #d1d5db;font-size:12px;}',
      '.pcs-print-summary-row:last-child{border-bottom:0;}',
      '.pcs-print-summary-row span{color:#374151;font-weight:700;}',
      '.pcs-print-summary-row strong{color:#111827;text-align:right;white-space:nowrap;}',
      '.pcs-print-note{margin-top:10px;padding:9px;border:1px dashed #6b7280;border-radius:8px;color:#111827;font-size:12px;line-height:1.4;white-space:pre-wrap;overflow-wrap:anywhere;word-break:break-word;}',
      '.pcs-print-qr{margin-top:12px;text-align:center;border-top:1px dashed #9ca3af;padding-top:10px;color:#111827;}',
      '.pcs-print-qr img{display:block;width:126px;height:126px;object-fit:contain;margin:0 auto 6px;image-rendering:pixelated;}',
      '.pcs-print-qr strong{display:block;font-size:12px;line-height:1.25;margin-bottom:4px;}',
      '.pcs-print-qr span{display:block;font-size:10px;color:#374151;line-height:1.25;overflow-wrap:anywhere;word-break:break-word;}',
      '.pcs-print-footer{margin-top:12px;color:#374151;font-size:11px;text-align:center;line-height:1.35;overflow-wrap:anywhere;word-break:break-word;}',
      '.pcs-print-signatures{display:grid;grid-template-columns:1fr 1fr;gap:20px;margin-top:30px;}',
      '.pcs-print-signatures div{border-top:1px solid #111827;padding-top:6px;text-align:center;color:#374151;font-size:12px;}',
      '@media print{.pcs-print-no-print{display:none!important;}}'
    ].join('');
    if (format === 'pos') {
      base += [
        'body{padding:0;font-size:11px;line-height:1.28;}',
        '.pcs-print-doc{font-family:"Courier New",monospace;}',
        '.pcs-print-head{display:block;text-align:center;border-bottom:1px dashed #111827;padding-bottom:7px;margin-bottom:8px;}',
        '.pcs-print-logos{justify-content:center;max-width:100%;margin:0 auto 6px;}',
        '.pcs-print-logo{max-width:80px;max-height:42px;}',
        '.pcs-print-brand h1{font-size:14px;margin-bottom:4px;}',
        '.pcs-print-brand p{font-size:10px;color:#111827;}',
        '.pcs-print-badge{display:inline-block;margin-top:5px;border-style:dashed;border-radius:0;padding:3px 6px;font-size:10px;}',
        '.pcs-print-meta{grid-template-columns:1fr;gap:4px;margin:8px 0;}',
        '.pcs-print-box{border:0;border-bottom:1px dashed #9ca3af;border-radius:0;background:#fff;padding:4px 0;}',
        '.pcs-print-box span{font-size:9px;color:#374151;margin:0;}',
        '.pcs-print-box strong{font-size:11px;}',
        '.pcs-print-table{font-size:10px;margin:7px 0;}',
        '.pcs-print-table th,.pcs-print-table td{border-bottom:1px dashed #9ca3af;padding:4px 2px;}',
        '.pcs-print-total{border:1px dashed #111827;border-radius:0;padding:7px 0;font-size:13px;}',
        '.pcs-print-summary{margin-top:6px;border-top:1px dashed #9ca3af;border-bottom:1px dashed #9ca3af;}',
        '.pcs-print-summary-row{font-size:10px;padding:3px 0;border-bottom:1px dashed #d1d5db;}',
        '.pcs-print-summary-row strong{font-size:10px;}',
        '.pcs-print-note{border:0;border-top:1px dashed #9ca3af;border-radius:0;padding:7px 0;font-size:10px;}',
        '.pcs-print-qr{margin-top:8px;padding-top:7px;}',
        '.pcs-print-qr img{width:92px;height:92px;margin-bottom:4px;}',
        '.pcs-print-qr strong{font-size:10px;}',
        '.pcs-print-qr span{font-size:8.5px;}',
        '.pcs-print-footer{border-top:1px dashed #9ca3af;padding-top:7px;font-size:10px;}',
        '.pcs-print-signatures{display:none;}'
      ].join('');
    } else {
      base += 'body{padding:0;font-size:13px;}.pcs-print-doc{border:1px solid #111827;border-radius:0;padding:18px;box-shadow:none;}';
    }
    base += [
      'body{font-size:' + fontSize + 'px;}',
      '.pcs-print-brand h1{font-size:' + titleFont + 'px;}',
      '.pcs-print-brand p,.pcs-print-footer,.pcs-print-signatures div{font-size:' + metaFont + 'px;}',
      '.pcs-print-badge,.pcs-print-table,.pcs-print-summary-row,.pcs-print-note,.pcs-print-qr strong{font-size:' + metaFont + 'px;}',
      '.pcs-print-box span,.pcs-print-table th,.pcs-print-qr span{font-size:' + smallFont + 'px;}',
      '.pcs-print-box strong{font-size:' + fontSize + 'px;}',
      '.pcs-print-total{font-size:' + totalFont + 'px;}',
      '.pcs-print-summary-row strong{font-size:' + metaFont + 'px;}'
    ].join('');
    return base;
  }

  function rowsHTML(rows) {
    return (Array.isArray(rows) ? rows : []).map(function(row) {
      return '<tr>' + row.map(function(cell, idx) {
        var value = typeof cell === 'object' && cell ? cell.value : cell;
        var hasNumberFlag = typeof cell === 'object' && cell && Object.prototype.hasOwnProperty.call(cell, 'number');
        var cls = hasNumberFlag ? (cell.number ? ' class="pcs-print-number"' : '') : (idx > 0 ? ' class="pcs-print-number"' : '');
        return '<td' + cls + '>' + escapeHTML(value) + '</td>';
      }).join('') + '</tr>';
    }).join('');
  }

  function metaHTML(meta) {
    return (Array.isArray(meta) ? meta : []).filter(function(item) {
      return item && text(item.value).trim() !== '';
    }).map(function(item) {
      return '<section class="pcs-print-box"><span>' + escapeHTML(item.label) + '</span><strong>' + escapeHTML(item.value) + '</strong></section>';
    }).join('');
  }

  function qrHTML(qr) {
    if (!qr || typeof qr !== 'object') return '';
    var src = safeQRImageURL(qr.src || qr.image || qr.imageUrl || qr.dataURL || '');
    var label = text(qr.label || 'Consultar documento electronico');
    var value = text(qr.value || qr.code || '');
    var url = text(qr.url || qr.href || '').trim();
    var caption = text(qr.caption || '').trim();
    if (!src && !url && !value) return '';
    return '<section class="pcs-print-qr">' +
      (src ? '<img src="' + escapeHTML(src) + '" alt="' + escapeHTML(label) + '">' : '') +
      '<strong>' + escapeHTML(label) + '</strong>' +
      (value ? '<span>' + escapeHTML(value) + '</span>' : '') +
      (url ? '<span>' + escapeHTML(url) + '</span>' : '') +
      (caption ? '<span>' + escapeHTML(caption) + '</span>' : '') +
      '</section>';
  }

  function summaryRowsHTML(rows) {
    return (Array.isArray(rows) ? rows : []).filter(function(row) {
      return row && text(row.value).trim() !== '';
    }).map(function(row) {
      return '<div class="pcs-print-summary-row"><span>' + escapeHTML(row.label || '') + '</span><strong>' + escapeHTML(row.value || '') + '</strong></div>';
    }).join('');
  }

  function buildDocument(options) {
    options = options || {};
    var format = normalizeFormat(options.format);
    var kind = normalizeKind(options.kind);
    var title = text(options.title || 'Documento');
    var company = text(options.company || '');
    var subtitle = text(options.subtitle || '');
    var badge = text(options.badge || (format === 'pos' ? 'POS' : 'Carta'));
    var headerLogos = Array.isArray(options.logos) ? options.logos : (options.logoUrl ? [{ src: options.logoUrl, alt: options.logoAlt || 'Logo' }] : []);
    var tableHeaders = Array.isArray(options.tableHeaders) ? options.tableHeaders : [];
    var table = '';
    if (tableHeaders.length || (Array.isArray(options.rows) && options.rows.length)) {
      table = '<table class="pcs-print-table"><thead><tr>' + tableHeaders.map(function(h, idx) {
        return '<th' + (idx > 0 ? ' class="pcs-print-number"' : '') + '>' + escapeHTML(h) + '</th>';
      }).join('') + '</tr></thead><tbody>' + rowsHTML(options.rows) + '</tbody></table>';
    }
    var body = text(options.bodyHTML || '');
    var qrBlock = qrHTML(options.qr);
    if (!body) {
      body = '<section class="pcs-print-meta">' + metaHTML(options.meta) + '</section>' + table;
      if (text(options.totalLabel || options.totalValue).trim()) {
        body += '<section class="pcs-print-total"><span>' + escapeHTML(options.totalLabel || 'Total') + '</span><span>' + escapeHTML(options.totalValue || '') + '</span></section>';
      }
      if (text(options.totalWords).trim()) {
        body += '<section class="pcs-print-note"><strong>Total en letras:</strong> ' + escapeHTML(options.totalWords) + '</section>';
      }
      var summary = summaryRowsHTML(options.summaryRows);
      if (summary) body += '<section class="pcs-print-summary">' + summary + '</section>';
      if (text(options.note).trim()) body += '<section class="pcs-print-note">' + escapeHTML(options.note) + '</section>';
      body += qrBlock;
      if (options.signatures !== false && format === 'carta') body += '<section class="pcs-print-signatures"><div>Recibe</div><div>Entrega / registra</div></section>';
    } else if (qrBlock) {
      body += qrBlock;
    }
    var auto = options.autoPrint === false ? '' : '<script>window.addEventListener("load",function(){setTimeout(function(){try{window.focus();window.print();' + (options.closeAfterPrint === false ? '' : 'window.close();') + '}catch(e){}},180);});<\/script>';
    return '<!doctype html><html lang="es"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>' + escapeHTML(title) + '</title><style>' + documentCSS(format, kind, options) + '</style></head><body><article class="pcs-print-doc pcs-print-' + escapeHTML(format) + ' pcs-print-' + escapeHTML(kind) + '"><header class="pcs-print-head">' + logosHTML(headerLogos) + '<div class="pcs-print-brand"><h1>' + escapeHTML(title) + '</h1>' + (company ? '<p><strong>' + escapeHTML(company) + '</strong></p>' : '') + (subtitle ? '<p>' + escapeHTML(subtitle) + '</p>' : '') + '</div><div class="pcs-print-badge">' + escapeHTML(badge) + '</div></header>' + body + (text(options.footer).trim() ? '<footer class="pcs-print-footer">' + escapeHTML(options.footer) + '</footer>' : '') + '</article>' + auto + '</body></html>';
  }

  function openDocument(options) {
    options = options || {};
    var format = normalizeFormat(options.format);
    var features = format === 'pos' ? 'width=340,height=760' : 'width=860,height=940';
    var win = global.open('', '_blank', options.features || features);
    if (!win) return null;
    win.document.open();
    win.document.write(buildDocument(options));
    win.document.close();
    return win;
  }

  function buildSharePayload(options) {
    options = options || {};
    var title = text(options.title || options.documentTitle || global.document && global.document.title || 'Documento');
    var code = text(options.code || options.documentCode || '').trim();
    var total = text(options.total || '').trim();
    var company = text(options.company || '').trim();
    var url = text(options.url || '').trim();
    if (!url) {
      try {
        url = global.location && global.location.href ? global.location.href : '';
      } catch (e) {
        url = '';
      }
    }
    var parts = [title];
    if (code) parts.push('Codigo: ' + code);
    if (company) parts.push('Empresa: ' + company);
    if (total) parts.push('Total: ' + total);
    if (text(options.message).trim()) parts.push(text(options.message).trim());
    if (url) parts.push('Enlace: ' + url);
    return {
      title: title,
      text: parts.join('\n'),
      url: url
    };
  }

  function openShareChannel(channel, payload) {
    var target = String(channel || '').toLowerCase();
    var subject = encodeURIComponent(payload.title || 'Documento');
    var body = encodeURIComponent(payload.text || payload.url || '');
    var href = '';
    if (target === 'whatsapp' || target === 'wa') {
      href = 'https://wa.me/?text=' + body;
    } else {
      href = 'mailto:?subject=' + subject + '&body=' + body;
    }
    try {
      global.open(href, '_blank', 'noopener,noreferrer');
    } catch (e) {
      try { global.location.href = href; } catch (_) {}
    }
  }

  function shareDocument(options) {
    options = options || {};
    var channel = String(options.channel || '').toLowerCase();
    var payload = buildSharePayload(options);
    if (!channel && global.navigator && typeof global.navigator.share === 'function') {
      return global.navigator.share(payload).catch(function() {
        openShareChannel('email', payload);
      });
    }
    openShareChannel(channel || 'email', payload);
    return Promise.resolve();
  }

  global.PCSPrint = {
    escapeHTML: escapeHTML,
    normalizeFormat: normalizeFormat,
    DEFAULT_PRINT_ITEMS: DEFAULT_PRINT_ITEMS,
    parsePrintItemsJSON: parsePrintItemsJSON,
    printItemsFromConfig: printItemsFromConfig,
    resolvePrintFontSize: resolvePrintFontSize,
    printItemEnabled: printItemEnabled,
    isElectronicDocument: isElectronicDocument,
    amountToWords: amountToWords,
    resolveDocumentLogos: resolveDocumentLogos,
    logosHTML: logosHTML,
    qrHTML: qrHTML,
    documentCSS: documentCSS,
    buildDocument: buildDocument,
    openDocument: openDocument,
    buildSharePayload: buildSharePayload,
    shareDocument: shareDocument
  };
})(window);
