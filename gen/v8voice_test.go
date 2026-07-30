package gen

import (
	"strings"
	"testing"

	"github.com/ditto-assistant/dittobench-datagen/internal/assistantvoice"
	"github.com/ditto-assistant/dittobench-datagen/protocol"
)

func TestV8TranscriptUsesWarmVariedAssistantVoice(t *testing.T) {
	prof, _ := ProfileForVersion("full", protocol.BenchVersionV8)
	artifact, err := GenerateDataset(3473949159349387300, prof, protocol.BenchVersionV8)
	if err != nil {
		t.Fatal(err)
	}

	counts := map[string]int{}
	total := 0
	check := func(source string, pair protocol.MemoryPair) {
		t.Helper()
		response := strings.TrimSpace(pair.Response)
		if assistantvoice.IsCold(response) {
			t.Fatalf("%s pair %s has cold assistant response %q", source, pair.PairID, response)
		}
		if strings.Contains(strings.ToLower(response), "subscribed graph") {
			t.Fatalf("%s pair %s leaks implementation language in response %q", source, pair.PairID, response)
		}
		counts[response]++
		total++
	}
	for _, toolCase := range artifact.ToolCases {
		for _, pair := range toolCase.PrerequisitePairs {
			check("tool prerequisite", pair)
		}
	}
	for _, wave := range artifact.MemoryWaves {
		for _, pair := range wave.Pairs {
			check("memory wave", pair)
		}
	}
	if total == 0 {
		t.Fatal("v8 artifact contains no assistant transcript rows")
	}
	if len(counts)*3 < total {
		t.Fatalf("assistant response diversity=%d/%d, want at least one unique response per three rows", len(counts), total)
	}
	for response, count := range counts {
		if count > 16 {
			t.Fatalf("assistant response repeated %d times, want at most 16: %q", count, response)
		}
	}
}
