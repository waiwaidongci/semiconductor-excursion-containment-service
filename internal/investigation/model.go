package investigation

import (
	"errors"
	"sort"
	"time"
)

type Chamber struct {
	ToolID    string
	ChamberID string
	RecipeID  string
}

func (chamber Chamber) Validate() error {
	if chamber.ToolID == "" || chamber.ChamberID == "" || chamber.RecipeID == "" {
		return errors.New("tool, chamber, and recipe are required")
	}
	return nil
}

type Finding struct {
	Chamber    Chamber
	Severity   string
	Summary    string
	ObservedAt time.Time
}

func (finding Finding) Validate() error {
	if err := finding.Chamber.Validate(); err != nil {
		return err
	}
	if finding.Severity == "" || finding.Summary == "" {
		return errors.New("finding severity and summary are required")
	}
	return nil
}

type Report struct {
	LotID      string
	Findings   []Finding
	StartedAt  time.Time
	FinishedAt time.Time
}

func (report Report) Clone() Report {
	clone := report
	clone.Findings = make([]Finding, len(report.Findings))
	copy(clone.Findings, report.Findings)
	return clone
}

func (report Report) Sorted() Report {
	sorted := report
	sorted.Findings = make([]Finding, len(report.Findings))
	copy(sorted.Findings, report.Findings)
	sort.SliceStable(sorted.Findings, func(left, right int) bool {
		if sorted.Findings[left].Chamber.ToolID == sorted.Findings[right].Chamber.ToolID {
			return sorted.Findings[left].Chamber.ChamberID < sorted.Findings[right].Chamber.ChamberID
		}
		return sorted.Findings[left].Chamber.ToolID < sorted.Findings[right].Chamber.ToolID
	})
	return sorted
}

func (report Report) Critical() []Finding {
	critical := make([]Finding, 0)
	for _, finding := range report.Findings {
		if finding.Severity == "critical" {
			critical = append(critical, finding)
		}
	}
	return critical
}
