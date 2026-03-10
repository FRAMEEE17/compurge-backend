package parser

import (
	"os"
	"strings"
	"testing"

	"compurge/internal/model"
)

func TestCSVParserParseSampleExport(t *testing.T) {
	file, err := os.Open("../../2026-03-09T17-03_export.csv")
	if err != nil {
		t.Fatalf("open sample csv: %v", err)
	}
	defer file.Close()

	parser := CSVParser{}
	events, err := parser.Parse(file)
	if err != nil {
		t.Fatalf("parse sample csv: %v", err)
	}

	if len(events) != 78 {
		t.Fatalf("expected 78 events, got %d", len(events))
	}

	first := events[0]
	if first.Minute != 0 || first.Second != 9 || first.Period != model.PeriodFirstHalf {
		t.Fatalf("unexpected first event: %+v", first)
	}
	if first.EventType != "Pass" || first.OutcomeType != "Unsuccessful" {
		t.Fatalf("unexpected first event type fields: %+v", first)
	}
	if first.X == nil || *first.X != 54.6 {
		t.Fatalf("unexpected x coordinate: %+v", first.X)
	}
}

func TestCSVParserParseMissingColumn(t *testing.T) {
	parser := CSVParser{}
	_, err := parser.Parse(strings.NewReader("minute,second,type\n0,1,Pass\n"))
	if err == nil {
		t.Fatal("expected error for missing period column")
	}
}
