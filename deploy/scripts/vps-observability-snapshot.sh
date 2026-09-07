#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/root/powerfulcontrolsystem}"
OUT_DIR="$PROJECT_DIR/backups/observability"
mkdir -p "$OUT_DIR"
OUT="$OUT_DIR/snapshot_$(date +%Y%m%d_%H%M%S).json"

disk="$(df -P / | awk 'NR==2 {print "{\"total\":\""$2"\",\"used\":\""$3"\",\"available\":\""$4"\",\"percent\":\""$5"\"}"}')"
mem="$(free -m | awk 'NR==2 {print "{\"total_mb\":"$2",\"used_mb\":"$3",\"free_mb\":"$4"}"}')"
load="$(cat /proc/loadavg | awk '{print "{\"1m\":\""$1"\",\"5m\":\""$2"\",\"15m\":\""$3"\"}"}')"
containers="$(docker ps --format '{{.Names}}|{{.Status}}' 2>/dev/null | sed 's/"/\\"/g' | awk 'BEGIN{print "["} {printf "%s{\"name\":\"%s\",\"status\":\"%s\"}", sep, $1, substr($0,index($0,"|")+1); sep=","} END{print "]"}' FS='|')"

cat > "$OUT" <<EOF
{
  "generated_at": "$(date -Iseconds)",
  "host": "$(hostname)",
  "disk": $disk,
  "memory": $mem,
  "load": $load,
  "docker_containers": $containers
}
EOF

ln -sfn "$OUT" "$OUT_DIR/latest.json"
cat "$OUT"
