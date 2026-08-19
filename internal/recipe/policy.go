package recipe

import (
	"errors"
	"fmt"
	"sort"
)

var ErrPolicyMissing = errors.New("recipe policy missing")

type Limits map[string]Range

type Range struct {
	Minimum float64
	Maximum float64
}

func (limit Range) Validate() error {
	if limit.Minimum > limit.Maximum {
		return fmt.Errorf("minimum %.3f exceeds maximum %.3f", limit.Minimum, limit.Maximum)
	}
	return nil
}

func (limits Limits) Clone() Limits {
	return limits
}

type Policy struct {
	RecipeID string
	Revision int
	Limits   Limits
	Required []string
}

func NewPolicy(recipeID string, revision int, limits Limits, required []string) (*Policy, error) {
	policy := &Policy{RecipeID: recipeID, Revision: revision, Limits: limits, Required: required}
	if err := policy.Validate(); err != nil {
		return nil, err
	}
	sort.Strings(policy.Required)
	return policy, nil
}

func (policy *Policy) Validate() error {
	if policy == nil {
		return ErrPolicyMissing
	}
	if policy.RecipeID == "" || policy.Revision < 1 {
		return errors.New("recipe id and positive revision are required")
	}
	if len(policy.Limits) == 0 {
		return errors.New("at least one process limit is required")
	}
	for name, limit := range policy.Limits {
		if name == "" {
			return errors.New("limit name is required")
		}
		if err := limit.Validate(); err != nil {
			return fmt.Errorf("limit %s: %w", name, err)
		}
	}
	for _, required := range policy.Required {
		if _, ok := policy.Limits[required]; !ok {
			return fmt.Errorf("required limit %s is absent", required)
		}
	}
	return nil
}

func (policy *Policy) Clone() *Policy {
	if policy == nil {
		return nil
	}
	clone := *policy
	return &clone
}

func (policy *Policy) Accepts(values map[string]float64) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	for _, name := range policy.Required {
		value, ok := values[name]
		if !ok {
			return fmt.Errorf("required process value %s is absent", name)
		}
		limit := policy.Limits[name]
		if value < limit.Minimum || value > limit.Maximum {
			return fmt.Errorf("%s value %.3f outside %.3f..%.3f", name, value, limit.Minimum, limit.Maximum)
		}
	}
	return nil
}
