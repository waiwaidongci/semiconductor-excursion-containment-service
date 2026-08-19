package evidence

import (
	"errors"
	"fmt"
)

type Resource interface {
	Write(item Item) error
	Close() error
}

type ResourceFactory interface {
	Open(item Item) (Resource, error)
}

type Transaction interface {
	Stage(item Item) error
	Commit() error
	Rollback() error
}

type TransactionFactory interface {
	Begin(lotID string) (Transaction, error)
}

type Processor struct {
	resources    ResourceFactory
	transactions TransactionFactory
}

func NewProcessor(resources ResourceFactory, transactions TransactionFactory) *Processor {
	return &Processor{resources: resources, transactions: transactions}
}

func (processor *Processor) Process(bundle Bundle) (Result, error) {
	if err := bundle.Validate(); err != nil {
		return Result{}, err
	}
	transaction, err := processor.transactions.Begin(bundle.LotID)
	if err != nil {
		return Result{}, fmt.Errorf("begin evidence transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = transaction.Rollback()
		}
	}()
	result := Result{IDs: make([]string, 0, len(bundle.Items))}
	for _, item := range bundle.Items {
		if err := processor.processItem(transaction, item); err != nil {
			return result.Clone(), err
		}
		result.Stored++
		result.IDs = append(result.IDs, item.ID)
	}
	if err := transaction.Commit(); err != nil {
		return result.Clone(), fmt.Errorf("commit evidence transaction: %w", err)
	}
	committed = true
	return result.Clone(), nil
}

func (processor *Processor) processItem(transaction Transaction, item Item) (err error) {
	resource, err := processor.resources.Open(item)
	if err != nil {
		return fmt.Errorf("open evidence %s: %w", item.ID, err)
	}
	defer func() {
		if closeErr := resource.Close(); closeErr != nil {
			if err == nil {
				err = fmt.Errorf("close evidence %s: %w", item.ID, closeErr)
			} else {
				err = errors.Join(err, fmt.Errorf("close evidence %s: %w", item.ID, closeErr))
			}
		}
	}()
	if err := resource.Write(item); err != nil {
		return fmt.Errorf("write evidence %s: %w", item.ID, err)
	}
	if err := transaction.Stage(item); err != nil {
		return fmt.Errorf("stage evidence %s: %w", item.ID, err)
	}
	return nil
}

func ProcessAll(processor *Processor, bundles []Bundle) ([]Result, error) {
	results := make([]Result, 0, len(bundles))
	for _, bundle := range bundles {
		result, err := processor.Process(bundle)
		if err != nil {
			return results, err
		}
		results = append(results, result.Clone())
	}
	return results, nil
}
