package recipe

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

type RawPolicy struct {
	RecipeID string              `json:"recipe_id"`
	Revision int                 `json:"revision"`
	Limits   map[string]RawRange `json:"limits"`
	Required []string            `json:"required"`
}

type RawRange struct {
	Minimum *float64 `json:"minimum"`
	Maximum *float64 `json:"maximum"`
}

func Decode(reader io.Reader) (*Policy, error) {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	var raw RawPolicy
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode recipe policy: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, errors.New("recipe policy contains trailing document")
		}
		return nil, fmt.Errorf("decode recipe policy trailer: %w", err)
	}
	return raw.Build()
}

func (raw RawPolicy) Build() (*Policy, error) {
	limits := make(Limits, len(raw.Limits))
	for name, rawLimit := range raw.Limits {
		if rawLimit.Minimum == nil || rawLimit.Maximum == nil {
			return nil, fmt.Errorf("limit %s needs minimum and maximum", name)
		}
		limits[name] = Range{Minimum: *rawLimit.Minimum, Maximum: *rawLimit.Maximum}
	}
	policy, err := NewPolicy(raw.RecipeID, raw.Revision, limits, raw.Required)
	if err != nil {
		return nil, fmt.Errorf("build recipe policy: %w", err)
	}
	return policy, nil
}

type Provider interface {
	Current(recipeID string) (*Policy, error)
}

type StaticProvider struct {
	policies map[string]*Policy
}

func NewStaticProvider(policies ...*Policy) *StaticProvider {
	provider := &StaticProvider{policies: make(map[string]*Policy)}
	for _, policy := range policies {
		if policy != nil {
			provider.policies[policy.RecipeID] = policy
		}
	}
	return provider
}

func (provider *StaticProvider) Current(recipeID string) (*Policy, error) {
	policy, ok := provider.policies[recipeID]
	if !ok {
		return nil, fmt.Errorf("recipe %s: %w", recipeID, ErrPolicyMissing)
	}
	return policy, nil
}

func ValidateValues(provider Provider, recipeID string, values map[string]float64) error {
	policy, err := provider.Current(recipeID)
	if err != nil {
		return err
	}
	return policy.Accepts(values)
}
