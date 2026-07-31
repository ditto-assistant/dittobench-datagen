package universe

import (
	"errors"
	"fmt"
	"hash/fnv"
	"math/rand"
	"sort"
	"strings"

	"github.com/ditto-assistant/dittobench-datagen/grade"
	"github.com/ditto-assistant/dittobench-datagen/protocol"
)

// QuestionPlan is the validator-internal bridge between world generation and a
// rendered MemoryCase. RequiredPairIDs identify the planted needles; Facts,
// Constraints, and Operations make the intended reasoning burden auditable.
// Only Case crosses the harness /run boundary.
type QuestionPlan struct {
	Case            protocol.MemoryCase
	RequiredPairIDs []string
	Facts           []string
	Constraints     []string
	Operations      []string

	oracleKind  string
	oracleIndex int
}

const (
	oracleContactCurrent         = "contact-current"
	oracleContactPrevious        = "contact-previous"
	oracleProjectOutstanding     = "project-outstanding"
	oracleProjectLeadCurrent     = "project-lead-current"
	oracleProjectLeadPrevious    = "project-lead-previous"
	oracleTripCurrent            = "trip-current"
	oracleTripChangedLegPrevious = "trip-changed-leg-previous"
	oracleTripChangedLegCurrent  = "trip-changed-leg-current"
	oracleTripLongestCurrent     = "trip-longest-current"
)

var errLexicalShortcut = errors.New("lexical shortcut")

// QuestionPlans executes the v8 contract in three explicit phases:
//
//  1. Generate produced one coherent world and planted its facts in Pairs.
//  2. candidates derives latent multi-fact plans from that world.
//  3. validatePlan proves every rendered question has one answer, all needles
//     exist, and the expected answer is the deterministic oracle result.
//
// The final seed-keyed shuffle varies the public mix without changing the
// answerability proof. Accidental lexical shortcuts are deterministically
// excluded; every structural ambiguity, missing needle, or oracle mismatch
// fails generation instead of entering a scored dataset.
func (w World) QuestionPlans(count int) ([]QuestionPlan, error) {
	if count <= 0 {
		return nil, nil
	}
	candidates := w.questionCandidates()
	if count > len(candidates) {
		return nil, fmt.Errorf("world has %d validated question candidates, need %d", len(candidates), count)
	}
	r := rand.New(rand.NewSource(questionSeed(w.Seed)))
	r.Shuffle(len(candidates), func(i, j int) { candidates[i], candidates[j] = candidates[j], candidates[i] })
	selected := make([]QuestionPlan, 0, count)
	seenQuestions := make(map[string]bool, len(candidates))
	for i := range candidates {
		if err := w.validatePlan(candidates[i]); err != nil {
			// A random surface can occasionally create an accidental lexical
			// shortcut. Exclude that surface deterministically; structural
			// ambiguity, missing evidence, and oracle drift still fail generation.
			if errors.Is(err, errLexicalShortcut) {
				continue
			}
			return nil, fmt.Errorf("candidate %s: %w", candidates[i].Case.QuestionType, err)
		}
		q := strings.ToLower(strings.TrimSpace(candidates[i].Case.Question))
		if seenQuestions[q] {
			return nil, fmt.Errorf("duplicate rendered question %q", candidates[i].Case.Question)
		}
		seenQuestions[q] = true
		selected = append(selected, candidates[i])
		if len(selected) == count {
			return selected, nil
		}
	}
	return nil, fmt.Errorf("world has only %d shortcut-free question candidates, need %d", len(selected), count)
}

func (w World) questionCandidates() []QuestionPlan {
	out := make([]QuestionPlan, 0, 2*len(w.People)+3*len(w.Projects)+4*len(w.Trips))
	for i, p := range w.People {
		current := []string{
			fmt.Sprintf("What address should I actually use now for %s, my %s in %s from the %s?", p.Nickname, p.Relation, p.City, p.Context),
			fmt.Sprintf("I need to reach %s, my %s in %s from the %s. Which email is current?", p.Nickname, p.Relation, p.City, p.Context),
			fmt.Sprintf("Which up-to-date email belongs to my %s %s, the person in %s who handled the %s?", p.Relation, p.Nickname, p.City, p.Context),
			fmt.Sprintf("For the %s follow-up, what is the corrected address for %s, my %s in %s?", p.Context, p.Nickname, p.Relation, p.City),
		}[i%4]
		out = append(out, w.personPlan(oracleContactCurrent, i, current, p.Email, p.PreviousEmail))

		previous := []string{
			fmt.Sprintf("Before %s changed addresses, which email had I saved for my %s in %s from the %s?", p.Nickname, p.Relation, p.City, p.Context),
			fmt.Sprintf("What was the earlier address for my %s in %s I call %s from the %s, before the contact correction?", p.Relation, p.City, p.Nickname, p.Context),
			fmt.Sprintf("I need the pre-correction email for %s — my %s who handled the %s in %s. What was it?", p.Nickname, p.Relation, p.Context, p.City),
			fmt.Sprintf("Looking back before the update, which address did I first have for %s, my %s from the %s who lives in %s?", p.Nickname, p.Relation, p.Context, p.City),
		}[i%4]
		out = append(out, w.personPlan(oracleContactPrevious, i, previous, p.PreviousEmail, p.Email))
	}

	for i, p := range w.Projects {
		lead := w.People[p.Lead]
		outstanding := []string{
			fmt.Sprintf("For %q, the %s work for %s, what is still owed to %s once the approved correction and the payment already sent are reconciled? Give cents.", p.Alias, p.Purpose, p.Client, p.Vendor),
			fmt.Sprintf("AP needs the remaining balance in cents for %s's invoice on %q for %s. Use the corrected total, not the draft, and account for our payment.", p.Vendor, p.Alias, p.Client),
			fmt.Sprintf("How many cents remain on the corrected %s bill tied to %q, the %s project for %s, after what we already paid?", p.Vendor, p.Alias, p.Purpose, p.Client),
			fmt.Sprintf("Reconcile %q for %s: after replacing the original %s invoice figure with the approved one and subtracting the partial payment, what balance remains in cents?", p.Alias, p.Client, p.Vendor),
		}[i%4]
		out = append(out, QuestionPlan{
			Case:            memoryCase(w.Seed, oracleProjectOutstanding, i, outstanding, fmt.Sprintf("%d", p.OutstandingCents), protocol.AnswerNumber, w.moneyDistractors(i, p.OutstandingCents)),
			RequiredPairIDs: []string{p.ContextPairID, p.LedgerPairID, p.CorrectionPairID},
			Facts:           []string{"project alias", "client and purpose", "draft invoice and payment", "approved correction", "vendor identity"},
			Constraints:     []string{p.Alias, p.Client, p.Vendor, p.Purpose}, Operations: []string{"resolve project alias", "replace draft with correction", "subtract payment"},
			oracleKind: oracleProjectOutstanding, oracleIndex: i,
		})

		current := []string{
			fmt.Sprintf("Who should get the %q handoff internally, and what current email should I use? I mean the %s work for %s, not the client contact.", p.Alias, p.Purpose, p.Client),
			fmt.Sprintf("For %s's %s project that we call %q, give me the corrected email for its internal owner.", p.Client, p.Purpose, p.Alias),
			fmt.Sprintf("What up-to-date address belongs to the person running %q on our side — %s's %s engagement?", p.Alias, p.Client, p.Purpose),
			fmt.Sprintf("I am sending the %q update. Resolve the internal owner from the %s work for %s, then use their current rather than original email.", p.Alias, p.Purpose, p.Client),
		}[i%4]
		out = append(out, w.projectLeadPlan(oracleProjectLeadCurrent, i, current, lead.Email, lead.PreviousEmail))

		previous := []string{
			fmt.Sprintf("Before the address correction, what email did I have for the internal lead on %q, the %s work for %s?", p.Alias, p.Purpose, p.Client),
			fmt.Sprintf("Find the earlier email for whoever owns %q, our %s project for %s — not their current one.", p.Alias, p.Purpose, p.Client),
			fmt.Sprintf("What was the original address for the internal owner of the %s project for %s that we call %q?", p.Purpose, p.Client, p.Alias),
			fmt.Sprintf("Looking back before the update, which email was saved for %q's internal owner on the %s work for %s?", p.Alias, p.Purpose, p.Client),
		}[i%4]
		out = append(out, w.projectLeadPlan(oracleProjectLeadPrevious, i, previous, lead.PreviousEmail, lead.Email))
	}

	for i, trip := range w.Trips {
		changed := changedLeg(trip)
		commonEvidence := []string{trip.ContextPairID, trip.PlanPairID, trip.CorrectionPairID}
		constraints := []string{trip.Alias, trip.Purpose, trip.When, trip.Countries[changed]}

		current := []string{
			fmt.Sprintf("How long does %s work out to now — the %s trip in %s through %s, %s, and %s — after the itinerary change?", trip.Alias, trip.Purpose, trip.When, trip.Countries[0], trip.Countries[1], trip.Countries[2]),
			fmt.Sprintf("After updating the %s leg, what is the total length of %s, our %s trip from %s?", trip.Countries[changed], trip.Alias, trip.Purpose, trip.When),
			fmt.Sprintf("Reconcile the revised country stays for %s, our %s trip in %s. How many days is the whole route now?", trip.Alias, trip.Purpose, trip.When),
			fmt.Sprintf("What is the current all-in duration for the %s route we called %s in %s, including the corrected %s stay?", trip.Purpose, trip.Alias, trip.When, trip.Countries[changed]),
		}[i%4]
		out = append(out, tripPlan(w, oracleTripCurrent, i, current, trip.CurrentDays, commonEvidence, constraints))

		previous := []string{
			fmt.Sprintf("Before the itinerary correction, how many days had we planned for the leg that later changed on %s, our %s route in %s?", trip.Alias, trip.Purpose, trip.When),
			fmt.Sprintf("For %s, the %s trip from %s, how long was the stay that was later corrected before it changed?", trip.Alias, trip.Purpose, trip.When),
			fmt.Sprintf("Give me the old duration for the leg we later revised on the %s route called %s in %s.", trip.Purpose, trip.Alias, trip.When),
			fmt.Sprintf("Looking at the first plan, how many days was the stay that later changed on %s, our %s trip in %s?", trip.Alias, trip.Purpose, trip.When),
		}[i%4]
		out = append(out, tripPlan(w, oracleTripChangedLegPrevious, i, previous, trip.OldLegDays[changed], commonEvidence, []string{trip.Alias, trip.Purpose, trip.When}))

		leg := []string{
			fmt.Sprintf("On %s, the %s trip in %s, how many days is the revised %s stay itself?", trip.Alias, trip.Purpose, trip.When, trip.Countries[changed]),
			fmt.Sprintf("What is the corrected number of days in %s for %s, the %s route from %s?", trip.Countries[changed], trip.Alias, trip.Purpose, trip.When),
			fmt.Sprintf("For the %s itinerary known as %s, how long is the changed %s leg now, not the whole trip?", trip.Purpose, trip.Alias, trip.Countries[changed]),
			fmt.Sprintf("After the update to %s on our %s trip %s, what duration applies to that country leg?", trip.Countries[changed], trip.When, trip.Alias),
		}[i%4]
		out = append(out, tripPlan(w, oracleTripChangedLegCurrent, i, leg, trip.LegDays[changed], commonEvidence, constraints))

		longest := trip.LegDays[0]
		for _, days := range trip.LegDays[1:] {
			if days > longest {
				longest = days
			}
		}
		change := []string{
			fmt.Sprintf("After the %s correction, how many days is the longest country stay on %s, our %s trip in %s?", trip.Countries[changed], trip.Alias, trip.Purpose, trip.When),
			fmt.Sprintf("Compare every revised country leg for %s, the %s route from %s. What is the longest stay in days after the %s change?", trip.Alias, trip.Purpose, trip.When, trip.Countries[changed]),
			fmt.Sprintf("How many days is the longest current leg of the %s route called %s in %s, once the %s update is applied?", trip.Purpose, trip.Alias, trip.When, trip.Countries[changed]),
			fmt.Sprintf("For %s in %s, our %s trip with the revised %s leg, what is the maximum country-stay duration?", trip.Alias, trip.When, trip.Purpose, trip.Countries[changed]),
		}[i%4]
		out = append(out, tripPlan(w, oracleTripLongestCurrent, i, change, longest, commonEvidence, constraints))
	}
	return out
}

func (w World) personPlan(kind string, index int, question, answer, extraDistractor string) QuestionPlan {
	p := w.People[index]
	evidence := []string{p.IdentityPairID, p.WorkPairID, p.EmailPairID}
	facts := []string{"full identity", "nickname and relationship", "work and city context", "original address"}
	operations := []string{"resolve nickname", "join identity to work context", "select prior address state"}
	if kind == oracleContactCurrent {
		evidence = append(evidence, p.CorrectionPairID)
		facts = append(facts, "address correction")
		operations[2] = "apply address correction"
	}
	return QuestionPlan{
		Case:            memoryCase(w.Seed, kind, index, question, answer, protocol.AnswerValue, w.emailDistractors(index, extraDistractor)),
		RequiredPairIDs: evidence,
		Facts:           facts,
		Constraints:     []string{p.Nickname, p.Relation, p.Context, p.Employer}, Operations: operations,
		oracleKind: kind, oracleIndex: index,
	}
}

func (w World) projectLeadPlan(kind string, index int, question, answer, extraDistractor string) QuestionPlan {
	p := w.Projects[index]
	lead := w.People[p.Lead]
	evidence := []string{p.ContextPairID, lead.IdentityPairID, lead.WorkPairID, lead.EmailPairID}
	facts := []string{"project alias", "client and purpose", "internal owner", "owner nickname and relationship", "owner original address"}
	operations := []string{"resolve project alias", "follow ownership edge", "resolve person identity", "select prior address state"}
	if kind == oracleProjectLeadCurrent {
		evidence = append(evidence, lead.CorrectionPairID)
		facts = append(facts, "owner correction")
		operations[3] = "apply address correction"
	}
	return QuestionPlan{
		Case:            memoryCase(w.Seed, kind, index, question, answer, protocol.AnswerValue, w.emailDistractors(p.Lead, extraDistractor)),
		RequiredPairIDs: evidence,
		Facts:           facts,
		Constraints:     []string{p.Alias, p.Client, p.Purpose}, Operations: operations,
		oracleKind: kind, oracleIndex: index,
	}
}

func tripPlan(w World, kind string, index int, question string, answer int, evidence, constraints []string) QuestionPlan {
	return QuestionPlan{
		Case:            memoryCase(w.Seed, kind, index, question, fmt.Sprintf("%d", answer), protocol.AnswerNumber, w.tripDistractors(index, answer)),
		RequiredPairIDs: append([]string(nil), evidence...),
		Facts:           []string{"trip alias", "purpose and time", "three original leg durations", "changed country", "corrected leg duration"},
		Constraints:     append([]string(nil), constraints...), Operations: []string{"resolve trip alias", "apply itinerary correction", "select requested state", "sum or compare legs"},
		oracleKind: kind, oracleIndex: index,
	}
}

func memoryCase(seed int64, kind string, index int, question, answer, answerKind string, distractors []string) protocol.MemoryCase {
	id := protocol.OpaqueCaseID(seed, "world-memory-"+kind, index)
	return protocol.MemoryCase{
		BenchVersion: protocol.BenchVersionV8, ID: id, QuestionID: id,
		QuestionType: "world-" + kind, Question: question, ExpectedAnswer: answer,
		AnswerKind: answerKind, DistractorAnswers: distractors,
	}
}

func (w World) validatePlan(plan QuestionPlan) error {
	if len(plan.Facts) < 4 || len(plan.Constraints) < 3 || len(plan.Operations) < 2 {
		return fmt.Errorf("under-specified plan: facts=%d constraints=%d operations=%d", len(plan.Facts), len(plan.Constraints), len(plan.Operations))
	}
	if grade.Hit(plan.Case.ExpectedAnswer, plan.Case.Question) {
		return fmt.Errorf("question leaks expected answer %q", plan.Case.ExpectedAnswer)
	}
	required := make(map[string]bool, len(plan.RequiredPairIDs))
	for _, pairID := range plan.RequiredPairIDs {
		required[pairID] = true
	}
	for _, pair := range w.Pairs {
		if !required[pair.PairID] {
			continue
		}
		body := pair.Prompt + " " + pair.Response
		if !grade.Hit(plan.Case.ExpectedAnswer, body) {
			continue
		}
		if overlap := contentOverlap(plan.Case.Question, body); overlap > 0.34 {
			return fmt.Errorf("%w: answer-bearing pair %s overlaps question at %.2f", errLexicalShortcut, pair.PairID, overlap)
		}
	}
	if got, ok := w.resolveWithEvidence(plan, required); !ok || got != plan.Case.ExpectedAnswer {
		return fmt.Errorf("oracle resolved %q, case expects %q", got, plan.Case.ExpectedAnswer)
	}
	if got := renderedConstraintCount(plan); got < 3 {
		return fmt.Errorf("rendered question contains %d declared constraints, want at least 3", got)
	}
	if got := w.subjectMatches(plan); got != 1 {
		return fmt.Errorf("constraints resolve %d subjects, want exactly one", got)
	}
	wantEvidence := w.oracleEvidence(plan)
	if !sameStrings(plan.RequiredPairIDs, wantEvidence) {
		return fmt.Errorf("declared evidence %v does not match oracle dependency %v", plan.RequiredPairIDs, wantEvidence)
	}
	wantOperations := w.oracleOperations(plan)
	if !sameStrings(plan.Operations, wantOperations) {
		return fmt.Errorf("declared operations %v do not match oracle operations %v", plan.Operations, wantOperations)
	}
	for _, omitted := range plan.RequiredPairIDs {
		counterfactual := make(map[string]bool, len(required)-1)
		for pairID := range required {
			if pairID != omitted {
				counterfactual[pairID] = true
			}
		}
		if got, ok := w.resolveWithEvidence(plan, counterfactual); ok && got == plan.Case.ExpectedAnswer {
			return fmt.Errorf("evidence pair %s is not causally required", omitted)
		}
	}
	pairPos := make(map[string]int, len(w.Pairs))
	for i, pair := range w.Pairs {
		pairPos[pair.PairID] = i
	}
	lo, hi := len(w.Pairs), -1
	seenEvidence := map[string]bool{}
	for _, id := range plan.RequiredPairIDs {
		if seenEvidence[id] {
			return fmt.Errorf("duplicate evidence pair %s", id)
		}
		seenEvidence[id] = true
		pos, ok := pairPos[id]
		if !ok {
			return fmt.Errorf("missing planted evidence pair %s", id)
		}
		if pos < lo {
			lo = pos
		}
		if pos > hi {
			hi = pos
		}
	}
	if len(plan.RequiredPairIDs) < 3 || hi-lo < len(w.Pairs)/6 {
		return fmt.Errorf("evidence is not long-distance: pairs=%d span=%d/%d", len(plan.RequiredPairIDs), hi-lo, len(w.Pairs))
	}
	if len(plan.Case.DistractorAnswers) != 3 {
		return fmt.Errorf("distractor count=%d, want 3", len(plan.Case.DistractorAnswers))
	}
	seenDistractors := map[string]bool{}
	for _, value := range plan.Case.DistractorAnswers {
		if value == plan.Case.ExpectedAnswer || seenDistractors[value] {
			return fmt.Errorf("invalid distractor %q", value)
		}
		seenDistractors[value] = true
	}
	return nil
}

func (w World) oracleEvidence(plan QuestionPlan) []string {
	switch plan.oracleKind {
	case oracleContactCurrent:
		p := w.People[plan.oracleIndex]
		return []string{p.IdentityPairID, p.WorkPairID, p.EmailPairID, p.CorrectionPairID}
	case oracleContactPrevious:
		p := w.People[plan.oracleIndex]
		return []string{p.IdentityPairID, p.WorkPairID, p.EmailPairID}
	case oracleProjectOutstanding:
		p := w.Projects[plan.oracleIndex]
		return []string{p.ContextPairID, p.LedgerPairID, p.CorrectionPairID}
	case oracleProjectLeadCurrent:
		p := w.Projects[plan.oracleIndex]
		lead := w.People[p.Lead]
		return []string{p.ContextPairID, lead.IdentityPairID, lead.WorkPairID, lead.EmailPairID, lead.CorrectionPairID}
	case oracleProjectLeadPrevious:
		p := w.Projects[plan.oracleIndex]
		lead := w.People[p.Lead]
		return []string{p.ContextPairID, lead.IdentityPairID, lead.WorkPairID, lead.EmailPairID}
	case oracleTripCurrent, oracleTripChangedLegPrevious, oracleTripChangedLegCurrent, oracleTripLongestCurrent:
		trip := w.Trips[plan.oracleIndex]
		return []string{trip.ContextPairID, trip.PlanPairID, trip.CorrectionPairID}
	default:
		return nil
	}
}

func (w World) oracleOperations(plan QuestionPlan) []string {
	switch plan.oracleKind {
	case oracleContactCurrent:
		return []string{"resolve nickname", "join identity to work context", "apply address correction"}
	case oracleContactPrevious:
		return []string{"resolve nickname", "join identity to work context", "select prior address state"}
	case oracleProjectOutstanding:
		return []string{"resolve project alias", "replace draft with correction", "subtract payment"}
	case oracleProjectLeadCurrent:
		return []string{"resolve project alias", "follow ownership edge", "resolve person identity", "apply address correction"}
	case oracleProjectLeadPrevious:
		return []string{"resolve project alias", "follow ownership edge", "resolve person identity", "select prior address state"}
	case oracleTripCurrent:
		return []string{"resolve trip alias", "apply itinerary correction", "select requested state", "sum or compare legs"}
	case oracleTripChangedLegPrevious:
		return []string{"resolve trip alias", "apply itinerary correction", "select requested state", "sum or compare legs"}
	case oracleTripChangedLegCurrent:
		return []string{"resolve trip alias", "apply itinerary correction", "select requested state", "sum or compare legs"}
	case oracleTripLongestCurrent:
		return []string{"resolve trip alias", "apply itinerary correction", "select requested state", "sum or compare legs"}
	default:
		return nil
	}
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	aa, bb := append([]string(nil), a...), append([]string(nil), b...)
	sort.Strings(aa)
	sort.Strings(bb)
	for i := range aa {
		if aa[i] != bb[i] {
			return false
		}
	}
	return true
}

func renderedConstraintCount(plan QuestionPlan) int {
	question := strings.ToLower(plan.Case.Question)
	n := 0
	for _, constraint := range plan.Constraints {
		if strings.Contains(question, strings.ToLower(constraint)) {
			n++
		}
	}
	return n
}

// resolveWithEvidence is the generation-time counterfactual oracle. It only
// exposes the deterministic answer after every source row in the canonical
// reasoning graph is available. validatePlan removes each row in turn and
// proves that the same answer can no longer be derived, catching decorative
// evidence declarations that inflate apparent depth without adding causality.
func (w World) resolveWithEvidence(plan QuestionPlan, available map[string]bool) (string, bool) {
	has := func(ids ...string) bool {
		for _, id := range ids {
			if !available[id] {
				return false
			}
		}
		return true
	}
	switch plan.oracleKind {
	case oracleContactCurrent:
		p := w.People[plan.oracleIndex]
		if !has(p.IdentityPairID, p.WorkPairID, p.EmailPairID, p.CorrectionPairID) {
			return "", false
		}
		return p.Email, true
	case oracleContactPrevious:
		p := w.People[plan.oracleIndex]
		if !has(p.IdentityPairID, p.WorkPairID, p.EmailPairID) {
			return "", false
		}
		return p.PreviousEmail, true
	case oracleProjectOutstanding:
		p := w.Projects[plan.oracleIndex]
		if !has(p.ContextPairID, p.LedgerPairID, p.CorrectionPairID) {
			return "", false
		}
		return fmt.Sprintf("%d", p.CorrectedCents-p.PaidCents), true
	case oracleProjectLeadCurrent, oracleProjectLeadPrevious:
		p := w.Projects[plan.oracleIndex]
		lead := w.People[p.Lead]
		if !has(p.ContextPairID, lead.IdentityPairID, lead.WorkPairID, lead.EmailPairID) {
			return "", false
		}
		if plan.oracleKind == oracleProjectLeadPrevious {
			return lead.PreviousEmail, true
		}
		if !has(lead.CorrectionPairID) {
			return "", false
		}
		return lead.Email, true
	case oracleTripCurrent, oracleTripChangedLegPrevious, oracleTripChangedLegCurrent, oracleTripLongestCurrent:
		trip := w.Trips[plan.oracleIndex]
		if !has(trip.ContextPairID, trip.PlanPairID, trip.CorrectionPairID) {
			return "", false
		}
		changed := changedLeg(trip) // identified by the correction row
		legs := trip.OldLegDays     // supplied by the original-plan row
		if plan.oracleKind == oracleTripChangedLegPrevious {
			return fmt.Sprintf("%d", legs[changed]), true
		}
		legs[changed] += trip.LegDays[changed] - trip.OldLegDays[changed]
		if plan.oracleKind == oracleTripChangedLegCurrent {
			return fmt.Sprintf("%d", legs[changed]), true
		}
		if plan.oracleKind == oracleTripCurrent {
			return fmt.Sprintf("%d", sum3(legs)), true
		}
		longest := legs[0]
		for _, days := range legs[1:] {
			if days > longest {
				longest = days
			}
		}
		return fmt.Sprintf("%d", longest), true
	default:
		return "", false
	}
}

func (w World) subjectMatches(plan QuestionPlan) int {
	matches := 0
	switch plan.oracleKind {
	case oracleContactCurrent, oracleContactPrevious:
		want := w.People[plan.oracleIndex]
		for _, p := range w.People {
			if p.Nickname == want.Nickname && p.Relation == want.Relation && p.Context == want.Context && p.Employer == want.Employer {
				matches++
			}
		}
	case oracleProjectOutstanding, oracleProjectLeadCurrent, oracleProjectLeadPrevious:
		want := w.Projects[plan.oracleIndex]
		for _, p := range w.Projects {
			if p.Alias == want.Alias && p.Client == want.Client && p.Purpose == want.Purpose && p.Vendor == want.Vendor {
				matches++
			}
		}
	case oracleTripCurrent, oracleTripChangedLegPrevious, oracleTripChangedLegCurrent, oracleTripLongestCurrent:
		want := w.Trips[plan.oracleIndex]
		for _, trip := range w.Trips {
			if trip.Alias == want.Alias && trip.Purpose == want.Purpose && trip.When == want.When && trip.Countries == want.Countries {
				matches++
			}
		}
	}
	return matches
}

// PairWaves partitions the already-spread memory stream into staged seed calls.
// A correction can therefore arrive much later than the identity or draft it
// supersedes, while timestamps preserve the underlying event chronology.
func (w World) PairWaves(n int) [][]protocol.MemoryPair {
	if n < 1 {
		n = 1
	}
	out := make([][]protocol.MemoryPair, n)
	for i, pair := range w.Pairs {
		wave := i * n / len(w.Pairs)
		if wave >= n {
			wave = n - 1
		}
		out[wave] = append(out[wave], pair)
	}
	return out
}

// UnlockWave is the first wave after which every planted needle is available.
func (w World) UnlockWave(plan QuestionPlan, n int) int {
	if n < 1 {
		n = 1
	}
	positions := make(map[string]int, len(w.Pairs))
	for i, pair := range w.Pairs {
		positions[pair.PairID] = i
	}
	last := 0
	for _, id := range plan.RequiredPairIDs {
		wave := positions[id] * n / len(w.Pairs)
		if wave >= n {
			wave = n - 1
		}
		if wave > last {
			last = wave
		}
	}
	return last
}

func changedLeg(trip Trip) int {
	for i := range trip.LegDays {
		if trip.LegDays[i] != trip.OldLegDays[i] {
			return i
		}
	}
	return 0
}

func questionSeed(seed int64) int64 {
	h := fnv.New64a()
	_, _ = fmt.Fprintf(h, "dittobench-v8-world-questions:%d", seed)
	return int64(h.Sum64() & ((1 << 63) - 1))
}

func contentOverlap(question, memory string) float64 {
	q := contentTokens(question)
	if len(q) == 0 {
		return 0
	}
	m := contentTokens(memory)
	hits := 0
	for token := range q {
		if m[token] {
			hits++
		}
	}
	return float64(hits) / float64(len(q))
}

func contentTokens(text string) map[string]bool {
	stop := map[string]bool{
		"what": true, "which": true, "where": true, "when": true, "have": true,
		"with": true, "that": true, "this": true, "from": true, "before": true,
		"after": true, "current": true, "earlier": true, "address": true, "email": true,
		"their": true, "person": true, "contact": true, "update": true, "saved": true,
	}
	out := map[string]bool{}
	for _, token := range strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9')
	}) {
		if len(token) >= 4 && !stop[token] {
			out[token] = true
		}
	}
	return out
}

// SortedEvidenceIDs is a public-safe audit helper used by tests and local
// probes. It exposes identities already present in the public dataset, never an
// answer or a hidden provider value.
func (p QuestionPlan) SortedEvidenceIDs() []string {
	out := append([]string(nil), p.RequiredPairIDs...)
	sort.Strings(out)
	return out
}
