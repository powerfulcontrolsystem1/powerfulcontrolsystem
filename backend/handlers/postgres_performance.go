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

// PostgresPerformanceHandler expone metricas operativas del motor PostgreSQL para el panel super.
func PostgresPerformanceHandler(dbEmpresas, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
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
