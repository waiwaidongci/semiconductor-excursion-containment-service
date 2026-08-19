package containment

import (
	"errors"
	"sort"
	"time"
)

type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskElevated RiskLevel = "elevated"
	RiskCritical RiskLevel = "critical"
)

type Signal struct {
	Name       string
	Weight     int
	ObservedAt time.Time
	ToolID     string
}

func (signal Signal) Validate() error {
	if signal.Name == "" || signal.ToolID == "" {
		return errors.New("signal name and tool are required")
	}
	if signal.Weight < 1 || signal.Weight > 10 {
		return errors.New("signal weight must be between one and ten")
	}
	if signal.ObservedAt.IsZero() {
		return errors.New("signal observation time is required")
	}
	return nil
}

type Assessment struct {
	LotID      string
	Signals    []Signal
	Score      int
	Level      RiskLevel
	AssessedAt time.Time
}

func Assess(lot *Lot, signals []Signal, now time.Time) (Assessment, error) {
	if lot == nil {
		return Assessment{}, ErrMissingLot
	}
	if err := lot.Validate(); err != nil {
		return Assessment{}, err
	}
	validated := make([]Signal, len(signals))
	copy(validated, signals)
	for _, signal := range validated {
		if err := signal.Validate(); err != nil {
			return Assessment{}, err
		}
	}
	sort.SliceStable(validated, func(left, right int) bool {
		if validated[left].ToolID == validated[right].ToolID {
			return validated[left].Name < validated[right].Name
		}
		return validated[left].ToolID < validated[right].ToolID
	})
	score := 0
	seenTools := make(map[string]struct{})
	for _, signal := range validated {
		score += signal.Weight
		seenTools[signal.ToolID] = struct{}{}
	}
	if len(seenTools) > 1 {
		score += len(seenTools) * 2
	}
	if lot.State == StateInvestigating {
		score += 3
	}
	return Assessment{LotID: lot.ID, Signals: validated, Score: score, Level: classifyRisk(score), AssessedAt: now}, nil
}

func classifyRisk(score int) RiskLevel {
	switch {
	case score >= 20:
		return RiskCritical
	case score >= 8:
		return RiskElevated
	default:
		return RiskLow
	}
}

func (assessment Assessment) RequiresImmediateContainment() bool {
	return assessment.Level == RiskCritical
}

func (assessment Assessment) Tools() []string {
	seen := make(map[string]struct{})
	for _, signal := range assessment.Signals {
		seen[signal.ToolID] = struct{}{}
	}
	tools := make([]string, 0, len(seen))
	for tool := range seen {
		tools = append(tools, tool)
	}
	sort.Strings(tools)
	return tools
}
