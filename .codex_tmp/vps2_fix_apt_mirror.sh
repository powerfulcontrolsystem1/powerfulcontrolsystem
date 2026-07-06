#!/usr/bin/env bash
set -euo pipefail

cp -a /etc/apt "/etc/apt.bak.$(date +%Y%m%d%H%M%S)"
while IFS= read -r file; do
  sed -i 's|http://co.archive.ubuntu.com/ubuntu|http://archive.ubuntu.com/ubuntu|g' "$file"
done < <(grep -RIl 'co.archive.ubuntu.com' /etc/apt/sources.list /etc/apt/sources.list.d 2>/dev/null || true)

apt-get clean
apt-get update
