package infra

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
)

// contextKey is a custom type for context keys to avoid collisions.
// clean-arch: Using typed context keys is a Go best practice
type contextKey int

const (
	// queryInfoKey is the context key for storing query information.
	queryInfoKey contextKey = iota
)

// SQLTracer logs all SQL queries executed through pgx.
// Implements pgx.QueryTracer interface.
type SQLTracer struct{}

// queryInfo stores information about a SQL query for logging.
type queryInfo struct {
	SQL       string
	Args      []any
	StartTime time.Time
}

// TraceQueryStart is called at the beginning of Query, QueryRow, and Exec calls.
func (t *SQLTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	info := queryInfo{
		SQL:       data.SQL,
		Args:      data.Args,
		StartTime: time.Now(),
	}
	return context.WithValue(ctx, queryInfoKey, info)
}

// TraceQueryEnd is called at the end of Query, QueryRow, and Exec calls.
func (t *SQLTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	info, ok := ctx.Value(queryInfoKey).(queryInfo)
	if !ok {
		return
	}

	duration := time.Since(info.StartTime)

	// Build log attributes
	attrs := []any{
		"sql", info.SQL,
		"duration_ms", duration.Milliseconds(),
		"rows_affected", data.CommandTag.RowsAffected(),
	}

	// Only log args in debug mode to avoid sensitive data exposure
	if slog.Default().Enabled(ctx, slog.LevelDebug) {
		attrs = append(attrs, "args", info.Args)
	}

	if data.Err != nil {
		attrs = append(attrs, "error", data.Err.Error())
		slog.Error("SQL Query Failed", attrs...)
	} else {
		// Log slow queries as warnings
		if duration > 100*time.Millisecond {
			slog.Warn("Slow SQL Query", attrs...)
		} else {
			slog.Debug("SQL Query", attrs...)
		}
	}
}

// =============================================================================
// Batch Tracing (optional, for batch operations)
// =============================================================================

// TraceBatchStart is called at the beginning of SendBatch calls.
func (t *SQLTracer) TraceBatchStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceBatchStartData) context.Context {
	info := queryInfo{
		StartTime: time.Now(),
	}
	return context.WithValue(ctx, queryInfoKey, info)
}

// TraceBatchQuery is called for each query in a batch.
func (t *SQLTracer) TraceBatchQuery(ctx context.Context, _ *pgx.Conn, data pgx.TraceBatchQueryData) {
	slog.Debug("Batch Query",
		"sql", data.SQL,
		"args_count", len(data.Args),
	)
}

// TraceBatchEnd is called at the end of SendBatch calls.
func (t *SQLTracer) TraceBatchEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceBatchEndData) {
	info, ok := ctx.Value(queryInfoKey).(queryInfo)
	if !ok {
		return
	}

	duration := time.Since(info.StartTime)

	if data.Err != nil {
		slog.Error("Batch Failed",
			"duration_ms", duration.Milliseconds(),
			"error", data.Err.Error(),
		)
	} else {
		slog.Debug("Batch Completed",
			"duration_ms", duration.Milliseconds(),
		)
	}
}

// =============================================================================
// Connection Tracing (optional, for connection lifecycle)
// =============================================================================

// TraceConnectStart is called at the beginning of Connect calls.
func (t *SQLTracer) TraceConnectStart(ctx context.Context, data pgx.TraceConnectStartData) context.Context {
	slog.Debug("Database Connect Start",
		"host", data.ConnConfig.Host,
		"port", data.ConnConfig.Port,
		"database", data.ConnConfig.Database,
	)
	info := queryInfo{
		StartTime: time.Now(),
	}
	return context.WithValue(ctx, queryInfoKey, info)
}

// TraceConnectEnd is called at the end of Connect calls.
func (t *SQLTracer) TraceConnectEnd(ctx context.Context, data pgx.TraceConnectEndData) {
	info, ok := ctx.Value(queryInfoKey).(queryInfo)
	if !ok {
		return
	}

	duration := time.Since(info.StartTime)

	if data.Err != nil {
		slog.Error("Database Connect Failed",
			"duration_ms", duration.Milliseconds(),
			"error", data.Err.Error(),
		)
	} else {
		slog.Info("Database Connected",
			"duration_ms", duration.Milliseconds(),
		)
	}
}
