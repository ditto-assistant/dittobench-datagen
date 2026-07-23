package persona

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/ditto-assistant/dittobench-datagen/grade"
	"github.com/ditto-assistant/dittobench-datagen/protocol"
)

// v8Opts mirrors the scored v8 full profile's persona sizing (gen's
// personaOptsForVersion). Kept local so this package's tests stay independent of
// gen, and small enough to run over many seeds.
func v8Opts() Opts {
	return Opts{Sessions: 24, Projects: 10, Trips: 8, Pets: 4, UpdateChains: 5, Reversals: 3, DecoyPeople: 12, DomainItems: 4, LongChain: 5, FillerBeats: 240}
}

// fillerVocabulary is every literal string the background-thread layer can put
// into the haystack, with template slots removed. Auditing the vocabulary rather
// than the assembled output is the stronger check: it holds for every seed and
// every assembly order at once, not just the ones a test happens to sample.
func fillerVocabulary() []string {
	var out []string
	for _, s := range fillerSubjects {
		out = append(out, s.label, s.coref, s.actor, s.thing, s.action)
		out = append(out, s.notes...)
	}
	strip := strings.NewReplacer("{label}", " ", "{coref}", " ", "{actor}", " ",
		"{thing}", " ", "{action}", " ", "{note}", " ")
	for stage := 0; stage < fillerStages; stage++ {
		for _, t := range fillerUserTmpls[stage] {
			out = append(out, strip.Replace(t))
		}
		for _, t := range fillerAsstTmpls[stage] {
			out = append(out, strip.Replace(t))
		}
	}
	out = append(out, elaborationClauses...)
	out = append(out, elaborationAcks...)
	return out
}

// answerPools is every pool a canonical answer value can be drawn from, across
// the universal scalars, all four professional domains, the sometimes-present
// attributes, the list and preference families, and the assistant-side
// recommendations. Checking the filler vocabulary against the POOLS rather than
// against a sample of generated plans makes the audit exhaustive: it covers
// every value the generator could ever emit, not the subset some seeds happened
// to draw.
func answerPools() map[string][]string {
	pools := map[string][]string{}
	add := func(name string, vals []string) { pools[name] = append(pools[name], vals...) }
	for _, s := range allScalarSpecs() {
		add(s.attr, s.pool)
	}
	for _, s := range prefSpecs {
		add(s.attr, s.pool)
	}
	for _, s := range []listSpec{projectSpec, tripSpec, petSpec} {
		add(s.attr, s.pool)
	}
	for _, d := range domains {
		for _, lc := range d.lists {
			add(lc.spec.attr, lc.spec.pool)
		}
	}
	for _, s := range asstRecSpecs {
		add(s.attr, s.pool)
	}
	add("hobby_opinion", hobbies)
	add("pet_type", petTypes)
	add("decoy_entity", firstNames)
	for _, s := range recurringSpecs {
		add(s.attr, []string{s.label})
	}
	return pools
}

// TestFillerNeverStatesAFact is the soundness gate on the deep-history layer.
//
// Background threads exist to make retrieval hard, not to make it ambiguous. If a
// filler clause happened to contain a canonical answer token, an honest reader
// could retrieve a turn that is not the evidence and be "correct" for the wrong
// reason -- or a genuinely correct harness could be marked down for surfacing the
// wrong occurrence of the value. The check uses the GRADER'S OWN matcher
// (grade.Hit), so it tests exactly the containment the scorer performs rather
// than an approximation of it.
func TestFillerNeverStatesAFact(t *testing.T) {
	vocab := fillerVocabulary()
	for attr, pool := range answerPools() {
		for _, value := range pool {
			for _, v := range vocab {
				if grade.Hit(value, v) {
					t.Errorf("filler vocabulary states a possible answer (%s=%q):\n  %q", attr, value, v)
				}
			}
		}
	}
}

// TestFillerIsDisjointFromGeneratedPlans is the end-to-end companion to the pool
// audit above: it confirms the property survives assembly, across seeds, on the
// facts a plan actually draws.
func TestFillerIsDisjointFromGeneratedPlans(t *testing.T) {
	vocab := fillerVocabulary()
	for seed := int64(1); seed <= 40; seed++ {
		plan, err := BuildPlanForVersion(seed, v8Opts(), protocol.BenchVersionV8)
		if err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		for _, f := range plan.Facts {
			for _, v := range vocab {
				if grade.Hit(f.Value, v) {
					t.Fatalf("seed %d: filler vocabulary states fact %s (%s=%q):\n  %q",
						seed, f.ID, f.Attribute, f.Value, v)
				}
			}
		}
	}
}

// TestFillerRendersEveryTemplateSlot guards against a typo'd or renamed slot
// silently reaching the haystack as literal "{thing}" text. fillerSlots has no
// unknown-symbol fallback, so this is the only thing standing between a bad
// template and a corrupted dataset.
func TestFillerRendersEveryTemplateSlot(t *testing.T) {
	r := rand.New(rand.NewSource(7))
	used := map[string]bool{}
	for _, s := range fillerSubjects {
		for stage := 0; stage < fillerStages; stage++ {
			for i := 0; i < 12; i++ {
				b := fillerBeat(r, s, stage, used)
				for _, text := range []string{b.UserText, b.AsstText} {
					if strings.ContainsAny(text, "{}") {
						t.Fatalf("unfilled slot in %s stage %d: %q", s.coref, stage, text)
					}
					if strings.TrimSpace(text) == "" {
						t.Fatalf("empty render for %s stage %d", s.coref, stage)
					}
				}
			}
		}
	}
}

// TestDeepHistoryIsVersionGated pins the contract split: pre-v8 plans must carry
// no background threads whatsoever, whatever Opts say. Without this, a stray
// ungated draw would move every frozen contract's bytes and break the
// reproducibility promise for every already-scored run.
func TestDeepHistoryIsVersionGated(t *testing.T) {
	opts := v8Opts()
	for _, version := range []int{protocol.BenchVersionV2, protocol.BenchVersionV5, protocol.BenchVersionV7} {
		plan, err := BuildPlanForVersion(99, opts, version)
		if err != nil {
			t.Fatalf("v%d: %v", version, err)
		}
		// Match on the notes and labels, not the corefs: a coref like "the roof"
		// is a substring of ordinary persona material ("the rooftop apiary" is a
		// project name), so it would false-positive. Notes and labels are long and
		// distinctive enough to be unambiguous evidence of a leak.
		for _, s := range plan.Sessions {
			for _, b := range s.Beats {
				text := b.UserText + " " + b.AsstText
				for _, subj := range fillerSubjects {
					for _, probe := range append([]string{subj.label}, subj.notes...) {
						if strings.Contains(text, probe) {
							t.Fatalf("v%d leaked a background thread (%q) into a pre-v8 plan: %q",
								version, probe, text)
						}
					}
				}
			}
		}
	}
	v8, err := BuildPlanForVersion(99, opts, protocol.BenchVersionV8)
	if err != nil {
		t.Fatal(err)
	}
	noise := 0
	for _, s := range v8.Sessions {
		for _, b := range s.Beats {
			if b.Kind == BeatNoise {
				noise++
			}
		}
	}
	if noise < opts.FillerBeats/2 {
		t.Fatalf("v8 emitted %d filler beats, want ~%d", noise, opts.FillerBeats)
	}
}

// TestBackgroundThreadsAreCoherent checks the property that separates a real
// distractor history from padding: a thread names its subject in full ONCE and
// refers back through a short coreferent afterwards. A haystack where every
// filler turn restates its topic is one a lexical retriever can partition
// cheaply, and it is not how a person talks about an ongoing problem.
func TestBackgroundThreadsAreCoherent(t *testing.T) {
	plan, err := BuildPlanForVersion(2026, v8Opts(), protocol.BenchVersionV8)
	if err != nil {
		t.Fatal(err)
	}
	total, restated := 0, 0
	distinct := map[string]bool{}
	perSessionSubjects := 0
	for _, s := range plan.Sessions {
		seen := map[string]bool{}
		for _, b := range s.Beats {
			if b.Kind != BeatNoise {
				continue
			}
			total++
			for _, subj := range fillerSubjects {
				if strings.Contains(b.UserText, subj.label) {
					restated++
					distinct[subj.coref] = true
					seen[subj.coref] = true
					break
				}
				if strings.Contains(b.UserText, subj.coref) {
					distinct[subj.coref] = true
					seen[subj.coref] = true
					break
				}
			}
		}
		if len(seen) > perSessionSubjects {
			perSessionSubjects = len(seen)
		}
	}
	if total == 0 {
		t.Fatal("no background-thread turns in a v8 plan")
	}
	if len(distinct) < 8 {
		t.Fatalf("only %d distinct threads across the history, want a running population", len(distinct))
	}
	if perSessionSubjects < 3 {
		t.Fatalf("busiest session touches only %d threads; sessions should read as a "+
			"life with several concerns in flight, not one monologue", perSessionSubjects)
	}
	// The full subject phrase is the kickoff turn only. If most turns restated it,
	// a lexical retriever could cluster and discount the whole filler layer.
	if got := float64(restated) / float64(total); got > 0.4 {
		t.Fatalf("%.0f%% of background turns restate their full subject phrase, want under 40%%", got*100)
	}
}

// TestElaboratePreservesValues pins the invariant the whole length-parity idea
// rests on: elaboration only APPENDS. Every canonical answer token must survive
// it verbatim, or containment grading breaks for every elaborated fact.
func TestElaboratePreservesValues(t *testing.T) {
	r := rand.New(rand.NewSource(11))
	plan, err := BuildPlanForVersion(5, v8Opts(), protocol.BenchVersionV8)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range plan.Facts {
		base := Beat{Kind: BeatFact, FactID: f.ID, UserText: f.UserText, AsstText: f.AsstText}
		for i := 0; i < 8; i++ {
			got := Elaborate(r, base)
			if !strings.HasPrefix(got.UserText, base.UserText) {
				t.Fatalf("fact %s: elaboration rewrote the user turn:\n base %q\n got  %q",
					f.ID, base.UserText, got.UserText)
			}
			if !strings.HasPrefix(got.AsstText, base.AsstText) {
				t.Fatalf("fact %s: elaboration rewrote the assistant turn", f.ID)
			}
		}
	}
}
