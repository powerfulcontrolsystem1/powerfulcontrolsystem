import fs from "node:fs";
import path from "node:path";

const scriptDir = path.dirname(new URL(import.meta.url).pathname).replace(/^\/([A-Za-z]:)/, "$1");
const root = path.resolve(scriptDir, "..");
const outDir = path.join(root, "web", "img", "brand", "pcs-logo-concepts");
fs.mkdirSync(outDir, { recursive: true });

function write(name, svg) {
  fs.writeFileSync(path.join(outDir, name), svg.trimStart(), "utf8");
}

function iconDefs(id) {
  return `
  <defs>
    <linearGradient id="${id}-blue" x1="0" x2="1" y1="0" y2="1">
      <stop offset="0" stop-color="#38d5ff"/>
      <stop offset="0.48" stop-color="#0878df"/>
      <stop offset="1" stop-color="#05306f"/>
    </linearGradient>
    <linearGradient id="${id}-steel" x1="0" x2="1" y1="0" y2="1">
      <stop offset="0" stop-color="#f8fbff"/>
      <stop offset="1" stop-color="#aeb8c5"/>
    </linearGradient>
    <linearGradient id="${id}-bolt" x1="0" x2="1" y1="0" y2="1">
      <stop offset="0" stop-color="#fff76b"/>
      <stop offset="0.42" stop-color="#ffb000"/>
      <stop offset="1" stop-color="#ff4d00"/>
    </linearGradient>
    <filter id="${id}-shadow" x="-20%" y="-20%" width="140%" height="140%">
      <feDropShadow dx="0" dy="24" stdDeviation="24" flood-color="#020617" flood-opacity="0.32"/>
    </filter>
  </defs>`;
}

const icon1 = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024" role="img" aria-label="PCS logo concepto control core">
${iconDefs("core")}
<rect width="1024" height="1024" rx="220" fill="#07111f"/>
<circle cx="512" cy="512" r="408" fill="url(#core-blue)" opacity=".18"/>
<g filter="url(#core-shadow)">
  <rect x="190" y="176" width="644" height="444" rx="78" fill="url(#core-blue)" stroke="#dff7ff" stroke-width="20"/>
  <rect x="250" y="240" width="524" height="296" rx="38" fill="#eaf8ff"/>
  <path d="M312 392c31-92 119-148 224-148 89 0 156 30 205 82l-91 88c-28-32-67-50-116-50-67 0-115 36-135 96h151l-44 106H286c1-61 10-119 26-174z" fill="#063d8f"/>
  <path d="M545 148 390 500h137l-62 376 184-446H516l29-282z" fill="url(#core-bolt)" stroke="#0b1220" stroke-width="13" stroke-linejoin="round"/>
  <rect x="394" y="650" width="236" height="70" rx="26" fill="#aeb8c5"/>
  <rect x="292" y="734" width="440" height="92" rx="40" fill="#d7dee8" stroke="#0b1220" stroke-width="18"/>
</g>
</svg>`;

const icon2 = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024" role="img" aria-label="PCS logo concepto nexus shield">
${iconDefs("shield")}
<rect width="1024" height="1024" rx="120" fill="#f6fbff"/>
<g filter="url(#shield-shadow)">
  <path d="M512 92 842 258v249c0 207-133 347-330 425-197-78-330-218-330-425V258L512 92z" fill="#07111f"/>
  <path d="M512 150 786 288v216c0 164-98 281-274 352-176-71-274-188-274-352V288l274-138z" fill="url(#shield-blue)"/>
  <path d="M292 522c0-142 92-250 232-250 89 0 156 38 199 96l-95 85c-25-39-63-59-109-59-66 0-111 50-111 128 0 77 45 128 111 128 48 0 86-22 112-62l94 86c-43 58-110 98-201 98-140 0-232-107-232-250z" fill="#eef9ff"/>
  <path d="M543 186 392 521h128l-58 304 178-390H519l24-249z" fill="url(#shield-bolt)" stroke="#07111f" stroke-width="12" stroke-linejoin="round"/>
  <text x="512" y="858" text-anchor="middle" font-family="Montserrat, Arial, sans-serif" font-size="84" font-weight="900" fill="#eef9ff" letter-spacing="10">PCS</text>
</g>
</svg>`;

const icon3 = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024" role="img" aria-label="PCS logo concepto pulse circuit">
${iconDefs("pulse")}
<rect width="1024" height="1024" rx="512" fill="#07111f"/>
<circle cx="512" cy="512" r="390" fill="none" stroke="url(#pulse-blue)" stroke-width="34"/>
<circle cx="512" cy="512" r="308" fill="#0d1b2f"/>
<g stroke="#35d5ff" stroke-width="18" stroke-linecap="round" opacity=".78">
  <path d="M150 512h124m476 0h124M512 150v118m0 488v118"/>
  <path d="M236 248l76 76m400 400 76 76M788 248l-76 76m-400 400-76 76"/>
</g>
<g fill="#35d5ff">
  <circle cx="150" cy="512" r="28"/><circle cx="874" cy="512" r="28"/><circle cx="512" cy="150" r="28"/><circle cx="512" cy="874" r="28"/>
</g>
<text x="512" y="580" text-anchor="middle" font-family="Montserrat, Arial Black, Arial, sans-serif" font-size="250" font-weight="900" fill="#eef9ff" letter-spacing="-18">PCS</text>
<path d="M536 214 396 528h122l-48 282 166-370H516l20-226z" fill="url(#pulse-bolt)" stroke="#07111f" stroke-width="13" stroke-linejoin="round"/>
</svg>`;

const icon4 = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024" role="img" aria-label="PCS logo concepto enterprise stack">
${iconDefs("stack")}
<rect width="1024" height="1024" rx="184" fill="#eef7ff"/>
<g filter="url(#stack-shadow)">
  <rect x="168" y="154" width="688" height="520" rx="72" fill="#07111f"/>
  <rect x="224" y="210" width="576" height="344" rx="34" fill="url(#stack-blue)"/>
  <path d="M306 428c20-88 96-147 196-147 82 0 145 31 190 84l-83 78c-27-31-60-47-105-47-58 0-98 33-113 84h132l-38 88H284c1-49 8-96 22-140z" fill="#eef9ff"/>
  <path d="M556 188 410 510h126l-54 326 173-388H532l24-260z" fill="url(#stack-bolt)" stroke="#07111f" stroke-width="12" stroke-linejoin="round"/>
  <rect x="214" y="702" width="596" height="146" rx="36" fill="#07111f"/>
  <rect x="274" y="738" width="210" height="72" rx="18" fill="#eef9ff"/>
  <rect x="542" y="738" width="54" height="54" rx="12" fill="#38d5ff"/>
  <rect x="628" y="738" width="54" height="54" rx="12" fill="#38d5ff"/>
  <rect x="714" y="738" width="54" height="54" rx="12" fill="#38d5ff"/>
</g>
</svg>`;

const icons = [
  ["pcs-icon-v1-control-core.svg", icon1],
  ["pcs-icon-v2-nexus-shield.svg", icon2],
  ["pcs-icon-v3-pulse-circuit.svg", icon3],
  ["pcs-icon-v4-enterprise-stack.svg", icon4],
];
for (const [name, svg] of icons) write(name, svg);

function wordmark(name, iconHref, variant, bg, fg, accent) {
  const tagline = variant === 2 ? "SaaS POS Multiempresa" : variant === 4 ? "Control empresarial inteligente" : "Sistema empresarial POS multiempresa";
  const layout = variant === 2 ? `
  <rect width="1800" height="720" rx="80" fill="${bg}"/>
  <image href="${iconHref}" x="676" y="56" width="448" height="448"/>
  <text x="900" y="575" text-anchor="middle" font-family="Montserrat, Arial, sans-serif" font-size="92" font-weight="900" fill="${fg}">Powerful Control System</text>
  <text x="900" y="638" text-anchor="middle" font-family="Montserrat, Arial, sans-serif" font-size="34" font-weight="700" fill="${accent}" letter-spacing="6">${tagline.toUpperCase()}</text>`
  : `
  <rect width="1800" height="520" rx="58" fill="${bg}"/>
  <image href="${iconHref}" x="96" y="72" width="376" height="376"/>
  <text x="520" y="228" font-family="Montserrat, Arial, sans-serif" font-size="${variant === 3 ? 96 : 88}" font-weight="900" fill="${fg}">Powerful Control System</text>
  <rect x="522" y="270" width="710" height="8" rx="4" fill="${accent}"/>
  <text x="520" y="350" font-family="Montserrat, Arial, sans-serif" font-size="42" font-weight="700" fill="${variant === 1 ? "#415166" : accent}" letter-spacing="3">${tagline}</text>
  <text x="1458" y="344" font-family="Montserrat, Arial, sans-serif" font-size="118" font-weight="900" fill="${accent}" opacity=".18">PCS</text>`;
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1800 ${variant === 2 ? 720 : 520}" role="img" aria-label="${name}">
${layout}
</svg>`;
}

const wordmarks = [
  ["pcs-wordmark-v1-control-core.svg", "pcs-icon-v1-control-core.svg", 1, "#f8fbff", "#07111f", "#0878df"],
  ["pcs-wordmark-v2-stacked-premium.svg", "pcs-icon-v2-nexus-shield.svg", 2, "#07111f", "#eef9ff", "#ffb000"],
  ["pcs-wordmark-v3-dark-enterprise.svg", "pcs-icon-v3-pulse-circuit.svg", 3, "#050b14", "#f8fbff", "#38d5ff"],
  ["pcs-wordmark-v4-clean-pos.svg", "pcs-icon-v4-enterprise-stack.svg", 4, "#eef7ff", "#07111f", "#0878df"],
];
for (const [file, href, variant, bg, fg, accent] of wordmarks) {
  write(file, wordmark(file, href, variant, bg, fg, accent));
}

const previewCards = [...icons.map(([name]) => name), ...wordmarks.map(([name]) => name)];
write("preview.html", `<!doctype html>
<html lang="es">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Propuestas de logo PCS</title>
  <style>
    :root { color-scheme: dark; font-family: Inter, Segoe UI, Arial, sans-serif; background:#070b12; color:#eaf2ff; }
    body { margin:0; padding:34px; }
    h1 { margin:0 0 8px; font-size:30px; }
    p { margin:0 0 26px; color:#9fb0c8; }
    .grid { display:grid; grid-template-columns: repeat(4, minmax(180px, 1fr)); gap:18px; }
    .card { border:1px solid rgba(148,163,184,.28); border-radius:18px; padding:16px; background:#0d1422; }
    .card.word { grid-column: span 2; }
    img { display:block; width:100%; height:auto; border-radius:14px; background:#fff; }
    strong { display:block; margin-top:12px; font-size:14px; color:#f8fbff; }
  </style>
</head>
<body>
  <h1>Propuestas de logo PCS</h1>
  <p>4 isotipos y 4 versiones con el nombre Powerful Control System.</p>
  <section class="grid">
    ${previewCards.map((file, idx) => `<article class="card ${file.includes("wordmark") ? "word" : ""}"><img src="${file}" alt="${file}"><strong>${idx + 1}. ${file.replace(".svg", "")}</strong></article>`).join("\n    ")}
  </section>
</body>
</html>`);

console.log(`generated=${previewCards.length} dir=${outDir}`);
