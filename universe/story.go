package universe

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"strings"

	"github.com/ditto-assistant/dittobench-datagen/protocol"
)

// StoryKind is the broad real-life context represented by a long conversation
// memory. Domains provide the finer-grained topic mix inside each kind.
type StoryKind string

const (
	StoryPersonal StoryKind = "personal"
	StoryBusiness StoryKind = "business"
)

// StorySection is the structured state from which story prose is compiled. The
// generator owns meaning first and surface text second: tests and the oracle can
// reason about the beginning, middle, and end without parsing generated prose.
type StorySection struct {
	Summary string   `json:"summary"`
	Events  []string `json:"events"`
}

// StoryFact is one exact, seed-derived fact rendered somewhere inside the story.
// Renderings are semantically equivalent sentence shapes; Value is retained for
// evidence-placement and exclusivity tests, not emitted as separate metadata.
type StoryFact struct {
	Key        string   `json:"key"`
	Value      string   `json:"value"`
	Renderings []string `json:"renderings"`
}

// Story is a deterministic story-to-prose program. A story compiles to one
// realistic user/agent MemoryPair, while the structured object keeps its state,
// themes, lessons, and planted facts auditable before prose exists.
type Story struct {
	ID             string       `json:"id"`
	PairID         string       `json:"pair_id"`
	SessionID      string       `json:"session_id"`
	Kind           StoryKind    `json:"kind"`
	Domain         string       `json:"domain"`
	Title          string       `json:"title"`
	Beginning      StorySection `json:"beginning"`
	Middle         StorySection `json:"middle"`
	End            StorySection `json:"end"`
	Themes         []string     `json:"themes"`
	LessonsLearned []string     `json:"lessons_learned"`
	Facts          []StoryFact  `json:"facts"`
	TargetBytes    int          `json:"target_bytes"`
}

// StoryArc joins a personal origin, a business decision, and a later outcome.
// None of the bridge/decision identities or story-only values are rendered in a
// short memory. Questions enter through personal context and must recover both
// hidden joins before they can apply the final state.
type StoryArc struct {
	ID                  string
	PersonIndex         int
	ProjectIndex        int
	TripIndex           int
	PreferredName       string
	BridgeCode          string
	DecisionCode        string
	OriginalContact     string
	CurrentContact      string
	BaseBudgetCents     int
	BudgetDeltaCents    int
	PaidCents           int
	UnexpectedCostCents int
	CreditCents         int
	CurrentBalanceCents int
	Lesson              string
	LessonAcceptAny     []string
	StoryPairIDs        [3]string
}

type lessonSet struct {
	canonical string
	accept    []string
}

var storyLessons = []lessonSet{
	{canonical: "confirm ownership before promising", accept: []string{"verify ownership before committing", "confirm the owner before making promises"}},
	{canonical: "separate draft numbers from approvals", accept: []string{"keep draft figures separate from approved numbers", "distinguish drafts from approvals"}},
	{canonical: "name the handoff before the deadline", accept: []string{"identify the handoff before the due date", "assign the handoff before deadline"}},
	{canonical: "verify the current channel before sending", accept: []string{"check the latest contact route before sending", "confirm the current channel before you send"}},
	{canonical: "reconcile personal context with work records", accept: []string{"connect personal context to the work record", "align personal context with business records"}},
	{canonical: "record corrections where decisions happen", accept: []string{"keep corrections with the decision", "attach corrections to decision records"}},
	{canonical: "trace commitments back to their source", accept: []string{"follow commitments to their source", "track promises back to the source"}},
	{canonical: "keep the client story and ledger aligned", accept: []string{"align the client narrative with the ledger", "keep client context consistent with the ledger"}},
}

var personalDomains = []string{
	"family and friendship", "community volunteering", "creative practice", "travel and place",
	"health and routines", "learning and mentorship", "neighborhood life", "personal milestones",
}

var businessDomains = []string{
	"client operations", "finance and approvals", "hiring and teams", "partnerships",
	"product launches", "research programs", "events and production", "fundraising",
}

var storyNameWords = []string{
	"Aven", "Brindle", "Caro", "Della", "Eris", "Faro", "Galen", "Hali",
	"Indra", "Jori", "Kiva", "Lumi", "Meri", "Niko", "Oren", "Pavi",
}

func storyArcCount(scale int) int {
	switch scale {
	case 1:
		return 1
	case 2:
		return 3
	default:
		return 8
	}
}

func buildStories(seed int64, scale int, w World) ([]StoryArc, []Story) {
	r := rand.New(rand.NewSource(storySeed(seed)))
	arcN := storyArcCount(scale)
	people := r.Perm(len(w.People))
	projects := r.Perm(len(w.Projects))
	trips := r.Perm(len(w.Trips))

	seenNames := map[string]bool{}
	seenEmails := map[string]bool{}
	seenAmounts := map[int]bool{}
	for _, p := range w.People {
		seenNames[strings.ToLower(p.Nickname)] = true
		seenNames[strings.ToLower(p.Name)] = true
		seenNames[strings.ToLower(strings.Fields(p.Name)[0])] = true
		seenEmails[p.Email] = true
		seenEmails[p.PreviousEmail] = true
	}
	for _, project := range w.Projects {
		seenAmounts[project.OriginalCents] = true
		seenAmounts[project.CorrectedCents] = true
		seenAmounts[project.PaidCents] = true
		seenAmounts[project.OutstandingCents] = true
	}

	arcs := make([]StoryArc, 0, arcN)
	stories := make([]Story, 0, 3*arcN)
	lessonOffset := r.Intn(len(storyLessons))
	for i := 0; i < arcN; i++ {
		personIndex := people[i%len(people)]
		projectIndex := projects[i%len(projects)]
		tripIndex := trips[i%len(trips)]
		person := w.People[personIndex]
		project := w.Projects[projectIndex]
		trip := w.Trips[tripIndex]

		preferred := uniqueStoryName(r, person.Nickname, seenNames)
		originalContact := uniqueEmail(person.Name, project.Name, 7000+i*2, true, seenEmails)
		currentContact := uniqueEmail(person.Name, project.Name, 7001+i*2, true, seenEmails)
		base := uniqueStoryAmount(r, seenAmounts, 1_200_000, 4_800_000, 12_500)
		deltaMagnitude := uniqueStoryAmount(r, seenAmounts, 50_000, 450_000, 12_500)
		delta := deltaMagnitude
		if i%3 == 1 {
			delta = -deltaMagnitude
		}
		paid := uniqueStoryAmount(r, seenAmounts, 175_000, 900_000, 12_500)
		cost := uniqueStoryAmount(r, seenAmounts, 50_000, 375_000, 12_500)
		credit := uniqueStoryAmount(r, seenAmounts, 25_000, 250_000, 12_500)
		if delta-cost+credit == 0 {
			credit = uniqueStoryAmount(r, seenAmounts, credit+12_500, 250_000, 12_500)
		}
		balance := base + delta - paid - cost + credit
		if balance <= 250_000 {
			base += 1_000_000
			balance += 1_000_000
		}

		bridge := "REF-" + strings.ToUpper(protocol.OpaqueCaseID(seed, "world-story-bridge", i)[:7])
		decision := "DEC-" + strings.ToUpper(protocol.OpaqueCaseID(seed, "world-story-decision", i)[:7])
		lesson := storyLessons[(lessonOffset+i)%len(storyLessons)]
		arc := StoryArc{
			ID:                  protocol.OpaqueCaseID(seed, "world-story-arc", i),
			PersonIndex:         personIndex,
			ProjectIndex:        projectIndex,
			TripIndex:           tripIndex,
			PreferredName:       preferred,
			BridgeCode:          bridge,
			DecisionCode:        decision,
			OriginalContact:     originalContact,
			CurrentContact:      currentContact,
			BaseBudgetCents:     base,
			BudgetDeltaCents:    delta,
			PaidCents:           paid,
			UnexpectedCostCents: cost,
			CreditCents:         credit,
			CurrentBalanceCents: balance,
			Lesson:              lesson.canonical,
			LessonAcceptAny:     append([]string(nil), lesson.accept...),
		}

		for part := 0; part < 3; part++ {
			pairID := protocol.OpaqueCaseID(seed, "world-story-pair", i*3+part)
			arc.StoryPairIDs[part] = pairID
		}
		arcStories := storiesForArc(seed, i, arc, person, project, trip, r)
		stories = append(stories, arcStories...)
		arcs = append(arcs, arc)
	}
	return arcs, stories
}

func storiesForArc(seed int64, index int, arc StoryArc, person Person, project Project, trip Trip, r *rand.Rand) []Story {
	personalDomain := personalDomains[(index+r.Intn(len(personalDomains)))%len(personalDomains)]
	businessDomain := businessDomains[(index+r.Intn(len(businessDomains)))%len(businessDomains)]
	origin := Story{
		ID: protocol.OpaqueCaseID(seed, "world-story-origin", index), PairID: arc.StoryPairIDs[0],
		SessionID: fmt.Sprintf("story-%02d-origin", index), Kind: StoryPersonal, Domain: personalDomain,
		Title: "how a personal conversation became a work thread",
		Beginning: StorySection{
			Summary: fmt.Sprintf("I reconnected with %s, my %s, around the %s while we were still talking about the %s. I think of this as part of my %s history.", person.Name, person.Relation, person.Context, trip.Alias, personalDomain),
			Events: []string{
				fmt.Sprintf("The conversation moved between %s, ordinary life in %s, and what had changed since the %s.", trip.Purpose, person.City, person.Context),
				fmt.Sprintf("I kept using %s, because that is the nickname I knew from the %s, even when the discussion drifted toward work.", person.Nickname, person.Context),
				"Nothing about the conversation began as a formal intake; it was the sort of loose personal history I would normally tell Ditto so the later references still make sense.",
			},
		},
		Middle: StorySection{
			Summary: "Halfway through, the personal recollection turned into a specific introduction, and I made a private bridge note so I would not flatten the relationship into a sales lead.",
			Events: []string{
				"There were several adjacent ideas, but only one became a real follow-up; I wrote down enough surrounding context to distinguish it from the other possibilities.",
				"The name used among friends was not quite the one requested for this working context, which is exactly the kind of correction that gets lost when notes are reduced to keywords.",
				"I wanted the later business records to remain connected to this personal origin without copying the entire conversation into every project note.",
				fmt.Sprintf("We talked through what %s had meant personally before deciding whether any part of it belonged in a professional follow-up.", trip.Alias),
				fmt.Sprintf("The fact that %s lived in %s and had handled the %s helped me separate this recollection from other introductions with similar names.", person.Name, person.City, person.Context),
				"I made the bridge note only after the conversation shifted; before that point it was simply a story about trust, timing, and how our shared context had changed.",
			},
		},
		End: StorySection{
			Summary: "By the end I had a follow-up path, but I also wanted the friendship, the trip, and the eventual work decision to remain distinct parts of one history.",
			Events: []string{
				"I left the conversation feeling that the context mattered as much as the task, because otherwise a future reminder would point at the right project for the wrong reason.",
				"The practical next step belonged in work notes, while the reason I trusted the introduction belonged here in the personal story.",
				fmt.Sprintf("I still associated the origin with %s and the %s, not with a pipeline stage or a client label.", trip.Alias, person.Context),
				"Writing the long version down let me preserve that distinction without relying on the shorthand I would use in a meeting.",
			},
		},
		Themes:         []string{"identity across contexts", "personal trust", "careful handoffs"},
		LessonsLearned: []string{"preserve why an introduction mattered"},
		Facts: []StoryFact{
			storyFact("preferred-name", arc.PreferredName,
				"For anything connected to that introduction, %s was the working name they explicitly asked me to use.",
				"They paused to clarify that the work-side name should be %s, even though friends used something else.",
				"The naming correction was simple but important: in this thread they wanted to be called %s."),
			storyFact("bridge-code", arc.BridgeCode,
				"I filed the private bridge between the personal conversation and any later work under %s.",
				"The cross-context reference I wrote down for this introduction was %s.",
				"To keep the introduction traceable without repeating the story, I marked its bridge as %s."),
		},
		TargetBytes: storyTargetBytes(r),
	}

	decision := Story{
		ID: protocol.OpaqueCaseID(seed, "world-story-decision", index), PairID: arc.StoryPairIDs[1],
		SessionID: fmt.Sprintf("story-%02d-decision", index), Kind: StoryBusiness, Domain: businessDomain,
		Title: "turning an informal introduction into a governed decision",
		Beginning: StorySection{
			Summary: fmt.Sprintf("A later planning session turned the introduction into %s for %s, the %s engagement we usually call %q. Operationally it sat inside our %s work.", project.Name, project.Client, project.Purpose, project.Alias, businessDomain),
			Events: []string{
				fmt.Sprintf("The team kept the formal project name, the client %s, and the vendor %s separate because similar names were already circulating.", project.Client, project.Vendor),
				"What looked like one decision was really a bundle of ownership, contact, budget, and payment details recorded at different moments in the meeting.",
				"I wrote this as a narrative rather than a table because the caveats explained which numbers were drafts and which commitments had actually been made.",
			},
		},
		Middle: StorySection{
			Summary: "The middle of the meeting established the governed record: it tied the informal referral to a decision identity, an initial approved envelope, an already-sent payment, and a reviewer channel.",
			Events: []string{
				"Several people repeated approximate totals, so I kept the authoritative amounts beside the reason for each one instead of copying every number into the summary line.",
				"The reviewer route was intentionally separate from the ordinary address book; it existed for this work thread and could later be corrected without changing the personal contact card.",
				"We agreed that later adjustments would point back to the decision identity, not merely repeat the project nickname, because nicknames collide too easily.",
				fmt.Sprintf("The %s purpose mattered because a superficially similar engagement for another client had a different owner, vendor, and approval trail.", project.Purpose),
				fmt.Sprintf("Someone used %q as shorthand while finance used %s, so I preserved both names but treated neither as the durable decision identity.", project.Alias, project.Name),
				fmt.Sprintf("The discussion separated the client %s from the vendor %s; mixing those roles would have sent the review request and the payment record in opposite directions.", project.Client, project.Vendor),
			},
		},
		End: StorySection{
			Summary: "The meeting ended with a usable baseline, while explicitly leaving room for corrections and follow-up costs that had not yet arrived.",
			Events: []string{
				"I did not treat the first envelope as a final balance; it was only the state from which later payments, credits, and surprises would have to be reconciled.",
				"That distinction became the main operational takeaway: preserve the chain of decisions instead of trusting the latest isolated number.",
				fmt.Sprintf("The closing recap kept %s, %s, and %s in their separate roles rather than merging them into one project label.", project.Client, project.Vendor, project.Name),
				"Everyone left knowing which pieces were authoritative now and which were merely the inputs to a later reconciliation.",
			},
		},
		Themes:         []string{"governed decisions", "financial state", "contextual identity"},
		LessonsLearned: []string{"keep draft and approved state distinguishable"},
		Facts: []StoryFact{
			storyFact("bridge-code", arc.BridgeCode,
				"The intake record explicitly traced this opportunity back to personal bridge %s.",
				"The personal-to-work reference attached to the intake was %s.",
				"This was the work item connected to bridge marker %s, not one of the similarly named leads."),
			storyFact("decision-code", arc.DecisionCode,
				"Once approved for tracking, the work received decision identity %s.",
				"The durable decision reference created in that meeting was %s.",
				"All subsequent corrections were supposed to point to %s, the decision identifier."),
			storyFact("original-contact", arc.OriginalContact,
				"The reviewer route originally written into the decision was %s.",
				"At that point the work-only contact channel on file was %s.",
				"The first review address recorded for the engagement was %s."),
			storyFact("base-budget", money(arc.BaseBudgetCents),
				"The initial approved working envelope was %s, before any later correction.",
				"We established %s as the starting budget, explicitly not the final reconciled balance.",
				"The governed baseline for the work came to %s."),
			storyFact("paid", money(arc.PaidCents),
				"Of that envelope, %s had already gone out as the first payment.",
				"The ledger note said the team had already paid %s against the work.",
				"A payment of %s was complete before the follow-up meeting."),
		},
		TargetBytes: storyTargetBytes(r),
	}

	deltaDirection := "increased"
	if arc.BudgetDeltaCents < 0 {
		deltaDirection = "reduced"
	}
	deltaValue := money(absInt(arc.BudgetDeltaCents))
	outcomeKind := StoryBusiness
	if index%2 == 1 {
		outcomeKind = StoryPersonal
	}
	outcome := Story{
		ID: protocol.OpaqueCaseID(seed, "world-story-outcome", index), PairID: arc.StoryPairIDs[2],
		SessionID: fmt.Sprintf("story-%02d-outcome", index), Kind: outcomeKind, Domain: "mixed personal and work follow-up",
		Title: "the correction, the consequence, and what I carried forward",
		Beginning: StorySection{
			Summary: "The follow-up arrived after the original meeting, when a few apparently minor changes forced me to reconstruct which promise, person, and financial state actually belonged together.",
			Events: []string{
				"The update came through ordinary conversation rather than a clean system notification, so the surrounding chronology mattered to interpreting it correctly.",
				"I first assumed the old channel and the headline budget were still safe, then realized both belonged to the earlier state of the decision.",
				"There were also a new cost and a separate credit; treating either as part of the budget correction would have produced the wrong balance.",
			},
		},
		Middle: StorySection{
			Summary: "Reconstructing the outcome required applying a budget change, preserving the payment, subtracting a new cost, adding a credit, and replacing the work-only contact route.",
			Events: []string{
				"Each adjustment came with a different explanation, which is why I do not want them collapsed into one unlabelled total or detached from the decision that caused them.",
				"The contact correction was operational rather than personal: it changed where this work should go without rewriting the person's ordinary identity.",
				"I checked the sequence twice because using the draft state, forgetting the credit, or subtracting the payment again would each yield a plausible but wrong answer.",
				"The budget change modified the prior envelope, while the later cost and credit remained separate ledger movements; combining all three as one replacement figure would erase the history.",
				"I also kept the stale route out of the final summary so that a future message would not accidentally revive it merely because it appeared earlier in the narrative.",
				"What finally made the account coherent was following the references in order, then applying each state transition according to what it represented rather than where it appeared in the prose.",
			},
		},
		End: StorySection{
			Summary: "Once the state was reconciled, the useful result was not just a number or address but a repeatable lesson about how personal introductions become accountable work.",
			Events: []string{
				"I want Ditto to retain that lesson alongside the outcome so a later question can recover both what happened and the reasoning habit it changed.",
				"The final state should therefore be read as the end of this particular arc, not as a global correction to every memory about the person or project.",
				"The personal origin still mattered, but the outcome was now governed by the explicit corrections rather than by what anyone first expected.",
				"I finished the note only after checking that the conclusion, the contact route, and the financial state all referred to the same chain of events.",
			},
		},
		Themes:         []string{"correction over time", "cross-domain reasoning", "accountable follow-through"},
		LessonsLearned: append([]string{arc.Lesson}, arc.LessonAcceptAny...),
		Facts: []StoryFact{
			storyFact("decision-code", arc.DecisionCode,
				"Every correction in this follow-up belonged to decision %s.",
				"The follow-up was attached to decision reference %s.",
				"The state being revised here was the one identified by %s."),
			storyFact("budget-delta", deltaValue,
				fmt.Sprintf("The approved envelope was %s by %%s; this was a delta, not a replacement total.", deltaDirection),
				fmt.Sprintf("Finance %s the existing budget by %%s rather than issuing a new standalone amount.", deltaDirection),
				fmt.Sprintf("The correction changed the earlier envelope by %%s, and the direction was %s.", deltaDirection)),
			storyFact("unexpected-cost", money(arc.UnexpectedCostCents),
				"A newly confirmed cost of %s also had to be deducted after the budget change.",
				"The follow-up introduced a separate %s expense that reduced what remained.",
				"Beyond the payment already made, another cost of %s became due."),
			storyFact("credit", money(arc.CreditCents),
				"A distinct credit of %s flowed back into the available amount.",
				"The reconciliation also added a %s credit, separate from the budget correction.",
				"There was a return credit worth %s that needed to be added rather than subtracted."),
			storyFact("current-contact", arc.CurrentContact,
				"The work-only reviewer route is now %s; the address in the original decision is stale.",
				"For this decision, the corrected current channel became %s.",
				"The follow-up replaced the earlier review address with %s."),
			storyFact("lesson", arc.Lesson,
				"The lesson I wrote down verbatim was: %s.",
				"I summarized the durable lesson as %s.",
				"The phrase I carried into future work was %s."),
		},
		TargetBytes: storyTargetBytes(r),
	}

	return []Story{origin, decision, outcome}
}

func storyFact(key, value string, renderings ...string) StoryFact {
	return StoryFact{Key: key, Value: value, Renderings: append([]string(nil), renderings...)}
}

func (s Story) render(seed int64) (string, string) {
	r := rand.New(rand.NewSource(storyRenderSeed(seed, s.ID)))
	target := s.TargetBytes
	if target < 1_800 {
		target = 1_800
	}
	var b strings.Builder
	b.WriteString([]string{
		"I want to put this whole story somewhere I can find it later, because the short version loses why the details connect.",
		"This is one of those longer memories where the chronology matters more than any single headline, so I am writing the full version down.",
		"Let me tell you the complete version rather than reducing it to a task note; several personal and practical details became important later.",
	}[r.Intn(3)])
	b.WriteString(" ")
	b.WriteString(s.Beginning.Summary)
	beginSeq, middleSeq, endSeq := 0, 0, 0
	fillStoryUntil(&b, s, s.Beginning, "beginning", target*22/100, &beginSeq)

	facts := append([]StoryFact(nil), s.Facts...)
	r.Shuffle(len(facts), func(i, j int) { facts[i], facts[j] = facts[j], facts[i] })
	for i, fact := range facts {
		point := 27
		if len(facts) > 1 {
			point += i * 42 / (len(facts) - 1)
		}
		fillStoryUntil(&b, s, s.Middle, "middle", target*point/100, &middleSeq)
		b.WriteString("\n\n")
		rendering := fact.Renderings[r.Intn(len(fact.Renderings))]
		b.WriteString(fmt.Sprintf(rendering, fact.Value))
	}
	fillStoryUntil(&b, s, s.Middle, "middle", target*75/100, &middleSeq)
	b.WriteString("\n\n")
	b.WriteString(s.End.Summary)
	fillStoryUntil(&b, s, s.End, "end", target, &endSeq)

	response := []string{
		"Thank you for telling me the complete version. I’ll keep the chronology, the personal context, the work state, and the later correction connected without flattening them into one generic note.",
		"I’ve got the arc of it. I’ll preserve how the beginning led to the decision and how the later update changed its meaning, while keeping personal and business context distinct.",
		"Understood. I’ll remember this as a connected story with an earlier state, a turning point, and a corrected outcome rather than as a bag of disconnected keywords.",
	}[r.Intn(3)]
	return b.String(), response
}

func (w World) validateStoryPlan(plan QuestionPlan) error {
	if !isStoryOracle(plan.oracleKind) {
		return nil
	}
	if len(plan.RequiredPairIDs) != 3 {
		return fmt.Errorf("story plan requires %d memories, want exactly 3", len(plan.RequiredPairIDs))
	}
	arc := w.StoryArcs[plan.oracleIndex]
	question := strings.ToLower(plan.Case.Question)
	for _, hidden := range []string{arc.BridgeCode, arc.DecisionCode, arc.OriginalContact, arc.CurrentContact} {
		if strings.Contains(question, strings.ToLower(hidden)) {
			return fmt.Errorf("story question leaks hidden join/state value %q", hidden)
		}
	}
	pairs := make(map[string]protocol.MemoryPair, len(w.Pairs))
	storyPair := make(map[string]bool, len(w.Stories))
	stories := make(map[string]Story, len(w.Stories))
	for _, pair := range w.Pairs {
		pairs[pair.PairID] = pair
	}
	for _, story := range w.Stories {
		storyPair[story.PairID] = true
		stories[story.PairID] = story
	}
	kinds := map[StoryKind]bool{}
	for _, pairID := range plan.RequiredPairIDs {
		story, ok := stories[pairID]
		if !ok {
			return fmt.Errorf("story evidence %s is not a story memory", pairID)
		}
		pair := pairs[pairID]
		if len(pair.Prompt)+len(pair.Response) < 1_800 {
			return fmt.Errorf("story evidence %s is only %d bytes", pairID, len(pair.Prompt)+len(pair.Response))
		}
		kinds[story.Kind] = true
		if len(story.Facts) < 2 {
			return fmt.Errorf("story evidence %s has only %d planted facts", pairID, len(story.Facts))
		}
		for _, fact := range story.Facts {
			pos := strings.Index(pair.Prompt, fact.Value)
			if pos < 0 {
				return fmt.Errorf("story evidence %s omits fact %s=%q", pairID, fact.Key, fact.Value)
			}
			ratio := float64(pos) / float64(len(pair.Prompt))
			if ratio < 0.15 || ratio > 0.85 {
				return fmt.Errorf("story evidence %s places fact %s at %.2f, outside interior", pairID, fact.Key, ratio)
			}
			if strings.Contains(pair.Response, fact.Value) {
				return fmt.Errorf("agent response duplicates story fact %s=%q", fact.Key, fact.Value)
			}
			for _, other := range w.Pairs {
				if storyPair[other.PairID] {
					continue
				}
				if strings.Contains(other.Prompt+" "+other.Response, fact.Value) {
					return fmt.Errorf("story-only fact %s=%q leaks into short memory %s", fact.Key, fact.Value, other.PairID)
				}
			}
		}
	}
	if !kinds[StoryPersonal] || !kinds[StoryBusiness] {
		return fmt.Errorf("story plan does not cross personal and business memories")
	}
	return nil
}

func fillStoryUntil(b *strings.Builder, story Story, section StorySection, phase string, target int, seq *int) {
	for b.Len() < target {
		b.WriteString("\n\n")
		b.WriteString(storyParagraph(story, section, phase, *seq))
		*seq++
	}
}

func storyParagraph(story Story, section StorySection, phase string, seq int) string {
	offset := storySurfaceIndex(story.ID+":"+phase, len(section.Events))
	event := section.Events[(seq+offset)%len(section.Events)]
	theme := story.Themes[(seq+storySurfaceIndex(story.Domain+phase, len(story.Themes)))%len(story.Themes)]
	patterns := []string{
		"What I remember most clearly: %s %s", "The detail that makes the sequence make sense is this: %s %s",
		"During the %s of the story: %s %s", "I nearly left this out: %s %s",
		"A terse note would miss the following: %s %s", "Another memory from the same stretch: %s %s",
	}
	reflections := []string{
		"That is why the memory belongs under %s rather than a generic reminder.",
		"It changed how I thought about %s, even though the practical consequence came later.",
		"For me, that detail is part of %s and not disposable scene-setting.",
		"Without it, a later retelling would lose the connection to %s.",
		"I associate that moment with %s more than with the headline event.",
	}
	reflection := fmt.Sprintf(reflections[(seq+storySurfaceIndex(story.Title+phase, len(reflections)))%len(reflections)], theme)
	pattern := patterns[(seq+storySurfaceIndex(story.ID+story.Title+phase, len(patterns)))%len(patterns)]
	if strings.Count(pattern, "%s") == 3 {
		return fmt.Sprintf(pattern, phase, event, reflection)
	}
	return fmt.Sprintf(pattern, event, reflection)
}

func storySurfaceIndex(value string, n int) int {
	if n <= 0 {
		return 0
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(value))
	return int(h.Sum64() % uint64(n))
}

func storyTargetBytes(r *rand.Rand) int {
	// Three independent draws form a deterministic bell-shaped distribution
	// around 2.8KB while retaining a realistic 1.8-3.9KB tail. Story count, not
	// length, is fixed per profile so difficulty stays comparable across seeds.
	return 1_800 + r.Intn(700) + r.Intn(700) + r.Intn(700)
}

func uniqueStoryName(r *rand.Rand, base string, seen map[string]bool) string {
	for attempt := 0; ; attempt++ {
		candidate := storyNameWords[(r.Intn(len(storyNameWords))+attempt)%len(storyNameWords)]
		if attempt >= len(storyNameWords) {
			candidate = fmt.Sprintf("%s %d", base, attempt-len(storyNameWords)+2)
		}
		key := strings.ToLower(candidate)
		if !seen[key] {
			seen[key] = true
			return candidate
		}
	}
}

func uniqueStoryAmount(r *rand.Rand, seen map[int]bool, min, spread, quantum int) int {
	for attempt := 0; ; attempt++ {
		value := min + (r.Intn(spread/quantum)+attempt)*quantum
		if !seen[value] {
			seen[value] = true
			return value
		}
	}
}

func storySeed(seed int64) int64 {
	h := fnv.New64a()
	_, _ = fmt.Fprintf(h, "dittobench-v8-stories:%d", seed)
	return int64(h.Sum64() & ((1 << 63) - 1))
}

func storyRenderSeed(seed int64, id string) int64 {
	h := fnv.New64a()
	_, _ = fmt.Fprintf(h, "dittobench-v8-story-render:%d:%s", seed, id)
	return int64(h.Sum64() & ((1 << 63) - 1))
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
