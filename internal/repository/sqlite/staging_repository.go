package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type StagingRepository struct {
	db *sql.DB
}

type StagingSchema struct {
	columnByField map[string]string
}

func NewStagingRepository(ctx context.Context, requestID string) (*StagingRepository, error) {
	if strings.TrimSpace(requestID) == "" {
		return nil, fmt.Errorf("request ID is required")
	}

	dsn := fmt.Sprintf("file:req_%s?mode=memory&cache=shared", sanitizeRequestID(requestID))
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	if _, err := db.ExecContext(ctx, "PRAGMA busy_timeout = 5000"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("set busy timeout: %w", err)
	}
	if _, err := db.ExecContext(ctx, "PRAGMA temp_store = MEMORY"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("set temp store: %w", err)
	}

	return &StagingRepository{db: db}, nil
}

func (r *StagingRepository) Close() error {
	if r == nil || r.db == nil {
		return nil
	}
	if err := r.db.Close(); err != nil {
		return fmt.Errorf("close sqlite database: %w", err)
	}
	return nil
}

func (r *StagingRepository) CreateRawStagingTable(ctx context.Context, columns []string) error {
	if len(columns) == 0 {
		return fmt.Errorf("staging columns are required")
	}

	columnDefs := make([]string, 0, len(columns))
	for i := range columns {
		columnDefs = append(columnDefs, fmt.Sprintf("col%d TEXT", i+1))
	}

	query := fmt.Sprintf("CREATE TABLE raw_staging (%s)", strings.Join(columnDefs, ", "))
	if _, err := r.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("create raw staging table: %w", err)
	}
	return nil
}

func (r *StagingRepository) InsertRawRow(ctx context.Context, values []string) error {
	if len(values) == 0 {
		return fmt.Errorf("raw row values are required")
	}

	placeholders := make([]string, 0, len(values))
	args := make([]any, 0, len(values))
	for _, value := range values {
		placeholders = append(placeholders, "?")
		args = append(args, value)
	}

	query := fmt.Sprintf("INSERT INTO raw_staging VALUES (%s)", strings.Join(placeholders, ", "))
	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("insert raw row: %w", err)
	}
	return nil
}

func (r *StagingRepository) QueryNormalizedEvents(ctx context.Context, header []string) (*sql.Rows, error) {
	schema, err := BuildStagingSchema(header)
	if err != nil {
		return nil, fmt.Errorf("build staging schema: %w", err)
	}

	query := fmt.Sprintf(`
SELECT
	%s AS period,
	CAST(%s AS INTEGER) AS minute,
	CAST(%s AS REAL) AS second,
	%s AS player_id,
	%s AS player_name,
	%s AS team_id,
	%s AS team_name,
	%s AS event_type,
	%s AS outcome_type,
	%s AS prog_pass,
	%s AS prog_carry,
	%s AS xt,
	%s AS x,
	%s AS y,
	%s AS end_x,
	%s AS end_y
FROM raw_staging
ORDER BY CAST(%s AS INTEGER), CAST(%s AS REAL)`,
		schema.requiredText("period"),
		schema.requiredText("minute"),
		schema.requiredText("second"),
		schema.optionalText("player_id"),
		schema.optionalText("player_name"),
		schema.optionalText("team_id"),
		schema.optionalText("team_name"),
		schema.optionalText("event_type"),
		schema.optionalText("outcome_type"),
		schema.optionalText("prog_pass"),
		schema.optionalText("prog_carry"),
		schema.optionalText("xt"),
		schema.optionalText("x"),
		schema.optionalText("y"),
		schema.optionalText("end_x"),
		schema.optionalText("end_y"),
		schema.requiredText("minute"),
		schema.requiredText("second"),
	)
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query normalized events: %w", err)
	}
	return rows, nil
}

func BuildStagingSchema(header []string) (StagingSchema, error) {
	columnByField := make(map[string]string)
	for i, raw := range header {
		field := canonicalFieldName(raw)
		if field == "" {
			continue
		}
		columnByField[field] = fmt.Sprintf("col%d", i+1)
	}

	required := []string{"period", "minute", "second"}
	for _, field := range required {
		if _, ok := columnByField[field]; !ok {
			return StagingSchema{}, fmt.Errorf("missing required field %q", field)
		}
	}

	return StagingSchema{columnByField: columnByField}, nil
}

func (s StagingSchema) requiredText(field string) string {
	column := s.columnByField[field]
	return fmt.Sprintf("TRIM(%s)", column)
}

func (s StagingSchema) optionalText(field string) string {
	column, ok := s.columnByField[field]
	if !ok {
		return "NULL"
	}
	return fmt.Sprintf("NULLIF(TRIM(%s), '')", column)
}

func canonicalFieldName(value string) string {
	normalized := normalizeHeaderToken(value)
	switch normalized {
	case "period", "half":
		return "period"
	case "minute", "matchminute", "matchmin", "match_minute", "min":
		return "minute"
	case "second", "seconds", "matchsecond", "matchsec", "match_second", "sec":
		return "second"
	case "type", "eventtype", "event", "actiontype", "primaryevent":
		return "event_type"
	case "outcometype", "outcome", "result", "eventoutcome":
		return "outcome_type"
	case "playerid", "player", "playeridentifier":
		return "player_id"
	case "playername", "playerfullname", "athletename":
		return "player_name"
	case "teamid", "teamidentifier", "squadid":
		return "team_id"
	case "teamname", "team", "squadname":
		return "team_name"
	case "progpass":
		return "prog_pass"
	case "progcarry":
		return "prog_carry"
	case "xt", "expectedthreat":
		return "xt"
	case "x", "startx":
		return "x"
	case "y", "starty":
		return "y"
	case "endx", "tox", "destinationx":
		return "end_x"
	case "endy", "toy", "destinationy":
		return "end_y"
	default:
		return ""
	}
}

func normalizeHeaderToken(value string) string {
	value = strings.TrimPrefix(value, "\ufeff")
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.NewReplacer("_", "", "-", "", " ", "").Replace(value)
	return value
}

func sanitizeRequestID(requestID string) string {
	var builder strings.Builder
	builder.Grow(len(requestID))
	for _, r := range requestID {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			builder.WriteRune(r + ('a' - 'A'))
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		default:
			builder.WriteByte('_')
		}
	}
	return builder.String()
}
