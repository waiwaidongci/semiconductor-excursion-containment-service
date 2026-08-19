package recipe

import (
	"errors"
	"testing"
)

func TestRecipePolicyNilAndAliasContracts(t *testing.T) {
	minimum, maximum := 10.0, 20.0
	raw := RawPolicy{RecipeID: "RCP-9", Revision: 3, Limits: map[string]RawRange{"pressure": {Minimum: &minimum, Maximum: &maximum}}, Required: []string{"pressure"}}
	policy, err := raw.Build()
	if err != nil {
		t.Fatal(err)
	}
	clone := policy.Clone()
	clone.Limits["pressure"] = Range{Minimum: 0, Maximum: 1}
	clone.Required[0] = "changed"
	if policy.Limits["pressure"].Minimum != 10 || policy.Required[0] != "pressure" {
		t.Fatalf("policy clone aliases source: %#v", policy)
	}
	provider := NewStaticProvider(policy)
	loaded, err := provider.Current(policy.RecipeID)
	if err != nil {
		t.Fatal(err)
	}
	loaded.Limits["pressure"] = Range{Minimum: -1, Maximum: -1}
	again, err := provider.Current(policy.RecipeID)
	if err != nil || again.Limits["pressure"].Minimum != 10 {
		t.Fatalf("provider leaked mutable policy: %#v %v", again, err)
	}
	var nilProvider *StaticProvider
	if err := ValidateValues(nilProvider, "RCP-9", map[string]float64{"pressure": 15}); !errors.Is(err, ErrPolicyMissing) {
		t.Fatalf("typed nil provider was not rejected: %v", err)
	}
}
