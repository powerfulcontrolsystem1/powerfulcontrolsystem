package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type postgresRecommendation struct {
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type postgresClusterInfo struct {
	CurrentDatabase              string  `json:"current_database"`
	PostgresVersion              string  `json:"postgres_version"`
	VersionLabel                 string  `json:"version_label"`
	InRecovery                   bool    `json:"in_recovery"`
	StartedAt                    string  `json:"started_at"`
	UptimeSeconds                int64   `json:"uptime_seconds"`
	MaxConnections               int64   `json:"max_connections"`
	TotalConnections             int64   `json:"total_connections"`
	ActiveConnections            int64   `json:"active_connections"`
	IdleConnections              int64   `json:"idle_connections"`
	WaitingConnections           int64   `json:"waiting_connections"`
	IdleInTransactionConnections int64   `json:"idle_in_transaction_connections"`
	BlockedConnections           int64   `json:"blocked_connections"`
	ConnectionUsagePct           float64 `json:"connection_usage_pct"`
	CacheHitRatioPct             float64 `json:"cache_hit_ratio_pct"`
}

type postgresDatabaseInfo struct {
	Key                string  `json:"key"`
	DisplayName        string  `json:"display_name"`
	Name               string  `json:"name"`
	SizeBytes          int64   `json:"size_bytes"`
	SizePretty         string  `json:"size_pretty"`
	TotalConnections   int64   `json:"total_connections"`
	ActiveConnections  int64   `json:"active_connections"`
	IdleConnections    int64   `json:"idle_connections"`
	WaitingConnections int64   `json:"waiting_connections"`
	BlockedConnections int64   `json:"blocked_connections"`
	XactCommit         int64   `json:"xact_commit"`
	XactRollback       int64   `json:"xact_rollback"`
	RollbackPct        float64 `json:"rollback_pct"`
	BlocksRead         int64   `json:"blocks_read"`
	BlocksHit          int64   `json:"blocks_hit"`
	CacheHitRatioPct   float64 `json:"cache_hit_ratio_pct"`
	TuplesReturned     int64   `json:"tuples_returned"`
	TuplesFetched      int64   `json:"tuples_fetched"`
	TuplesInserted     int64   `json:"tuples_inserted"`
	TuplesUpdated      int64   `json:"tuples_updated"`
	TuplesDeleted      int64   `json:"tuples_deleted"`
	TempFiles          int64   `json:"temp_files"`
	TempBytes          int64   `json:"temp_bytes"`
	Deadlocks          int64   `json:"deadlocks"`
	BlockReadTimeMs    float64 `json:"block_read_time_ms"`
	BlockWriteTimeMs   float64 `json:"block_write_time_ms"`
}

type postgresBGWriterInfo struct {
	Available             bool    `json:"available"`
	Error                 string  `json:"error,omitempty"`
	CheckpointsTimed      int64   `json:"checkpoints_timed"`
	CheckpointsReq        int64   `json:"checkpoints_req"`
	BuffersCheckpoint     int64   `json:"buffers_checkpoint"`
	BuffersClean          int64   `json:"buffers_clean"`
	MaxwrittenClean       int64   `json:"maxwritten_clean"`
	BuffersBackend        int64   `json:"buffers_backend"`
	BuffersBackendFsync   int64   `json:"buffers_backend_fsync"`
	BuffersAlloc          int64   `json:"buffers_alloc"`
	CheckpointWriteTimeMs float64 `json:"checkpoint_write_time_ms"`
	CheckpointSyncTimeMs  float64 `json:"checkpoint_sync_time_ms"`
}

type postgresLongRunningQuery struct {
	PID            int64  `json:"pid"`
	User           string `json:"user"`
	State          string `json:"state"`
	WaitEventType  string `json:"wait_event_type"`
	WaitEvent      string `json:"wait_event"`
	RunningSeconds int64  `json:"running_seconds"`
	QuerySnippet   string `json:"query_snippet"`
}

type postgresPerformanceResponse struct {
	Ok                 bool                       `json:"ok"`
	Engine             string                     `json:"engine"`
	GeneratedAt        string                     `json:"generated_at"`
	Health             string                     `json:"health"`
	Cluster            postgresClusterInfo        `json:"cluster"`
	Databases          []postgresDatabaseInfo     `json:"databases"`
	BGWriter           postgresBGWriterInfo       `json:"bgwriter"`
	LongRunningQueries []postgresLongRunningQuery `json:"long_running_queries"`
	Recommendations    []postgresRecommendation   `json:"recommendations"`
}

type postgresEmpresaStorageItem struct {
	EmpresaID          int64   `json:"empresa_id"`
	Nombre             string  `json:"nombre"`
	Nit                string  `json:"nit"`
	Estado             string  `json:"estado"`
	TotalBytes         int64   `json:"total_bytes"`
	TotalPretty        string  `json:"total_pretty"`
	TotalMB            float64 `json:"total_mb"`
	RowsCount          int64   `json:"rows_count"`
	TablesWithData     int64   `json:"tables_with_data"`
	LargestTable       string  `json:"largest_table"`
	LargestTableBytes  int64   `json:"largest_table_bytes"`
	LargestTablePretty string  `json:"largest_table_pretty"`
	LargestTableMB     float64 `json:"largest_table_mb"`
	QuotaGB            int64   `json:"quota_gb"`
	QuotaBytes         int64   `json:"quota_bytes"`
	QuotaPretty        string  `json:"quota_pretty"`
	QuotaUsagePct      float64 `json:"quota_usage_pct"`
	QuotaStatus        string  `json:"quota_status"`
}

type postgresEmpresaStorageResponse struct {
	Ok            bool                         `json:"ok"`
	Engine        string                       `json:"engine"`
	GeneratedAt   string                       `json:"generated_at"`
	TablesScanned int                          `json:"tables_scanned"`
	TotalEmpresas int                          `json:"total_empresas"`
	TotalBytes    int64                        `json:"total_bytes"`
	TotalPretty   string                       `json:"total_pretty"`
	Empresas      []postgresEmpresaStorageItem `json:"empresas"`
}

// PostgresPerformanceHandler expone metricas operativas del motor PostgreSQL para el panel super.
func PostgresPerformanceHandler(dbEmpresas, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch action {
		case "", "performance":
			// continua con la ruta actual
		case "empresas_storage":
			if dbEmpresas == nil {
				writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
					"ok":    false,
					"error": "conexion de base de datos no disponible",
				})
				return
			}
		default:
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"ok":    false,
				"error": "accion no soportada para el panel PostgreSQL",
			})
			return
		}

		if dbEmpresas == nil || dbSuper == nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"ok":    false,
				"error": "conexion de base de datos no disponible",
			})
			return
		}
		if !dbpkg.IsPostgresDialect() {
			writeJSON(w, http.StatusConflict, map[string]interface{}{
				"ok":    false,
				"error": "este panel solo esta disponible cuando el runtime opera con PostgreSQL",
			})
			return
		}
		switch action {
		case "", "performance":
			// continua con la ruta actual
		case "empresas_storage":
			handlePostgresEmpresaStorage(w, r, dbEmpresas, dbSuper)
			return
		default:
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"ok":    false,
				"error": "accion no soportada para el panel PostgreSQL",
			})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
		defer cancel()

		cluster, clusterRecs, err := collectPostgresClusterInfo(ctx, dbSuper)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"ok":    false,
				"error": "no se pudo leer el estado del cluster PostgreSQL: " + err.Error(),
			})
			return
		}

		superDB, superRecs, err := collectPostgresDatabaseInfo(ctx, dbSuper, "super", "Superadministrador")
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"ok":    false,
				"error": "no se pudo leer metricas de superadministrador: " + err.Error(),
			})
			return
		}
		empresasDB, empresasRecs, err := collectPostgresDatabaseInfo(ctx, dbEmpresas, "empresas", "Empresas")
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"ok":    false,
				"error": "no se pudo leer metricas de empresas: " + err.Error(),
			})
			return
		}

		bgwriter, bgwriterRecs := collectPostgresBGWriterInfo(ctx, dbSuper)
		longQueries, longQueryRecs := collectPostgresLongRunningQueries(ctx, dbSuper)

		recommendations := mergePostgresRecommendations(
			clusterRecs,
			superRecs,
			empresasRecs,
			bgwriterRecs,
			longQueryRecs,
		)

		payload := postgresPerformanceResponse{
			Ok:                 true,
			Engine:             "postgresql",
			GeneratedAt:        time.Now().Format(time.RFC3339),
			Health:             resolvePostgresHealth(recommendations, cluster),
			Cluster:            cluster,
			Databases:          []postgresDatabaseInfo{superDB, empresasDB},
			BGWriter:           bgwriter,
			LongRunningQueries: longQueries,
			Recommendations:    recommendations,
		}
		writeJSON(w, http.StatusOK, payload)
	}
}

func handlePostgresEmpresaStorage(w http.ResponseWriter, r *http.Request, dbEmpresas, dbSuper *sql.DB) {
	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()

	items, tablesScanned, totalBytes, err := collectPostgresEmpresaStorage(ctx, dbEmpresas)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"ok":    false,
			"error": "no se pudo calcular el tamano por empresa: " + err.Error(),
		})
		return
	}
	applyPostgresEmpresaStorageQuotas(dbSuper, items)

	writeJSON(w, http.StatusOK, postgresEmpresaStorageResponse{
		Ok:            true,
		Engine:        "postgresql",
		GeneratedAt:   time.Now().Format(time.RFC3339),
		TablesScanned: tablesScanned,
		TotalEmpresas: len(items),
		TotalBytes:    totalBytes,
		TotalPretty:   humanizeBytesBinary(totalBytes),
		Empresas:      items,
	})
}

func applyPostgresEmpresaStorageQuotas(dbSuper *sql.DB, items []postgresEmpresaStorageItem) {
	if len(items) == 0 {
		return
	}
	quotaGB, _, _, err := getLimitacionInt64WithLegacy(dbSuper, superEmpresaLimitDBMaxGBKey, superEmpresaLimitLegacyNextcloudMaxGBKey, defaultEmpresaDBMaxGB)
	if err != nil {
		quotaGB = defaultEmpresaDBMaxGB
	}
	if quotaGB < 0 {
		quotaGB = 0
	}
	quotaBytes := quotaGB * 1024 * 1024 * 1024
	quotaPretty := humanizeBytesBinary(quotaBytes)
	for i := range items {
		items[i].QuotaGB = quotaGB
		items[i].QuotaBytes = quotaBytes
		items[i].QuotaPretty = quotaPretty
		items[i].QuotaStatus = "sin_limite"
		if quotaBytes <= 0 {
			continue
		}
		items[i].QuotaUsagePct = round2(percentFloat(float64(items[i].TotalBytes), float64(quotaBytes)))
		switch {
		case items[i].TotalBytes >= quotaBytes:
			items[i].QuotaStatus = "excedido"
		case items[i].QuotaUsagePct >= 85:
			items[i].QuotaStatus = "alerta"
		default:
			items[i].QuotaStatus = "ok"
		}
	}
}

type postgresEmpresaStorageAccumulator struct {
	RowsCount         int64
	TablesWithData    int64
	TotalBytes        int64
	LargestTable      string
	LargestTableBytes int64
}

func collectPostgresEmpresaStorage(ctx context.Context, dbConn *sql.DB) ([]postgresEmpresaStorageItem, int, int64, error) {
	empresas, err := dbpkg.GetEmpresas(dbConn)
	if err != nil {
		return nil, 0, 0, err
	}

	tables, err := listPostgresEmpresaScopedTables(ctx, dbConn)
	if err != nil {
		return nil, 0, 0, err
	}

	usageByEmpresa := make(map[int64]*postgresEmpresaStorageAccumulator)
	for _, tableName := range tables {
		rows, err := dbConn.QueryContext(ctx, fmt.Sprintf(`
			SELECT empresa_id::bigint, COUNT(*)::bigint, COALESCE(SUM(pg_column_size(t)),0)::bigint
			FROM public.%s t
			WHERE empresa_id IS NOT NULL
			GROUP BY empresa_id
		`, quotePostgresIdentifier(tableName)))
		if err != nil {
			return nil, 0, 0, err
		}

		for rows.Next() {
			var empresaID int64
			var rowCount int64
			var bytes int64
			if err := rows.Scan(&empresaID, &rowCount, &bytes); err != nil {
				rows.Close()
				return nil, 0, 0, err
			}
			acc := usageByEmpresa[empresaID]
			if acc == nil {
				acc = &postgresEmpresaStorageAccumulator{}
				usageByEmpresa[empresaID] = acc
			}
			acc.RowsCount += rowCount
			acc.TotalBytes += bytes
			if rowCount > 0 || bytes > 0 {
				acc.TablesWithData += 1
			}
			if bytes > acc.LargestTableBytes {
				acc.LargestTableBytes = bytes
				acc.LargestTable = tableName
			}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, 0, 0, err
		}
		rows.Close()
	}

	items := make([]postgresEmpresaStorageItem, 0, len(empresas))
	var totalBytes int64
	for _, empresa := range empresas {
		empresaKey := empresa.EmpresaID
		if empresaKey <= 0 {
			empresaKey = empresa.ID
		}
		acc := usageByEmpresa[empresaKey]
		item := postgresEmpresaStorageItem{
			EmpresaID: empresaKey,
			Nombre:    strings.TrimSpace(empresa.Nombre),
			Nit:       strings.TrimSpace(empresa.Nit),
			Estado:    strings.TrimSpace(empresa.Estado),
		}
		if acc != nil {
			item.TotalBytes = acc.TotalBytes
			item.TotalPretty = humanizeBytesBinary(acc.TotalBytes)
			item.TotalMB = round2(float64(acc.TotalBytes) / (1024 * 1024))
			item.RowsCount = acc.RowsCount
			item.TablesWithData = acc.TablesWithData
			item.LargestTable = acc.LargestTable
			item.LargestTableBytes = acc.LargestTableBytes
			item.LargestTablePretty = humanizeBytesBinary(acc.LargestTableBytes)
			item.LargestTableMB = round2(float64(acc.LargestTableBytes) / (1024 * 1024))
			totalBytes += acc.TotalBytes
		} else {
			item.TotalPretty = humanizeBytesBinary(0)
			item.LargestTablePretty = humanizeBytesBinary(0)
		}
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].TotalBytes != items[j].TotalBytes {
			return items[i].TotalBytes > items[j].TotalBytes
		}
		if items[i].RowsCount != items[j].RowsCount {
			return items[i].RowsCount > items[j].RowsCount
		}
		return strings.ToLower(items[i].Nombre) < strings.ToLower(items[j].Nombre)
	})

	return items, len(tables), totalBytes, nil
}

func listPostgresEmpresaScopedTables(ctx context.Context, dbConn *sql.DB) ([]string, error) {
	rows, err := dbConn.QueryContext(ctx, `
		SELECT c.table_name
		FROM information_schema.columns c
		JOIN information_schema.tables t
		  ON t.table_schema = c.table_schema
		 AND t.table_name = c.table_name
		WHERE c.table_schema = 'public'
		  AND c.column_name = 'empresa_id'
		  AND t.table_type = 'BASE TABLE'
		ORDER BY c.table_name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tableName = strings.TrimSpace(tableName)
		if tableName == "" {
			continue
		}
		tables = append(tables, tableName)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tables, nil
}

func quotePostgresIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(strings.TrimSpace(identifier), `"`, `""`) + `"`
}

func humanizeBytesBinary(totalBytes int64) string {
	if totalBytes <= 0 {
		return "0 B"
	}
	units := []string{"B", "KB", "MB", "GB", "TB"}
	value := float64(totalBytes)
	unitIdx := 0
	for value >= 1024 && unitIdx < len(units)-1 {
		value = value / 1024
		unitIdx += 1
	}
	if unitIdx == 0 {
		return fmt.Sprintf("%d %s", totalBytes, units[unitIdx])
	}
	return fmt.Sprintf("%.2f %s", value, units[unitIdx])
}

func collectPostgresClusterInfo(ctx context.Context, dbConn *sql.DB) (postgresClusterInfo, []postgresRecommendation, error) {
	var info postgresClusterInfo
	var recs []postgresRecommendation

	var startedAt time.Time
	err := dbConn.QueryRowContext(ctx, `
		SELECT
			version(),
			current_database(),
			pg_is_in_recovery(),
			pg_postmaster_start_time(),
			COALESCE(EXTRACT(EPOCH FROM (now() - pg_postmaster_start_time())),0)::bigint
	`).Scan(
		&info.PostgresVersion,
		&info.CurrentDatabase,
		&info.InRecovery,
		&startedAt,
		&info.UptimeSeconds,
	)
	if err != nil {
		return info, recs, err
	}
	info.StartedAt = startedAt.Format(time.RFC3339)
	info.VersionLabel = extractPostgresVersionLabel(info.PostgresVersion)

	var maxConnectionsRaw string
	if err := dbConn.QueryRowContext(ctx, `SHOW max_connections`).Scan(&maxConnectionsRaw); err == nil {
		parsed, parseErr := strconv.ParseInt(strings.TrimSpace(maxConnectionsRaw), 10, 64)
		if parseErr == nil {
			info.MaxConnections = parsed
		}
	}

	err = dbConn.QueryRowContext(ctx, `
		SELECT
			COUNT(*)::bigint,
			COUNT(*) FILTER (WHERE state = 'active')::bigint,
			COUNT(*) FILTER (WHERE state = 'idle')::bigint,
			COUNT(*) FILTER (WHERE wait_event_type IS NOT NULL AND state <> 'idle')::bigint,
			COUNT(*) FILTER (WHERE state = 'idle in transaction')::bigint,
			COUNT(*) FILTER (WHERE wait_event_type = 'Lock')::bigint
		FROM pg_stat_activity
	`).Scan(
		&info.TotalConnections,
		&info.ActiveConnections,
		&info.IdleConnections,
		&info.WaitingConnections,
		&info.IdleInTransactionConnections,
		&info.BlockedConnections,
	)
	if err != nil {
		return info, recs, err
	}

	var blocksHit float64
	var blocksRead float64
	err = dbConn.QueryRowContext(ctx, `
		SELECT
			COALESCE(SUM(blks_hit),0)::float8,
			COALESCE(SUM(blks_read),0)::float8
		FROM pg_stat_database
	`).Scan(&blocksHit, &blocksRead)
	if err == nil {
		info.CacheHitRatioPct = cacheHitRatio(blocksHit, blocksRead)
	}

	if info.MaxConnections > 0 {
		info.ConnectionUsagePct = round2(percentFloat(float64(info.TotalConnections), float64(info.MaxConnections)))
	}

	if info.ConnectionUsagePct >= 95 {
		recs = append(recs, postgresRecommendation{Severity: "critical", Message: "Uso de conexiones del cluster por encima del 95%; aumentar capacidad o cerrar sesiones ociosas"})
	} else if info.ConnectionUsagePct >= 80 {
		recs = append(recs, postgresRecommendation{Severity: "warning", Message: "Uso de conexiones del cluster por encima del 80%; revisar pool de conexiones"})
	}
	if info.BlockedConnections > 0 {
		recs = append(recs, postgresRecommendation{Severity: "critical", Message: fmt.Sprintf("Se detectaron %d conexiones bloqueadas por locks", info.BlockedConnections)})
	}
	if info.WaitingConnections > 0 {
		recs = append(recs, postgresRecommendation{Severity: "warning", Message: fmt.Sprintf("Hay %d conexiones en espera de eventos del motor", info.WaitingConnections)})
	}
	if info.CacheHitRatioPct > 0 && info.CacheHitRatioPct < 95 {
		recs = append(recs, postgresRecommendation{Severity: "warning", Message: fmt.Sprintf("Cache hit ratio global bajo (%.2f%%); revisar memoria compartida y patrones de consulta", info.CacheHitRatioPct)})
	}
	if info.InRecovery {
		recs = append(recs, postgresRecommendation{Severity: "info", Message: "La instancia esta en modo recovery/replica"})
	}

	return info, recs, nil
}

func collectPostgresDatabaseInfo(ctx context.Context, dbConn *sql.DB, key, displayName string) (postgresDatabaseInfo, []postgresRecommendation, error) {
	info := postgresDatabaseInfo{Key: key, DisplayName: displayName}
	var recs []postgresRecommendation

	if err := dbConn.QueryRowContext(ctx, `SELECT current_database()`).Scan(&info.Name); err != nil {
		return info, recs, err
	}

	err := dbConn.QueryRowContext(ctx, `
		SELECT
			pg_database_size(current_database())::bigint,
			pg_size_pretty(pg_database_size(current_database()))
	`).Scan(&info.SizeBytes, &info.SizePretty)
	if err != nil {
		return info, recs, err
	}

	err = dbConn.QueryRowContext(ctx, `
		SELECT
			COALESCE(xact_commit,0)::bigint,
			COALESCE(xact_rollback,0)::bigint,
			COALESCE(blks_read,0)::bigint,
			COALESCE(blks_hit,0)::bigint,
			COALESCE(tup_returned,0)::bigint,
			COALESCE(tup_fetched,0)::bigint,
			COALESCE(tup_inserted,0)::bigint,
			COALESCE(tup_updated,0)::bigint,
			COALESCE(tup_deleted,0)::bigint,
			COALESCE(temp_files,0)::bigint,
			COALESCE(temp_bytes,0)::bigint,
			COALESCE(deadlocks,0)::bigint,
			COALESCE(blk_read_time,0)::float8,
			COALESCE(blk_write_time,0)::float8
		FROM pg_stat_database
		WHERE datname = current_database()
	`).Scan(
		&info.XactCommit,
		&info.XactRollback,
		&info.BlocksRead,
		&info.BlocksHit,
		&info.TuplesReturned,
		&info.TuplesFetched,
		&info.TuplesInserted,
		&info.TuplesUpdated,
		&info.TuplesDeleted,
		&info.TempFiles,
		&info.TempBytes,
		&info.Deadlocks,
		&info.BlockReadTimeMs,
		&info.BlockWriteTimeMs,
	)
	if err != nil {
		return info, recs, err
	}

	err = dbConn.QueryRowContext(ctx, `
		SELECT
			COUNT(*)::bigint,
			COUNT(*) FILTER (WHERE state = 'active')::bigint,
			COUNT(*) FILTER (WHERE state = 'idle')::bigint,
			COUNT(*) FILTER (WHERE wait_event_type IS NOT NULL AND state <> 'idle')::bigint,
			COUNT(*) FILTER (WHERE wait_event_type = 'Lock')::bigint
		FROM pg_stat_activity
		WHERE datname = current_database()
	`).Scan(
		&info.TotalConnections,
		&info.ActiveConnections,
		&info.IdleConnections,
		&info.WaitingConnections,
		&info.BlockedConnections,
	)
	if err != nil {
		return info, recs, err
	}

	info.CacheHitRatioPct = cacheHitRatio(float64(info.BlocksHit), float64(info.BlocksRead))
	txTotal := info.XactCommit + info.XactRollback
	if txTotal > 0 {
		info.RollbackPct = round2(percentFloat(float64(info.XactRollback), float64(txTotal)))
	}
	info.BlockReadTimeMs = round2(info.BlockReadTimeMs)
	info.BlockWriteTimeMs = round2(info.BlockWriteTimeMs)

	if info.CacheHitRatioPct > 0 && info.CacheHitRatioPct < 95 {
		recs = append(recs, postgresRecommendation{
			Severity: "warning",
			Message:  fmt.Sprintf("%s: cache hit ratio bajo (%.2f%%)", displayName, info.CacheHitRatioPct),
		})
	}
	if info.Deadlocks > 0 {
		recs = append(recs, postgresRecommendation{
			Severity: "critical",
			Message:  fmt.Sprintf("%s: se detectaron deadlocks (%d)", displayName, info.Deadlocks),
		})
	}
	if txTotal >= 100 && info.RollbackPct > 3 {
		recs = append(recs, postgresRecommendation{
			Severity: "warning",
			Message:  fmt.Sprintf("%s: porcentaje de rollback alto (%.2f%%)", displayName, info.RollbackPct),
		})
	}
	if info.TempFiles >= 100 || info.TempBytes >= 536870912 {
		recs = append(recs, postgresRecommendation{
			Severity: "warning",
			Message:  fmt.Sprintf("%s: alto uso de archivos temporales (files=%d)", displayName, info.TempFiles),
		})
	}
	if info.WaitingConnections > 0 {
		recs = append(recs, postgresRecommendation{
			Severity: "warning",
			Message:  fmt.Sprintf("%s: conexiones en espera (%d)", displayName, info.WaitingConnections),
		})
	}
	if info.BlockedConnections > 0 {
		recs = append(recs, postgresRecommendation{
			Severity: "critical",
			Message:  fmt.Sprintf("%s: conexiones bloqueadas por locks (%d)", displayName, info.BlockedConnections),
		})
	}

	return info, recs, nil
}

func collectPostgresBGWriterInfo(ctx context.Context, dbConn *sql.DB) (postgresBGWriterInfo, []postgresRecommendation) {
	info := postgresBGWriterInfo{Available: true}
	var recs []postgresRecommendation

	err := dbConn.QueryRowContext(ctx, `
		SELECT
			COALESCE(checkpoints_timed,0)::bigint,
			COALESCE(checkpoints_req,0)::bigint,
			COALESCE(buffers_checkpoint,0)::bigint,
			COALESCE(buffers_clean,0)::bigint,
			COALESCE(maxwritten_clean,0)::bigint,
			COALESCE(buffers_backend,0)::bigint,
			COALESCE(buffers_backend_fsync,0)::bigint,
			COALESCE(buffers_alloc,0)::bigint,
			COALESCE(checkpoint_write_time,0)::float8,
			COALESCE(checkpoint_sync_time,0)::float8
		FROM pg_stat_bgwriter
	`).Scan(
		&info.CheckpointsTimed,
		&info.CheckpointsReq,
		&info.BuffersCheckpoint,
		&info.BuffersClean,
		&info.MaxwrittenClean,
		&info.BuffersBackend,
		&info.BuffersBackendFsync,
		&info.BuffersAlloc,
		&info.CheckpointWriteTimeMs,
		&info.CheckpointSyncTimeMs,
	)
	if err != nil {
		info.Available = false
		info.Error = err.Error()
		recs = append(recs, postgresRecommendation{Severity: "warning", Message: "No fue posible leer pg_stat_bgwriter"})
		return info, recs
	}

	info.CheckpointWriteTimeMs = round2(info.CheckpointWriteTimeMs)
	info.CheckpointSyncTimeMs = round2(info.CheckpointSyncTimeMs)

	if info.BuffersBackendFsync > 0 {
		recs = append(recs, postgresRecommendation{Severity: "warning", Message: "Se detectaron fsync en backend; revisar configuracion de checkpoints"})
	}
	if info.CheckpointsReq > info.CheckpointsTimed {
		recs = append(recs, postgresRecommendation{Severity: "warning", Message: "Hay mas checkpoints por demanda que por temporizador"})
	}

	return info, recs
}

func collectPostgresLongRunningQueries(ctx context.Context, dbConn *sql.DB) ([]postgresLongRunningQuery, []postgresRecommendation) {
	rows, err := dbConn.QueryContext(ctx, `
		SELECT
			COALESCE(pid,0)::bigint,
			COALESCE(usename,''),
			COALESCE(state,''),
			COALESCE(wait_event_type,''),
			COALESCE(wait_event,''),
			COALESCE(EXTRACT(EPOCH FROM (now() - query_start)),0)::bigint,
			LEFT(regexp_replace(COALESCE(query,''), '\\s+', ' ', 'g'), 220)
		FROM pg_stat_activity
		WHERE state = 'active'
		  AND query_start IS NOT NULL
		  AND pid <> pg_backend_pid()
		  AND query NOT ILIKE '%pg_stat_activity%'
		ORDER BY 6 DESC
		LIMIT 10
	`)
	if err != nil {
		return nil, []postgresRecommendation{{Severity: "warning", Message: "No fue posible consultar las consultas activas de PostgreSQL"}}
	}
	defer rows.Close()

	queries := make([]postgresLongRunningQuery, 0, 10)
	for rows.Next() {
		var q postgresLongRunningQuery
		if err := rows.Scan(&q.PID, &q.User, &q.State, &q.WaitEventType, &q.WaitEvent, &q.RunningSeconds, &q.QuerySnippet); err != nil {
			continue
		}
		q.QuerySnippet = strings.TrimSpace(q.QuerySnippet)
		if q.QuerySnippet == "" {
			q.QuerySnippet = "(consulta sin texto)"
		}
		queries = append(queries, q)
	}

	if err := rows.Err(); err != nil {
		return queries, []postgresRecommendation{{Severity: "warning", Message: "Lectura parcial de consultas activas"}}
	}

	var recs []postgresRecommendation
	if len(queries) > 0 {
		longest := queries[0]
		if longest.RunningSeconds >= 120 {
			recs = append(recs, postgresRecommendation{Severity: "critical", Message: fmt.Sprintf("Consulta activa con %d segundos de ejecucion", longest.RunningSeconds)})
		} else if longest.RunningSeconds >= 60 {
			recs = append(recs, postgresRecommendation{Severity: "warning", Message: fmt.Sprintf("Consulta activa prolongada (%d segundos)", longest.RunningSeconds)})
		}
	}

	return queries, recs
}

func mergePostgresRecommendations(groups ...[]postgresRecommendation) []postgresRecommendation {
	seen := map[string]bool{}
	out := make([]postgresRecommendation, 0)

	for _, group := range groups {
		for _, item := range group {
			item.Severity = normalizeRecommendationSeverity(item.Severity)
			item.Message = strings.TrimSpace(item.Message)
			if item.Message == "" {
				continue
			}
			key := item.Severity + "|" + strings.ToLower(item.Message)
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, item)
		}
	}

	sort.Slice(out, func(i, j int) bool {
		ri := recommendationSeverityRank(out[i].Severity)
		rj := recommendationSeverityRank(out[j].Severity)
		if ri != rj {
			return ri < rj
		}
		return out[i].Message < out[j].Message
	})

	return out
}

func resolvePostgresHealth(recs []postgresRecommendation, cluster postgresClusterInfo) string {
	hasWarning := false
	for _, rec := range recs {
		severity := normalizeRecommendationSeverity(rec.Severity)
		if severity == "critical" {
			return "critical"
		}
		if severity == "warning" {
			hasWarning = true
		}
	}

	if cluster.ConnectionUsagePct >= 95 || cluster.BlockedConnections > 0 {
		return "critical"
	}
	if hasWarning || cluster.ConnectionUsagePct >= 80 {
		return "warning"
	}
	return "healthy"
}

func normalizeRecommendationSeverity(severity string) string {
	s := strings.ToLower(strings.TrimSpace(severity))
	switch s {
	case "critical", "warning", "info":
		return s
	default:
		return "info"
	}
}

func recommendationSeverityRank(severity string) int {
	switch normalizeRecommendationSeverity(severity) {
	case "critical":
		return 0
	case "warning":
		return 1
	default:
		return 2
	}
}

func extractPostgresVersionLabel(full string) string {
	version := strings.TrimSpace(full)
	if idx := strings.Index(version, ","); idx > 0 {
		version = strings.TrimSpace(version[:idx])
	}
	parts := strings.Fields(version)
	if len(parts) >= 2 && strings.EqualFold(parts[0], "PostgreSQL") {
		return parts[0] + " " + parts[1]
	}
	if version == "" {
		return "PostgreSQL"
	}
	return version
}

func percentFloat(numerator, denominator float64) float64 {
	if denominator <= 0 {
		return 0
	}
	return (numerator / denominator) * 100
}

func cacheHitRatio(blocksHit, blocksRead float64) float64 {
	total := blocksHit + blocksRead
	if total <= 0 {
		return 100
	}
	return round2(percentFloat(blocksHit, total))
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
