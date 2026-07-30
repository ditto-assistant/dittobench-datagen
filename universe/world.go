// Package universe builds the deterministic fictional world a v8 user lives in.
//
// The world is public and fully reproducible from the benchmark seed. Difficulty
// comes from joining facts across an evolving personal and business history, not
// from hiding a finite answer table. Both tool and memory cases consume the same
// people, aliases, projects, corrections, trips, and messages.
package universe

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/ditto-assistant/dittobench-datagen/protocol"
)

// Person is one seeded person in the user's world. Pair IDs identify the source
// memories an action case must resolve before it can mutate or contact anyone.
type Person struct {
	Name             string
	Nickname         string
	Relation         string
	Employer         string
	Role             string
	City             string
	Email            string
	PreviousEmail    string
	Context          string
	IdentityPairID   string
	WorkPairID       string
	EmailPairID      string
	CorrectionPairID string
	ToolNotePairID   string
}

// Project is a messy business thread: the user uses an informal alias, while
// pasted records use the formal project and vendor names. Outstanding is never
// stored directly; it is derived from the corrected invoice and partial payment.
type Project struct {
	Name             string
	Alias            string
	RecordID         string
	Purpose          string
	Client           string
	Vendor           string
	Lead             int
	OriginalCents    int
	CorrectedCents   int
	PaidCents        int
	OutstandingCents int
	ContextPairID    string
	LedgerPairID     string
	CorrectionPairID string
	ToolNotePairID   string
}

// Trip is a multi-leg event with a later correction. Companion provides the
// natural human link among memories; generated conversations never invent an
// itinerary ID. CurrentDays and PreviousDays are computed from the legs and
// are never written as totals.
type Trip struct {
	Alias            string
	Companion        int
	Purpose          string
	When             string
	Countries        [3]string
	OldLegDays       [3]int
	LegDays          [3]int
	PreviousDays     int
	CurrentDays      int
	ContextPairID    string
	PlanPairID       string
	CorrectionPairID string
}

// World is the shared state used throughout one v8 dataset.
type World struct {
	Seed           int64
	UserCompany    string
	People         []Person
	Projects       []Project
	Trips          []Trip
	StoryArcs      []StoryArc
	Stories        []Story
	Pairs          []protocol.MemoryPair
	BusinessPairID string
	Accent         string
}

var ordinaryFirst = []string{
	"Avery", "Jordan", "Morgan", "Riley", "Samira", "Tomas", "Nadia", "Jules",
	"Dorian", "Mina", "Caleb", "Priya", "Leonie", "Hector", "Sasha", "Inez",
}

var coinedStarts = []string{"Bel", "Cor", "Dra", "Eli", "Fen", "Har", "Ivo", "Kes", "Lor", "Mey", "Nor", "Ori", "Pera", "Qua", "Ryn", "Sel", "Tal", "Vae", "Wen", "Zor"}
var coinedEnds = []string{"adin", "ara", "elis", "enne", "ian", "ira", "orin", "essa", "ovan", "urel", "yra", "eth"}
var familyStarts = []string{"Ash", "Bram", "Cinder", "Dun", "Ever", "Fair", "Glen", "Holl", "Kest", "Mar", "Norr", "Pen", "Quill", "Raven", "Stone", "Tal", "Vale", "West"}
var familyEnds = []string{"brook", "den", "fall", "ford", "mere", "mont", "ridge", "row", "vale", "wick", "wyn", "croft"}
var relations = []string{"friend from university", "cousin", "former roommate", "accountant", "neighbor", "design collaborator", "client contact", "family friend", "old manager", "running-club friend", "bookkeeper", "event producer"}
var roles = []string{"operations lead", "account director", "studio manager", "bookkeeper", "producer", "research lead", "finance manager", "designer", "venue coordinator", "procurement lead", "tax adviser", "editor"}
var cities = []string{"Baltimore", "Providence", "Montreal", "Lisbon", "Osaka", "Durham", "Edinburgh", "Melbourne", "Chicago", "Nairobi", "Portland", "Valencia"}
var contexts = []string{"winter gallery opening", "Harborlight launch", "2025 tax filing", "neighborhood fundraiser", "Northpoint workshop", "spring catalog shoot", "Foundry conference", "Juniper client dinner", "museum benefit", "summer product pilot", "regional planning retreat", "community arts grant"}
var colors = []string{"teal", "indigo", "amber", "emerald", "crimson", "violet", "cobalt", "coral"}
var countryPool = []string{"France", "Spain", "Portugal", "Belgium", "the Netherlands", "Italy", "Switzerland", "Japan", "South Korea", "Kenya", "Tanzania", "Chile", "Argentina", "Norway", "Sweden"}

// Generate returns one deterministic world. scale is a run-size hint (small =
// 1, medium = 2, full = 3); it changes population, never the generation rules.
func Generate(seed int64, scale int) World {
	if scale < 1 {
		scale = 1
	}
	if scale > 3 {
		scale = 3
	}
	r := rand.New(rand.NewSource(worldSeed(seed)))
	w := World{Seed: seed, UserCompany: coinedCompany(r), Accent: colors[r.Intn(len(colors))]}
	// V8's scored memory surface is a world, not a pile of independent cards.
	// The full profile deliberately exceeds the context a harness can solve by
	// dumping every memory into one prompt; successful agents must retrieve and
	// join the right records. Small remains a cheap compatibility smoke path.
	peopleN := []int{6, 14, 28}[scale-1]
	projectN := []int{3, 8, 18}[scale-1]
	tripN := []int{2, 5, 12}[scale-1]
	w.BusinessPairID = protocol.OpaqueCaseID(seed, "world-business-wall", 0)

	seenNames := map[string]bool{}
	seenNicknames := map[string]bool{}
	seenCompanies := map[string]bool{w.UserCompany: true}
	seenEmails := map[string]bool{}
	for i := 0; i < peopleN; i++ {
		name := uniquePersonName(r, seenNames, i)
		nick := uniqueNickname(name, r, seenNicknames)
		employer := uniqueString(r, seenCompanies, coinedCompany)
		previous := uniqueEmail(name, employer, 2*i, false, seenEmails)
		current := uniqueEmail(name, employer, 2*i+1, i%3 == 0, seenEmails)
		if previous == current {
			previous = "old." + previous
		}
		p := Person{
			Name: name, Nickname: nick, Relation: relations[(i+r.Intn(len(relations)))%len(relations)],
			Employer: employer, Role: roles[(i+r.Intn(len(roles)))%len(roles)],
			City: cities[(i+r.Intn(len(cities)))%len(cities)], Email: current,
			PreviousEmail: previous, Context: contexts[(i+r.Intn(len(contexts)))%len(contexts)],
			IdentityPairID:   protocol.OpaqueCaseID(seed, "world-person-identity", i),
			WorkPairID:       protocol.OpaqueCaseID(seed, "world-person-work", i),
			EmailPairID:      protocol.OpaqueCaseID(seed, "world-person-email", i),
			CorrectionPairID: protocol.OpaqueCaseID(seed, "world-person-email-correction", i),
			ToolNotePairID:   protocol.OpaqueCaseID(seed, "world-person-tool-note", i),
		}
		w.People = append(w.People, p)
	}

	seenProjects := map[string]bool{}
	seenProjectAliases := map[string]bool{}
	for i := 0; i < projectN; i++ {
		original := 180000 + r.Intn(2600000)
		deltas := []int{-4, -3, -2, -1, 1, 2, 3, 4}
		corrected := original + deltas[r.Intn(len(deltas))]*12500
		if corrected < 50000 {
			corrected = 50000
		}
		paid := corrected * (2 + r.Intn(5)) / 10
		p := Project{
			Name: uniqueString(r, seenProjects, coinedProject), Alias: uniqueString(r, seenProjectAliases, projectAlias),
			RecordID: "AP-" + strings.ToUpper(protocol.OpaqueCaseID(seed, "world-project-record", i)[:8]), Purpose: projectPurpose(r),
			Client: uniqueString(r, seenCompanies, coinedCompany), Vendor: uniqueString(r, seenCompanies, coinedCompany), Lead: i % len(w.People),
			OriginalCents: original, CorrectedCents: corrected, PaidCents: paid,
			OutstandingCents: corrected - paid,
			ContextPairID:    protocol.OpaqueCaseID(seed, "world-project-context", i),
			LedgerPairID:     protocol.OpaqueCaseID(seed, "world-project-ledger", i),
			CorrectionPairID: protocol.OpaqueCaseID(seed, "world-project-correction", i),
			ToolNotePairID:   protocol.OpaqueCaseID(seed, "world-project-tool-note", i),
		}
		w.Projects = append(w.Projects, p)
	}

	permCountries := r.Perm(len(countryPool))
	seenTripAliases := map[string]bool{}
	for i := 0; i < tripN; i++ {
		countries := [3]string{countryPool[permCountries[(i*3)%len(permCountries)]], countryPool[permCountries[(i*3+1)%len(permCountries)]], countryPool[permCountries[(i*3+2)%len(permCountries)]]}
		oldLegs := [3]int{4 + r.Intn(9), 4 + r.Intn(9), 3 + r.Intn(8)}
		legs := oldLegs
		changed := r.Intn(3)
		legs[changed] += []int{-2, -1, 2, 3}[r.Intn(4)]
		if legs[changed] < 2 {
			legs[changed] = 2
		}
		t := Trip{
			Alias:     uniqueString(r, seenTripAliases, tripAlias),
			Companion: (i*2 + 1) % len(w.People),
			Purpose:   tripPurpose(r), When: tripWhen(r), Countries: countries,
			OldLegDays: oldLegs, LegDays: legs, PreviousDays: sum3(oldLegs), CurrentDays: sum3(legs),
			ContextPairID:    protocol.OpaqueCaseID(seed, "world-trip-context", i),
			PlanPairID:       protocol.OpaqueCaseID(seed, "world-trip-plan", i),
			CorrectionPairID: protocol.OpaqueCaseID(seed, "world-trip-correction", i),
		}
		w.Trips = append(w.Trips, t)
	}

	// Story generation owns an independent seed stream: adding surface variety
	// cannot perturb the people/projects/trips already established above. Each
	// arc contributes three long memories and story-only join/state facts.
	w.StoryArcs, w.Stories = buildStories(seed, scale, w)
	w.Pairs = w.renderPairs(r)
	return w
}

func (w World) renderPairs(r *rand.Rand) []protocol.MemoryPair {
	pairs := make([]protocol.MemoryPair, 0, len(w.People)*4+len(w.Projects)*3+len(w.Trips)*3+2)
	base := time.Date(2024, 1, 8, 9, 0, 0, 0, time.UTC)
	add := func(id, session, prompt, response string) {
		pairs = append(pairs, protocol.MemoryPair{PairID: id, SessionID: session, Timestamp: base.Add(time.Duration(len(pairs)*137) * time.Hour).Format(time.RFC3339), Prompt: prompt, Response: response})
	}
	for i, p := range w.People {
		add(p.IdentityPairID, fmt.Sprintf("people-%02d-a", i), shortLead(r)+p.Name+" is my "+p.Relation+". Everyone around "+p.Context+" calls "+p.Name+" “"+p.Nickname+".”", warmResponse(w.Seed, p.IdentityPairID,
			"Aw, I love the nickname story. I’ll remember who you mean.",
			"That history helps. I’ll remember the nickname and your connection.",
			"I’ve got you — nickname, person, and shared history together."))
		add(p.WorkPairID, fmt.Sprintf("people-%02d-b", i), fmt.Sprintf("For context, %s works as the %s at %s in %s and handled the %s.", p.Name, p.Role, p.Employer, p.City, p.Context), warmResponse(w.Seed, p.WorkPairID,
			"Got you — I’ll connect their work life to the person you know.",
			"That makes sense. I’ll remember their work and personal context together.",
			"I’m with you — I’ll keep both sides of their story connected."))
		// Contact values deliberately do not repeat the person's name, nickname,
		// relationship, or event context. Resolving one requires the identity ->
		// work -> address -> correction chain instead of one lexical lookup.
		add(p.EmailPairID, fmt.Sprintf("people-%02d-c", i), fmt.Sprintf("For the %s at %s, the address I first had on file was %s.", p.Role, p.Employer, p.PreviousEmail), warmResponse(w.Seed, p.EmailPairID,
			"I’ll remember this as the first address, in case the trail matters.",
			"Okay, I’ll keep that as the original address, not necessarily the current one.",
			"Got it — this stays as the earlier contact point if an update comes."))
		add(p.CorrectionPairID, fmt.Sprintf("people-%02d-d", i), fmt.Sprintf("Quick contact cleanup: %s is stale now. Replace it with %s.", p.PreviousEmail, p.Email), warmResponse(w.Seed, p.CorrectionPairID,
			"Good catch — I’ll use the new address and keep the old one as history.",
			"Thanks for the heads-up. I’ll switch over and remember the old one is stale.",
			"All right, I’ve got the change: new one now, old one in the history."))
		// Destructive tool cases target this disposable note, never a fact row
		// required by later memory scoring. Tool execution and memory evaluation
		// therefore share one universe without mutating its canonical evidence.
		add(p.ToolNotePairID, fmt.Sprintf("people-%02d-tool", i), fmt.Sprintf("Contact-maintenance receipt for %s after the %s: I finished reconciling the stale address. This receipt can be deleted once reviewed.", p.Nickname, p.Context), warmResponse(w.Seed, p.ToolNotePairID,
			"Perfect — this can stay temporary while the real contact history stays safe.",
			"Sounds good. I’ll keep this disposable and leave the meaningful history alone.",
			"Got you — this cleanup receipt stays separate from the memories that matter."))
	}

	// One deliberately long, messy paste mixes prose, headings, shorthand, and
	// ledger rows. It represents the wall-of-text messages real business users
	// paste, without increasing every pair to the same artificial size.
	var wall strings.Builder
	fmt.Fprintf(&wall, "Dumping my %s ops notes here because they are scattered across email and a spreadsheet. Please reconcile names, project nicknames, vendor bills, and payments rather than treating every line as a separate company.\n\n", w.UserCompany)
	for i, p := range w.Projects {
		fmt.Fprintf(&wall, "PROJECT %d — %s (we usually call it %q)\nClient: %s. Purpose: %s. Vendor line: %s. Ownership and figures live in their separate project/AP records, not this pasted summary.\n\n", i+1, p.Name, p.Alias, p.Client, p.Purpose, p.Vendor)
	}
	add(w.BusinessPairID, "business-import", "Here is the raw operations paste:\n\n"+wall.String(), "Send it my way — I’ll untangle this without losing how everything connects.")
	for i, p := range w.Projects {
		lead := w.People[p.Lead]
		add(p.ContextPairID, fmt.Sprintf("project-%02d-context", i), fmt.Sprintf("When I say “%s” I mean %s for %s, not the similarly named client work. %q owns it internally; %s is the vendor, and the AP record is %s.", p.Alias, p.Name, p.Client, lead.Nickname, p.Vendor, p.RecordID), warmResponse(w.Seed, p.ContextPairID,
			"Got you — I’ll keep every name and person in the right role.",
			"That distinction makes sense. I’ll keep the shorthand mapped correctly.",
			"I’m following — I won’t mix up the project, client, vendor, or owner."))
		add(p.LedgerPairID, fmt.Sprintf("project-%02d-ledger", i), fmt.Sprintf("Accounts payable record %s: the original invoice was %s; we have already paid %s against it.", p.RecordID, money(p.OriginalCents), money(p.PaidCents)), warmResponse(w.Seed, p.LedgerPairID,
			"I’ve got the starting total and what has gone out so far.",
			"Okay, I’ll keep the original amount and payment together.",
			"I’m with you — I’ll remember the starting bill and what was paid."))
		add(p.CorrectionPairID, fmt.Sprintf("project-%02d-correction", i), fmt.Sprintf("Approval correction for AP record %s: the approved invoice total is %s, replacing the earlier %s figure. The partial payment is unchanged.", p.RecordID, money(p.CorrectedCents), money(p.OriginalCents)), warmResponse(w.Seed, p.CorrectionPairID,
			"Thanks for catching that — I’ll use the approved total now.",
			"Got it. The new total is current, and the payment stays put.",
			"That’s clear — approved figure now, with the same payment."))
		add(p.ToolNotePairID, fmt.Sprintf("project-%02d-tool", i), fmt.Sprintf("Handoff scratchpad for %q at %s: add scheduling details here rather than editing the project identity or ownership record.", p.Alias, p.Client), warmResponse(w.Seed, p.ToolNotePairID,
			"Sounds good — scheduling stays here and the project details stay clean.",
			"Perfect, the moving pieces can live here without muddying the project.",
			"Got you — loose notes here, important project history untouched."))
	}

	for i, t := range w.Trips {
		companion := w.People[t.Companion]
		add(t.ContextPairID, fmt.Sprintf("trip-%02d-context", i), fmt.Sprintf("Our %s was the %s trip in %s — the one through %s, %s, and %s that I planned with %s, my %s.", t.Alias, t.Purpose, t.When, t.Countries[0], t.Countries[1], t.Countries[2], companion.Nickname, companion.Relation), warmResponse(w.Seed, t.ContextPairID,
			"Oh yes, I know the one. I’ll remember who planned it with you.",
			"I’m with you — I’ll keep the route and your travel companion connected.",
			"That trip has a whole story. I’ll remember who helped you dream it up."))
		add(t.PlanPairID, fmt.Sprintf("trip-%02d-plan", i), fmt.Sprintf("When %s and I first mapped that trip out, we had %d days in %s, %d in %s, then %d in %s.", companion.Nickname, t.OldLegDays[0], t.Countries[0], t.OldLegDays[1], t.Countries[1], t.OldLegDays[2], t.Countries[2]), warmResponse(w.Seed, t.PlanPairID,
			"That sounds lovely. I’ll remember this first version if the plan shifts.",
			"Ooh, I can picture it. I’ll keep this first plan for context.",
			"I’ve got the original shape of the trip if anything changes."))
		changed := 0
		for j := range t.LegDays {
			if t.LegDays[j] != t.OldLegDays[j] {
				changed = j
				break
			}
		}
		delta := t.LegDays[changed] - t.OldLegDays[changed]
		change := fmt.Sprintf("adding %d days to our time in %s", delta, t.Countries[changed])
		if delta < 0 {
			delta = -delta
			change = fmt.Sprintf("cutting %d days from our time in %s", delta, t.Countries[changed])
		}
		add(t.CorrectionPairID, fmt.Sprintf("trip-%02d-correction", i), fmt.Sprintf("Quick update on the trip %s and I planned: we’re %s, but leaving the other two stays alone.", companion.Nickname, change), warmResponse(w.Seed, t.CorrectionPairID,
			"Got you — both versions remembered, with this as the current plan.",
			"That makes sense. I’ll keep the old plan as history and use this one now.",
			"Okay, I’ve got the change — everything else stays put."))
	}
	for _, story := range w.Stories {
		prompt, response := story.render(w.Seed)
		add(story.PairID, story.SessionID, prompt, response)
	}
	add(protocol.OpaqueCaseID(w.Seed, "world-preference", 0), "preferences", fmt.Sprintf("For Ditto itself I like a %s accent, but client brand colors should never override my app preference.", w.Accent), fmt.Sprintf("Love it — %s for your Ditto, and client palettes can stay in their own lane.", w.Accent))
	return spreadPairs(pairs)
}

// MemoryCases is the protocol-only view retained for generator callers. The
// construction and validation contract lives in QuestionPlans: generate the
// world, seed its facts, derive a latent plan, then validate the rendered
// question before exposing the case.
func (w World) MemoryCases(count int) []protocol.MemoryCase {
	plans, err := w.QuestionPlans(count)
	if err != nil {
		panic(err)
	}
	out := make([]protocol.MemoryCase, 0, len(plans))
	for _, plan := range plans {
		out = append(out, plan.Case)
	}
	return out
}

// DumpGuardValues returns real, seeded world values used by integrity cases to
// detect broad memory dumping. V8 replaces the discarded legacy-persona guard
// values with these so the guard remains causally grounded without retaining an
// orphaned v7 haystack solely for negative scoring.
func (w World) DumpGuardValues() []string {
	out := make([]string, 0, 12)
	for i := 0; i < len(w.People) && len(out) < 8; i++ {
		out = append(out, w.People[i].Email)
	}
	for i := 0; i < len(w.Projects) && len(out) < 12; i++ {
		out = append(out, w.Projects[i].Alias)
	}
	return out
}

func (w World) emailDistractors(index int, extra string) []string {
	out := []string{extra}
	for j := 1; len(out) < 3; j++ {
		candidate := w.People[(index+j)%len(w.People)].Email
		if !contains(out, candidate) {
			out = append(out, candidate)
		}
	}
	return out
}

func (w World) moneyDistractors(index, correct int) []string {
	out := []string{}
	for j := 1; len(out) < 3 && j <= len(w.Projects)*3; j++ {
		p := w.Projects[(index+j)%len(w.Projects)]
		for _, v := range []int{p.OutstandingCents, p.CorrectedCents, p.OriginalCents} {
			s := fmt.Sprintf("%d", v)
			if v != correct && !contains(out, s) {
				out = append(out, s)
				if len(out) == 3 {
					break
				}
			}
		}
	}
	for delta := 12500; len(out) < 3; delta += 12500 {
		candidate := correct + delta
		if delta%25000 == 0 && correct > delta {
			candidate = correct - delta
		}
		s := fmt.Sprintf("%d", candidate)
		if candidate > 0 && !contains(out, s) {
			out = append(out, s)
		}
	}
	return out
}

func (w World) tripDistractors(index, correct int) []string {
	out := []string{}
	for j := 0; len(out) < 3 && j < len(w.Trips)*6; j++ {
		t := w.Trips[(index+j)%len(w.Trips)]
		for _, v := range []int{t.PreviousDays, t.CurrentDays, t.LegDays[j%3]} {
			s := fmt.Sprintf("%d", v)
			if v != correct && !contains(out, s) {
				out = append(out, s)
				if len(out) == 3 {
					break
				}
			}
		}
	}
	for delta := 1; len(out) < 3; delta++ {
		candidate := correct + delta
		if delta%2 == 0 && correct > delta {
			candidate = correct - delta
		}
		s := fmt.Sprintf("%d", candidate)
		if candidate > 0 && !contains(out, s) {
			out = append(out, s)
		}
	}
	return out
}

func worldSeed(seed int64) int64 {
	h := fnv.New64a()
	_, _ = fmt.Fprintf(h, "dittobench-v8-world:%d", seed)
	return int64(h.Sum64() & ((1 << 63) - 1))
}

func warmResponse(seed int64, pairID string, variants ...string) string {
	if len(variants) == 0 {
		return ""
	}
	h := fnv.New64a()
	_, _ = fmt.Fprintf(h, "dittobench-v8-warm-response:%d:%s", seed, pairID)
	return variants[h.Sum64()%uint64(len(variants))]
}

func uniquePersonName(r *rand.Rand, seen map[string]bool, index int) string {
	for {
		first := ordinaryFirst[r.Intn(len(ordinaryFirst))]
		if index%3 == 0 {
			first = coinedStarts[r.Intn(len(coinedStarts))] + coinedEnds[r.Intn(len(coinedEnds))]
		}
		last := familyStarts[r.Intn(len(familyStarts))] + familyEnds[r.Intn(len(familyEnds))]
		name := first + " " + last
		if !seen[name] {
			seen[name] = true
			return name
		}
	}
}

func nicknameFor(name string, r *rand.Rand) string {
	first := strings.Fields(name)[0]
	if len(first) > 5 {
		return first[:3+randIndex(r, 2)]
	}
	return []string{"Sam", "Jules", "Ren", "Mick", "Dee", "Ro", "Ash", "Kit"}[r.Intn(8)]
}

func uniqueNickname(name string, r *rand.Rand, seen map[string]bool) string {
	parts := strings.Fields(name)
	first, last := parts[0], parts[len(parts)-1]
	for attempt := 0; ; attempt++ {
		candidate := nicknameFor(name, r)
		if attempt > 0 {
			width := 3 + attempt
			if width <= len(first) {
				candidate = first[:width]
			} else {
				lastWidth := width - len(first)
				if lastWidth > len(last) {
					lastWidth = len(last)
				}
				candidate = first + " " + last[:lastWidth]
			}
		}
		if !seen[candidate] {
			seen[candidate] = true
			return candidate
		}
	}
}

func emailFor(name, employer string, index int, business bool) string {
	// The address itself must not encode the person's name or employer: doing so
	// turns an intended identity join into substring matching. Opaque but
	// plausible mailbox handles mirror real users whose contact address bears no
	// resemblance to the name saved in an address book.
	h := fnv.New64a()
	_, _ = fmt.Fprintf(h, "%s|%s|%d|%t", name, employer, index, business)
	sum := h.Sum64()
	left := []string{"quiet", "silver", "maple", "moss", "copper", "lunar", "cedar", "paper", "amber", "violet", "river", "north"}
	right := []string{"orchard", "lantern", "window", "harbor", "sparrow", "studio", "meadow", "signal", "bridge", "thimble", "garden", "atlas"}
	local := fmt.Sprintf("%s.%s%d", left[sum%uint64(len(left))], right[(sum/17)%uint64(len(right))], 10+(sum/257)%90)
	domain := []string{"gmail.com", "outlook.com", "proton.me", "fastmail.com", "hey.com"}[(sum/4099)%5]
	if business {
		domain = []string{"post.team", "contact.work", "mail.studio"}[(sum/8191)%3]
	}
	return local + "@" + domain
}

func uniqueEmail(name, employer string, index int, business bool, seen map[string]bool) string {
	for attempt := 0; ; attempt++ {
		candidate := emailFor(name, employer, index+attempt*1009, business)
		if !seen[candidate] {
			seen[candidate] = true
			return candidate
		}
	}
}

func coinedCompany(r *rand.Rand) string {
	return familyStarts[r.Intn(len(familyStarts))] + familyEnds[r.Intn(len(familyEnds))] + " " + []string{"Studio", "Works", "Partners", "Labs", "Collective", "Group"}[r.Intn(6)]
}

func uniqueString(r *rand.Rand, seen map[string]bool, generate func(*rand.Rand) string) string {
	for {
		candidate := generate(r)
		if !seen[candidate] {
			seen[candidate] = true
			return candidate
		}
	}
}

func coinedProject(r *rand.Rand) string {
	return []string{"Project", "Program", "Initiative", "Campaign"}[r.Intn(4)] + " " + familyStarts[r.Intn(len(familyStarts))] + coinedEnds[r.Intn(len(coinedEnds))]
}
func projectAlias(r *rand.Rand) string {
	first := []string{"Northpoint", "Lantern", "Bluebird", "Foundry", "Harborlight", "Orchard", "Kestrel", "Juniper"}[r.Intn(8)]
	second := []string{"ledger", "room", "line", "file", "track", "brief", "cycle", "desk"}[r.Intn(8)]
	return strings.ToLower(first + " " + second)
}
func projectPurpose(r *rand.Rand) string {
	return []string{"retail launch", "annual report", "museum installation", "client migration", "brand research", "regional workshop", "fundraising campaign", "supplier transition"}[r.Intn(8)]
}
func tripAlias(r *rand.Rand) string {
	first := []string{"museum", "food", "field", "festival", "archive", "research", "family", "studio"}[r.Intn(8)]
	second := []string{"loop", "trip", "week", "run", "visit", "circuit", "route", "tour"}[r.Intn(8)]
	return strings.ToLower(first + " " + second)
}
func tripPurpose(r *rand.Rand) string {
	return []string{"food research", "museum research", "wildlife fieldwork", "family reunion", "music festival", "supplier meetings", "architecture tour", "archive project"}[r.Intn(8)]
}
func tripWhen(r *rand.Rand) string {
	return []string{"last spring", "last autumn", "the year before the move", "summer 2025", "the winter after the launch", "our 2024 break"}[r.Intn(6)]
}

func shortLead(r *rand.Rand) string {
	return []string{"FYI: ", "Tiny note — ", "Remember: ", "One thing: ", ""}[r.Intn(5)]
}
func sum3(v [3]int) int      { return v[0] + v[1] + v[2] }
func money(cents int) string { return fmt.Sprintf("$%d.%02d", cents/100, cents%100) }
func randIndex(r *rand.Rand, n int) int {
	if n <= 0 {
		return 0
	}
	return r.Intn(n)
}
func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}
func slug(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// spreadPairs keeps semantic time in each pair's timestamp while separating
// records that were generated together in the seed payload. A benchmark-aware
// harness should not receive a person's identity, job, old address, and
// correction as four neighboring rows it can solve with a fixed local window.
func spreadPairs(in []protocol.MemoryPair) []protocol.MemoryPair {
	if len(in) < 3 {
		return in
	}
	stride := len(in)/2 + 1
	for gcd(stride, len(in)) != 1 {
		stride++
	}
	out := make([]protocol.MemoryPair, len(in))
	for i, pair := range in {
		out[(i*stride)%len(in)] = pair
	}
	return out
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// SortedPairIDs is a test/debug helper proving the world never emits duplicate
// memory identities.
func (w World) SortedPairIDs() []string {
	ids := make([]string, 0, len(w.Pairs))
	for _, p := range w.Pairs {
		ids = append(ids, p.PairID)
	}
	sort.Strings(ids)
	return ids
}
