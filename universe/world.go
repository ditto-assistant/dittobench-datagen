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

	"github.com/ditto-assistant/dittobench-datagen/internal/humandata"
	"github.com/ditto-assistant/dittobench-datagen/protocol"
)

// Person is one seeded person in the user's world. Pair IDs identify the source
// memories an action case must resolve before it can mutate or contact anyone.
type Person struct {
	Name             string
	Nickname         string
	Relation         string
	Employer         string
	PreviousEmployer string
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
	seenCompanies := map[string]bool{w.UserCompany: true}
	seenEmails := map[string]bool{}
	for i := 0; i < peopleN; i++ {
		name := uniquePersonName(r, seenNames, w.People, i)
		given := strings.Fields(name)[0]
		nick := humandata.PreferredName(given, r)
		previousEmployer := uniqueString(r, seenCompanies, coinedCompany)
		employer := uniqueString(r, seenCompanies, coinedCompany)
		previous := uniqueEmail(name, previousEmployer, 2*i, true, seenEmails)
		current := uniqueEmail(name, employer, 2*i+1, true, seenEmails)
		if previous == current {
			previous = "old." + previous
		}
		p := Person{
			Name: name, Nickname: nick, Relation: relations[(i+r.Intn(len(relations)))%len(relations)],
			Employer: employer, PreviousEmployer: previousEmployer,
			Role: roles[(i+r.Intn(len(roles)))%len(roles)],
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
		projectName, projectAlias := uniqueProjectIdentity(r, seenProjects, seenProjectAliases)
		original := 180000 + r.Intn(2600000)
		deltas := []int{-4, -3, -2, -1, 1, 2, 3, 4}
		corrected := original + deltas[r.Intn(len(deltas))]*12500
		if corrected < 50000 {
			corrected = 50000
		}
		paid := corrected * (2 + r.Intn(5)) / 10
		p := Project{
			Name: projectName, Alias: projectAlias,
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
			"Aw, I love that nickname. I’ll remember them.",
			"That history helps — I know who you mean.",
			"I’ve got the person and nickname together."))
		add(p.WorkPairID, fmt.Sprintf("people-%02d-b", i), fmt.Sprintf("%s used to work at %s. These days they’re the %s at %s in %s; that’s how they ended up handling the %s.", p.Name, p.PreviousEmployer, p.Role, p.Employer, p.City, p.Context), warmResponse(w.Seed, p.WorkPairID,
			"Got you — work and person connected.",
			"I’ll remember both sides of their life.",
			"I’ll keep their whole story together."))
		// Contact values deliberately do not repeat the person's name, nickname,
		// relationship, or event context. Resolving one requires the identity ->
		// work -> address -> correction chain instead of one lexical lookup.
		add(p.EmailPairID, fmt.Sprintf("people-%02d-c", i), fmt.Sprintf("Back when they were at %s, the work address I had saved was %s.", p.PreviousEmployer, p.PreviousEmail), warmResponse(w.Seed, p.EmailPairID,
			"I’ll keep this as their first address.",
			"I’ll remember this as the original address, not necessarily the current one.",
			"I’ve got the earlier contact point."))
		add(p.CorrectionPairID, fmt.Sprintf("people-%02d-d", i), fmt.Sprintf("Since the move to %s, %s is stale. Their current work email is %s.", p.Employer, p.PreviousEmail, p.Email), warmResponse(w.Seed, p.CorrectionPairID,
			"Good catch — I’ll use the new address.",
			"Thanks — switched, with the old one kept.",
			"I’ve got the new one now; the old one stays in history."))
		// Destructive tool cases target this disposable note, never a fact row
		// required by later memory scoring. Tool execution and memory evaluation
		// therefore share one universe without mutating its canonical evidence.
		add(p.ToolNotePairID, fmt.Sprintf("people-%02d-tool", i), fmt.Sprintf("Contact-maintenance receipt for %s after the %s: I finished reconciling the stale address. This receipt can be deleted once reviewed.", p.Nickname, p.Context), warmResponse(w.Seed, p.ToolNotePairID,
			"Perfect — temporary note, real history safe.",
			"Sounds good — disposable note, history intact.",
			"Got you — receipt separate from what matters."))
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
		add(p.ContextPairID, fmt.Sprintf("project-%02d-context", i), fmt.Sprintf("When I say “%s” I mean %s for %s, not the similarly named client work. %s owns it internally; %s is the vendor, and the AP record is %s.", p.Alias, p.Name, p.Client, lead.Name, p.Vendor, p.RecordID), warmResponse(w.Seed, p.ContextPairID,
			"Got you — every name in the right role.",
			"I’ll keep the shorthand mapped correctly.",
			"I won’t mix up the project or people."))
		add(p.LedgerPairID, fmt.Sprintf("project-%02d-ledger", i), fmt.Sprintf("Accounts payable record %s: the original invoice was %s; we have already paid %s against it.", p.RecordID, money(p.OriginalCents), money(p.PaidCents)), warmResponse(w.Seed, p.LedgerPairID,
			"I’ve got the total and payments so far.",
			"I’ve got the original amount and payment together.",
			"I’ll remember the bill and what was paid."))
		add(p.CorrectionPairID, fmt.Sprintf("project-%02d-correction", i), fmt.Sprintf("Approval correction for AP record %s: the approved invoice total is %s, replacing the earlier %s figure. The partial payment is unchanged.", p.RecordID, money(p.CorrectedCents), money(p.OriginalCents)), warmResponse(w.Seed, p.CorrectionPairID,
			"Thanks — I’ll use the approved total.",
			"I’ve got the new total; the payment stays put.",
			"I’ve got the approved figure now, with the same payment."))
		add(p.ToolNotePairID, fmt.Sprintf("project-%02d-tool", i), fmt.Sprintf("Handoff scratchpad for %q at %s: add scheduling details here rather than editing the project identity or ownership record.", p.Alias, p.Client), warmResponse(w.Seed, p.ToolNotePairID,
			"Sounds good — scheduling stays separate.",
			"Perfect — moving pieces stay in this note.",
			"Got you — scratchpad separate from history."))
	}

	for i, t := range w.Trips {
		companion := w.People[t.Companion]
		add(t.ContextPairID, fmt.Sprintf("trip-%02d-context", i), fmt.Sprintf("Our %s was the %s trip in %s — the one through %s, %s, and %s that I planned with %s, my %s.", t.Alias, t.Purpose, t.When, t.Countries[0], t.Countries[1], t.Countries[2], companion.Nickname, companion.Relation), warmResponse(w.Seed, t.ContextPairID,
			"Oh yes — I’ll remember who planned it with you.",
			"I’ll keep the route and companion connected.",
			"I’ll remember who helped dream it up."))
		add(t.PlanPairID, fmt.Sprintf("trip-%02d-plan", i), fmt.Sprintf("When %s and I first mapped that trip out, we had %d days in %s, %d in %s, then %d in %s.", companion.Nickname, t.OldLegDays[0], t.Countries[0], t.OldLegDays[1], t.Countries[1], t.OldLegDays[2], t.Countries[2]), warmResponse(w.Seed, t.PlanPairID,
			"Lovely — I’ll remember this first version.",
			"Ooh, I can picture it. First plan saved.",
			"I’ve got the trip’s original shape."))
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
			"Got you — this is the current plan.",
			"Old plan kept; this one is current.",
			"I’ve got the change; everything else stays."))
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

func uniquePersonName(r *rand.Rand, seen map[string]bool, prior []Person, index int) string {
	for {
		first := humandata.GivenName(r, index)
		// Deliberately repeat ordinary first names in full worlds. A real user
		// knows multiple Sams, Marias, and Alexes; the event/employer graph must
		// disambiguate them instead of a globally unique preferred name doing it.
		if index >= 5 && index%6 == 5 {
			first = strings.Fields(prior[index-5].Name)[0]
		}
		last := humandata.Surname(r, index)
		name := first + " " + last
		if !seen[name] {
			seen[name] = true
			return name
		}
	}
}

func emailFor(name, employer string, index int, business bool) string {
	h := fnv.New64a()
	_, _ = fmt.Fprintf(h, "%s|%s|%d|%t", name, employer, index, business)
	sum := h.Sum64()
	parts := strings.Fields(strings.ToLower(name))
	first, last := slug(parts[0]), slug(parts[len(parts)-1])
	locals := []string{first, first + "." + last, first[:1] + last, first + "." + last[:1]}
	local := locals[sum%uint64(len(locals))]
	domain := []string{"gmail.com", "outlook.com", "proton.me", "fastmail.com", "hey.com"}[(sum/4099)%5]
	if business {
		domain = companyDomain(employer)
	}
	return local + "@" + domain
}

func companyDomain(employer string) string {
	words := strings.Fields(employer)
	if len(words) > 1 {
		words = words[:len(words)-1]
	}
	joined := slug(strings.Join(words, ""))
	if joined == "" {
		joined = "company"
	}
	return joined + ".com"
}

func uniqueEmail(name, employer string, index int, business bool, seen map[string]bool) string {
	base := emailFor(name, employer, index, business)
	for attempt := 0; ; attempt++ {
		candidate := base
		if attempt > 0 {
			parts := strings.SplitN(base, "@", 2)
			candidate = fmt.Sprintf("%s%d@%s", parts[0], attempt+1, parts[1])
		}
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

type projectNameFamily struct {
	formal string
	alias  string
}

// These are close enough to be confused in real conversation while remaining
// different words. A project identity draws both names from one family instead
// of independently combining unrelated fantasy words.
var projectNameFamilies = []projectNameFamily{
	{"Bluehaven", "Bluebird"}, {"Bluelake", "Bluebell"}, {"Harborline", "Harborlight"},
	{"Harborview", "Harborside"}, {"Northstar", "Northpoint"}, {"Northfield", "Northgate"},
	{"Juniper Grove", "Juniper Lane"}, {"Juniper House", "Juniper Hill"},
	{"Lantern House", "Lantern Room"}, {"Lantern Field", "Lantern Lane"},
	{"Orchard Row", "Orchard Road"}, {"Orchard House", "Orchard Hall"},
	{"Kestrel Field", "Kestrel Flight"}, {"Kestrel House", "Kestrel Hill"},
	{"Foundry Lane", "Foundry Line"}, {"Foundry House", "Foundry Hall"},
	{"Cedar Grove", "Cedar Gate"}, {"Cedar House", "Cedar Hill"},
	{"Riverstone", "Riverside"}, {"Riverlight", "Riverline"},
	{"Westbridge", "Westbrook"}, {"Westfield", "Westford"},
	{"Greenhaven", "Greenhouse"}, {"Greenfield", "Greenlight"},
}

func uniqueProjectIdentity(r *rand.Rand, seenNames, seenAliases map[string]bool) (string, string) {
	prefixes := []string{"Project", "Program", "Initiative", "Campaign"}
	nouns := []string{"brief", "ledger", "workstream", "rollout", "plan", "review", "track", "file"}
	for {
		family := projectNameFamilies[r.Intn(len(projectNameFamilies))]
		name := prefixes[r.Intn(len(prefixes))] + " " + family.formal
		alias := strings.ToLower(family.alias + " " + nouns[r.Intn(len(nouns))])
		if !seenNames[name] && !seenAliases[alias] {
			seenNames[name] = true
			seenAliases[alias] = true
			return name, alias
		}
	}
}

func projectNamesAreRelated(name, alias string) bool {
	formalWords := strings.Fields(name)
	if len(formalWords) < 2 || len(strings.Fields(alias)) == 0 {
		return false
	}
	formal := slug(strings.Join(formalWords[1:], ""))
	aliasHead := slug(strings.Fields(alias)[0])
	limit := len(formal)
	if len(aliasHead) < limit {
		limit = len(aliasHead)
	}
	shared := 0
	for shared < limit && formal[shared] == aliasHead[shared] {
		shared++
	}
	return shared >= 4
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
