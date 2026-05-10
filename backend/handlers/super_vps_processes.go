package handlers

import (
	"bufio"
	"database/sql"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

type vpsProcessRow struct {
	PID     int    `json:"pid"`
	Command string `json:"command"`
	Args    string `json:"args"`
	RSSKB   int64  `json:"rss_kb"`
	MemPct  string `json:"mem_pct"`
	CpuPct  string `json:"cpu_pct"`
}

// SuperVPSProcessesHandler lista procesos del VPS por consumo de memoria (RSS) desc.
// Nota: esto corre en el mismo host del backend (VPS). En Windows retorna error.
func SuperVPSProcessesHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper); !ok {
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if runtime.GOOS == "windows" {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{
				"ok":    false,
				"error": "Este backend corre en Windows. Para listar procesos del VPS, ejecute el backend en el VPS Linux.",
			})
			return
		}

		limit := 80
		if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
			if n, err := strconv.Atoi(raw); err == nil && n > 0 && n <= 400 {
				limit = n
			}
		}

		// BusyBox/Alpine no soporta --sort ni columnas pmem/pcpu; Go ordena despues.
		cmd := exec.Command("sh", "-lc", fmt.Sprintf("ps -o pid,comm,rss,args | head -n %d", limit+8))
		out, err := cmd.Output()
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": "No se pudo ejecutar ps en el VPS"})
			return
		}

		rows := make([]vpsProcessRow, 0, limit)
		sc := bufio.NewScanner(strings.NewReader(string(out)))
		lineNo := 0
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" {
				continue
			}
			lineNo++
			if lineNo == 1 && strings.Contains(line, "PID") {
				continue // header
			}
			// Split: pid comm rss args...
			parts := strings.Fields(line)
			if len(parts) < 4 {
				continue
			}
			pid, _ := strconv.Atoi(parts[0])
			rss := parseProcessRSSKB(parts[2])
			args := strings.Join(parts[3:], " ")
			rows = append(rows, vpsProcessRow{
				PID:     pid,
				Command: parts[1],
				RSSKB:   rss,
				MemPct:  "",
				CpuPct:  "",
				Args:    args,
			})
		}

		// Asegurar orden por RSS desc.
		sort.SliceStable(rows, func(i, j int) bool { return rows[i].RSSKB > rows[j].RSSKB })

		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "items": rows})
	}
}

func parseProcessRSSKB(raw string) int64 {
	v := strings.TrimSpace(strings.ToLower(raw))
	if v == "" {
		return 0
	}
	multiplier := int64(1)
	switch {
	case strings.HasSuffix(v, "g"):
		multiplier = 1024 * 1024
		v = strings.TrimSuffix(v, "g")
	case strings.HasSuffix(v, "m"):
		multiplier = 1024
		v = strings.TrimSuffix(v, "m")
	case strings.HasSuffix(v, "k"):
		v = strings.TrimSuffix(v, "k")
	}
	n, _ := strconv.ParseFloat(strings.TrimSpace(v), 64)
	if n <= 0 {
		return 0
	}
	return int64(n * float64(multiplier))
}
