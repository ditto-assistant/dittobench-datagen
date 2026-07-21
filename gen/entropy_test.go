package gen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ditto-assistant/dittobench-datagen/protocol"
)

// TestV5FamilySurfaceEntropy is the anti-overfit / anti-enumeration guard. The v5
// memory-case families must NOT be a small closed set of templates a miner can
// enumerate into a static cue list (the v4 phrase-list exploit, one level up). This
// pins a high distinct-surface count across seeds for each family, so shrinking the
// grammars/pools (regressing toward memorizable templates) fails the build.
//
// The declarative WRITE turn is the highest-risk family (it is the direct analog of
// the v4 durable-write cue list), so it carries the strictest floor.
func TestV5FamilySurfaceEntropy(t *testing.T) {
	// question_type -> set of distinct rendered surfaces across many seeds. For the
	// declarative WRITE (which has no question), collect the seeded /run prompt.
	surfaces := map[string]map[string]struct{}{}
	add := func(kind, s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		if surfaces[kind] == nil {
			surfaces[kind] = map[string]struct{}{}
		}
		surfaces[kind][s] = struct{}{}
	}
	for i := 0; i < 40; i++ {
		seed := int64(i*7919 + 11)
		r, err := NewRNGForVersion(seed, protocol.BenchVersionV5)
		if err != nil {
			t.Fatal(err)
		}
		s, err := GenerateMemorySuiteForVersion(r, seed, 90, 4, 0.2, protocol.BenchVersionV5)
		if err != nil {
			t.Fatal(err)
		}
		for _, sc := range s.Cases {
			add(sc.Case.QuestionType, sc.Case.Question)
		}
		// Declarative-write prompts are the /run write turns; collect their text.
		for _, sc := range s.Cases {
			if sc.Case.QuestionType == QTDeclarativeWrite {
				add(QTDeclarativeWrite, sc.Case.Question)
			}
		}
	}
	// Minimum distinct surfaces per family (across 40 seeds). Floors chosen well
	// above the small-template regime and below the current combinatorial space, so
	// they catch a regression without being brittle to phrasing tweaks.
	floors := map[string]int{
		QTMultiHop:            30,
		QTTemporalDepth:       25,
		QTDeclarativeWrite:    20,
		QTAbstainConfab:       12,
		QTDeclarativeAck:      8,
		QTChitchat:            8,
		QTDeclarativeBehavior: 8,
	}
	for kind, floor := range floors {
		got := len(surfaces[kind])
		if got < floor {
			t.Errorf("%s: only %d distinct surfaces across 40 seeds (want >= %d) — family is too enumerable (overfit risk)", kind, got, floor)
		}
	}
	if !t.Failed() {
		var parts []string
		for kind := range floors {
			parts = append(parts, fmt.Sprintf("%s=%d", kind, len(surfaces[kind])))
		}
		t.Logf("distinct v5 surfaces across 40 seeds: %s", strings.Join(parts, " "))
	}
}

// TestV6FamilySurfaceEntropy is the content-variety guard: v6 triples the six
// content pools, so every family must show a STRICTLY larger distinct-surface
// count than v5 and clear a raised floor. Shrinking the v6 pools back toward the
// v5 sizes (regressing the anti-enumeration property this release exists to add)
// fails the build.
func TestV6FamilySurfaceEntropy(t *testing.T) {
	count := func(version int) map[string]int {
		surfaces := map[string]map[string]struct{}{}
		add := func(kind, s string) {
			s = strings.TrimSpace(s)
			if s == "" {
				return
			}
			if surfaces[kind] == nil {
				surfaces[kind] = map[string]struct{}{}
			}
			surfaces[kind][s] = struct{}{}
		}
		for i := 0; i < 40; i++ {
			seed := int64(i*7919 + 11)
			r, err := NewRNGForVersion(seed, version)
			if err != nil {
				t.Fatal(err)
			}
			s, err := GenerateMemorySuiteForVersion(r, seed, 90, 4, 0.2, version)
			if err != nil {
				t.Fatal(err)
			}
			for _, sc := range s.Cases {
				add(sc.Case.QuestionType, sc.Case.Question)
				if sc.Case.QuestionType == QTDeclarativeWrite {
					add(QTDeclarativeWrite, sc.Case.Question)
				}
			}
		}
		out := map[string]int{}
		for k, set := range surfaces {
			out[k] = len(set)
		}
		return out
	}
	v5, v6 := count(protocol.BenchVersionV5), count(protocol.BenchVersionV6)
	// Raised floors: comfortably above v5's floors and below the v6 combinatorial
	// space, so a pool-size regression trips without brittleness to phrasing tweaks.
	floors := map[string]int{
		QTMultiHop:            60,
		QTTemporalDepth:       45,
		QTDeclarativeWrite:    30,
		QTAbstainConfab:       25,
		QTDeclarativeBehavior: 12,
	}
	// Families whose QUESTION text is pool-driven (relation/noun/domain, no coined
	// token in the prompt) must show STRICTLY more distinct surfaces under the
	// tripled v6 pools. temporal-depth and declarative-write are deliberately
	// excluded: their question embeds a per-case coined value, so every surface is
	// already unique and the count saturates at the case count in both versions —
	// their pool growth is proven by the multi-hop/abstention families and the
	// v6 known vector, not by a saturated distinct-surface count.
	strictExceed := map[string]bool{
		QTMultiHop:            true,
		QTAbstainConfab:       true,
		QTDeclarativeBehavior: true,
	}
	for kind, floor := range floors {
		if v6[kind] < floor {
			t.Errorf("%s: v6 has %d distinct surfaces across 40 seeds (want >= %d) — content pool regressed", kind, v6[kind], floor)
		}
		if v6[kind] < v5[kind] {
			t.Errorf("%s: v6 (%d) regressed below v5 (%d) distinct surfaces", kind, v6[kind], v5[kind])
		}
		if strictExceed[kind] && v6[kind] <= v5[kind] {
			t.Errorf("%s: v6 (%d) must exceed v5 (%d) distinct surfaces — the tripled pool is not reaching this family", kind, v6[kind], v5[kind])
		}
	}
	if !t.Failed() {
		var parts []string
		for kind := range floors {
			parts = append(parts, fmt.Sprintf("%s v5=%d v6=%d", kind, v5[kind], v6[kind]))
		}
		t.Logf("v6 vs v5 distinct surfaces across 40 seeds: %s", strings.Join(parts, " | "))
	}
}
