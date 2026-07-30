package gen

import (
	"strings"
	"testing"

	"github.com/ditto-assistant/dittobench-datagen/persona"
	"github.com/ditto-assistant/dittobench-datagen/protocol"
)

func TestGenerateIsolationStructure(t *testing.T) {
	const seed = 778899
	iso := GenerateIsolation(seed, 20, 2, 4)

	if len(iso.Cases) == 0 {
		t.Fatal("expected isolation cases")
	}
	if len(iso.SecondaryWave.Pairs) == 0 {
		t.Fatal("secondary graph should have seeded pairs")
	}
	if iso.SecondaryWave.UserID != SecondaryUser {
		t.Fatalf("secondary graph must be seeded under %q, got %q", SecondaryUser, iso.SecondaryWave.UserID)
	}

	sawA, sawB := false, false
	for _, sc := range iso.Cases {
		// The suite also carries the cross-user LIFECYCLE probe (B3), whose cases
		// are lifecycle-typed by design: they mutate under A and read under B.
		// They are covered by gen/crossuser_test.go.
		if xuNoun(sc.Case.Question) != "" {
			continue
		}
		if sc.Case.QuestionType != "isolation" {
			t.Fatalf("case %s: want type isolation, got %q", sc.Case.ID, sc.Case.QuestionType)
		}
		if sc.Case.ExpectedAnswer == "" {
			t.Fatalf("case %s: empty expected answer", sc.Case.ID)
		}
		switch sc.UserID {
		case PrimaryUser:
			sawA = true
		case SecondaryUser:
			sawB = true
		default:
			t.Fatalf("case %s: unexpected user_id %q", sc.Case.ID, sc.UserID)
		}
	}
	if !sawA || !sawB {
		t.Fatalf("expected both A-scoped and B-scoped isolation cases (A=%v B=%v)", sawA, sawB)
	}
}

// The crux of isolation: for each case, the OTHER user's graph must hold a
// DIFFERENT value for the queried attribute; otherwise there is no leak to catch.
func TestIsolationCasesHaveConflictingValue(t *testing.T) {
	const seed = 4242
	iso := GenerateIsolation(seed, 20, 2, 4)

	pCur := currentScalars(persona.BuildPlan(seed, personaOptsFor(20)))
	sCur := currentScalars(persona.BuildPlan(seed^isolationSalt, isolationOpts()))

	for _, sc := range iso.Cases {
		// Cross-user lifecycle cases (B3) key off a coined token, not a persona
		// attribute, so the conflict-axis check below does not apply to them.
		if xuNoun(sc.Case.Question) != "" {
			continue
		}
		// The wire ID is opaque (no attribute); the validator-internal QuestionID
		// still names the conflict axis ("iso-a-<attr>").
		attr := attrFromISOQID(sc.Case.QuestionID)
		if attr == "" {
			t.Fatalf("could not parse attr from question id %s", sc.Case.QuestionID)
		}
		pv, sv := pCur[attr], sCur[attr]
		if pv == "" || sv == "" || pv == sv {
			t.Fatalf("case %s attr %q: both graphs must hold a distinct value (A=%q B=%q)", sc.Case.ID, attr, pv, sv)
		}
		// The expected answer is the QUERIED user's value, never the other's.
		want := pv
		if sc.UserID == SecondaryUser {
			want = sv
		}
		if sc.Case.ExpectedAnswer != want {
			t.Fatalf("case %s (user %s): expected %q (its own graph), got %q", sc.Case.ID, sc.UserID, want, sc.Case.ExpectedAnswer)
		}
	}
}

func TestGenerateIsolationDeterministic(t *testing.T) {
	a := GenerateIsolation(31337, 20, 2, 4)
	b := GenerateIsolation(31337, 20, 2, 4)
	if len(a.Cases) != len(b.Cases) || len(a.SecondaryWave.Pairs) != len(b.SecondaryWave.Pairs) {
		t.Fatal("isolation generation not deterministic (counts differ)")
	}
	for i := range a.Cases {
		if a.Cases[i].Case.ID != b.Cases[i].Case.ID ||
			a.Cases[i].Case.ExpectedAnswer != b.Cases[i].Case.ExpectedAnswer ||
			a.Cases[i].UserID != b.Cases[i].UserID {
			t.Fatalf("case %d differs across identical (seed, rng)", i)
		}
	}
}

func TestGenerateIsolationDisabled(t *testing.T) {
	iso := GenerateIsolation(5, 20, 2, 0)
	if len(iso.Cases) != 0 || len(iso.SecondaryWave.Pairs) != 0 {
		t.Fatal("isoCases=0 should produce no isolation content")
	}
}

func TestV8IsolationUsesWorldMemoriesAndExactGraphConflicts(t *testing.T) {
	const seed int64 = 123456789
	iso, err := GenerateIsolationForVersion(seed, 225, 5, 9, protocol.BenchVersionV8)
	if err != nil {
		t.Fatal(err)
	}
	if len(iso.Cases) != 9 || len(iso.ReviewPlans) != 9 || len(iso.SecondaryWave.Pairs) != 27 {
		t.Fatalf("v8 isolation cases/plans/pairs=%d/%d/%d, want 9/9/27", len(iso.Cases), len(iso.ReviewPlans), len(iso.SecondaryWave.Pairs))
	}
	sawPrimary, sawSecondary := false, false
	pairs := map[string]bool{}
	for _, pair := range iso.SecondaryWave.Pairs {
		if strings.HasPrefix(pair.SessionID, "sess-") {
			t.Fatalf("v8 retained legacy isolation session %q", pair.SessionID)
		}
		pairs[pair.PairID] = true
	}
	for i, staged := range iso.Cases {
		if staged.Case.QuestionType != "world-isolation-contact-current" {
			t.Fatalf("case %d type=%q", i, staged.Case.QuestionType)
		}
		if staged.Case.ExpectedAnswer == "" || staged.Case.ForbiddenAnswer == "" || staged.Case.ExpectedAnswer == staged.Case.ForbiddenAnswer {
			t.Fatalf("case %d lacks an exact cross-graph conflict: %+v", i, staged.Case)
		}
		switch staged.UserID {
		case PrimaryUser:
			sawPrimary = true
		case SecondaryUser:
			sawSecondary = true
		default:
			t.Fatalf("case %d has unexpected user %q", i, staged.UserID)
		}
		if len(iso.ReviewPlans[i].RequiredPairIDs) != 3 {
			t.Fatalf("case %d evidence=%v", i, iso.ReviewPlans[i].RequiredPairIDs)
		}
		if staged.UserID == SecondaryUser {
			for _, pairID := range iso.ReviewPlans[i].RequiredPairIDs {
				if !pairs[pairID] {
					t.Fatalf("secondary case %d evidence %s is not seeded", i, pairID)
				}
			}
		}
	}
	if !sawPrimary || !sawSecondary {
		t.Fatalf("v8 isolation directions primary=%v secondary=%v", sawPrimary, sawSecondary)
	}
}

// attrFromISOQID pulls the attribute suffix out of an "iso-a-<attr>" QuestionID.
func attrFromISOQID(qid string) string {
	parts := strings.SplitN(qid, "-", 3)
	if len(parts) < 3 {
		return ""
	}
	return parts[2]
}
