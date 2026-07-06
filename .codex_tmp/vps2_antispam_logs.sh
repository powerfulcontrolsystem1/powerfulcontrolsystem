#!/usr/bin/env bash
set -u

echo "---ANTISPAM LOGS---"
docker logs --tail 220 pcs-mailu-antispam 2>&1 || true

echo "---ANTISPAM INSPECT---"
docker inspect pcs-mailu-antispam --format '{{.State.ExitCode}} {{.State.Error}} {{if .State.Health}}{{.State.Health.Status}}{{end}}' 2>/dev/null || true

echo "---ANTISPAM STATUS---"
docker ps -a --filter name=pcs-mailu-antispam --format 'table {{.Names}}\t{{.Status}}\t{{.Image}}'
