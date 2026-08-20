// Lightweight export helpers — CSV (Excel) and printable PDF.

export function exportCSV(filename, columns, rows) {
  const header = columns.map((c) => `"${c.label}"`).join(',');
  const body = rows
    .map((r) => columns.map((c) => `"${String(r[c.key] ?? '').replace(/"/g, '""')}"`).join(','))
    .join('\n');
  const csv = header + '\n' + body;
  const blob = new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename.endsWith('.csv') ? filename : `${filename}.csv`;
  a.click();
  URL.revokeObjectURL(url);
}

export function exportPDF(title, columns, rows) {
  const w = window.open('', '_blank');
  if (!w) { alert('Please allow pop-ups to export PDF.'); return; }
  const th = columns.map((c) => `<th>${c.label}</th>`).join('');
  const trs = rows
    .map((r) => `<tr>${columns.map((c) => `<td>${r[c.key] ?? ''}</td>`).join('')}</tr>`)
    .join('');
  w.document.write(`
    <html><head><title>${title}</title>
    <style>
      body { font-family: Arial, sans-serif; color:#000; padding:24px; }
      h1 { font-size:18px; border-bottom:2px solid #000; padding-bottom:8px; text-transform:uppercase; letter-spacing:1px; }
      .meta { font-size:11px; color:#555; margin-bottom:16px; }
      table { width:100%; border-collapse:collapse; font-size:12px; }
      th { background:#000; color:#fff; text-align:left; padding:8px; text-transform:uppercase; font-size:10px; }
      td { padding:7px 8px; border-bottom:1px solid #ddd; }
    </style></head><body>
    <h1>${title}</h1>
    <div class="meta">Generated ${new Date().toLocaleString('en-IN')} · PG·Commission System</div>
    <table><thead><tr>${th}</tr></thead><tbody>${trs}</tbody></table>
    <script>window.onload=function(){window.print();}</script>
    </body></html>
  `);
  w.document.close();
}
