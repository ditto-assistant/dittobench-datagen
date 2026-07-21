package gen

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/ditto-assistant/dittobench-datagen/persona"
	"github.com/ditto-assistant/dittobench-datagen/protocol"
)

// Multi-hop relational retrieval (bench_version 5; v5 plan section 4.9 — the KG
// moat). Every other memory case is answerable from a SINGLE stored pair; a
// grounded champion with strong single-shot retrieval aces them. A multi-hop case
// is answerable only by JOINING two or more memories that no single pair contains:
//
//	session A (early):  "My sister is Dana."
//	session B (later):  "Dana adopted a puppy and named it <coined>."
//	question:           "What did my sister name her puppy?"  -> <coined>
//
// The leaf token exists in the store but is reachable only by traversing the
// relation link (mine -> sister -> Dana -> puppy). A grep parser and single-shot
// vector recall both miss it: the query "my sister's puppy" lexically matches the
// RELATION pair (no leaf) and the LEAF pair only via the intermediary entity that
// the query never names. Resolving it needs the linked knowledge graph subject
// promotion builds (backend kg.go), which is the production moat the aligned
// champion stack ships and the single strongest discriminator between a shallow
// retriever and a real one.
//
// Anti-shotgun: each case seeds a WRONG-RELATIVE decoy on the same leaf attribute
// ("my cousin adopted a puppy named <other>"), so a one-hop match on the leaf
// entity ("puppy") surfaces the cousin's value and zeros. The join must resolve
// the correct relative.
const (
	// QTMultiHop is a relational-join question answerable only by traversing a link
	// no single stored pair contains.
	QTMultiHop = "multi-hop-relational"
)

type multiHopSuite struct {
	Cases []StagedCase
	Pairs []protocol.MemoryPair
}

func multiHopEnabled(benchVersion int) bool { return benchVersion >= protocol.BenchVersionV5 }

// multiHopCasesFor is the seed-independent multi-hop quota. The chain needs its
// relation intro and leaf fact seeded before the question, and benefits from
// spanning sessions/waves, so single-wave (small) runs carry none.
func multiHopCasesFor(n, nWaves int) int {
	switch {
	case nWaves < 2:
		return 0
	case n >= 70:
		return 4
	case n >= 40:
		return 3
	case n >= 20:
		return 2
	default:
		return 0
	}
}

// relativeNames are uncommon given names for the intermediary entity. Uncommon so
// a collision with a rendered decoy person is unlikely; the join is by exact name
// match between the relation pair and the leaf pair.
var relativeNames = []string{
	"Dana", "Priya", "Marcus", "Ingrid", "Rafael", "Yuki", "Nadia", "Emeka",
	"Lena", "Tomas", "Anouk", "Darius", "Freya", "Kwame", "Sofia", "Bjorn",
}

// relationPair is a (target, decoy) relation on the same leaf attribute. The
// question asks about the TARGET relative; naming the DECOY relative's leaf value
// is the shallow one-hop error.
type relationPair struct {
	target string
	decoy  string
}

var relationPairs = []relationPair{
	{target: "sister", decoy: "cousin"},
	{target: "brother", decoy: "uncle"},
	{target: "best friend", decoy: "neighbor"},
	{target: "aunt", decoy: "coworker"},
	{target: "mentor", decoy: "landlord"},
}

// leafFact is a joinable attribute whose value is a COINED token (a pet/boat/plant
// name — plausibly any string, so a coined value reads naturally). intro states
// the relation; fact attaches the coined value to the intermediary; ask queries
// the target relative's leaf.
type leafFact struct {
	// factTmpl: %s=person, %s=value. askTmpl: (no args) the join question.
	factTmpl string
	askTmpl  string // %s = target relation
}

var leafFacts = []leafFact{
	{factTmpl: "%s adopted a puppy last month and named it %s.", askTmpl: "What did my %s name their puppy?"},
	{factTmpl: "%s finally named their sailboat %s.", askTmpl: "What's the name of my %s's sailboat?"},
	{factTmpl: "%s got a kitten and is calling it %s.", askTmpl: "What did my %s name their kitten?"},
	{factTmpl: "%s named their new houseplant %s, of all things.", askTmpl: "What did my %s name their houseplant?"},
}

// buildMultiHop generates the multi-hop relational cases. Deterministic per (seed,
// draw position): all picks draw from the shared suite rng at a fixed point in
// GenerateMemorySuite, gated so a pre-v5 contract's draw sequence is untouched.
func buildMultiHop(r *rand.Rand, seed int64, plan *persona.Plan, n, nWaves int) multiHopSuite {
	var out multiHopSuite
	nCases := multiHopCasesFor(n, nWaves)
	if nCases == 0 {
		return out
	}
	pairIdx := 0
	ordinal := 0
	// The relation intro and the leaf fact are seeded in DIFFERENT sessions, so the
	// join is genuinely cross-session (the answer is in no single session window).
	seedPair := func(session int, prompt, response string) {
		ts := persona.TimeAnchor(plan).Add(time.Duration(6+session)*time.Hour + time.Duration(pairIdx*beatSpacingMinutes)*time.Minute)
		out.Pairs = append(out.Pairs, protocol.MemoryPair{
			PairID:    fmt.Sprintf("p-mh-%d", pairIdx),
			SessionID: fmt.Sprintf("sess-mh-%d", session),
			Timestamp: ts.Format(time.RFC3339),
			Prompt:    prompt,
			Response:  response,
		})
		pairIdx++
	}

	for c := 0; c < nCases; c++ {
		rel := relationPairs[r.Intn(len(relationPairs))]
		leaf := leafFacts[r.Intn(len(leafFacts))]
		// Distinct intermediary names for target and decoy.
		tPerm := r.Perm(len(relativeNames))
		targetPerson := relativeNames[tPerm[0]]
		decoyPerson := relativeNames[tPerm[1]]
		targetVal := persona.CoinShaped(seed, fmt.Sprintf("mh|val|%d", c))
		decoyVal := persona.CoinShaped(seed, fmt.Sprintf("mh|dec|%d", c))

		// Target chain: relation intro and leaf fact, in two distinct sessions.
		seedPair(4*c, "My "+rel.target+" is "+targetPerson+".", "Good to know.")
		seedPair(4*c+1, fmt.Sprintf(leaf.factTmpl, targetPerson, targetVal), "Noted.")
		// Decoy chain: wrong relative on the SAME leaf attribute, two more sessions.
		seedPair(4*c+2, "My "+rel.decoy+" is "+decoyPerson+".", "Got it.")
		seedPair(4*c+3, fmt.Sprintf(leaf.factTmpl, decoyPerson, decoyVal), "Nice.")

		mc := protocol.MemoryCase{
			ID:                protocol.OpaqueCaseID(seed, "memmh", ordinal),
			QuestionType:      QTMultiHop,
			Question:          fmt.Sprintf(leaf.askTmpl, rel.target),
			ExpectedAnswer:    targetVal,
			AnswerKind:        protocol.AnswerValue,
			DistractorAnswers: []string{decoyVal},
		}
		out.Cases = append(out.Cases, StagedCase{Case: mc, RunAfterWave: nWaves - 1})
		ordinal++
	}
	return out
}
