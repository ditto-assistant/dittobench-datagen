package datagen

import (
	"strings"
	"testing"
)

// TestWrappedCategoriesHaveWideSurface pins the P1 surface-variety fix: the
// audited low-variety categories must present many distinct prompts across
// seeds (templates x lead-ins x trailers), not a memorizable handful.
func TestWrappedCategoriesHaveWideSurface(t *testing.T) {
	wrapped := map[string]bool{}
	for _, c := range categories {
		if c.wrap {
			wrapped[c.name] = true
		}
	}
	if len(wrapped) == 0 {
		t.Fatal("no wrap-flagged categories — the audited low-variety categories must be wrapped")
	}
	prompts := map[string]map[string]bool{}
	for seed := int64(1); seed <= 80; seed++ {
		for _, c := range Generate(seed, 60).ToolCases {
			if !wrapped[c.Category] {
				continue
			}
			if prompts[c.Category] == nil {
				prompts[c.Category] = map[string]bool{}
			}
			prompts[c.Category][c.Prompt] = true
		}
	}
	for cat, seen := range prompts {
		if len(seen) < 12 {
			t.Errorf("category %s: only %d distinct prompts across 80 seeds — template-memorizable", cat, len(seen))
		}
	}
	for cat := range wrapped {
		if prompts[cat] == nil {
			t.Errorf("category %s never emitted in 80 seeds", cat)
		}
	}
}

// TestMemoryFetchPromptNotVerbatim pins the P2 hardening: the memory_fetch
// prompt must not state the tool call's own keywords verbatim (the no-model
// keyword router solved the old surface at floor 1.0).
func TestMemoryFetchPromptNotVerbatim(t *testing.T) {
	for seed := int64(1); seed <= 40; seed++ {
		for _, c := range Generate(seed, 60).ToolCases {
			if c.Category != "memory_fetch" {
				continue
			}
			p := strings.ToLower(c.Prompt)
			for _, kw := range []string{"fetch", "memory", "memories", "pair"} {
				if strings.Contains(p, kw) {
					t.Fatalf("seed %d: memory_fetch prompt states tool keyword %q verbatim: %q", seed, kw, c.Prompt)
				}
			}
			// The pinned pair-ID argument must still be present in the prompt.
			if !strings.Contains(p, "mem-") {
				t.Fatalf("seed %d: memory_fetch prompt lost its pair ID: %q", seed, c.Prompt)
			}
		}
	}
}
