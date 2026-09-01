"use strict";

const fs = require("fs");
const vm = require("vm");

if (process.argv.length < 3) {
  console.error("Uso: node scripts/check_inline_js.js archivo.html [...]");
  process.exit(2);
}

let failed = false;
for (const filePath of process.argv.slice(2)) {
  const html = fs.readFileSync(filePath, "utf8");
  const scriptPattern = /<script\b([^>]*)>([\s\S]*?)<\/script>/gi;
  let match;
  let inlineIndex = 0;
  while ((match = scriptPattern.exec(html)) !== null) {
    if (/\bsrc\s*=/i.test(match[1])) continue;
    inlineIndex += 1;
    try {
      new vm.Script(match[2], { filename: filePath + "#inline-" + inlineIndex });
    } catch (error) {
      failed = true;
      console.error(error && error.stack ? error.stack : String(error));
    }
  }
  if (inlineIndex === 0) {
    console.error(filePath + ": no contiene scripts inline");
    failed = true;
  } else if (!failed) {
    console.log(filePath + ": " + inlineIndex + " script(s) inline validos");
  }
}

if (failed) process.exit(1);
