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
	clone.Findings = report.Findings
	return clone
}

func (report Report) Sorted() Report {
	sort.SliceStable(report.Findings, func(left, right int) bool {
		if report.Findings[left].Chamber.ToolID == report.Findings[right].Chamber.ToolID {
			return report.Findings[left].Chamber.ChamberID < report.Findings[right].Chamber.ChamberID
		}
		return report.Findings[left].Chamber.ToolID < report.Findings[right].Chamber.ToolID
	})
	return report
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
