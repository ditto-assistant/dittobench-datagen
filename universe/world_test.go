package universe

import (
	"reflect"
	"strings"
	"testing"
)

func TestWorldIsDeterministicAndSeedVarying(t *testing.T) {
	a := Generate(918273, 3)
	b := Generate(918273, 3)
	c := Generate(918274, 3)
	if !reflect.DeepEqual(a, b) {
		t.Fatal("same seed and scale did not reproduce the same world")
	}
	if reflect.DeepEqual(a, c) {
		t.Fatal("adjacent seeds generated an identical world")
	}
}

func TestWorldIdentitiesAreUnambiguous(t *testing.T) {
	for seed := int64(1); seed <= 100; seed++ {
		w := Generate(seed, 3)
		assertUnique(t, seed, "pair id", w.SortedPairIDs())

		var names, nicknames, emails, projectNames, projectAliases, tripAliases []string
		for _, p := range w.People {
			names = append(names, p.Name)
			nicknames = append(nicknames, p.Nickname)
			emails = append(emails, p.Email, p.PreviousEmail)
			if p.Email == p.PreviousEmail {
				t.Fatalf("seed %d person %q has no effective email correction", seed, p.Name)
			}
		}
		for _, p := range w.Projects {
			projectNames = append(projectNames, p.Name)
			projectAliases = append(projectAliases, p.Alias)
			if p.OriginalCents == p.CorrectedCents {
				t.Fatalf("seed %d project %q has a no-op invoice correction", seed, p.Alias)
			}
			if p.OutstandingCents != p.CorrectedCents-p.PaidCents {
				t.Fatalf("seed %d project %q has inconsistent ledger math", seed, p.Alias)
			}
		}
		for _, trip := range w.Trips {
			tripAliases = append(tripAliases, trip.Alias)
			if trip.PreviousDays != sum3(trip.OldLegDays) || trip.CurrentDays != sum3(trip.LegDays) {
				t.Fatalf("seed %d trip %q has inconsistent leg totals", seed, trip.Alias)
			}
			if trip.PreviousDays == trip.CurrentDays {
				t.Fatalf("seed %d trip %q has a no-op itinerary correction", seed, trip.Alias)
			}
		}
		assertUnique(t, seed, "person name", names)
		assertUnique(t, seed, "nickname", nicknames)
		assertUnique(t, seed, "email", emails)
		assertUnique(t, seed, "project name", projectNames)
		assertUnique(t, seed, "project alias", projectAliases)
		assertUnique(t, seed, "trip alias", tripAliases)
	}
}

func TestWorldRepresentsShortNotesAndMessyBusinessPastes(t *testing.T) {
	w := Generate(42, 3)
	minLen, maxLen := int(^uint(0)>>1), 0
	for _, pair := range w.Pairs {
		n := len(strings.TrimSpace(pair.Prompt))
		if n < minLen {
			minLen = n
		}
		if n > maxLen {
			maxLen = n
		}
	}
	if minLen > 100 {
		t.Fatalf("shortest user memory is %d bytes; expected terse real-user notes", minLen)
	}
	if maxLen < 1200 || maxLen < 8*minLen {
		t.Fatalf("message lengths do not include a realistic wall-of-text tail: min=%d max=%d", minLen, maxLen)
	}
}

func TestWorldQuestionsHaveThreeNearMissesAndDoNotLeakAnswers(t *testing.T) {
	for _, scale := range []int{1, 2, 3} {
		w := Generate(991+int64(scale), scale)
		for _, c := range w.MemoryCases(24) {
			if len(c.DistractorAnswers) != 3 {
				t.Fatalf("scale %d case %s has %d distractors", scale, c.ID, len(c.DistractorAnswers))
			}
			assertUnique(t, w.Seed, "distractor", c.DistractorAnswers)
			for _, distractor := range c.DistractorAnswers {
				if distractor == c.ExpectedAnswer {
					t.Fatalf("case %s includes its answer as a distractor", c.ID)
				}
			}
			if strings.Contains(strings.ToLower(c.Question), strings.ToLower(c.ExpectedAnswer)) {
				t.Fatalf("case %s leaks its answer in the question", c.ID)
			}
		}
	}
}

func assertUnique(t *testing.T, seed int64, label string, values []string) {
	t.Helper()
	seen := map[string]bool{}
	for _, value := range values {
		if seen[value] {
			t.Fatalf("seed %d generated duplicate %s %q", seed, label, value)
		}
		seen[value] = true
	}
}
