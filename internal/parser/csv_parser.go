package parser

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"compurge/internal/model"
)

type CSVParser struct{}

func (p CSVParser) Parse(reader io.Reader) ([]model.Event, error) {
	r := csv.NewReader(bufio.NewReader(reader))

	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	indexes := buildHeaderIndex(header)
	events := make([]model.Event, 0, 64)

	for {
		record, err := r.Read()
		if err == io.EOF {
			return events, nil
		}
		if err != nil {
			return nil, fmt.Errorf("read record: %w", err)
		}

		event, err := parseEvent(record, indexes)
		if err != nil {
			return nil, fmt.Errorf("parse record: %w", err)
		}
		events = append(events, event)
	}
}

func buildHeaderIndex(header []string) map[string]int {
	indexes := make(map[string]int, len(header))
	for i, column := range header {
		indexes[normalizeHeader(column)] = i
	}
	return indexes
}

func normalizeHeader(value string) string {
	value = strings.TrimPrefix(value, "\ufeff")
	return strings.ToLower(strings.TrimSpace(value))
}

func parseEvent(record []string, indexes map[string]int) (model.Event, error) {
	minute, err := parseRequiredInt(record, indexes, "minute")
	if err != nil {
		return model.Event{}, fmt.Errorf("minute: %w", err)
	}
	second, err := parseRequiredFloat(record, indexes, "second")
	if err != nil {
		return model.Event{}, fmt.Errorf("second: %w", err)
	}
	periodRaw, err := getRequired(record, indexes, "period")
	if err != nil {
		return model.Event{}, fmt.Errorf("period: %w", err)
	}
	period, err := model.ParsePeriod(periodRaw)
	if err != nil {
		return model.Event{}, fmt.Errorf("period: %w", err)
	}

	event := model.Event{
		Minute:      minute,
		Second:      second,
		Period:      period,
		PlayerID:    parseOptionalString(record, indexes, "player_id"),
		PlayerName:  parseOptionalString(record, indexes, "player_name"),
		TeamID:      parseOptionalString(record, indexes, "team_id"),
		TeamName:    parseOptionalString(record, indexes, "team_name"),
		EventType:   getOptional(record, indexes, "type"),
		OutcomeType: getOptional(record, indexes, "outcometype"),
	}

	if event.ProgPass, err = parseOptionalFloat(record, indexes, "prog_pass"); err != nil {
		return model.Event{}, fmt.Errorf("prog_pass: %w", err)
	}
	if event.ProgCarry, err = parseOptionalFloat(record, indexes, "prog_carry"); err != nil {
		return model.Event{}, fmt.Errorf("prog_carry: %w", err)
	}
	if event.XT, err = parseOptionalFloat(record, indexes, "xt"); err != nil {
		return model.Event{}, fmt.Errorf("xt: %w", err)
	}
	if event.X, err = parseOptionalFloat(record, indexes, "x"); err != nil {
		return model.Event{}, fmt.Errorf("x: %w", err)
	}
	if event.Y, err = parseOptionalFloat(record, indexes, "y"); err != nil {
		return model.Event{}, fmt.Errorf("y: %w", err)
	}
	if event.EndX, err = parseOptionalFloat(record, indexes, "endx"); err != nil {
		return model.Event{}, fmt.Errorf("endx: %w", err)
	}
	if event.EndY, err = parseOptionalFloat(record, indexes, "endy"); err != nil {
		return model.Event{}, fmt.Errorf("endy: %w", err)
	}

	return event, nil
}

func getRequired(record []string, indexes map[string]int, column string) (string, error) {
	index, ok := indexes[column]
	if !ok {
		return "", fmt.Errorf("missing column %q", column)
	}
	if index >= len(record) {
		return "", fmt.Errorf("column %q out of range", column)
	}
	value := strings.TrimSpace(record[index])
	if value == "" {
		return "", fmt.Errorf("empty value for %q", column)
	}
	return value, nil
}

func getOptional(record []string, indexes map[string]int, column string) string {
	index, ok := indexes[column]
	if !ok || index >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[index])
}

func parseRequiredInt(record []string, indexes map[string]int, column string) (int, error) {
	value, err := getRequired(record, indexes, column)
	if err != nil {
		return 0, err
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse int %q: %w", value, err)
	}
	return parsed, nil
}

func parseRequiredFloat(record []string, indexes map[string]int, column string) (float64, error) {
	value, err := getRequired(record, indexes, column)
	if err != nil {
		return 0, err
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("parse float %q: %w", value, err)
	}
	return parsed, nil
}

func parseOptionalFloat(record []string, indexes map[string]int, column string) (*float64, error) {
	value := getOptional(record, indexes, column)
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, fmt.Errorf("parse float %q: %w", value, err)
	}
	return &parsed, nil
}

func parseOptionalString(record []string, indexes map[string]int, column string) *string {
	value := getOptional(record, indexes, column)
	if value == "" {
		return nil
	}
	return &value
}
