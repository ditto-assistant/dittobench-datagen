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
		if hist["set_effort"] == 0 {
			t.Fatalf("seed %d omitted set_effort", seed)
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

func TestV8EveryProfileCarriesSetEffort(t *testing.T) {
	for _, runSize := range []string{"small", "medium", "full"} {
		prof, _ := ProfileForVersion(runSize, protocol.BenchVersionV8)
		for seed := int64(1); seed <= 40; seed++ {
			artifact, err := GenerateDataset(seed, prof, protocol.BenchVersionV8)
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, tc := range artifact.ToolCases {
				found = found || tc.Category == "set_effort"
			}
			if !found {
				t.Fatalf("%s seed %d omitted set_effort", runSize, seed)
			}
		}
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
		if tc.Category != "state_routed_action" {
			if len(tc.PrerequisitePairs) != 0 {
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
		if !strings.Contains(tc.PrerequisitePairs[0].Response, tc.ExpectedTools[0].Name) {
			t.Fatalf("private policy omitted route %q", tc.ExpectedTools[0].Name)
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

func TestV8RequiredArgsAreNotPromptSpans(t *testing.T) {
	prof, _ := ProfileForVersion("full", protocol.BenchVersionV8)
	var total, verbatim int
	for seed := int64(1); seed <= 40; seed++ {
		artifact, err := GenerateDataset(seed, prof, protocol.BenchVersionV8)
		if err != nil {
			t.Fatal(err)
		}
		for _, tc := range artifact.ToolCases {
			for _, spec := range tc.ExpectedTools {
				for _, value := range spec.RequiredArgs {
					total++
					if strings.Contains(strings.ToLower(tc.Prompt), strings.ToLower(value)) {
						verbatim++
					}
				}
			}
		}
	}
	if total == 0 {
		t.Fatal("v8 generated no required arguments")
	}
	if 4*verbatim >= total {
		t.Fatalf("v8 prompt exposes %d/%d required arguments verbatim, want <25%%", verbatim, total)
	}
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
