package service

import (
	"sort"

	"compurge/internal/model"
)

type HighlightService struct{}

func (s HighlightService) FilterEvents(events []model.Event, filter model.EventFilter) []model.Event {
	filtered := make([]model.Event, 0, len(events))
	for _, event := range events {
		if !matchesFilter(event, filter) {
			continue
		}
		filtered = append(filtered, event)
	}
	return filtered
}

func (s HighlightService) BuildClipRanges(events []model.Event, filter model.EventFilter, options model.ClipOptions) []model.ClipRange {
	filtered := s.FilterEvents(events, filter)
	clips := make([]model.ClipRange, 0, len(filtered))

	for _, event := range filtered {
		eventTime := event.MatchSecond()
		resolvedEventTime := eventTime + resolveTimelineOffset(event, options)
		start := resolvedEventTime - options.PreRollSeconds
		if start < 0 {
			start = 0
		}
		end := resolvedEventTime + options.PostRollSeconds
		clips = append(clips, model.ClipRange{
			EventTime:         eventTime,
			ResolvedEventTime: resolvedEventTime,
			ResolvedStartTime: start,
			ResolvedEndTime:   end,
			SourceEventType:   event.EventType,
			SourceOutcomeType: event.OutcomeType,
		})
	}

	return MergeClipRanges(clips, options.MergeGapSeconds)
}

func resolveTimelineOffset(event model.Event, options model.ClipOptions) float64 {
	switch event.Period {
	case model.PeriodFirstHalf:
		if options.FirstHalfOffset != nil {
			return *options.FirstHalfOffset
		}
	case model.PeriodSecondHalf:
		if options.SecondHalfOffset != nil {
			return *options.SecondHalfOffset
		}
	}
	return options.TimelineOffset
}

func matchesFilter(event model.Event, filter model.EventFilter) bool {
	if filter.Period != model.PeriodUnknown && event.Period != filter.Period {
		return false
	}
	if filter.EventType != "" && event.EventType != filter.EventType {
		return false
	}
	if filter.OutcomeType != "" && event.OutcomeType != filter.OutcomeType {
		return false
	}
	if filter.MinMinute != nil && event.Minute < *filter.MinMinute {
		return false
	}
	if filter.MaxMinute != nil && event.Minute > *filter.MaxMinute {
		return false
	}
	if filter.MinX != nil {
		if event.X == nil || *event.X < *filter.MinX {
			return false
		}
	}
	if filter.MaxX != nil {
		if event.X == nil || *event.X > *filter.MaxX {
			return false
		}
	}
	if filter.MinY != nil {
		if event.Y == nil || *event.Y < *filter.MinY {
			return false
		}
	}
	if filter.MaxY != nil {
		if event.Y == nil || *event.Y > *filter.MaxY {
			return false
		}
	}
	if filter.MinEndX != nil {
		if event.EndX == nil || *event.EndX < *filter.MinEndX {
			return false
		}
	}
	if filter.MaxEndX != nil {
		if event.EndX == nil || *event.EndX > *filter.MaxEndX {
			return false
		}
	}
	if filter.MinEndY != nil {
		if event.EndY == nil || *event.EndY < *filter.MinEndY {
			return false
		}
	}
	if filter.MaxEndY != nil {
		if event.EndY == nil || *event.EndY > *filter.MaxEndY {
			return false
		}
	}
	if filter.MinXT != nil {
		if event.XT == nil || *event.XT < *filter.MinXT {
			return false
		}
	}
	if filter.MaxXT != nil {
		if event.XT == nil || *event.XT > *filter.MaxXT {
			return false
		}
	}
	if filter.MinProgPass != nil {
		if event.ProgPass == nil || *event.ProgPass < *filter.MinProgPass {
			return false
		}
	}
	if filter.MaxProgPass != nil {
		if event.ProgPass == nil || *event.ProgPass > *filter.MaxProgPass {
			return false
		}
	}
	if filter.MinProgCarry != nil {
		if event.ProgCarry == nil || *event.ProgCarry < *filter.MinProgCarry {
			return false
		}
	}
	if filter.MaxProgCarry != nil {
		if event.ProgCarry == nil || *event.ProgCarry > *filter.MaxProgCarry {
			return false
		}
	}
	return true
}

func MergeClipRanges(clips []model.ClipRange, mergeGapSeconds float64) []model.ClipRange {
	if len(clips) == 0 {
		return nil
	}

	sorted := append([]model.ClipRange(nil), clips...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].ResolvedStartTime == sorted[j].ResolvedStartTime {
			return sorted[i].ResolvedEndTime < sorted[j].ResolvedEndTime
		}
		return sorted[i].ResolvedStartTime < sorted[j].ResolvedStartTime
	})

	merged := make([]model.ClipRange, 0, len(sorted))
	current := sorted[0]

	for _, clip := range sorted[1:] {
		if clip.ResolvedStartTime <= current.ResolvedEndTime+mergeGapSeconds {
			if clip.ResolvedEndTime > current.ResolvedEndTime {
				current.ResolvedEndTime = clip.ResolvedEndTime
			}
			continue
		}
		merged = append(merged, current)
		current = clip
	}

	merged = append(merged, current)
	return merged
}
