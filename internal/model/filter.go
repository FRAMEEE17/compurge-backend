package model

type EventFilter struct {
	Period       Period
	EventType    string
	OutcomeType  string
	MinMinute    *int
	MaxMinute    *int
	MinX         *float64
	MaxX         *float64
	MinY         *float64
	MaxY         *float64
	MinEndX      *float64
	MaxEndX      *float64
	MinEndY      *float64
	MaxEndY      *float64
	MinXT        *float64
	MaxXT        *float64
	MinProgPass  *float64
	MaxProgPass  *float64
	MinProgCarry *float64
	MaxProgCarry *float64
}

type ClipOptions struct {
	PreRollSeconds   float64
	PostRollSeconds  float64
	TimelineOffset   float64
	FirstHalfOffset  *float64
	SecondHalfOffset *float64
	MergeGapSeconds  float64
}

type ClipRange struct {
	EventTime         float64 `json:"eventTime"`
	ResolvedEventTime float64 `json:"resolvedEventTime"`
	ResolvedStartTime float64 `json:"resolvedStartTime"`
	ResolvedEndTime   float64 `json:"resolvedEndTime"`
	SourceEventType   string  `json:"sourceEventType"`
	SourceOutcomeType string  `json:"sourceOutcomeType"`
}
