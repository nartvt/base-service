package infra

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
)

// SQLTracer logs all SQL queries executed through pgx
type SQLTracer struct{}

type queryInfo struct {
	SQL       string
	Args      []any
	StartTime time.Time
}

// TraceQueryStart is called at the beginning of Query, QueryRow, and Exec calls
func (t *SQLTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	// Store query info in context
	info := queryInfo{
		SQL:       data.SQL,
		Args:      data.Args,
		StartTime: time.Now(),
	}
	return context.WithValue(ctx, "query_info", info)
}

// TraceQueryEnd is called at the end of Query, QueryRow, and Exec calls
func (t *SQLTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	// Retrieve query info from context
	info, ok := ctx.Value("query_info").(queryInfo)
	if !ok {
		return
	}

	// Calculate query duration
	duration := time.Since(info.StartTime)

	// Log the SQL query
	if data.Err != nil {
		// Log errors
		slog.Error("SQL Query Failed",
			"sql", info.SQL,
			"args", info.Args,
			"duration", duration,
			"error", data.Err,
		)
	} else {
		// Log successful queries
		slog.Info("SQL Query",
			"sql", info.SQL,
			"args", info.Args,
			"duration", duration,
			"rows_affected", data.CommandTag.RowsAffected(),
		)
	}
}
