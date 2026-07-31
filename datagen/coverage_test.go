package datagen

import (
	"math/rand"
	"testing"

	"github.com/ditto-assistant/dittobench-datagen/catalog"
	"github.com/ditto-assistant/dittobench-datagen/protocol"
)

// TestFullCatalogCoverage checks the property: every tool in the catalog is
// the correct answer (an expected tool) for at least one generated case. A large
// n makes the stratified draw hit every category.
func TestFullCatalogCoverage(t *testing.T) {
	// Generate at the latest contract (v5) so the coverage property holds over the
	// full catalog including the Code Mode tools (run_code / search_tools), which
	// are only reachable under v5's category set. The catalog is advertised to the
	// harness for every version, so every tool must be a correct answer somewhere.
	rotated, err := protocol.RotateSeedForVersion(2026, protocol.BenchVersionV5)
	if err != nil {
		t.Fatal(err)
	}
	r := rand.New(rand.NewSource(rotated))
	toolCases, _ := GenerateCasesWithFillersForVersion(r, 2026, 200, protocol.BenchVersionV5)
	reachable := map[string]bool{}
	for _, c := range toolCases {
		for _, ts := range c.ExpectedTools {
			reachable[ts.Name] = true
		}
	}
	var missing []string
	for _, tool := range catalog.Catalog() {
		if !reachable[tool.Name] {
			missing = append(missing, tool.Name)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("catalog tools never a correct answer: %v", missing)
	}
}

func TestV8CatalogUsesCurrentWorkflowSurface(t *testing.T) {
	reachable := map[string]bool{}
	for seed := int64(1); seed <= 40; seed++ {
		rotated, err := protocol.RotateSeedForVersion(seed, protocol.BenchVersionV8)
		if err != nil {
			t.Fatal(err)
		}
		r := rand.New(rand.NewSource(rotated))
		toolCases, _ := GenerateCasesWithFillersForVersion(r, seed, 84, protocol.BenchVersionV8)
		for _, c := range toolCases {
			for _, ts := range c.ExpectedTools {
				reachable[ts.Name] = true
			}
		}
	}
	for _, tool := range catalog.CatalogForVersion(protocol.BenchVersionV8) {
		if !reachable[tool.Name] {
			t.Errorf("v8 catalog tool %q is never a correct answer", tool.Name)
		}
	}
	for _, retired := range []string{"execute_agent_workflow", "create_automation", "list_automations", "create_recipe", "apply_recipe"} {
		if reachable[retired] {
			t.Errorf("v8 still generates retired tool %q", retired)
		}
		for _, tool := range catalog.CatalogForVersion(protocol.BenchVersionV8) {
			if tool.Name == retired {
				t.Errorf("v8 still advertises retired tool %q", retired)
			}
		}
	}
}
