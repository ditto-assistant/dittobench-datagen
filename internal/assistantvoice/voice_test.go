package assistantvoice

import (
	"strings"
	"testing"
)

func TestRenderReplacesColdAcknowledgementsDeterministically(t *testing.T) {
	got := Render(42, "pair-a", "trip-1", "The trip changed yesterday.", "Noted.")
	again := Render(42, "pair-a", "trip-1", "The trip changed yesterday.", "Noted.")
	if got != again {
		t.Fatalf("same identity rendered different responses: %q != %q", got, again)
	}
	if IsCold(got) {
		t.Fatalf("rendered response remains cold: %q", got)
	}
	if !strings.Contains(got, "state") && !strings.Contains(got, "changed") && !strings.Contains(got, "versions") && !strings.Contains(got, "update") && !strings.Contains(got, "before and after") {
		t.Fatalf("update acknowledgement lost its context: %q", got)
	}
}

func TestRenderPreservesSubstantiveFactsWhileWarmingPrefix(t *testing.T) {
	got := Render(42, "pair-b", "person", "My partner is Anika.", "Noted your partner, Anika.")
	if IsCold(got) {
		t.Fatalf("rendered response remains cold: %q", got)
	}
	if !strings.Contains(got, "partner, Anika.") {
		t.Fatalf("rendered response lost its factual suffix: %q", got)
	}

	direct := "Morgan Lee's office number is +1-212-555-0192."
	if got := Render(42, "pair-c", "accountant", "I need the number.", direct); got != direct {
		t.Fatalf("substantive direct answer changed: %q", got)
	}

	got = Render(42, "pair-d", "security", "My spare key code is 2194.", "Got it. Your spare key code is 2194.")
	if IsCold(got) || !strings.Contains(got, "Your spare key code is 2194.") {
		t.Fatalf("period-prefixed fact was not preserved and warmed: %q", got)
	}
}

func TestRenderVariesGenericAcknowledgementsByPair(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 80; i++ {
		pairID := "pair-" + strings.Repeat("x", i) + string(rune('a'+i%26))
		seen[Render(7, pairID, "general", "Here is another detail.", "Got it.")] = true
	}
	if len(seen) < 24 {
		t.Fatalf("generic acknowledgement diversity=%d, want at least 24", len(seen))
	}
}

func TestIsColdRejectsTransactionalSurfaces(t *testing.T) {
	for _, response := range []string{"Noted.", "Noted. Your safe code is 2194.", "Good to know.", "Got it. Your spare key code is 2194.", "Got it — size 10.", "Understood, that city belongs to Hana.", "Updated."} {
		if !IsCold(response) {
			t.Fatalf("cold response accepted: %q", response)
		}
	}
	for _, response := range []string{"Got you — I’ll remember that.", "Lovely — say hi to Luca.", "Morgan Lee's office number is 555-0192."} {
		if IsCold(response) {
			t.Fatalf("human response rejected: %q", response)
		}
	}
}
