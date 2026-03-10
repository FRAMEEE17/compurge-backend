package service

import (
	"reflect"
	"testing"

	"compurge/internal/model"
)

func TestHighlightServiceFilterEvents(t *testing.T) {
	minMinute := 0
	maxMinute := 10
	tests := []struct {
		name     string
		filter   model.EventFilter
		expected []float64
	}{
		{
			name: "first half successful passes in minute range",
			filter: model.EventFilter{
				Period:      model.PeriodFirstHalf,
				EventType:   "Pass",
				OutcomeType: "Successful",
				MinMinute:   &minMinute,
				MaxMinute:   &maxMinute,
			},
			expected: []float64{24, 105, 199, 394, 485},
		},
		{
			name: "x filter only",
			filter: model.EventFilter{
				MinX: floatPtr(40),
			},
			expected: []float64{9, 466, 660},
		},
		{
			name: "xT threshold",
			filter: model.EventFilter{
				MinXT: floatPtr(0.01),
			},
			expected: []float64{660},
		},
		{
			name: "progressive pass threshold",
			filter: model.EventFilter{
				EventType:   "Pass",
				MinProgPass: floatPtr(10),
			},
			expected: []float64{660},
		},
		{
			name: "progressive carry threshold",
			filter: model.EventFilter{
				EventType:    "Carry",
				MinProgCarry: floatPtr(5),
			},
			expected: []float64{466},
		},
		{
			name: "start y zone filter",
			filter: model.EventFilter{
				MinY: floatPtr(45),
			},
			expected: []float64{466, 485},
		},
		{
			name: "end x final third filter",
			filter: model.EventFilter{
				MinEndX: floatPtr(45),
			},
			expected: []float64{9, 660},
		},
		{
			name: "end y box-side filter",
			filter: model.EventFilter{
				MinEndY: floatPtr(35),
				MaxEndY: floatPtr(61),
			},
			expected: []float64{24, 105, 466, 660},
		},
	}

	service := HighlightService{}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			events := sampleEvents()
			filtered := service.FilterEvents(events, tc.filter)
			got := make([]float64, 0, len(filtered))
			for _, event := range filtered {
				got = append(got, event.MatchSecond())
			}
			if !reflect.DeepEqual(tc.expected, got) {
				t.Fatalf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestMergeClipRanges(t *testing.T) {
	tests := []struct {
		name     string
		input    []model.ClipRange
		gap      float64
		expected []model.ClipRange
	}{
		{
			name: "merge overlapping and near-adjacent ranges",
			gap:  1,
			input: []model.ClipRange{
				{ResolvedStartTime: 10, ResolvedEndTime: 20},
				{ResolvedStartTime: 19.5, ResolvedEndTime: 30},
				{ResolvedStartTime: 31, ResolvedEndTime: 35},
				{ResolvedStartTime: 40, ResolvedEndTime: 45},
			},
			expected: []model.ClipRange{
				{ResolvedStartTime: 10, ResolvedEndTime: 35},
				{ResolvedStartTime: 40, ResolvedEndTime: 45},
			},
		},
		{
			name:     "empty input",
			gap:      0,
			input:    nil,
			expected: nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := MergeClipRanges(tc.input, tc.gap)
			if !reflect.DeepEqual(tc.expected, got) {
				t.Fatalf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestHighlightServiceBuildClipRanges(t *testing.T) {
	service := HighlightService{}
	clips := service.BuildClipRanges(
		sampleEvents(),
		model.EventFilter{
			Period:      model.PeriodFirstHalf,
			EventType:   "Pass",
			OutcomeType: "Successful",
		},
		model.ClipOptions{
			PreRollSeconds:  2,
			PostRollSeconds: 3,
			TimelineOffset:  5,
			MergeGapSeconds: 0,
		},
	)

	if len(clips) == 0 {
		t.Fatal("expected clip ranges")
	}

	first := clips[0]
	if first.ResolvedStartTime != 27 || first.ResolvedEndTime != 32 {
		t.Fatalf("unexpected first clip: %+v", first)
	}
}

func TestHighlightServiceBuildClipRangesUsesHalfSpecificOffsets(t *testing.T) {
	service := HighlightService{}
	clips := service.BuildClipRanges(
		[]model.Event{
			{Minute: 0, Second: 24, Period: model.PeriodFirstHalf, EventType: "Pass", OutcomeType: "Successful"},
			{Minute: 46, Second: 0, Period: model.PeriodSecondHalf, EventType: "Pass", OutcomeType: "Successful"},
		},
		model.EventFilter{
			EventType: "Pass",
		},
		model.ClipOptions{
			PreRollSeconds:   2,
			PostRollSeconds:  3,
			TimelineOffset:   5,
			FirstHalfOffset:  floatPtr(10),
			SecondHalfOffset: floatPtr(120),
			MergeGapSeconds:  0,
		},
	)

	if len(clips) != 2 {
		t.Fatalf("expected 2 clip ranges, got %d", len(clips))
	}

	if clips[0].ResolvedEventTime != 34 {
		t.Fatalf("expected first-half offset to apply, got %+v", clips[0])
	}
	if clips[1].ResolvedEventTime != 2880 {
		t.Fatalf("expected second-half offset to apply, got %+v", clips[1])
	}
}

func sampleEvents() []model.Event {
	return []model.Event{
		{Minute: 0, Second: 9, Period: model.PeriodFirstHalf, EventType: "Pass", OutcomeType: "Unsuccessful", X: floatPtr(54.6), Y: floatPtr(5.37), EndX: floatPtr(62.16), EndY: floatPtr(21.76)},
		{Minute: 0, Second: 24, Period: model.PeriodFirstHalf, EventType: "Pass", OutcomeType: "Successful", X: floatPtr(27.09), Y: floatPtr(19.92), EndX: floatPtr(20.47), EndY: floatPtr(35.56), ProgPass: floatPtr(-5.36)},
		{Minute: 1, Second: 45, Period: model.PeriodFirstHalf, EventType: "Pass", OutcomeType: "Successful", X: floatPtr(28.98), Y: floatPtr(16.04), EndX: floatPtr(30.03), EndY: floatPtr(35.29), ProgPass: floatPtr(-2.91)},
		{Minute: 3, Second: 19, Period: model.PeriodFirstHalf, EventType: "Pass", OutcomeType: "Successful", X: floatPtr(30.87), Y: floatPtr(19.92), EndX: floatPtr(24.15), EndY: floatPtr(34.13), ProgPass: floatPtr(-5.39)},
		{Minute: 6, Second: 34, Period: model.PeriodFirstHalf, EventType: "Pass", OutcomeType: "Successful", X: floatPtr(20.05), Y: floatPtr(17.88), EndX: floatPtr(14.80), EndY: floatPtr(33.11), ProgPass: floatPtr(-3.73)},
		{Minute: 7, Second: 46, Period: model.PeriodFirstHalf, EventType: "Carry", OutcomeType: "Successful", X: floatPtr(44.62), Y: floatPtr(52.90), EndX: floatPtr(33.70), EndY: floatPtr(47.12), ProgCarry: floatPtr(9.22)},
		{Minute: 8, Second: 5, Period: model.PeriodFirstHalf, EventType: "Pass", OutcomeType: "Successful", X: floatPtr(33.7), Y: floatPtr(47.12), EndX: floatPtr(24.99), EndY: floatPtr(34.34), ProgPass: floatPtr(-7.51)},
		{Minute: 11, Second: 0, Period: model.PeriodFirstHalf, EventType: "Pass", OutcomeType: "Successful", X: floatPtr(42.94), Y: floatPtr(40.32), EndX: floatPtr(88.62), EndY: floatPtr(35.29), ProgPass: floatPtr(45.94), XT: floatPtr(0.095)},
	}
}

func floatPtr(value float64) *float64 {
	return &value
}
