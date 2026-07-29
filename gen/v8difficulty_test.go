package gen

import (
	"reflect"
	"strings"
	"testing"

	"github.com/ditto-assistant/dittobench-datagen/persona"
	"github.com/ditto-assistant/dittobench-datagen/protocol"
)

func TestV8ToolMixVariesAndCoversEveryFamily(t *testing.T) {
	prof, _ := ProfileForVersion("full", protocol.BenchVersionV8)
	var first map[string]int
	seen := map[string]bool{}
	varied := false
	for seed := int64(1); seed <= 40; seed++ {
		artifact, err := GenerateDataset(seed, prof, protocol.BenchVersionV8)
		if err != nil {
			t.Fatal(err)
		}
		hist := map[string]int{}
		for _, tc := range artifact.ToolCases {
			hist[tc.Category]++
			seen[tc.Category] = true
		}
		if first == nil {
			first = hist
		} else if !reflect.DeepEqual(first, hist) {
			varied = true
		}
	}
	if !varied {
		t.Fatal("v8 tool histogram was identical across 40 seeds")
	}
	if len(seen) != 54 {
		t.Fatalf("v8 full covered %d tool families across 40 seeds, want 54", len(seen))
	}
}

func TestV8StateRoutedCasesCarryPrivateSeedFacts(t *testing.T) {
	prof, _ := ProfileForVersion("full", protocol.BenchVersionV8)
	artifact, err := GenerateDataset(77, prof, protocol.BenchVersionV8)
	if err != nil {
		t.Fatal(err)
	}
	routed := 0
	for _, tc := range artifact.ToolCases {
		if tc.Category != "context_routed_action" {
			if len(tc.PrerequisitePairs) != 0 && tc.Category != "memory_fetch" {
				t.Fatalf("ordinary case %s carried prerequisite pairs", tc.ID)
			}
			continue
		}
		routed++
		if len(tc.ExpectedTools) != 1 || len(tc.PrerequisitePairs) != 1 {
			t.Fatalf("routed case shape: %+v", tc)
		}
		if strings.Contains(tc.Prompt, tc.ExpectedTools[0].Name) {
			t.Fatalf("visible prompt leaked route %q: %q", tc.ExpectedTools[0].Name, tc.Prompt)
		}
		joined := tc.Prompt + " " + tc.PrerequisitePairs[0].Prompt + " " + tc.PrerequisitePairs[0].Response
		if strings.Contains(joined, "exact words") || strings.Contains(joined, "formed from") || strings.Contains(joined, "operation ") {
			t.Fatalf("routed case uses benchmark language instead of a user request: %q", joined)
		}
	}
	if routed != 63 {
		t.Fatalf("full v8 routed cases = %d, want 63", routed)
	}

	v7, err := GenerateDataset(77, prof, protocol.BenchVersionV7)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range v7.ToolCases {
		if len(tc.PrerequisitePairs) != 0 {
			t.Fatalf("v7 case %s changed wire shape", tc.ID)
		}
	}
}

func TestV8FreeFormArgumentsDoNotRequireMagicStrings(t *testing.T) {
	prof, _ := ProfileForVersion("full", protocol.BenchVersionV8)
	freeForm := map[string]bool{
		"agent_job": true, "workflow_not_job": true, "agent_workflow": true,
		"feedback": true, "set_tool_prefs": true, "automation_not_job": true,
		"recipe_create": true, "recipe_apply": true, "calendar_create": true,
		"calendar_search": true,
	}
	seen := map[string]bool{}
	for seed := int64(1); seed <= 40; seed++ {
		artifact, err := GenerateDataset(seed, prof, protocol.BenchVersionV8)
		if err != nil {
			t.Fatal(err)
		}
		for _, tc := range artifact.ToolCases {
			if !freeForm[tc.Category] {
				continue
			}
			seen[tc.Category] = true
			for _, spec := range tc.ExpectedTools {
				if len(spec.RequiredArgs) != 0 {
					t.Fatalf("%s requires one magic free-form payload: %+v", tc.Category, spec.RequiredArgs)
				}
			}
		}
	}
	if len(seen) != len(freeForm) {
		t.Fatalf("did not exercise every free-form family: got %v", seen)
	}
}

func TestV8MemoryFetchUsesNaturalSearchThenFetch(t *testing.T) {
	prof, _ := ProfileForVersion("full", protocol.BenchVersionV8)
	for seed := int64(1); seed <= 80; seed++ {
		artifact, err := GenerateDataset(seed, prof, protocol.BenchVersionV8)
		if err != nil {
			t.Fatal(err)
		}
		for _, tc := range artifact.ToolCases {
			if tc.Category != "memory_fetch" {
				continue
			}
			if len(tc.ExpectedTools) != 2 || tc.ExpectedTools[0].Name != "search_memories" || tc.ExpectedTools[1].Name != "fetch_memories" {
				t.Fatalf("memory_fetch trajectory = %+v", tc.ExpectedTools)
			}
			if len(tc.PrerequisitePairs) != 1 || strings.Contains(strings.ToLower(tc.Prompt), "pair id") {
				t.Fatalf("memory_fetch is not a natural grounded request: %+v", tc)
			}
			return
		}
	}
	t.Fatal("no memory_fetch case appeared across seeds")
}

func TestV8ReasoningMemoryFloor(t *testing.T) {
	prof, _ := ProfileForVersion("full", protocol.BenchVersionV8)
	for seed := int64(1); seed <= 40; seed++ {
		artifact, err := GenerateDataset(seed, prof, protocol.BenchVersionV8)
		if err != nil {
			t.Fatal(err)
		}
		reasoning := 0
		hist := map[string]int{}
		for _, mc := range artifact.MemoryCases {
			hist[mc.QuestionType]++
			switch mc.QuestionType {
			case persona.QTComputed, persona.QTMultiSession, persona.QTAggregation, persona.QTTemporal,
				"temporal-arithmetic", "nonverbatim-computed":
				reasoning++
			}
			if mc.BenchVersion != protocol.BenchVersionV8 {
				t.Fatalf("v8 memory case omitted version: %+v", mc)
			}
		}
		if 10*reasoning < 3*len(artifact.MemoryCases) {
			t.Fatalf("seed %d reasoning share %d/%d is below 30%%: %v", seed, reasoning, len(artifact.MemoryCases), hist)
		}
	}
}
