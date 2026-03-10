package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"

	"compurge/internal/model"
	"compurge/internal/service"
)

const (
	DefaultMaxUploadSize       = 10 << 20
	DefaultMaxConcurrentIngest = 50
	DefaultPreviewLimit        = 50
	MaxPreviewLimit            = 100
)

type TimestampHandler struct {
	service       service.TimestampService
	maxUploadSize int64
	ingestSem     chan struct{}
	requestSeq    atomic.Uint64
}

func NewTimestampHandler(maxUploadSize int64, maxConcurrentIngest int) *TimestampHandler {
	if maxUploadSize <= 0 {
		maxUploadSize = DefaultMaxUploadSize
	}
	if maxConcurrentIngest <= 0 {
		maxConcurrentIngest = DefaultMaxConcurrentIngest
	}

	return &TimestampHandler{
		service:       service.NewTimestampService(),
		maxUploadSize: maxUploadSize,
		ingestSem:     make(chan struct{}, maxConcurrentIngest),
	}
}

func (h *TimestampHandler) Router() http.Handler {
	router := chi.NewRouter()
	router.Handle("/timestamps", h.withIngestSlot(http.HandlerFunc(h.handleTimestamps)))
	router.Post("/timestamps/json", h.handleTimestampsJSON)
	router.Handle("/parse-preview", h.withIngestSlot(http.HandlerFunc(h.handleParsePreview)))
	router.Post("/parse-preview/json", h.handleParsePreviewJSON)
	return router
}

func (h *TimestampHandler) withIngestSlot(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case h.ingestSem <- struct{}{}:
			defer func() { <-h.ingestSem }()
			next.ServeHTTP(w, r)
		default:
			http.Error(w, "server busy", http.StatusTooManyRequests)
		}
	})
}

func (h *TimestampHandler) handleTimestamps(w http.ResponseWriter, r *http.Request) {
	limitReader := http.MaxBytesReader(w, r.Body, h.maxUploadSize)
	r.Body = limitReader
	if err := r.ParseMultipartForm(h.maxUploadSize); err != nil {
		http.Error(w, "file too large or invalid multipart request", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("eventData")
	if err != nil {
		http.Error(w, "eventData file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filter, options, err := parseRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	requestID := strings.TrimSpace(r.FormValue("requestId"))
	if requestID == "" {
		requestID = h.nextRequestID()
	}

	clips, err := h.service.GenerateClipRanges(r.Context(), requestID, header.Filename, file, filter, options)
	if err != nil {
		http.Error(w, "failed to generate timestamps", http.StatusInternalServerError)
		return
	}

	response := timestampResponse{
		RequestID:  requestID,
		ClipRanges: clips,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *TimestampHandler) nextRequestID() string {
	seq := h.requestSeq.Add(1)
	return fmt.Sprintf("req_%d_%d", time.Now().UnixNano(), seq)
}

type timestampResponse struct {
	RequestID  string            `json:"requestId"`
	ClipRanges []model.ClipRange `json:"clipRanges"`
}

type timestampJSONRequest struct {
	RequestID        string         `json:"requestId"`
	Limit            *int           `json:"limit"`
	EventDataCSV     string         `json:"eventDataCsv"`
	Events           []eventPayload `json:"events"`
	Period           string         `json:"period"`
	EventType        string         `json:"eventType"`
	OutcomeType      string         `json:"outcomeType"`
	MinMinute        *int           `json:"minMinute"`
	MaxMinute        *int           `json:"maxMinute"`
	MinX             *float64       `json:"minX"`
	MaxX             *float64       `json:"maxX"`
	MinY             *float64       `json:"minY"`
	MaxY             *float64       `json:"maxY"`
	MinEndX          *float64       `json:"minEndX"`
	MaxEndX          *float64       `json:"maxEndX"`
	MinEndY          *float64       `json:"minEndY"`
	MaxEndY          *float64       `json:"maxEndY"`
	MinXT            *float64       `json:"minXT"`
	MaxXT            *float64       `json:"maxXT"`
	MinProgPass      *float64       `json:"minProgPass"`
	MaxProgPass      *float64       `json:"maxProgPass"`
	MinProgCarry     *float64       `json:"minProgCarry"`
	MaxProgCarry     *float64       `json:"maxProgCarry"`
	PreRollSeconds   *float64       `json:"preRollSeconds"`
	PostRollSeconds  *float64       `json:"postRollSeconds"`
	TimelineOffset   *float64       `json:"timelineOffset"`
	FirstHalfOffset  *float64       `json:"firstHalfOffset"`
	SecondHalfOffset *float64       `json:"secondHalfOffset"`
	MergeGapSeconds  *float64       `json:"mergeGapSeconds"`
}

type eventPayload struct {
	Minute      int      `json:"minute"`
	Second      float64  `json:"second"`
	Period      string   `json:"period"`
	PlayerID    *string  `json:"playerId"`
	PlayerName  *string  `json:"playerName"`
	TeamID      *string  `json:"teamId"`
	TeamName    *string  `json:"teamName"`
	EventType   string   `json:"eventType"`
	OutcomeType string   `json:"outcomeType"`
	ProgPass    *float64 `json:"progPass"`
	ProgCarry   *float64 `json:"progCarry"`
	XT          *float64 `json:"xT"`
	X           *float64 `json:"x"`
	Y           *float64 `json:"y"`
	EndX        *float64 `json:"endX"`
	EndY        *float64 `json:"endY"`
}

type previewResponse struct {
	RequestID string         `json:"requestId"`
	Events    []eventPreview `json:"events"`
}

type eventPreview struct {
	Minute      int      `json:"minute"`
	Second      float64  `json:"second"`
	Period      string   `json:"period"`
	MatchSecond float64  `json:"matchSecond"`
	PlayerName  *string  `json:"playerName,omitempty"`
	TeamName    *string  `json:"teamName,omitempty"`
	EventType   string   `json:"eventType"`
	OutcomeType string   `json:"outcomeType"`
	XT          *float64 `json:"xT,omitempty"`
}

func parseRequest(r *http.Request) (model.EventFilter, model.ClipOptions, error) {
	var (
		filter  model.EventFilter
		options model.ClipOptions
		err     error
	)

	if periodValue := strings.TrimSpace(r.FormValue("period")); periodValue != "" {
		filter.Period, err = model.ParsePeriod(periodValue)
		if err != nil {
			return model.EventFilter{}, model.ClipOptions{}, fmt.Errorf("invalid period: %w", err)
		}
	}

	filter.EventType = strings.TrimSpace(r.FormValue("eventType"))
	filter.OutcomeType = strings.TrimSpace(r.FormValue("outcomeType"))

	if filter.MinMinute, err = parseOptionalInt(r.FormValue("minMinute")); err != nil {
		return model.EventFilter{}, model.ClipOptions{}, fmt.Errorf("invalid minMinute: %w", err)
	}
	if filter.MaxMinute, err = parseOptionalInt(r.FormValue("maxMinute")); err != nil {
		return model.EventFilter{}, model.ClipOptions{}, fmt.Errorf("invalid maxMinute: %w", err)
	}
	if filter.MinX, err = parseOptionalFloat(r.FormValue("minX")); err != nil {
		return model.EventFilter{}, model.ClipOptions{}, fmt.Errorf("invalid minX: %w", err)
	}
	if filter.MaxX, err = parseOptionalFloat(r.FormValue("maxX")); err != nil {
		return model.EventFilter{}, model.ClipOptions{}, fmt.Errorf("invalid maxX: %w", err)
	}
	if filter.MinY, err = parseOptionalFloat(r.FormValue("minY")); err != nil {
		return model.EventFilter{}, model.ClipOptions{}, fmt.Errorf("invalid minY: %w", err)
	}
	if filter.MaxY, err = parseOptionalFloat(r.FormValue("maxY")); err != nil {
		return model.EventFilter{}, model.ClipOptions{}, fmt.Errorf("invalid maxY: %w", err)
	}
	if filter.MinEndX, err = parseOptionalFloat(r.FormValue("minEndX")); err != nil {
		return model.EventFilter{}, model.ClipOptions{}, fmt.Errorf("invalid minEndX: %w", err)
	}
	if filter.MaxEndX, err = parseOptionalFloat(r.FormValue("maxEndX")); err != nil {
		return model.EventFilter{}, model.ClipOptions{}, fmt.Errorf("invalid maxEndX: %w", err)
	}
	if filter.MinEndY, err = parseOptionalFloat(r.FormValue("minEndY")); err != nil {
		return model.EventFilter{}, model.ClipOptions{}, fmt.Errorf("invalid minEndY: %w", err)
	}
	if filter.MaxEndY, err = parseOptionalFloat(r.FormValue("maxEndY")); err != nil {
		return model.EventFilter{}, model.ClipOptions{}, fmt.Errorf("invalid maxEndY: %w", err)
	}
	if filter.MinXT, err = parseOptionalFloat(r.FormValue("minXT")); err != nil {
		return model.EventFilter{}, model.ClipOptions{}, fmt.Errorf("invalid minXT: %w", err)
	}
	if filter.MaxXT, err = parseOptionalFloat(r.FormValue("maxXT")); err != nil {
		return model.EventFilter{}, model.ClipOptions{}, fmt.Errorf("invalid maxXT: %w", err)
	}
	if filter.MinProgPass, err = parseOptionalFloat(r.FormValue("minProgPass")); err != nil {
		return model.EventFilter{}, model.ClipOptions{}, fmt.Errorf("invalid minProgPass: %w", err)
	}
	if filter.MaxProgPass, err = parseOptionalFloat(r.FormValue("maxProgPass")); err != nil {
		return model.EventFilter{}, model.ClipOptions{}, fmt.Errorf("invalid maxProgPass: %w", err)
	}
	if filter.MinProgCarry, err = parseOptionalFloat(r.FormValue("minProgCarry")); err != nil {
		return model.EventFilter{}, model.ClipOptions{}, fmt.Errorf("invalid minProgCarry: %w", err)
	}
	if filter.MaxProgCarry, err = parseOptionalFloat(r.FormValue("maxProgCarry")); err != nil {
		return model.EventFilter{}, model.ClipOptions{}, fmt.Errorf("invalid maxProgCarry: %w", err)
	}

	options.PreRollSeconds, err = parseFloatWithDefault(r.FormValue("preRollSeconds"), 3)
	if err != nil {
		return model.EventFilter{}, model.ClipOptions{}, fmt.Errorf("invalid preRollSeconds: %w", err)
	}
	options.PostRollSeconds, err = parseFloatWithDefault(r.FormValue("postRollSeconds"), 3)
	if err != nil {
		return model.EventFilter{}, model.ClipOptions{}, fmt.Errorf("invalid postRollSeconds: %w", err)
	}
	options.TimelineOffset, err = parseFloatWithDefault(r.FormValue("timelineOffset"), 0)
	if err != nil {
		return model.EventFilter{}, model.ClipOptions{}, fmt.Errorf("invalid timelineOffset: %w", err)
	}
	if options.FirstHalfOffset, err = parseOptionalFloat(r.FormValue("firstHalfOffset")); err != nil {
		return model.EventFilter{}, model.ClipOptions{}, fmt.Errorf("invalid firstHalfOffset: %w", err)
	}
	if options.SecondHalfOffset, err = parseOptionalFloat(r.FormValue("secondHalfOffset")); err != nil {
		return model.EventFilter{}, model.ClipOptions{}, fmt.Errorf("invalid secondHalfOffset: %w", err)
	}
	options.MergeGapSeconds, err = parseFloatWithDefault(r.FormValue("mergeGapSeconds"), 0)
	if err != nil {
		return model.EventFilter{}, model.ClipOptions{}, fmt.Errorf("invalid mergeGapSeconds: %w", err)
	}

	return filter, options, nil
}

func parseOptionalInt(value string) (*int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseOptionalFloat(value string) (*float64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseFloatWithDefault(value string, defaultValue float64) (float64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultValue, nil
	}
	return strconv.ParseFloat(value, 64)
}

func (h *TimestampHandler) handleTimestampsJSON(w http.ResponseWriter, r *http.Request) {
	var request timestampJSONRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid JSON request", http.StatusBadRequest)
		return
	}

	filter, options, err := parseJSONRequest(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	requestID := strings.TrimSpace(request.RequestID)
	if requestID == "" {
		requestID = h.nextRequestID()
	}

	var clips []model.ClipRange
	switch {
	case strings.TrimSpace(request.EventDataCSV) != "":
		clips, err = h.service.GenerateClipRangesFromCSVText(r.Context(), requestID, request.EventDataCSV, filter, options)
	case len(request.Events) > 0:
		events, conversionErr := convertEventPayloads(request.Events)
		if conversionErr != nil {
			http.Error(w, conversionErr.Error(), http.StatusBadRequest)
			return
		}
		clips, err = h.service.GenerateClipRangesFromEvents(r.Context(), events, filter, options)
	default:
		http.Error(w, "eventDataCsv or events is required", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "failed to generate timestamps", http.StatusInternalServerError)
		return
	}

	response := timestampResponse{
		RequestID:  requestID,
		ClipRanges: clips,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *TimestampHandler) handleParsePreview(w http.ResponseWriter, r *http.Request) {
	limitReader := http.MaxBytesReader(w, r.Body, h.maxUploadSize)
	r.Body = limitReader
	if err := r.ParseMultipartForm(h.maxUploadSize); err != nil {
		http.Error(w, "file too large or invalid multipart request", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("eventData")
	if err != nil {
		http.Error(w, "eventData file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	requestID := strings.TrimSpace(r.FormValue("requestId"))
	if requestID == "" {
		requestID = h.nextRequestID()
	}

	limit, err := parsePreviewLimit(r.FormValue("limit"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	events, err := h.service.PreviewEvents(r.Context(), requestID, header.Filename, file, limit)
	if err != nil {
		http.Error(w, "failed to parse preview", http.StatusInternalServerError)
		return
	}

	writePreviewResponse(w, requestID, events)
}

func (h *TimestampHandler) handleParsePreviewJSON(w http.ResponseWriter, r *http.Request) {
	var request timestampJSONRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid JSON request", http.StatusBadRequest)
		return
	}

	requestID := strings.TrimSpace(request.RequestID)
	if requestID == "" {
		requestID = h.nextRequestID()
	}

	limit := DefaultPreviewLimit
	if request.Limit != nil {
		if *request.Limit <= 0 || *request.Limit > MaxPreviewLimit {
			http.Error(w, fmt.Sprintf("limit must be between 1 and %d", MaxPreviewLimit), http.StatusBadRequest)
			return
		}
		limit = *request.Limit
	}

	var (
		events []model.Event
		err    error
	)
	switch {
	case strings.TrimSpace(request.EventDataCSV) != "":
		events, err = h.service.PreviewEventsFromCSVText(r.Context(), requestID, request.EventDataCSV, limit)
	case len(request.Events) > 0:
		var converted []model.Event
		converted, err = convertEventPayloads(request.Events)
		if err == nil {
			events, err = h.service.PreviewEventsFromEvents(r.Context(), converted, limit)
		}
	default:
		http.Error(w, "eventDataCsv or events is required", http.StatusBadRequest)
		return
	}
	if err != nil {
		if strings.Contains(err.Error(), "invalid event period") || strings.Contains(err.Error(), "required") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "failed to parse preview", http.StatusInternalServerError)
		return
	}

	writePreviewResponse(w, requestID, events)
}

func writePreviewResponse(w http.ResponseWriter, requestID string, events []model.Event) {
	response := previewResponse{
		RequestID: requestID,
		Events:    toEventPreviews(events),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func parseJSONRequest(request timestampJSONRequest) (model.EventFilter, model.ClipOptions, error) {
	var (
		filter  model.EventFilter
		options model.ClipOptions
		err     error
	)

	if strings.TrimSpace(request.Period) != "" {
		filter.Period, err = model.ParsePeriod(strings.TrimSpace(request.Period))
		if err != nil {
			return model.EventFilter{}, model.ClipOptions{}, fmt.Errorf("invalid period: %w", err)
		}
	}

	filter.EventType = strings.TrimSpace(request.EventType)
	filter.OutcomeType = strings.TrimSpace(request.OutcomeType)
	filter.MinMinute = request.MinMinute
	filter.MaxMinute = request.MaxMinute
	filter.MinX = request.MinX
	filter.MaxX = request.MaxX
	filter.MinY = request.MinY
	filter.MaxY = request.MaxY
	filter.MinEndX = request.MinEndX
	filter.MaxEndX = request.MaxEndX
	filter.MinEndY = request.MinEndY
	filter.MaxEndY = request.MaxEndY
	filter.MinXT = request.MinXT
	filter.MaxXT = request.MaxXT
	filter.MinProgPass = request.MinProgPass
	filter.MaxProgPass = request.MaxProgPass
	filter.MinProgCarry = request.MinProgCarry
	filter.MaxProgCarry = request.MaxProgCarry

	options.PreRollSeconds = floatOrDefault(request.PreRollSeconds, 3)
	options.PostRollSeconds = floatOrDefault(request.PostRollSeconds, 3)
	options.TimelineOffset = floatOrDefault(request.TimelineOffset, 0)
	options.FirstHalfOffset = request.FirstHalfOffset
	options.SecondHalfOffset = request.SecondHalfOffset
	options.MergeGapSeconds = floatOrDefault(request.MergeGapSeconds, 0)

	return filter, options, nil
}

func convertEventPayloads(payloads []eventPayload) ([]model.Event, error) {
	events := make([]model.Event, 0, len(payloads))
	for _, payload := range payloads {
		period, err := model.ParsePeriod(strings.TrimSpace(payload.Period))
		if err != nil {
			return nil, fmt.Errorf("invalid event period: %w", err)
		}

		events = append(events, model.Event{
			Minute:      payload.Minute,
			Second:      payload.Second,
			Period:      period,
			PlayerID:    payload.PlayerID,
			PlayerName:  payload.PlayerName,
			TeamID:      payload.TeamID,
			TeamName:    payload.TeamName,
			EventType:   strings.TrimSpace(payload.EventType),
			OutcomeType: strings.TrimSpace(payload.OutcomeType),
			ProgPass:    payload.ProgPass,
			ProgCarry:   payload.ProgCarry,
			XT:          payload.XT,
			X:           payload.X,
			Y:           payload.Y,
			EndX:        payload.EndX,
			EndY:        payload.EndY,
		})
	}
	return events, nil
}

func floatOrDefault(value *float64, defaultValue float64) float64 {
	if value == nil {
		return defaultValue
	}
	return *value
}

func parsePreviewLimit(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return DefaultPreviewLimit, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid limit: %w", err)
	}
	if parsed <= 0 || parsed > MaxPreviewLimit {
		return 0, fmt.Errorf("limit must be between 1 and %d", MaxPreviewLimit)
	}
	return parsed, nil
}

func toEventPreviews(events []model.Event) []eventPreview {
	previews := make([]eventPreview, 0, len(events))
	for _, event := range events {
		previews = append(previews, eventPreview{
			Minute:      event.Minute,
			Second:      event.Second,
			Period:      string(event.Period),
			MatchSecond: event.MatchSecond(),
			PlayerName:  event.PlayerName,
			TeamName:    event.TeamName,
			EventType:   event.EventType,
			OutcomeType: event.OutcomeType,
			XT:          event.XT,
		})
	}
	return previews
}
