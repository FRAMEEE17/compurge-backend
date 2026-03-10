package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"compurge/internal/model"
)

type TimestampService struct {
	ingest    IngestService
	highlight HighlightService
}

func NewTimestampService() TimestampService {
	return TimestampService{
		ingest:    IngestService{},
		highlight: HighlightService{},
	}
}

func (s TimestampService) GenerateClipRanges(
	ctx context.Context,
	requestID string,
	filename string,
	reader io.Reader,
	filter model.EventFilter,
	options model.ClipOptions,
) ([]model.ClipRange, error) {
	events, err := s.ingest.IngestFileToEvents(ctx, requestID, filename, reader)
	if err != nil {
		return nil, fmt.Errorf("ingest events: %w", err)
	}

	clips := s.highlight.BuildClipRanges(events, filter, options)
	return clips, nil
}

func (s TimestampService) GenerateClipRangesFromCSVText(
	ctx context.Context,
	requestID string,
	csvText string,
	filter model.EventFilter,
	options model.ClipOptions,
) ([]model.ClipRange, error) {
	if strings.TrimSpace(csvText) == "" {
		return nil, fmt.Errorf("eventDataCsv is required")
	}

	reader := strings.NewReader(csvText)
	return s.GenerateClipRanges(ctx, requestID, "events.csv", reader, filter, options)
}

func (s TimestampService) GenerateClipRangesFromEvents(
	_ context.Context,
	events []model.Event,
	filter model.EventFilter,
	options model.ClipOptions,
) ([]model.ClipRange, error) {
	if len(events) == 0 {
		return nil, fmt.Errorf("events are required")
	}
	clips := s.highlight.BuildClipRanges(events, filter, options)
	return clips, nil
}

func (s TimestampService) PreviewEvents(
	ctx context.Context,
	requestID string,
	filename string,
	reader io.Reader,
	limit int,
) ([]model.Event, error) {
	events, err := s.ingest.IngestFileToEvents(ctx, requestID, filename, reader)
	if err != nil {
		return nil, fmt.Errorf("ingest events: %w", err)
	}
	return limitEvents(events, limit), nil
}

func (s TimestampService) PreviewEventsFromCSVText(
	ctx context.Context,
	requestID string,
	csvText string,
	limit int,
) ([]model.Event, error) {
	if strings.TrimSpace(csvText) == "" {
		return nil, fmt.Errorf("eventDataCsv is required")
	}
	reader := strings.NewReader(csvText)
	return s.PreviewEvents(ctx, requestID, "events.csv", reader, limit)
}

func (s TimestampService) PreviewEventsFromEvents(
	_ context.Context,
	events []model.Event,
	limit int,
) ([]model.Event, error) {
	if len(events) == 0 {
		return nil, fmt.Errorf("events are required")
	}
	return limitEvents(events, limit), nil
}

func limitEvents(events []model.Event, limit int) []model.Event {
	if limit <= 0 || limit >= len(events) {
		return events
	}
	return events[:limit]
}

func CSVHeader(csvText string) ([]string, error) {
	reader := csv.NewReader(strings.NewReader(csvText))
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read csv header: %w", err)
	}
	return header, nil
}
