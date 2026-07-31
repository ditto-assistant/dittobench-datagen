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
}

// Project is a messy business thread: the user uses an informal alias, while
// pasted records use the formal project and vendor names. Outstanding is never
// stored directly; it is derived from the corrected invoice and partial payment.
type Project struct {
	Name             string
	Alias            string
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
}

// Trip is a multi-leg event with a later correction. CurrentDays and
// PreviousDays are computed from the legs and are never written as totals.
type Trip struct {
	Alias            string
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
	Seed        int64
	UserCompany string
	People      []Person
	Projects    []Project
	Trips       []Trip
	Pairs       []protocol.MemoryPair
	Accent      string
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
	peopleN := []int{5, 9, 14}[scale-1]
	projectN := []int{2, 4, 7}[scale-1]
	tripN := []int{1, 3, 5}[scale-1]

	seenNames := map[string]bool{}
	seenNicknames := map[string]bool{}
	seenCompanies := map[string]bool{w.UserCompany: true}
	for i := 0; i < peopleN; i++ {
		name := uniquePersonName(r, seenNames, i)
		nick := uniqueNickname(name, r, seenNicknames)
		employer := uniqueString(r, seenCompanies, coinedCompany)
		previous := emailFor(name, employer, i, false)
		current := emailFor(name, employer, i, i%3 == 0)
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
			Name: uniqueString(r, seenProjects, coinedProject), Alias: uniqueString(r, seenProjectAliases, projectAlias), Purpose: projectPurpose(r),
			Client: uniqueString(r, seenCompanies, coinedCompany), Vendor: uniqueString(r, seenCompanies, coinedCompany), Lead: i % len(w.People),
			OriginalCents: original, CorrectedCents: corrected, PaidCents: paid,
			OutstandingCents: corrected - paid,
			ContextPairID:    protocol.OpaqueCaseID(seed, "world-project-context", i),
			LedgerPairID:     protocol.OpaqueCaseID(seed, "world-project-ledger", i),
			CorrectionPairID: protocol.OpaqueCaseID(seed, "world-project-correction", i),
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
			Alias: uniqueString(r, seenTripAliases, tripAlias), Purpose: tripPurpose(r), When: tripWhen(r), Countries: countries,
			OldLegDays: oldLegs, LegDays: legs, PreviousDays: sum3(oldLegs), CurrentDays: sum3(legs),
			ContextPairID:    protocol.OpaqueCaseID(seed, "world-trip-context", i),
			PlanPairID:       protocol.OpaqueCaseID(seed, "world-trip-plan", i),
			CorrectionPairID: protocol.OpaqueCaseID(seed, "world-trip-correction", i),
		}
		w.Trips = append(w.Trips, t)
	}

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
		add(p.IdentityPairID, fmt.Sprintf("people-%02d-a", i), shortLead(r)+p.Name+" is my "+p.Relation+". Everyone around "+p.Context+" calls "+p.Name+" “"+p.Nickname+".”", "Got it.")
		add(p.WorkPairID, fmt.Sprintf("people-%02d-b", i), fmt.Sprintf("For context, %s works as the %s at %s in %s and handled the %s.", p.Name, p.Role, p.Employer, p.City, p.Context), "I’ll keep that context connected.")
		add(p.EmailPairID, fmt.Sprintf("people-%02d-c", i), fmt.Sprintf("The address I first saved for %s was %s.", p.Nickname, p.PreviousEmail), "Saved.")
		add(p.CorrectionPairID, fmt.Sprintf("people-%02d-d", i), fmt.Sprintf("Small correction after the %s: %s no longer uses that address. Their current email is %s.", p.Context, p.Name, p.Email), "Updated — I’ll use the current address unless you ask about the old one.")
	}

	// One deliberately long, messy paste mixes prose, headings, shorthand, and
	// ledger rows. It represents the wall-of-text messages real business users
	// paste, without increasing every pair to the same artificial size.
	var wall strings.Builder
	fmt.Fprintf(&wall, "Dumping my %s ops notes here because they are scattered across email and a spreadsheet. Please reconcile names, project nicknames, vendor bills, and payments rather than treating every line as a separate company.\n\n", w.UserCompany)
	for i, p := range w.Projects {
		lead := w.People[p.Lead]
		fmt.Fprintf(&wall, "PROJECT %d — %s (we usually call it %q)\nClient: %s. Purpose: %s. Internal lead: %s / %q, the %s at %s. Vendor line: %s. Initial invoice: %s. Partial payment already sent: %s.\n\n", i+1, p.Name, p.Alias, p.Client, p.Purpose, lead.Name, lead.Nickname, lead.Role, lead.Employer, p.Vendor, money(p.OriginalCents), money(p.PaidCents))
	}
	add(protocol.OpaqueCaseID(w.Seed, "world-business-wall", 0), "business-import", "Here is the raw operations paste:\n\n"+wall.String(), "Imported. I’ll treat the aliases, people, clients, vendors, and payments as linked context.")
	for i, p := range w.Projects {
		lead := w.People[p.Lead]
		add(p.ContextPairID, fmt.Sprintf("project-%02d-context", i), fmt.Sprintf("When I say “%s” I mean %s for %s, not the similarly named client work. %s (%s) owns it internally.", p.Alias, p.Name, p.Client, lead.Nickname, lead.Name), "Understood.")
		add(p.LedgerPairID, fmt.Sprintf("project-%02d-ledger", i), fmt.Sprintf("Accounts payable note: %s invoiced %s for %s; we have already paid %s against it.", p.Vendor, w.UserCompany, money(p.OriginalCents), money(p.PaidCents)), "Recorded.")
		add(p.CorrectionPairID, fmt.Sprintf("project-%02d-correction", i), fmt.Sprintf("Correction for %s / %s: the approved invoice total is %s, replacing the earlier %s figure. The partial payment is unchanged.", p.Alias, p.Vendor, money(p.CorrectedCents), money(p.OriginalCents)), "Corrected the total while retaining the payment.")
	}

	for i, t := range w.Trips {
		add(t.ContextPairID, fmt.Sprintf("trip-%02d-context", i), fmt.Sprintf("Our %s was the %s trip in %s — the one through %s, %s, and %s.", t.Alias, t.Purpose, t.When, t.Countries[0], t.Countries[1], t.Countries[2]), "I know which trip you mean.")
		add(t.PlanPairID, fmt.Sprintf("trip-%02d-plan", i), fmt.Sprintf("Original itinerary: %d days in %s, %d in %s, then %d in %s.", t.OldLegDays[0], t.Countries[0], t.OldLegDays[1], t.Countries[1], t.OldLegDays[2], t.Countries[2]), "Saved the original legs.")
		changed := 0
		for j := range t.LegDays {
			if t.LegDays[j] != t.OldLegDays[j] {
				changed = j
				break
			}
		}
		add(t.CorrectionPairID, fmt.Sprintf("trip-%02d-correction", i), fmt.Sprintf("We changed the %s leg of %s from %d days to %d. The other country stays are unchanged.", t.Countries[changed], t.Alias, t.OldLegDays[changed], t.LegDays[changed]), "Updated that leg only.")
	}
	add(protocol.OpaqueCaseID(w.Seed, "world-preference", 0), "preferences", fmt.Sprintf("For Ditto itself I like a %s accent, but client brand colors should never override my app preference.", w.Accent), "I’ll keep your personal app preference separate from client branding.")
	return pairs
}

// MemoryCases returns deterministic outcome questions derived from the world.
// Each question joins at least four facts and carries three plausible near-miss
// answers. The caller chooses how many replace simpler cases in the fixed budget.
func (w World) MemoryCases(count int) []protocol.MemoryCase {
	if count <= 0 {
		return nil
	}
	out := make([]protocol.MemoryCase, 0, count)
	for i := 0; i < count; i++ {
		id := protocol.OpaqueCaseID(w.Seed, "world-memory", i)
		switch i % 6 {
		case 0:
			p := w.People[i%len(w.People)]
			out = append(out, protocol.MemoryCase{BenchVersion: protocol.BenchVersionV8, ID: id, QuestionID: id, QuestionType: "world-contact-current", Question: fmt.Sprintf("What is the current email for %s — my %s in %s who handled the %s?", p.Nickname, p.Relation, p.City, p.Context), ExpectedAnswer: p.Email, AnswerKind: protocol.AnswerValue, DistractorAnswers: w.emailDistractors(i, p.PreviousEmail)})
		case 1:
			p := w.People[(i+2)%len(w.People)]
			out = append(out, protocol.MemoryCase{BenchVersion: protocol.BenchVersionV8, ID: id, QuestionID: id, QuestionType: "world-contact-previous", Question: fmt.Sprintf("Before the correction after the %s, which email did I have for %s, the %s at %s?", p.Context, p.Nickname, p.Role, p.Employer), ExpectedAnswer: p.PreviousEmail, AnswerKind: protocol.AnswerValue, DistractorAnswers: w.emailDistractors(i+2, p.Email)})
		case 2:
			p := w.Projects[i%len(w.Projects)]
			out = append(out, protocol.MemoryCase{BenchVersion: protocol.BenchVersionV8, ID: id, QuestionID: id, QuestionType: "world-business-reconciliation", Question: fmt.Sprintf("After the corrected invoice and the payment already sent, how many cents do we still owe %s for %q, the %s work for %s?", p.Vendor, p.Alias, p.Purpose, p.Client), ExpectedAnswer: fmt.Sprintf("%d", p.OutstandingCents), AnswerKind: protocol.AnswerNumber, DistractorAnswers: w.moneyDistractors(i, p.OutstandingCents)})
		case 3:
			p := w.Projects[(i+1)%len(w.Projects)]
			lead := w.People[p.Lead]
			out = append(out, protocol.MemoryCase{BenchVersion: protocol.BenchVersionV8, ID: id, QuestionID: id, QuestionType: "world-business-contact", Question: fmt.Sprintf("Which current email should I use for the person who owns %q internally — the %s project for %s, not the client contact?", p.Alias, p.Purpose, p.Client), ExpectedAnswer: lead.Email, AnswerKind: protocol.AnswerValue, DistractorAnswers: w.emailDistractors(p.Lead, lead.PreviousEmail)})
		case 4:
			t := w.Trips[i%len(w.Trips)]
			out = append(out, protocol.MemoryCase{BenchVersion: protocol.BenchVersionV8, ID: id, QuestionID: id, QuestionType: "world-trip-current", Question: fmt.Sprintf("After the itinerary correction, how many days is %s — our %s trip in %s across all three countries?", t.Alias, t.Purpose, t.When), ExpectedAnswer: fmt.Sprintf("%d days", t.CurrentDays), AnswerKind: protocol.AnswerDuration, DistractorAnswers: w.tripDistractors(i, t.CurrentDays)})
		default:
			t := w.Trips[(i+1)%len(w.Trips)]
			out = append(out, protocol.MemoryCase{BenchVersion: protocol.BenchVersionV8, ID: id, QuestionID: id, QuestionType: "world-trip-previous-state", Question: fmt.Sprintf("What was the total planned length of %s before we changed the %s itinerary?", t.Alias, t.Purpose), ExpectedAnswer: fmt.Sprintf("%d days", t.PreviousDays), AnswerKind: protocol.AnswerDuration, DistractorAnswers: w.tripDistractors(i+1, t.PreviousDays)})
		}
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
			s := fmt.Sprintf("%d days", v)
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
		s := fmt.Sprintf("%d days", candidate)
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
	parts := strings.Fields(strings.ToLower(name))
	local := slug(parts[0]) + "." + slug(parts[len(parts)-1])
	domain := []string{"gmail.com", "outlook.com", "proton.me"}[index%3]
	if business {
		domain = slug(employer) + ".co"
	}
	return local + "@" + domain
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
	return strings.ToLower([]string{"Northpoint", "Lantern", "Bluebird", "Foundry", "Harborlight", "Orchard", "Kestrel", "Juniper"}[r.Intn(8)])
}
func projectPurpose(r *rand.Rand) string {
	return []string{"retail launch", "annual report", "museum installation", "client migration", "brand research", "regional workshop", "fundraising campaign", "supplier transition"}[r.Intn(8)]
}
func tripAlias(r *rand.Rand) string {
	return strings.ToLower([]string{"museum loop", "food trip", "field week", "festival run", "archive visit", "research circuit", "family trip", "studio tour"}[r.Intn(8)])
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
