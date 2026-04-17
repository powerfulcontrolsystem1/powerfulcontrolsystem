#!/usr/bin/env bash
# Simple wrapper for spellcheck: prefer cSpell (npx) else python fallback
PATH_ARG=${1:-.}
PATTERNS=("web/**/*.html" "web/**/*.htm" "web/**/*.md" "documentos/**/*.md" "web/**/*.js")

if command -v npx >/dev/null 2>&1; then
  echo "Using cSpell (npx)"
  npx -y cspell@6 --config .cspell.json "${PATTERNS[@]}"
  exit $?
fi

if command -v python3 >/dev/null 2>&1; then
  echo "Using python fallback"
  python3 scripts/spellcheck.py "$PATH_ARG"
  exit $?
fi

echo "Install Node.js (for cSpell) or Python 3 to run the spellcheck script."
exit 2
