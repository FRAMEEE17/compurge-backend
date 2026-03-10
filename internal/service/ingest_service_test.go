package service

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"compurge/internal/model"
	"github.com/xuri/excelize/v2"
)

func TestIngestServiceIngestCSVToEvents(t *testing.T) {
	file, err := os.Open("../../2026-03-09T17-03_export.csv")
	if err != nil {
		t.Fatalf("open sample csv: %v", err)
	}
	defer file.Close()

	service := IngestService{}
	events, err := service.IngestCSVToEvents(context.Background(), "req-test-1", file)
	if err != nil {
		t.Fatalf("ingest csv: %v", err)
	}

	if len(events) != 78 {
		t.Fatalf("expected 78 events, got %d", len(events))
	}

	first := events[0]
	if first.Period != model.PeriodFirstHalf || first.EventType != "Pass" {
		t.Fatalf("unexpected first event: %+v", first)
	}
	if first.X == nil || *first.X != 54.6 {
		t.Fatalf("unexpected first event X: %+v", first.X)
	}
}

func TestIngestServiceRequiresRequestID(t *testing.T) {
	service := IngestService{}
	_, err := service.IngestCSVToEvents(context.Background(), "", strings.NewReader("minute,second,period\n0,1,FirstHalf\n"))
	if err == nil {
		t.Fatal("expected error for empty request ID")
	}
}

func TestIngestServiceSupportsReorderedHeaderMapping(t *testing.T) {
	input := strings.NewReader(strings.Join([]string{
		"type,period,second,minute,outcomeType,x,endX,y,endY",
		"Pass,FirstHalf,24,0,Successful,27.09,20.475,19.924,35.564",
		"Carry,FirstHalf,23,0,Successful,28.455,27.09,15.028,19.924",
	}, "\n"))

	service := IngestService{}
	events, err := service.IngestCSVToEvents(context.Background(), "req-reordered", input)
	if err != nil {
		t.Fatalf("ingest reordered csv: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	first := events[0]
	if first.Minute != 0 || first.Second != 23 || first.EventType != "Carry" {
		t.Fatalf("unexpected first event after ordering: %+v", first)
	}

	second := events[1]
	if second.EventType != "Pass" || second.OutcomeType != "Successful" {
		t.Fatalf("unexpected second event: %+v", second)
	}
	if second.X == nil || *second.X != 27.09 {
		t.Fatalf("unexpected x coordinate: %+v", second.X)
	}
}

func TestIngestServiceRejectsMissingRequiredField(t *testing.T) {
	input := strings.NewReader(strings.Join([]string{
		"type,period,minute,outcomeType",
		"Pass,FirstHalf,0,Successful",
	}, "\n"))

	service := IngestService{}
	_, err := service.IngestCSVToEvents(context.Background(), "req-missing-second", input)
	if err == nil {
		t.Fatal("expected error for missing required second field")
	}
}

func TestIngestServiceSupportsProviderStyleAliases(t *testing.T) {
	input := strings.NewReader(strings.Join([]string{
		"match_minute,match_second,half,playerName,teamName,eventType,eventOutcome,startX,startY,to_x,to_y,expectedThreat",
		"12,10.5,H1,Player A,Team A,Pass,Successful,30.1,40.2,50.3,60.4,0.12",
	}, "\n"))

	service := IngestService{}
	events, err := service.IngestCSVToEvents(context.Background(), "req-provider-aliases", input)
	if err != nil {
		t.Fatalf("ingest provider-style aliases: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.Period != model.PeriodFirstHalf {
		t.Fatalf("expected first half, got %q", event.Period)
	}
	if event.PlayerName == nil || *event.PlayerName != "Player A" {
		t.Fatalf("unexpected player name: %+v", event.PlayerName)
	}
	if event.TeamName == nil || *event.TeamName != "Team A" {
		t.Fatalf("unexpected team name: %+v", event.TeamName)
	}
	if event.X == nil || *event.X != 30.1 {
		t.Fatalf("unexpected x coordinate: %+v", event.X)
	}
	if event.EndX == nil || *event.EndX != 50.3 {
		t.Fatalf("unexpected endX coordinate: %+v", event.EndX)
	}
	if event.XT == nil || *event.XT != 0.12 {
		t.Fatalf("unexpected xT: %+v", event.XT)
	}
}

func TestIngestServiceIngestXLSXToEvents(t *testing.T) {
	workbook := excelize.NewFile()
	sheet := workbook.GetSheetName(0)
	rows := [][]string{
		{"minute", "second", "period", "type", "outcomeType", "playerName"},
		{"0", "24", "FirstHalf", "Pass", "Successful", "Player A"},
		{"1", "45", "FirstHalf", "Pass", "Successful", "Player B"},
	}
	for i, row := range rows {
		cell, err := excelize.CoordinatesToCellName(1, i+1)
		if err != nil {
			t.Fatalf("cell name: %v", err)
		}
		if err := workbook.SetSheetRow(sheet, cell, &row); err != nil {
			t.Fatalf("set sheet row: %v", err)
		}
	}

	buffer, err := workbook.WriteToBuffer()
	if err != nil {
		t.Fatalf("write workbook: %v", err)
	}

	service := IngestService{}
	events, err := service.IngestXLSXToEvents(context.Background(), "req-xlsx", bytes.NewReader(buffer.Bytes()))
	if err != nil {
		t.Fatalf("ingest xlsx: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].PlayerName == nil || *events[0].PlayerName != "Player A" {
		t.Fatalf("unexpected first player name: %+v", events[0].PlayerName)
	}
}
