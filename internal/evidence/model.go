package evidence

import (
	"errors"
	"fmt"
	"time"
)

var ErrEvidenceRejected = errors.New("evidence rejected")

type Item struct {
	ID        string
	LotID     string
	Kind      string
	Location  string
	Checksum  string
	CreatedAt time.Time
}

func (item Item) Validate() error {
	if item.ID == "" || item.LotID == "" || item.Kind == "" {
		return errors.New("evidence identity is incomplete")
	}
	if item.Location == "" || item.Checksum == "" {
		return errors.New("evidence location and checksum are required")
	}
	return nil
}

type Bundle struct {
	LotID string
	Items []Item
}

func (bundle Bundle) Validate() error {
	if bundle.LotID == "" {
		return errors.New("bundle lot id is required")
	}
	if len(bundle.Items) == 0 {
		return errors.New("bundle requires evidence")
	}
	seen := make(map[string]struct{}, len(bundle.Items))
	for _, item := range bundle.Items {
		if err := item.Validate(); err != nil {
			return err
		}
		if item.LotID != bundle.LotID {
			return fmt.Errorf("evidence %s belongs to lot %s", item.ID, item.LotID)
		}
		if _, exists := seen[item.ID]; exists {
			return fmt.Errorf("duplicate evidence %s", item.ID)
		}
		seen[item.ID] = struct{}{}
	}
	return nil
}

func (bundle Bundle) Clone() Bundle {
	clone := bundle
	clone.Items = append([]Item(nil), bundle.Items...)
	return clone
}

type Result struct {
	Stored   int
	Rejected int
	IDs      []string
}

func (result Result) Clone() Result {
	clone := result
	clone.IDs = append([]string(nil), result.IDs...)
	return clone
}
