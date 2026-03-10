package handler

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestTimestampHandlerTimestamps(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fileWriter, err := writer.CreateFormFile("eventData", "events.csv")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}

	file, err := os.Open("../../2026-03-09T17-03_export.csv")
	if err != nil {
		t.Fatalf("open sample csv: %v", err)
	}
	defer file.Close()

	if _, err := file.WriteTo(fileWriter); err != nil {
		t.Fatalf("copy sample csv: %v", err)
	}

	_ = writer.WriteField("period", "FirstHalf")
	_ = writer.WriteField("eventType", "Pass")
	_ = writer.WriteField("outcomeType", "Successful")
	_ = writer.WriteField("preRollSeconds", "2")
	_ = writer.WriteField("postRollSeconds", "3")
	_ = writer.WriteField("timelineOffset", "5")

	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/timestamps", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()

	handler := NewTimestampHandler(DefaultMaxUploadSize, 1)
	handler.Router().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		RequestID  string `json:"requestId"`
		ClipRanges []struct {
			ResolvedStartTime float64 `json:"resolvedStartTime"`
			ResolvedEndTime   float64 `json:"resolvedEndTime"`
		} `json:"clipRanges"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.RequestID == "" {
		t.Fatal("expected request ID")
	}
	if len(response.ClipRanges) == 0 {
		t.Fatal("expected clip ranges")
	}
}

func TestTimestampHandlerMissingFile(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("eventType", "Pass")
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/timestamps", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()

	handler := NewTimestampHandler(DefaultMaxUploadSize, 1)
	handler.Router().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestTimestampHandlerRejectsOversizedBody(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fileWriter, err := writer.CreateFormFile("eventData", "events.csv")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}

	largeContent := strings.Repeat("a", 1024)
	for i := 0; i < 64; i++ {
		if _, err := fileWriter.Write([]byte(largeContent)); err != nil {
			t.Fatalf("write large content: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/timestamps", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()

	handler := NewTimestampHandler(1024, 1)
	handler.Router().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestTimestampHandlerXLSXUpload(t *testing.T) {
	workbook := excelize.NewFile()
	sheet := workbook.GetSheetName(0)
	rows := [][]string{
		{"minute", "second", "period", "type", "outcomeType"},
		{"0", "24", "FirstHalf", "Pass", "Successful"},
		{"1", "45", "FirstHalf", "Pass", "Successful"},
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

	xlsxBody, err := workbook.WriteToBuffer()
	if err != nil {
		t.Fatalf("write workbook: %v", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, err := writer.CreateFormFile("eventData", "events.xlsx")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := fileWriter.Write(xlsxBody.Bytes()); err != nil {
		t.Fatalf("write xlsx body: %v", err)
	}
	_ = writer.WriteField("period", "FirstHalf")
	_ = writer.WriteField("eventType", "Pass")
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/timestamps", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()

	handler := NewTimestampHandler(DefaultMaxUploadSize, 1)
	handler.Router().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestTimestampHandlerJSONWithCSVText(t *testing.T) {
	body := strings.NewReader(`{
		"eventDataCsv": "minute,second,period,type,outcomeType\n0,24,FirstHalf,Pass,Successful\n1,45,FirstHalf,Pass,Successful",
		"period": "FirstHalf",
		"eventType": "Pass",
		"preRollSeconds": 2,
		"postRollSeconds": 3,
		"timelineOffset": 5
	}`)

	req := httptest.NewRequest(http.MethodPost, "/timestamps/json", body)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler := NewTimestampHandler(DefaultMaxUploadSize, 1)
	handler.Router().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestTimestampHandlerJSONWithEvents(t *testing.T) {
	body := strings.NewReader(`{
		"events": [
			{"minute": 0, "second": 24, "period": "FirstHalf", "eventType": "Pass", "outcomeType": "Successful"},
			{"minute": 1, "second": 45, "period": "FirstHalf", "eventType": "Pass", "outcomeType": "Successful"}
		],
		"period": "FirstHalf",
		"eventType": "Pass"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/timestamps/json", body)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler := NewTimestampHandler(DefaultMaxUploadSize, 1)
	handler.Router().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestTimestampHandlerParsePreview(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fileWriter, err := writer.CreateFormFile("eventData", "events.csv")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}

	file, err := os.Open("../../2026-03-09T17-03_export.csv")
	if err != nil {
		t.Fatalf("open sample csv: %v", err)
	}
	defer file.Close()

	if _, err := file.WriteTo(fileWriter); err != nil {
		t.Fatalf("copy sample csv: %v", err)
	}

	_ = writer.WriteField("limit", "3")
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/parse-preview", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()

	handler := NewTimestampHandler(DefaultMaxUploadSize, 1)
	handler.Router().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		RequestID string `json:"requestId"`
		Events    []struct {
			Minute      int     `json:"minute"`
			MatchSecond float64 `json:"matchSecond"`
			EventType   string  `json:"eventType"`
		} `json:"events"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.RequestID == "" {
		t.Fatal("expected request ID")
	}
	if len(response.Events) != 3 {
		t.Fatalf("expected 3 preview events, got %d", len(response.Events))
	}
}

func TestTimestampHandlerParsePreviewJSON(t *testing.T) {
	body := strings.NewReader(`{
		"limit": 2,
		"eventDataCsv": "minute,second,period,type,outcomeType\n0,24,FirstHalf,Pass,Successful\n1,45,FirstHalf,Carry,Successful\n46,10,SecondHalf,Pass,Successful"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/parse-preview/json", body)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler := NewTimestampHandler(DefaultMaxUploadSize, 1)
	handler.Router().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Events []struct {
			Minute int    `json:"minute"`
			Period string `json:"period"`
		} `json:"events"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(response.Events) != 2 {
		t.Fatalf("expected 2 preview events, got %d", len(response.Events))
	}
	if response.Events[0].Period != "FirstHalf" {
		t.Fatalf("expected first event period FirstHalf, got %q", response.Events[0].Period)
	}
}
