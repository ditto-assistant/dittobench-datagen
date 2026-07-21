package gen

import "github.com/ditto-assistant/dittobench-datagen/protocol"

// bench_version 6 CONTENT-VARIETY expansion. v6 triples every v5 content-variety
// pool that feeds the high-entropy memory families (multi-hop, temporal-depth,
// declarative-preference, abstention). The v5 pools stay BYTE-IDENTICAL below —
// v6 only ADDS entries (the extra* slices here) and the generator selects the
// larger pool when benchVersion >= V6. Because a bench version is an immutable
// contract, v5's already-pinned bytes and scores are untouched: v6 is a new
// rotated surface (RotateSeedForVersion) drawing from ~3x the surface area, so a
// harness that overfit the v5 pools to a static cue list loses that purchase.
//
// The additions are real-world-grounded generic CATEGORY words only (nameable
// possessions, generic life-admin service domains, relationship kinds, evolving
// personal attributes, uncommon given names for anonymous intermediaries). No
// brand names, no real people, no private data: every graded VALUE is still a
// per-seed coined token (contamination-proof), exactly as in v5.
//
// Counts (v5 -> v6): relativeNames 40->120, relationPairs 14->42, leafNouns
// 20->60, tdNouns 20->60, confabSpecs 18->54, declPrefDomains 10->30.

// extraRelativeNames — v6 additions (80).
var extraRelativeNames = []string{
	"Aarav", "Adaeze", "Aleksei", "Alessia", "Amias", "Anders", "Aneta", "Arash",
	"Beatrix", "Beren", "Bram", "Caius", "Caspian", "Celestine", "Chiara", "Cosima",
	"Dagny", "Dalia", "Delphine", "Dmitri", "Eero", "Efua", "Elke", "Enzo",
	"Esben", "Fabiola", "Fenna", "Ferran", "Florian", "Giselle", "Gunnar", "Hedda",
	"Henrik", "Ilya", "Imara", "Imran", "Indira", "Ines", "Ishan", "Jarrah",
	"Juno", "Kaito", "Kalinda", "Katya", "Kestrel", "Kiran", "Lior", "Livia",
	"Ludo", "Magnus", "Marisol", "Matthias", "Meira", "Nikolai", "Nuria", "Odile",
	"Oona", "Orla", "Osei", "Pilar", "Ramiro", "Renske", "Rhian", "Roald",
	"Saanvi", "Sanne", "Selin", "Senna", "Solveig", "Suvi", "Tadeo", "Talia",
	"Tariq", "Thandiwe", "Tinashe", "Tobias", "Xiomara", "Yohan", "Zephyr", "Zuri",
}

// extraLeafNouns — v6 additions (40).
var extraLeafNouns = []string{
	"aquarium", "kayak", "drone", "3d printer", "mechanical keyboard", "synthesizer",
	"telescope", "beehive", "pottery kiln", "home server", "chess set", "espresso machine",
	"record player", "film camera", "sewing machine", "gaming pc", "ham radio", "weather station",
	"canoe", "motorcycle", "houseboat", "paddleboard", "snowmobile", "go-kart",
	"pickup truck", "teardrop trailer", "ferret", "goldfish", "hamster", "tortoise",
	"bearded dragon", "guinea pig", "rabbit", "axolotl", "cockatiel", "bonsai tree",
	"koi pond", "herb garden", "ukulele", "quilt",
}

// extraTdNouns — v6 additions (40).
var extraTdNouns = []string{
	"daily coffee beans", "UI theme", "keyboard layout", "Linux distro",
	"standing-desk height", "lock-screen photo", "gym class", "home-screen layout",
	"default browser", "go-to smoothie", "preferred workout time", "favorite houseplant",
	"current book", "weekend brunch spot", "gym locker number", "bedtime routine",
	"morning alarm tone", "daily step goal", "preferred meeting room", "go-to karaoke song",
	"favorite hiking trail", "default map app", "preferred cloud storage", "VPN provider",
	"streaming service", "meditation app", "calendar app", "web host",
	"mobile carrier", "meal-kit service", "dry cleaner", "auto mechanic",
	"house-cleaning service", "pet sitter", "terminal color scheme", "code editor theme",
	"shell prompt style", "go-to breakfast", "favorite sandwich order", "preferred running route",
}

// extraRelationPairs — v6 additions (28).
var extraRelationPairs = []relationPair{
	{target: "partner", decoy: "hairdresser"},
	{target: "uncle", decoy: "electrician"},
	{target: "godfather", decoy: "optometrist"},
	{target: "niece", decoy: "contractor"},
	{target: "stepbrother", decoy: "mechanic"},
	{target: "stepsister", decoy: "babysitter"},
	{target: "grandmother", decoy: "gardener"},
	{target: "grandfather", decoy: "locksmith"},
	{target: "father-in-law", decoy: "masseuse"},
	{target: "mother-in-law", decoy: "real estate agent"},
	{target: "old classmate", decoy: "dermatologist"},
	{target: "teammate", decoy: "financial advisor"},
	{target: "bandmate", decoy: "notary"},
	{target: "study partner", decoy: "bank teller"},
	{target: "college roommate", decoy: "receptionist"},
	{target: "running partner", decoy: "bookkeeper"},
	{target: "former colleague", decoy: "house painter"},
	{target: "second cousin", decoy: "surveyor"},
	{target: "great-aunt", decoy: "chimney sweep"},
	{target: "foster brother", decoy: "window cleaner"},
	{target: "former roommate", decoy: "dry cleaner"},
	{target: "camp friend", decoy: "caterer"},
	{target: "school friend", decoy: "valet"},
	{target: "work friend", decoy: "bellhop"},
	{target: "travel buddy", decoy: "florist"},
	{target: "penpal", decoy: "courier"},
	{target: "old flame", decoy: "concierge"},
	{target: "goddaughter", decoy: "upholsterer"},
}

// extraConfabSpecs — v6 additions (36).
var extraConfabSpecs = []confabSpec{
	{seededNoun: "telescope", askedNoun: "binoculars"},
	{seededNoun: "film camera", askedNoun: "instant camera"},
	{seededNoun: "synthesizer", askedNoun: "drum machine"},
	{seededNoun: "terrarium", askedNoun: "fish tank"},
	{seededNoun: "beehive", askedNoun: "ant farm"},
	{seededNoun: "pottery wheel", askedNoun: "kiln"},
	{seededNoun: "sewing machine", askedNoun: "knitting machine"},
	{seededNoun: "mechanical keyboard", askedNoun: "gaming mouse"},
	{seededNoun: "drone", askedNoun: "model airplane"},
	{seededNoun: "3d printer", askedNoun: "laser cutter"},
	{seededNoun: "e-bike", askedNoun: "electric scooter"},
	{seededNoun: "aquarium", askedNoun: "birdcage"},
	{seededNoun: "hammock", askedNoun: "porch swing"},
	{seededNoun: "treadmill", askedNoun: "rowing machine"},
	{seededNoun: "gaming console", askedNoun: "arcade cabinet"},
	{seededNoun: "mandolin", askedNoun: "fiddle"},
	{seededNoun: "grand piano", askedNoun: "upright piano"},
	{seededNoun: "drum kit", askedNoun: "xylophone"},
	{seededNoun: "surfboard", askedNoun: "paddleboard"},
	{seededNoun: "snowboard", askedNoun: "skis"},
	{seededNoun: "climbing harness", askedNoun: "climbing rope"},
	{seededNoun: "hot tub", askedNoun: "plunge pool"},
	{seededNoun: "fire pit", askedNoun: "chiminea"},
	{seededNoun: "bonsai tree", askedNoun: "cactus"},
	{seededNoun: "herb garden", askedNoun: "berry bushes"},
	{seededNoun: "koi pond", askedNoun: "birdbath"},
	{seededNoun: "treehouse", askedNoun: "garden shed"},
	{seededNoun: "wine rack", askedNoun: "liquor cabinet"},
	{seededNoun: "pool table", askedNoun: "foosball table"},
	{seededNoun: "ping pong table", askedNoun: "air hockey table"},
	{seededNoun: "saxophone", askedNoun: "clarinet"},
	{seededNoun: "cello", askedNoun: "double bass"},
	{seededNoun: "harmonica", askedNoun: "accordion"},
	{seededNoun: "jukebox", askedNoun: "boombox"},
	{seededNoun: "projector", askedNoun: "flat-screen tv"},
	{seededNoun: "slow cooker", askedNoun: "pressure cooker"},
}

// extraDeclPrefDomains — v6 additions (20).
var extraDeclPrefDomains = []declPrefDomain{
	{
		write: []string{
			"For meal kits I'm on %s now — never %s, half the ingredients showed up wilted.",
			"I get my dinner kits from %s these days. I cancelled %s for good.",
			"Standing preference: my meal-kit box comes from %s, not %s.",
		},
		read:     []string{"Which meal-kit service do I use?", "Where do I get my dinner kits now?", "What's my go-to meal-kit box again?"},
		behavior: []string{"Line up next week's dinners for me — which meal-kit service are you using?", "Add a couple of recipe boxes to my order; what service are you on?", "Set up a meal-kit delivery for the week — where are you placing it?"},
	},
	{
		write: []string{
			"For cloud storage I keep everything on %s now — never %s, they throttled my uploads.",
			"All my files live on %s these days. I moved off %s entirely.",
			"Standing choice: my cloud storage is %s, not %s.",
		},
		read:     []string{"Which cloud storage service do I use?", "Where do I keep my files these days?", "What's my go-to cloud storage again?"},
		behavior: []string{"Back up these documents for me — which cloud storage are you using?", "Share that folder from my drive; what service is it on?", "Save this to my cloud — where are you putting it?"},
	},
	{
		write: []string{
			"For passwords I've switched to %s — never %s, it kept locking me out.",
			"I store all my logins in %s now. I ditched %s completely.",
			"Standing preference: my password manager is %s, not %s.",
		},
		read:     []string{"Which password manager do I use?", "Where do I keep my logins now?", "What's my go-to password manager again?"},
		behavior: []string{"Pull up my login for that site — which password manager are you checking?", "Save this new password for me; what manager are you using?", "Generate a strong password and store it — where are you saving it?"},
	},
	{
		write: []string{
			"For my VPN I use %s now — never %s, the connection dropped constantly.",
			"I route everything through %s these days. I cancelled %s.",
			"Standing choice: my VPN is %s, not %s.",
		},
		read:     []string{"Which VPN service do I use?", "What VPN am I on these days?", "What's my go-to VPN again?"},
		behavior: []string{"Connect me to a secure server before I browse — which VPN are you using?", "Turn on my VPN for this session; what service is it?", "Set me up on a private connection — where are you routing it?"},
	},
	{
		write: []string{
			"For streaming shows I'm on %s now — never %s, the catalog got thin.",
			"I watch everything on %s these days. I cancelled %s.",
			"Standing preference: my streaming service is %s, not %s.",
		},
		read:     []string{"Which streaming service do I use?", "Where do I watch my shows now?", "What's my go-to streaming service again?"},
		behavior: []string{"Queue up something for tonight — which streaming service are you using?", "Find me a documentary; what streaming app are you opening?", "Put on a movie for the kids — where are you streaming it?"},
	},
	{
		write: []string{
			"For e-books I read on %s now — never %s, the app was clunky.",
			"I buy all my books through %s these days. I left %s behind.",
			"Standing choice: my e-book service is %s, not %s.",
		},
		read:     []string{"Which e-book service do I use?", "Where do I read my books now?", "What's my go-to reading service again?"},
		behavior: []string{"Grab me that new release to read — which e-book service are you using?", "Add a book to my library; what reading app are you on?", "Find me something for the flight — where are you buying it?"},
	},
	{
		write: []string{
			"For meditation I use %s now — never %s, the sessions felt generic.",
			"I do my daily sit on %s these days. I dropped %s.",
			"Standing preference: my meditation app is %s, not %s.",
		},
		read:     []string{"Which meditation app do I use?", "Where do I do my meditation now?", "What's my go-to meditation service again?"},
		behavior: []string{"Start me a wind-down session tonight — which meditation app are you using?", "Play me a short breathing exercise; what app are you opening?", "Set up my morning meditation — where are you pulling it from?"},
	},
	{
		write: []string{
			"For prescriptions I fill at %s now — never %s, the wait was endless.",
			"I get my meds from %s these days. I switched off %s.",
			"Standing choice: my pharmacy is %s, not %s.",
		},
		read:     []string{"Which pharmacy do I use?", "Where do I fill my prescriptions now?", "What's my go-to pharmacy again?"},
		behavior: []string{"Refill my prescription for me — which pharmacy are you sending it to?", "Order my meds for pickup; what pharmacy are you using?", "Set up a refill reminder — where are you filling it?"},
	},
	{
		write: []string{
			"For dry cleaning I use %s now — never %s, they scorched a shirt.",
			"I take my clothes to %s these days. I stopped using %s.",
			"Standing preference: my dry cleaner is %s, not %s.",
		},
		read:     []string{"Which dry cleaner do I use?", "Where do I take my dry cleaning now?", "What's my go-to dry cleaner again?"},
		behavior: []string{"Send out my suit for cleaning — which dry cleaner are you using?", "Schedule a pickup for my coats; what service are you booking?", "Drop off these shirts for pressing — where are you taking them?"},
	},
	{
		write: []string{
			"For car repairs I go to %s now — never %s, they overcharged me twice.",
			"I take my car to %s these days. I'm done with %s.",
			"Standing choice: my mechanic is %s, not %s.",
		},
		read:     []string{"Which mechanic do I use?", "Where do I take my car now?", "What's my go-to auto shop again?"},
		behavior: []string{"Book my car in for a service — which mechanic are you calling?", "Schedule an oil change; what shop are you using?", "Get that rattle looked at — where are you taking the car?"},
	},
	{
		write: []string{
			"For moving and storage I use %s now — never %s, they showed up late.",
			"I book my moves through %s these days. I won't touch %s again.",
			"Standing preference: my movers are %s, not %s.",
		},
		read:     []string{"Which moving service do I use?", "Who do I book my moves through now?", "What's my go-to moving-and-storage company again?"},
		behavior: []string{"Arrange help hauling my furniture — which moving service are you booking?", "Set up a storage unit for me; what company are you using?", "Book movers for next month — who are you going with?"},
	},
	{
		write: []string{
			"For home internet I'm with %s now — never %s, the outages were constant.",
			"I get my internet from %s these days. I dropped %s.",
			"Standing choice: my home internet provider is %s, not %s.",
		},
		read:     []string{"Which internet provider do I use?", "Who supplies my home internet now?", "What's my go-to internet service again?"},
		behavior: []string{"Report an outage on my line — which internet provider are you contacting?", "Upgrade my home plan; what provider am I with?", "Set up a service call for the router — who is my internet through?"},
	},
	{
		write: []string{
			"For my phone plan I use %s now — never %s, the coverage was spotty.",
			"I'm on %s for mobile these days. I left %s.",
			"Standing preference: my mobile carrier is %s, not %s.",
		},
		read:     []string{"Which mobile carrier do I use?", "Who's my phone plan with now?", "What's my go-to mobile carrier again?"},
		behavior: []string{"Add a line to my plan — which carrier are you calling?", "Check my data usage; what carrier am I on?", "Sort out an international add-on — who is my phone through?"},
	},
	{
		write: []string{
			"For coffee beans I subscribe to %s now — never %s, the roast was stale.",
			"I get my beans from %s these days. I cancelled %s.",
			"Standing choice: my coffee subscription is %s, not %s.",
		},
		read:     []string{"Which coffee subscription do I use?", "Where do I get my beans now?", "What's my go-to coffee-subscription service again?"},
		behavior: []string{"Reorder my beans before I run out — which coffee subscription are you using?", "Add a bag of decaf to my next box; what service is it?", "Set up a recurring coffee delivery — where are you placing it?"},
	},
	{
		write: []string{
			"For house cleaning I use %s now — never %s, they kept rescheduling.",
			"I book my cleaners through %s these days. I stopped using %s.",
			"Standing preference: my cleaning service is %s, not %s.",
		},
		read:     []string{"Which cleaning service do I use?", "Who do I book my house cleaning through now?", "What's my go-to cleaning service again?"},
		behavior: []string{"Book a deep clean before the guests arrive — which cleaning service are you using?", "Schedule my weekly cleaner; what service are you booking?", "Arrange someone to tidy the place — who are you calling?"},
	},
	{
		write: []string{
			"For pet sitting I use %s now — never %s, they flaked on a booking.",
			"I book sitters through %s these days. I'm done with %s.",
			"Standing choice: my pet-sitting service is %s, not %s.",
		},
		read:     []string{"Which pet-sitting service do I use?", "Who do I book my pet sitters through now?", "What's my go-to pet-sitting service again?"},
		behavior: []string{"Line up someone to watch the dog this weekend — which pet-sitting service are you using?", "Book a sitter for while I'm away; what service are you on?", "Arrange care for the cat over the holiday — who are you booking?"},
	},
	{
		write: []string{
			"For insurance I'm with %s now — never %s, their claims process was a nightmare.",
			"I get my coverage through %s these days. I switched off %s.",
			"Standing preference: my insurance provider is %s, not %s.",
		},
		read:     []string{"Which insurance provider do I use?", "Who is my coverage through now?", "What's my go-to insurance company again?"},
		behavior: []string{"Start a claim for me — which insurance provider are you contacting?", "Check what my policy covers; who am I insured with?", "Get me a quote for adding coverage — which provider are you calling?"},
	},
	{
		write: []string{
			"For web hosting I use %s now — never %s, the uptime was terrible.",
			"I host my sites on %s these days. I migrated off %s.",
			"Standing choice: my web host is %s, not %s.",
		},
		read:     []string{"Which web host do I use?", "Where are my sites hosted now?", "What's my go-to web-hosting service again?"},
		behavior: []string{"Spin up a new site for me — which web host are you using?", "Check why my page is down; what host is it on?", "Set up staging for the project — where are you hosting it?"},
	},
	{
		write: []string{
			"For photo storage I use %s now — never %s, it kept re-compressing my shots.",
			"I keep all my photos on %s these days. I left %s.",
			"Standing preference: my photo storage is %s, not %s.",
		},
		read:     []string{"Which photo storage service do I use?", "Where do I keep my photos now?", "What's my go-to photo-storage service again?"},
		behavior: []string{"Back up the photos from my trip — which photo storage are you using?", "Make an album from last night; what service are they on?", "Free up space by archiving my pictures — where are you storing them?"},
	},
	{
		write: []string{
			"For laundry I use %s now — never %s, they shrank a sweater.",
			"I send my wash out through %s these days. I stopped using %s.",
			"Standing choice: my laundry service is %s, not %s.",
		},
		read:     []string{"Which laundry service do I use?", "Where do I send my wash now?", "What's my go-to laundry service again?"},
		behavior: []string{"Schedule a wash-and-fold pickup — which laundry service are you using?", "Send out my bedding to be cleaned; what service are you booking?", "Arrange a laundry pickup for tomorrow — who are you using?"},
	},
}

// v6 merged pools, precomputed once at init from the frozen v5 slice plus the v6
// additions. Kept as distinct vars so the v5 slices above never reallocate — the
// v5 generator draws from exactly the same backing arrays it always has.
// Each v6 pool is the frozen v5 slice, plus the model-authored extra* tier, plus
// the web-entropy webExtra* tier (gen/poolsv6_web.go — real-world terms harvested
// from public web lists and selected by a random.org TRNG draw, so the variety is
// grounded outside model priors). All three tiers are static literals; nothing here
// runs at generation time.
var (
	relativeNamesV6   = mergeStrings(relativeNames, extraRelativeNames, webExtraRelativeNames)
	leafNounsV6       = mergeStrings(leafNouns, extraLeafNouns, webExtraLeafNouns)
	tdNounsV6         = mergeStrings(tdNouns, extraTdNouns, webExtraTdNouns)
	relationPairsV6   = mergeRelationPairs(relationPairs, extraRelationPairs, webExtraRelationPairs)
	confabSpecsV6     = mergeConfabSpecs(confabSpecs, extraConfabSpecs, webExtraConfabSpecs)
	declPrefDomainsV6 = mergeDeclPrefDomains(declPrefDomains, extraDeclPrefDomains, webExtraDeclPrefDomains)
)

func mergeStrings(tiers ...[]string) []string {
	var n int
	for _, t := range tiers {
		n += len(t)
	}
	out := make([]string, 0, n)
	for _, t := range tiers {
		out = append(out, t...)
	}
	return out
}

func mergeRelationPairs(tiers ...[]relationPair) []relationPair {
	var out []relationPair
	for _, t := range tiers {
		out = append(out, t...)
	}
	return out
}

func mergeConfabSpecs(tiers ...[]confabSpec) []confabSpec {
	var out []confabSpec
	for _, t := range tiers {
		out = append(out, t...)
	}
	return out
}

func mergeDeclPrefDomains(tiers ...[]declPrefDomain) []declPrefDomain {
	var out []declPrefDomain
	for _, t := range tiers {
		out = append(out, t...)
	}
	return out
}

// Pool selectors: v6+ draws from the tripled pool; earlier versions from the
// frozen v5 pool. Threaded through the build* functions so pool size is a pure
// function of bench_version and the v5 draw sequence is preserved.
func relativeNamesFor(v int) []string {
	if v >= protocol.BenchVersionV6 {
		return relativeNamesV6
	}
	return relativeNames
}

func relationPairsFor(v int) []relationPair {
	if v >= protocol.BenchVersionV6 {
		return relationPairsV6
	}
	return relationPairs
}

func leafNounsFor(v int) []string {
	if v >= protocol.BenchVersionV6 {
		return leafNounsV6
	}
	return leafNouns
}

func tdNounsFor(v int) []string {
	if v >= protocol.BenchVersionV6 {
		return tdNounsV6
	}
	return tdNouns
}

func confabSpecsFor(v int) []confabSpec {
	if v >= protocol.BenchVersionV6 {
		return confabSpecsV6
	}
	return confabSpecs
}

func declPrefDomainsFor(v int) []declPrefDomain {
	if v >= protocol.BenchVersionV6 {
		return declPrefDomainsV6
	}
	return declPrefDomains
}
