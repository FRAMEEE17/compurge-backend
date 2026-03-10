package model

import "fmt"

type Period string

const (
	PeriodUnknown    Period = ""
	PeriodFirstHalf  Period = "FirstHalf"
	PeriodSecondHalf Period = "SecondHalf"
)

type Event struct {
	Minute      int
	Second      float64
	Period      Period
	PlayerID    *string
	PlayerName  *string
	TeamID      *string
	TeamName    *string
	EventType   string
	OutcomeType string
	ProgPass    *float64
	ProgCarry   *float64
	XT          *float64
	X           *float64
	Y           *float64
	EndX        *float64
	EndY        *float64
}

func (e Event) MatchSecond() float64 {
	base := float64(e.Minute)*60 + e.Second
	if e.Period == PeriodSecondHalf && e.Minute < 45 {
		return 45*60 + base
	}
	return base
}

func ParsePeriod(value string) (Period, error) {
	switch value {
	case "", "Unknown":
		return PeriodUnknown, nil
	case "FirstHalf", "1", "H1":
		return PeriodFirstHalf, nil
	case "SecondHalf", "2", "H2":
		return PeriodSecondHalf, nil
	default:
		return PeriodUnknown, fmt.Errorf("unknown period %q", value)
	}
}
