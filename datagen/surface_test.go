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
	// Grammar-driven categories must clear a much higher floor than the plain
	// wrapped ones: the CFG expander exists to lift them from a memorizable
	// template list to hundreds of surfaces (audit follow-up item A).
	grammarDriven := map[string]bool{}
	for _, c := range categories {
		if c.grammar != nil {
			grammarDriven[c.name] = true
		}
	}
	if len(grammarDriven) < 5 {
		t.Fatalf("only %d grammar-driven categories — the five audited categories must use grammars", len(grammarDriven))
	}
	for cat, seen := range prompts {
		floor := 12
		if grammarDriven[cat] {
			floor = 60
		}
		if len(seen) < floor {
			t.Errorf("category %s: only %d distinct prompts across 80 seeds (floor %d) — template-memorizable", cat, len(seen), floor)
		}
	}
	for cat := range wrapped {
		if prompts[cat] == nil {
			t.Errorf("category %s never emitted in 80 seeds", cat)
		}
	}
}

// TestGrammarPromptsWellFormed sweeps every grammar-driven category's emitted
// prompts for leaked grammar syntax and unfilled placeholders.
func TestGrammarPromptsWellFormed(t *testing.T) {
	grammarDriven := map[string]bool{}
	for _, c := range categories {
		if c.grammar != nil {
			grammarDriven[c.name] = true
		}
	}
	for seed := int64(1); seed <= 60; seed++ {
		for _, c := range Generate(seed, 60).ToolCases {
			if !grammarDriven[c.Category] {
				continue
			}
			if strings.Contains(c.Prompt, "#") {
				t.Fatalf("seed %d %s: prompt leaked grammar syntax: %q", seed, c.Category, c.Prompt)
			}
			if strings.Contains(c.Prompt, "%s") {
				t.Fatalf("seed %d %s: prompt has unfilled placeholder: %q", seed, c.Category, c.Prompt)
			}
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
			for _, kw := range []string{"fetch", "memory", "memories", "pair", "conversation"} {
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

// TestAutomationListPromptNotRoutable pins the leak fix for the automation_list
// grammar: the prompt must not state list_automations' own name/description
// content words ("list", "scheduled", "automations"). The no-model reference
// router (refharness) signals on those tokens, so reusing any of them lets it
// solve the read-vs-create trap with no retrieval (a grammar production once
// stacked all three and drove the benchcal floor to 1.0 on a seed).
func TestAutomationListPromptNotRoutable(t *testing.T) {
	for seed := int64(1); seed <= 40; seed++ {
		for _, c := range Generate(seed, 60).ToolCases {
			if c.Category != "automation_list" {
				continue
			}
			p := strings.ToLower(c.Prompt)
			for _, kw := range []string{"list", "scheduled", "automations"} {
				if strings.Contains(p, kw) {
					t.Fatalf("seed %d: automation_list prompt states router keyword %q: %q", seed, kw, c.Prompt)
				}
			}
		}
	}
}
