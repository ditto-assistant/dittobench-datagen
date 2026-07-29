package gen

import (
	"fmt"
	"sort"
	"time"

	"github.com/ditto-assistant/dittobench-datagen/persona"
	"github.com/ditto-assistant/dittobench-datagen/protocol"
)

// User graph identifiers for multi-graph isolation. The
// primary persona is seeded under PrimaryUser (default "miner");
// a second persona is seeded under SecondaryUser. Isolation cases then query one
// user while the OTHER user's graph holds a conflicting value for the same
// attribute. A harness that respects the user_id scoping answers correctly; one
// that leaks across graphs returns the other user's value and is wrong.
const (
	PrimaryUser   = "miner"
	SecondaryUser = "colleague"
)

// isolationSalt perturbs the master seed to draw a DISTINCT second persona
// (different names/values) from the same pools, deterministically.
const isolationSalt = 0x5ec0_11ab_0000_0001

// IsolationSuite is the secondary user graph plus the cross-user isolation cases
// (both A-scoped and B-scoped). It is layered on top of the primary MemorySuite
// by the pipeline: the secondary graph is seeded once under SecondaryUser, and
// the cases are run interleaved with the primary cases (each carrying the user_id
// it must be answered under).
type IsolationSuite struct {
	SecondaryWave protocol.SeedRequest // B's haystack, seeded under SecondaryUser
	Cases         []StagedCase         // isolation cases, each with StagedCase.UserID set
}

// isolationOpts keeps the secondary (contamination) persona small: it exists to
// hold conflicting values, not to be the primary test surface.
func isolationOpts() persona.Opts {
	return persona.Opts{Sessions: 4, Projects: 3, Trips: 2, Pets: 1, UpdateChains: 1, Reversals: 1, DecoyPeople: 3, DomainItems: 3, LongChain: 0}
}

// GenerateIsolation builds the multi-graph isolation layer. It is a pure
// function of (seed, primaryN, nWaves, isoCases): it rebuilds the primary plan
// from the seed (BuildPlan is pure) to align isolation cases with the primary
// seeding waves, draws a distinct secondary persona, and emits up to isoCases
// isolation cases split between A-scoped (query PrimaryUser, the conflicting
// value lives in B) and B-scoped (query SecondaryUser, the conflict lives in A).
// The secondary haystack is TEMPLATE-rendered (no LLM): it is contamination, so
// it adds zero generator token cost. isoCases<=0 returns an empty suite.
func GenerateIsolation(seed int64, primaryN, nWaves, isoCases int) IsolationSuite {
	suite, _ := GenerateIsolationForVersion(seed, primaryN, nWaves, isoCases, protocol.BenchVersionV2)
	return suite
}

// GenerateIsolationForVersion builds the isolation layer for an explicit
// benchmark contract.
func GenerateIsolationForVersion(seed int64, primaryN, nWaves, isoCases, benchVersion int) (IsolationSuite, error) {
	if isoCases <= 0 {
		return IsolationSuite{}, nil
	}
	if nWaves < 1 {
		nWaves = 1
	}
	pPlan, err := persona.BuildPlanForVersion(seed, personaOptsFor(primaryN), benchVersion)
	if err != nil {
		return IsolationSuite{}, err
	}
	sPlan, err := persona.BuildPlanForVersion(seed^isolationSalt, isolationOpts(), benchVersion)
	if err != nil {
		return IsolationSuite{}, err
	}

	// Secondary haystack: template-rendered (non-LLM), fully Tier-A (prepared
	// subjects) so it is retrievable: a cross-graph leak actually surfaces it.
	sPairs, sEvidence := RenderHaystack(sPlan)
	sSubjects, sLinks := synthesizeSubjects(sPlan, sEvidence, nil)
	secondary := protocol.SeedRequest{
		UserID:   SecondaryUser,
		Pairs:    sPairs,
		Subjects: sSubjects,
		Links:    sLinks,
	}

	pCur, sCur := currentScalars(pPlan), currentScalars(sPlan)
	pRecall, sRecall := recallByAttr(pPlan), recallByAttr(sPlan)
	pFW := factWaves(pPlan, nWaves)

	// Attributes both personas hold as a current scalar self-fact with DIFFERENT
	// values: the conflict axis. Sorted for determinism.
	var attrs []string
	for a, v := range pCur {
		if sv, ok := sCur[a]; ok && sv != "" && sv != v {
			attrs = append(attrs, a)
		}
	}
	sort.Strings(attrs)

	cases := make([]StagedCase, 0, isoCases)
	if benchVersion >= protocol.BenchVersionV8 {
		// V7's alternating selector could emit fewer than IsoCases because one
		// direction lacked a recall question for a particular attribute. That made
		// total case count seed-dependent. V8 considers both directions, then takes
		// the exact public-profile quota while preserving the five lifecycle cases
		// appended below.
		candidates := make([]StagedCase, 0, len(attrs)*2)
		for i, a := range attrs {
			q, ok := pRecall[a]
			if ok {
				candidates = append(candidates, StagedCase{
					Case: protocol.MemoryCase{
						ID:              protocol.OpaqueCaseID(seed, "iso-a", i),
						QuestionID:      "iso-a-" + a,
						QuestionType:    "isolation",
						Question:        q.Text,
						ExpectedAnswer:  q.Answer,
						ForbiddenAnswer: sCur[a],
					},
					RunAfterWave: caseUnlockWave(q, pFW),
					UserID:       PrimaryUser,
				})
			}
			q, ok = sRecall[a]
			if ok {
				candidates = append(candidates, StagedCase{
					Case: protocol.MemoryCase{
						ID:              protocol.OpaqueCaseID(seed, "iso-b", i),
						QuestionID:      "iso-b-" + a,
						QuestionType:    "isolation",
						Question:        q.Text,
						ExpectedAnswer:  q.Answer,
						ForbiddenAnswer: pCur[a],
					},
					RunAfterWave: 0,
					UserID:       SecondaryUser,
				})
			}
		}
		target := v8ScalarIsolationBudget(primaryN, isoCases)
		if len(candidates) < target {
			return IsolationSuite{}, fmt.Errorf("v8 isolation quota needs %d cases, generated %d", target, len(candidates))
		}
		cases = append(cases, candidates[:target]...)
	} else {
		i := 0
		for _, a := range attrs {
			if len(cases) >= isoCases {
				break
			}
			// Frozen v3-v7 selector: alternate query direction.
			if len(cases)%2 == 0 {
				q, ok := pRecall[a]
				if !ok {
					continue
				}
				cases = append(cases, StagedCase{
					Case: protocol.MemoryCase{
						ID:              protocol.OpaqueCaseID(seed, "iso-a", i),
						QuestionID:      "iso-a-" + a,
						QuestionType:    "isolation",
						Question:        q.Text,
						ExpectedAnswer:  q.Answer,
						ForbiddenAnswer: sCur[a],
					},
					RunAfterWave: caseUnlockWave(q, pFW),
					UserID:       PrimaryUser,
				})
			} else {
				q, ok := sRecall[a]
				if !ok {
					continue
				}
				cases = append(cases, StagedCase{
					Case: protocol.MemoryCase{
						ID:              protocol.OpaqueCaseID(seed, "iso-b", i),
						QuestionID:      "iso-b-" + a,
						QuestionType:    "isolation",
						Question:        q.Text,
						ExpectedAnswer:  q.Answer,
						ForbiddenAnswer: pCur[a],
					},
					RunAfterWave: 0,
					UserID:       SecondaryUser,
				})
			}
			i++
		}
	}

	// B3 cross-user lifecycle, v3 ONLY. The read-path leak above is probed by the
	// cases just built, but lifecycle chains (gen/lifecycle.go) carry no UserID at
	// all, so a WRITE or DELETE under user A followed by a read of the same
	// attribute under user B was never exercised. A harness keeping a global
	// (rather than per-user) saved/deleted map passes every existing case and
	// leaks here.
	//
	// Gated on the version because v2 is a frozen contract: adding cases would
	// change its bytes and break the reproducibility of every run already scored
	// under it.
	if benchVersion >= protocol.BenchVersionV3 {
		xPairs, xCases := crossUserLifecycle(seed, nWaves, benchVersion)
		secondary.Pairs = append(secondary.Pairs, xPairs...)
		cases = append(cases, xCases...)
	}
	if benchVersion >= protocol.BenchVersionV8 {
		for i := range cases {
			cases[i].Case.BenchVersion = benchVersion
		}
	}

	return IsolationSuite{SecondaryWave: secondary, Cases: cases}, nil
}

func v8ScalarIsolationBudget(primaryN, requested int) int {
	// The fixed full-profile total is more valuable than preserving v7's
	// seed-variable scalar slice. Medium spends its isolation budget on the five
	// stronger write/delete/read cross-user cases; full retains four scalar
	// conflicts in addition to those five.
	switch primaryN {
	case 185:
		return 4
	case 52:
		return 0
	default:
		return requested
	}
}

// Cross-user lifecycle nouns. Distinct from the single-user lifecycle nouns
// (lcSaveNoun and friends) so the two chains can never be confused for one
// another, and distinct from every persona attribute so no distractor pool moves.
const (
	xuWriteNoun = "safe combination"
	xuDelNoun   = "spare key code"
)

// crossUserLifecycle builds the B3 probe: state is mutated under the PRIMARY
// user, then the SAME attribute is read under the SECONDARY user, where the
// correct answer is B's own value and A's post-mutation state is forbidden.
//
// Both chains establish A's value by INSTRUCTION rather than by a seeded pair,
// so this needs to add pairs only to B's haystack (which the isolation suite
// owns) and does not have to reach into the primary wave.
//
//   - write chain: A is told a value; B already has its own. Reading under B
//     must return B's value. A harness with a global saved map returns A's.
//   - delete chain: A is told a value and then told to erase it; B's value for
//     the same noun must be UNAFFECTED. A harness with a global deleted map
//     drops B's value too and cannot answer, which is the leak in the other
//     direction -- a deletion crossing a user boundary.
func crossUserLifecycle(seed int64, nWaves, benchVersion int) ([]protocol.MemoryPair, []StagedCase) {
	if nWaves < 2 {
		return nil, nil // reads land in the last wave; a single-wave run has no room
	}
	readWave := nWaves - 1

	aWrite := persona.CoinShaped(seed, "xu-a-write")
	bWrite := persona.CoinShaped(seed, "xu-b-write")
	aDel := persona.CoinShaped(seed, "xu-a-del")
	bDel := persona.CoinShaped(seed, "xu-b-del")

	pairs := []protocol.MemoryPair{
		xuPair(seed, 0, "Keep this for me: my "+xuWriteNoun+" is "+bWrite+".",
			"Noted. Your "+xuWriteNoun+" is "+bWrite+"."),
		xuPair(seed, 1, "For your records, my "+xuDelNoun+" is "+bDel+".",
			"Got it. Your "+xuDelNoun+" is "+bDel+"."),
	}

	cases := []StagedCase{
		// A writes its own value for the shared noun.
		{
			Case: protocol.MemoryCase{
				ID:             protocol.OpaqueCaseID(seed, "xu", 0),
				QuestionID:     "xu-write-w",
				QuestionType:   QTLifecycleWrite,
				Question:       "Please remember that my " + xuWriteNoun + " is " + aWrite + ".",
				ExpectedAnswer: aWrite,
			},
			RunAfterWave: 0,
			UserID:       PrimaryUser,
		},
		// B reads the same noun. B's own value is correct; A's is a cross-user leak.
		{
			Case: protocol.MemoryCase{
				ID:              protocol.OpaqueCaseID(seed, "xu", 1),
				QuestionID:      "xu-write-r",
				QuestionType:    QTLifecycleRead,
				Question:        "What's my " + xuWriteNoun + "?",
				ExpectedAnswer:  bWrite,
				ForbiddenAnswer: aWrite,
			},
			RunAfterWave: readWave,
			UserID:       SecondaryUser,
		},
		// A writes, then erases, its own value for the second noun.
		{
			Case: protocol.MemoryCase{
				ID:             protocol.OpaqueCaseID(seed, "xu", 2),
				QuestionID:     "xu-del-w",
				QuestionType:   QTLifecycleWrite,
				Question:       "Hold onto this for me: my " + xuDelNoun + " is " + aDel + ".",
				ExpectedAnswer: aDel,
			},
			RunAfterWave: 0,
			UserID:       PrimaryUser,
		},
		{
			Case: protocol.MemoryCase{
				ID:             protocol.OpaqueCaseID(seed, "xu", 3),
				QuestionID:     "xu-del-d",
				QuestionType:   QTLifecycleWrite,
				Question:       "Please forget my " + xuDelNoun + " entirely.",
				ExpectedAnswer: xuDelNoun,
				// Instruction, not a question -- see lc-del-w. The cross-user
				// survival check on the next case carries the real signal.
				AnswerKind: acknowledgeKindFor(benchVersion),
			},
			RunAfterWave: 0,
			UserID:       PrimaryUser,
		},
		// B's value for that noun must have survived A's deletion.
		{
			Case: protocol.MemoryCase{
				ID:              protocol.OpaqueCaseID(seed, "xu", 4),
				QuestionID:      "xu-del-r",
				QuestionType:    QTLifecycleRead,
				Question:        "What's my " + xuDelNoun + "?",
				ExpectedAnswer:  bDel,
				ForbiddenAnswer: aDel,
			},
			RunAfterWave: readWave,
			UserID:       SecondaryUser,
		},
	}
	return pairs, cases
}

// xuPair renders one pair of B's cross-user lifecycle haystack. It shares the
// secondary graph's session so it is retrievable exactly like the rest of B's
// contamination.
func xuPair(seed int64, idx int, prompt, response string) protocol.MemoryPair {
	ts := protocol.DatasetEpoch.Add(time.Duration(idx*beatSpacingMinutes) * time.Minute)
	return protocol.MemoryPair{
		PairID:    fmt.Sprintf("p-xu-%d", idx),
		SessionID: "sess-xu",
		Timestamp: ts.Format(time.RFC3339),
		Prompt:    prompt,
		Response:  response,
	}
}

// currentScalars maps each current (latest-value-wins) scalar self attribute to
// its value for a plan.
func currentScalars(plan *persona.Plan) map[string]string {
	out := map[string]string{}
	for _, f := range plan.Facts {
		if f.Kind == persona.KindScalar && f.Entity == "self" && f.Current {
			out[f.Attribute] = f.Value
		}
	}
	return out
}

// recallByAttr maps each scalar attribute to a single-session-recall question for
// it (the first one derived). Used to phrase an isolation question in the
// persona's own recall wording.
func recallByAttr(plan *persona.Plan) map[string]persona.Question {
	factAttr := map[string]string{}
	for _, f := range plan.Facts {
		factAttr[f.ID] = f.Attribute
	}
	out := map[string]persona.Question{}
	for _, q := range persona.DeriveQuestions(plan) {
		if q.Abstain || q.Type != persona.QTSingleSession || len(q.Evidence) == 0 {
			continue
		}
		attr := factAttr[q.Evidence[0]]
		if attr == "" {
			continue
		}
		if _, seen := out[attr]; !seen {
			out[attr] = q
		}
	}
	return out
}
