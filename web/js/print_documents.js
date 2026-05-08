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
    return 'recibo';
  }

  function documentCSS(format, kind) {
    format = normalizeFormat(format);
    kind = normalizeKind(kind);
    var accent = kind === 'factura' ? '#0f766e' : (kind === 'orden' ? '#7c3aed' : (kind === 'comprobante' ? '#b45309' : '#2563eb'));
    var page = format === 'pos'
      ? '@page{size:80mm auto;margin:3mm;}html,body{width:74mm;max-width:74mm;}'
      : '@page{size:letter;margin:12mm;}html,body{max-width:190mm;}';
    var base = [
      page,
      'html,body{margin:0 auto;background:#fff;color:#111827;font-family:Arial,Helvetica,sans-serif;-webkit-print-color-adjust:exact;print-color-adjust:exact;}',
      'body{box-sizing:border-box;}*{box-sizing:border-box;}',
      '.pcs-print-doc{width:100%;margin:0 auto;background:#fff;color:#111827;}',
      '.pcs-print-head{display:flex;align-items:flex-start;justify-content:space-between;gap:12px;border-bottom:2px solid ' + accent + ';padding-bottom:10px;margin-bottom:12px;}',
      '.pcs-print-brand h1{margin:0 0 4px;font-size:22px;line-height:1.1;color:#111827;}',
      '.pcs-print-brand p{margin:2px 0;color:#4b5563;font-size:12px;line-height:1.35;}',
      '.pcs-print-badge{border:1px solid ' + accent + ';color:' + accent + ';border-radius:8px;padding:6px 9px;font-weight:900;font-size:12px;text-transform:uppercase;white-space:nowrap;}',
      '.pcs-print-meta{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:8px;margin:10px 0 12px;}',
      '.pcs-print-box{border:1px solid #d9e0ea;border-radius:10px;padding:9px;background:#f8fafc;}',
      '.pcs-print-box span{display:block;color:#64748b;font-size:10px;text-transform:uppercase;font-weight:800;letter-spacing:.04em;margin-bottom:3px;}',
      '.pcs-print-box strong{display:block;color:#111827;font-size:13px;line-height:1.25;word-break:break-word;}',
      '.pcs-print-table{width:100%;border-collapse:collapse;margin:10px 0;font-size:12px;}',
      '.pcs-print-table th{color:#334155;text-align:left;font-size:11px;text-transform:uppercase;border-bottom:1px solid #94a3b8;padding:6px 4px;}',
      '.pcs-print-table td{border-bottom:1px solid #e2e8f0;padding:6px 4px;vertical-align:top;}',
      '.pcs-print-number{text-align:right;white-space:nowrap;}',
      '.pcs-print-total{display:flex;justify-content:space-between;gap:12px;border:2px solid ' + accent + ';border-radius:10px;padding:10px 12px;margin-top:10px;font-size:16px;font-weight:900;}',
      '.pcs-print-note{margin-top:10px;padding:9px;border:1px dashed #94a3b8;border-radius:10px;color:#334155;font-size:12px;line-height:1.4;white-space:pre-wrap;}',
      '.pcs-print-footer{margin-top:12px;color:#64748b;font-size:11px;text-align:center;line-height:1.35;}',
      '.pcs-print-signatures{display:grid;grid-template-columns:1fr 1fr;gap:20px;margin-top:30px;}',
      '.pcs-print-signatures div{border-top:1px solid #111827;padding-top:6px;text-align:center;color:#374151;font-size:12px;}',
      '@media print{.pcs-print-no-print{display:none!important;}}'
    ].join('');
    if (format === 'pos') {
      base += [
        'body{padding:0;font-size:11px;line-height:1.28;}',
        '.pcs-print-doc{font-family:"Courier New",monospace;}',
        '.pcs-print-head{display:block;text-align:center;border-bottom:1px dashed #111827;padding-bottom:7px;margin-bottom:8px;}',
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
        '.pcs-print-note{border:0;border-top:1px dashed #9ca3af;border-radius:0;padding:7px 0;font-size:10px;}',
        '.pcs-print-footer{border-top:1px dashed #9ca3af;padding-top:7px;font-size:10px;}',
        '.pcs-print-signatures{display:none;}'
      ].join('');
    } else {
      base += 'body{padding:0;font-size:13px;}.pcs-print-doc{border:1px solid #d9e0ea;border-radius:14px;padding:18px;box-shadow:0 16px 44px rgba(15,23,42,.08);}';
    }
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

  function buildDocument(options) {
    options = options || {};
    var format = normalizeFormat(options.format);
    var kind = normalizeKind(options.kind);
    var title = text(options.title || 'Documento');
    var company = text(options.company || '');
    var subtitle = text(options.subtitle || '');
    var badge = text(options.badge || (format === 'pos' ? 'POS' : 'Carta'));
    var tableHeaders = Array.isArray(options.tableHeaders) ? options.tableHeaders : [];
    var table = '';
    if (tableHeaders.length || (Array.isArray(options.rows) && options.rows.length)) {
      table = '<table class="pcs-print-table"><thead><tr>' + tableHeaders.map(function(h, idx) {
        return '<th' + (idx > 0 ? ' class="pcs-print-number"' : '') + '>' + escapeHTML(h) + '</th>';
      }).join('') + '</tr></thead><tbody>' + rowsHTML(options.rows) + '</tbody></table>';
    }
    var body = text(options.bodyHTML || '');
    if (!body) {
      body = '<section class="pcs-print-meta">' + metaHTML(options.meta) + '</section>' + table;
      if (text(options.totalLabel || options.totalValue).trim()) {
        body += '<section class="pcs-print-total"><span>' + escapeHTML(options.totalLabel || 'Total') + '</span><span>' + escapeHTML(options.totalValue || '') + '</span></section>';
      }
      if (text(options.note).trim()) body += '<section class="pcs-print-note">' + escapeHTML(options.note) + '</section>';
      if (options.signatures !== false && format === 'carta') body += '<section class="pcs-print-signatures"><div>Recibe</div><div>Entrega / registra</div></section>';
    }
    var auto = options.autoPrint === false ? '' : '<script>window.addEventListener("load",function(){setTimeout(function(){try{window.focus();window.print();' + (options.closeAfterPrint === false ? '' : 'window.close();') + '}catch(e){}},180);});<\/script>';
    return '<!doctype html><html lang="es"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>' + escapeHTML(title) + '</title><style>' + documentCSS(format, kind) + '</style></head><body><article class="pcs-print-doc pcs-print-' + escapeHTML(format) + ' pcs-print-' + escapeHTML(kind) + '"><header class="pcs-print-head"><div class="pcs-print-brand"><h1>' + escapeHTML(title) + '</h1>' + (company ? '<p><strong>' + escapeHTML(company) + '</strong></p>' : '') + (subtitle ? '<p>' + escapeHTML(subtitle) + '</p>' : '') + '</div><div class="pcs-print-badge">' + escapeHTML(badge) + '</div></header>' + body + (text(options.footer).trim() ? '<footer class="pcs-print-footer">' + escapeHTML(options.footer) + '</footer>' : '') + '</article>' + auto + '</body></html>';
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

  global.PCSPrint = {
    escapeHTML: escapeHTML,
    normalizeFormat: normalizeFormat,
    documentCSS: documentCSS,
    buildDocument: buildDocument,
    openDocument: openDocument
  };
})(window);
