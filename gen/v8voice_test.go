package gen

import (
	"strings"
	"testing"

	"github.com/ditto-assistant/dittobench-datagen/internal/assistantvoice"
	"github.com/ditto-assistant/dittobench-datagen/protocol"
	"github.com/ditto-assistant/dittobench-datagen/universe"
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
		lowerPrompt := strings.ToLower(pair.Prompt)
		for _, banned := range []string{"for reference", "nothing urgent", "not something about you", "worth a mention"} {
			if strings.Contains(lowerPrompt, banned) || strings.Contains(strings.ToLower(response), banned) {
				t.Fatalf("%s pair %s contains benchmark-like phrase %q: %q / %q", source, pair.PairID, banned, pair.Prompt, response)
			}
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
	for _, memoryCase := range artifact.MemoryCases {
		if memoryCase.QuestionType == QTDeclarativeAck && strings.Contains(strings.ToLower(memoryCase.Question), "nickname") {
			t.Fatalf("token-shaped integrity value is mislabeled as a nickname: %q", memoryCase.Question)
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

func TestV8TranscriptAddressesEachSeededUserByFirstNameAtHumanCadence(t *testing.T) {
	const seed = int64(3473949159349387300)
	prof, _ := ProfileForVersion("full", protocol.BenchVersionV8)
	artifact, err := GenerateDataset(seed, prof, protocol.BenchVersionV8)
	if err != nil {
		t.Fatal(err)
	}
	firstName := func(full string) string { return strings.Fields(full)[0] }
	names := map[string]string{
		PrimaryUser:   firstName(universe.UserName(seed)),
		SecondaryUser: firstName(universe.UserName(seed ^ isolationSalt)),
	}

	eligible, addressed := 0, 0
	for _, toolCase := range artifact.ToolCases {
		for _, pair := range toolCase.PrerequisitePairs {
			eligible++
			if responseAddressesName(pair.Response, names[PrimaryUser]) {
				addressed++
			}
		}
	}
	for _, wave := range artifact.MemoryWaves {
		name := names[wave.UserID]
		if name == "" {
			continue
		}
		for _, pair := range wave.Pairs {
			eligible++
			if responseAddressesName(pair.Response, name) {
				addressed++
			}
		}
	}
	if addressed < eligible*18/100 || addressed > eligible*30/100 {
		t.Fatalf("seeded-name cadence=%d/%d, want 18-30%%", addressed, eligible)
	}
	if addressed == eligible {
		t.Fatal("every V8 reply addresses the user by name")
	}
}

func responseAddressesName(response, name string) bool {
	if strings.HasPrefix(response, name+", ") || strings.Contains(response, ", "+name+" —") {
		return true
	}
	for _, ending := range []string{".", "!", "?"} {
		if strings.HasSuffix(response, ", "+name+ending) {
			return true
		}
	}
	return false
}
