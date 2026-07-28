package persona

import (
	"strings"
	"testing"

	"github.com/ditto-assistant/dittobench-datagen/grade"
	"github.com/ditto-assistant/dittobench-datagen/protocol"
)

func TestV8AmbiguousAnswerPoolsAreQualified(t *testing.T) {
	for attr, legacy := range map[string][]string{
		"favorite_color":   colors,
		"primary_language": softwareLanguages,
	} {
		pool := answerPoolForVersion(attr, legacy, protocol.BenchVersionV8)
		for _, value := range pool {
			if !strings.Contains(value, " ") && grade.Hit(value, "Let me "+strings.ToLower(value)+" check that.") {
				t.Fatalf("v8 %s value %q is creditable as incidental prose", attr, value)
			}
		}
	}
	if got := answerPoolForVersion("primary_language", softwareLanguages, protocol.BenchVersionV7); &got[0] != &softwareLanguages[0] {
		t.Fatal("v7 software-language pool changed")
	}
}
