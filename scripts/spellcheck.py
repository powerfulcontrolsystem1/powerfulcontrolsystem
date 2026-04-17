#!/usr/bin/env python3
"""
Spellcheck fallback script (Python).

Extrae texto visible de archivos HTML y Markdown y reporta posibles palabras mal escritas
usando `pyspellchecker`. Está pensado como fallback cuando no está disponible `cspell`.

Instalación de dependencias:
  pip install pyspellchecker beautifulsoup4

Uso:
  python scripts/spellcheck.py [ruta]

Ejemplo:
  python scripts/spellcheck.py .
"""
import sys
import os
import glob
import re
import argparse
from collections import defaultdict

try:
    from spellchecker import SpellChecker
except Exception:
    print("Falta la dependencia 'pyspellchecker'. Instala con: pip install pyspellchecker")
    sys.exit(2)

try:
    from bs4 import BeautifulSoup
except Exception:
    print("Falta la dependencia 'beautifulsoup4'. Instala con: pip install beautifulsoup4")
    sys.exit(2)


def get_text_from_html(path):
    with open(path, 'r', encoding='utf-8', errors='ignore') as f:
        raw = f.read()
    soup = BeautifulSoup(raw, 'html.parser')
    for t in soup(['script', 'style', 'noscript']):
        t.extract()
    text = soup.get_text(separator='\n')
    return text


def get_text_from_md(path):
    with open(path, 'r', encoding='utf-8', errors='ignore') as f:
        return f.read()


WORD_RE = re.compile(r"[A-Za-zÁÉÍÓÚÜÑáéíóúüñ']{2,}")


def tokenize(text):
    return [w.lower() for w in WORD_RE.findall(text)]


def load_whitelist(path):
    if not os.path.exists(path):
        return set()
    words = set()
    with open(path, 'r', encoding='utf-8') as f:
        for line in f:
            s = line.strip()
            if not s or s.startswith('#'):
                continue
            words.add(s.lower())
    return words


def find_files(base_path):
    patterns = [
        os.path.join(base_path, '**', '*.html'),
        os.path.join(base_path, '**', '*.htm'),
        os.path.join(base_path, '**', '*.md'),
        os.path.join(base_path, '**', '*.txt'),
    ]
    files = []
    for p in patterns:
        files.extend(glob.glob(p, recursive=True))
    # remove common directories we don't want to scan
    files = [f for f in files if '/node_modules/' not in f.replace('\\', '/') and '\\node_modules\\' not in f]
    return sorted(files)


def main():
    parser = argparse.ArgumentParser(description='Spellcheck fallback (Python)')
    parser.add_argument('path', nargs='?', default='.')
    args = parser.parse_args()

    base = args.path
    whitelist = load_whitelist(os.path.join('scripts', 'spell_whitelist.txt'))

    spell = SpellChecker(language='es')

    files = find_files(base)
    missings = defaultdict(set)

    for f in files:
        try:
            if f.lower().endswith(('.html', '.htm')):
                text = get_text_from_html(f)
            else:
                text = get_text_from_md(f)

            words = tokenize(text)
            unique = set(words)
            to_check = [w for w in unique if w not in whitelist and not w.isdigit()]
            if not to_check:
                continue
            unknown = spell.unknown(to_check)
            if unknown:
                for w in sorted(unknown):
                    missings[f].add(w)
        except Exception as e:
            print(f"Error procesando {f}: {e}", file=sys.stderr)

    if not missings:
        print('No se detectaron posibles errores ortográficos.')
        return 0

    print('\nPosibles errores ortográficos encontrados:')
    total = 0
    for f, words in missings.items():
        total += len(words)
        print(f"\nArchivo: {f}")
        for w in sorted(words):
            suggestions = spell.candidates(w)
            sugg_list = ', '.join(list(suggestions)[:5])
            print(f"  - {w}  -> sugerencias: {sugg_list}")

    print(f"\nTotal palabras marcadas: {total}")
    print("Puedes añadir términos al archivo scripts/spell_whitelist.txt para ignorarlos.")
    return 1


if __name__ == '__main__':
    sys.exit(main())
