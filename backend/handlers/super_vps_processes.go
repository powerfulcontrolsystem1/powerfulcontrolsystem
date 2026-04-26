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

		// ps output: pid comm rss pmem pcpu args
		cmd := exec.Command("bash", "-lc", fmt.Sprintf("ps -eo pid,comm,rss,pmem,pcpu,args --sort=-rss | head -n %d", limit+1))
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
			// Split into at most 6 fields.
			parts := strings.Fields(line)
			if len(parts) < 6 {
				continue
			}
			pid, _ := strconv.Atoi(parts[0])
			rss, _ := strconv.ParseInt(parts[2], 10, 64)
			args := strings.Join(parts[5:], " ")
			rows = append(rows, vpsProcessRow{
				PID:     pid,
				Command: parts[1],
				RSSKB:   rss,
				MemPct:  parts[3],
				CpuPct:  parts[4],
				Args:    args,
			})
		}

		// Asegurar orden por RSS desc.
		sort.SliceStable(rows, func(i, j int) bool { return rows[i].RSSKB > rows[j].RSSKB })

		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "items": rows})
	}
}

