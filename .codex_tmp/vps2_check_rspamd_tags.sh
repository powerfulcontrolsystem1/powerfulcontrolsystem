set -euo pipefail
for tag in 2023.01 2022.12 1.9; do
  echo "== manifest $tag =="
  docker manifest inspect ghcr.io/mailu/rspamd:$tag >/dev/null 2>&1 && echo "OK $tag" || echo "NO $tag"
done
