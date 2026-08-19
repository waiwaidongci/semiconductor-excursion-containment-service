package evidence

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

var errResourceClose = errors.New("resource close failed")

type trackedResource struct {
	open *int
	fail bool
}

func (resource *trackedResource) Write(item Item) error {
	if resource.fail {
		return ErrEvidenceRejected
	}
	return nil
}

func (resource *trackedResource) Close() error {
	*resource.open--
	if resource.fail {
		return errResourceClose
	}
	return nil
}

type trackedResources struct {
	open        int
	maximumOpen int
	failID      string
}

func (factory *trackedResources) Open(item Item) (Resource, error) {
	factory.open++
	if factory.open > factory.maximumOpen {
		factory.maximumOpen = factory.open
	}
	return &trackedResource{open: &factory.open, fail: item.ID == factory.failID}, nil
}

type trackedTransaction struct {
	staged     []string
	committed  bool
	rolledBack bool
}

func (transaction *trackedTransaction) Stage(item Item) error {
	transaction.staged = append(transaction.staged, item.ID)
	return nil
}
func (transaction *trackedTransaction) Commit() error   { transaction.committed = true; return nil }
func (transaction *trackedTransaction) Rollback() error { transaction.rolledBack = true; return nil }

type trackedTransactions struct{ current *trackedTransaction }

func (factory *trackedTransactions) Begin(lotID string) (Transaction, error) {
	factory.current = &trackedTransaction{}
	return factory.current, nil
}

func TestEvidenceResourcesCloseAndErrorsSurvive(t *testing.T) {
	items := make([]Item, 0, 12)
	for index := 0; index < 12; index++ {
		items = append(items, Item{ID: fmt.Sprintf("E-%02d", index), LotID: "LOT-E", Kind: "image", Location: fmt.Sprintf("s3://e/%d", index), Checksum: "sha256", CreatedAt: time.Now()})
	}
	resources := &trackedResources{}
	transactions := &trackedTransactions{}
	processor := NewProcessor(resources, transactions)
	result, err := processor.Process(Bundle{LotID: "LOT-E", Items: items})
	if err != nil {
		t.Fatal(err)
	}
	if resources.open != 0 || resources.maximumOpen != 1 || result.Stored != len(items) || !transactions.current.committed {
		t.Fatalf("resource lifecycle mismatch: open=%d max=%d result=%#v tx=%#v", resources.open, resources.maximumOpen, result, transactions.current)
	}
	resources.failID = "E-05"
	_, err = processor.Process(Bundle{LotID: "LOT-E", Items: items})
	if !errors.Is(err, ErrEvidenceRejected) || !errors.Is(err, errResourceClose) {
		t.Fatalf("write and close errors were not both preserved: %v", err)
	}
	if resources.open != 0 || !transactions.current.rolledBack || transactions.current.committed {
		t.Fatalf("failed bundle leaked resource or transaction: open=%d tx=%#v", resources.open, transactions.current)
	}
}
