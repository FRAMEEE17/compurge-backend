package service

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"compurge/internal/model"
	"compurge/internal/parser"
	"compurge/internal/repository/sqlite"
)

type IngestService struct{}

func (s IngestService) IngestCSVToEvents(ctx context.Context, requestID string, reader io.Reader) ([]model.Event, error) {
	csvReader := csv.NewReader(bufio.NewReader(reader))
	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("read csv header: %w", err)
	}

	next := func() ([]string, error) {
		record, err := csvReader.Read()
		if err != nil {
			return nil, err
		}
		return record, nil
	}

	return s.ingestRecords(ctx, requestID, header, next)
}

func (s IngestService) IngestXLSXToEvents(ctx context.Context, requestID string, reader io.Reader) ([]model.Event, error) {
	xlsxParser := parser.XLSXParser{}
	header, next, err := xlsxParser.Open(reader)
	if err != nil {
		return nil, fmt.Errorf("open xlsx parser: %w", err)
	}

	return s.ingestRecords(ctx, requestID, header, next)
}

func (s IngestService) IngestFileToEvents(ctx context.Context, requestID, filename string, reader io.Reader) ([]model.Event, error) {
	switch strings.ToLower(filepath.Ext(strings.TrimSpace(filename))) {
	case ".xlsx":
		return s.IngestXLSXToEvents(ctx, requestID, reader)
	default:
		return s.IngestCSVToEvents(ctx, requestID, reader)
	}
}

func (s IngestService) ingestRecords(
	ctx context.Context,
	requestID string,
	header []string,
	next func() ([]string, error),
) ([]model.Event, error) {
	repo, err := sqlite.NewStagingRepository(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("create staging repository: %w", err)
	}
	defer repo.Close()

	if err := repo.CreateRawStagingTable(ctx, header); err != nil {
		return nil, fmt.Errorf("create staging table: %w", err)
	}

	for {
		record, err := next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read record: %w", err)
		}
		if err := repo.InsertRawRow(ctx, record); err != nil {
			return nil, fmt.Errorf("insert raw record: %w", err)
		}
	}

	rows, err := repo.QueryNormalizedEvents(ctx, header)
	if err != nil {
		return nil, fmt.Errorf("query normalized events: %w", err)
	}
	defer rows.Close()

	events := make([]model.Event, 0, 64)
	for rows.Next() {
		event, err := scanNormalizedEvent(rows)
		if err != nil {
			return nil, fmt.Errorf("scan normalized event: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate normalized events: %w", err)
	}

	return events, nil
}

func scanNormalizedEvent(scanner interface {
	Scan(dest ...any) error
}) (model.Event, error) {
	var (
		periodRaw   string
		minute      int
		second      float64
		playerID    sqlNullString
		playerName  sqlNullString
		teamID      sqlNullString
		teamName    sqlNullString
		eventType   string
		outcomeType string
		progPass    sqlNullString
		progCarry   sqlNullString
		xt          sqlNullString
		x           sqlNullString
		y           sqlNullString
		endX        sqlNullString
		endY        sqlNullString
	)

	if err := scanner.Scan(
		&periodRaw,
		&minute,
		&second,
		&playerID,
		&playerName,
		&teamID,
		&teamName,
		&eventType,
		&outcomeType,
		&progPass,
		&progCarry,
		&xt,
		&x,
		&y,
		&endX,
		&endY,
	); err != nil {
		return model.Event{}, fmt.Errorf("scan row: %w", err)
	}

	period, err := model.ParsePeriod(strings.TrimSpace(periodRaw))
	if err != nil {
		return model.Event{}, fmt.Errorf("parse period: %w", err)
	}

	return model.Event{
		Minute:      minute,
		Second:      second,
		Period:      period,
		PlayerID:    playerID.StringPtr(),
		PlayerName:  playerName.StringPtr(),
		TeamID:      teamID.StringPtr(),
		TeamName:    teamName.StringPtr(),
		EventType:   eventType,
		OutcomeType: outcomeType,
		ProgPass:    progPass.Float64Ptr(),
		ProgCarry:   progCarry.Float64Ptr(),
		XT:          xt.Float64Ptr(),
		X:           x.Float64Ptr(),
		Y:           y.Float64Ptr(),
		EndX:        endX.Float64Ptr(),
		EndY:        endY.Float64Ptr(),
	}, nil
}

type sqlNullString struct {
	String string
	Valid  bool
}

func (n *sqlNullString) Scan(value any) error {
	if value == nil {
		n.String = ""
		n.Valid = false
		return nil
	}
	switch v := value.(type) {
	case string:
		n.String = v
	case []byte:
		n.String = string(v)
	default:
		return fmt.Errorf("unsupported scan type %T", value)
	}
	n.Valid = true
	return nil
}

func (n sqlNullString) Float64Ptr() *float64 {
	if !n.Valid {
		return nil
	}
	trimmed := strings.TrimSpace(n.String)
	if trimmed == "" {
		return nil
	}
	value, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return nil
	}
	return &value
}

func (n sqlNullString) StringPtr() *string {
	if !n.Valid {
		return nil
	}
	trimmed := strings.TrimSpace(n.String)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
