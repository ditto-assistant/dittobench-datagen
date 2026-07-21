package gen

// bench_version 6 WEB-ENTROPY tier. These pool additions were sourced OUTSIDE the
// model's weights: candidate category words were harvested from public web lists
// and their inclusion/order was selected by TRUE random entropy (random.org
// atmospheric-noise draw, 2026-07-21), not model priors. They are then frozen as
// static literals here — generation itself never touches the network or a RNG it
// does not control, so the (seed, bench_version) determinism contract is intact.
//
// Sources: en.wikipedia.org List_of_musical_instruments / List_of_construction_trades
// / List_of_most_popular_given_names, mantelligence.com list-of-hobbies, and a
// nameberry/familyeducation unique-names roundup. TRNG selection seed (first draw): 452461.
// Privacy: generic category words only; every graded value stays a coined token.

// webExtraLeafNouns — nameable possessions/instruments/vehicles/pets.
var webExtraLeafNouns = []string{
	"unicycle", "harmonica", "trumpet", "hot rod", "canary", "concertina",
	"saxophone", "hedgehog", "dulcimer", "violin", "accordion", "cockapoo",
	"clarinet", "budgie", "surfboard", "betta fish", "harp", "dirt bike",
	"double bass", "mandolin", "dinghy", "banjo", "gecko", "chinchilla",
	"tandem bike", "model train", "vintage tractor", "jet ski", "model rocket", "cello",
	"corn snake", "flute",
}

// webExtraRelativeNames — diverse, distinctive intermediary given names.
var webExtraRelativeNames = []string{
	"Linnea", "Charbel", "Celestia", "Farhan", "Narek", "Vusal",
	"Fiorella", "Ayaan", "Vedant", "Iskra", "Ammar", "Isolde",
	"Milena", "Malin", "Kabir", "Astrid", "Dahlia", "Vespera",
	"Hani", "Hayk", "Ilgar", "Nabeel", "Samvel", "Dhruv",
	"Solara", "Santiago", "Mira", "Elnur", "Elias", "Tigran",
	"Kimani", "Zorina", "Mateo", "Violetta", "Viola", "Saeed",
	"Rashad", "Viraj", "Armen", "Hisham",
}

// webExtraTdNouns — evolving preference/setting attributes.
var webExtraTdNouns = []string{
	"preferred pen", "budgeting app", "weekend project",
	"language-learning app", "desk plant", "go-to delivery app",
	"current jigsaw puzzle", "default search engine", "favorite ice cream flavor",
	"favorite board game", "phone lock code style", "default video-call background",
	"daily vitamin", "morning news source", "preferred grocery store",
	"go-to takeout order", "preferred airline seat", "gym playlist",
}

// webExtraConfabSpecs — adjacent-but-distinct possession pairs.
var webExtraConfabSpecs = []confabSpec{
	{seededNoun: "model train", askedNoun: "model tram"},
	{seededNoun: "unicycle", askedNoun: "tricycle"},
	{seededNoun: "banjo", askedNoun: "mandolin"},
	{seededNoun: "corn snake", askedNoun: "ball python"},
	{seededNoun: "accordion", askedNoun: "melodica"},
	{seededNoun: "dirt bike", askedNoun: "pit bike"},
	{seededNoun: "trumpet", askedNoun: "cornet"},
	{seededNoun: "violin", askedNoun: "viola"},
	{seededNoun: "jet ski", askedNoun: "dinghy"},
	{seededNoun: "gecko", askedNoun: "chameleon"},
	{seededNoun: "tandem bike", askedNoun: "recumbent bike"},
	{seededNoun: "surfboard", askedNoun: "bodyboard"},
	{seededNoun: "canary", askedNoun: "budgie"},
	{seededNoun: "saxophone", askedNoun: "oboe"},
	{seededNoun: "flute", askedNoun: "piccolo"},
	{seededNoun: "hedgehog", askedNoun: "chinchilla"},
	{seededNoun: "betta fish", askedNoun: "guppy"},
	{seededNoun: "harp", askedNoun: "lyre"},
}

// webExtraRelationPairs — personal relation (target) vs trade/service decoy.
var webExtraRelationPairs = []relationPair{
	{target: "flatmate", decoy: "carpenter"},
	{target: "great-uncle", decoy: "landscaper"},
	{target: "half-sister", decoy: "plasterer"},
	{target: "stepfather", decoy: "glazier"},
	{target: "godson", decoy: "tiler"},
	{target: "half-brother", decoy: "welder"},
	{target: "host brother", decoy: "hvac technician"},
	{target: "dorm-mate", decoy: "mason"},
	{target: "stepmother", decoy: "roofer"},
	{target: "foster sister", decoy: "fencer"},
}

// webExtraDeclPrefDomains — generic life-admin service domains.
var webExtraDeclPrefDomains = []declPrefDomain{
	{
		write: []string{
			"For budgeting I use %s now — never %s, it kept miscategorizing everything.",
			"I track my spending in %s these days. I dropped %s.",
			"Standing choice: my budgeting app is %s, not %s.",
		},
		read:     []string{"Which budgeting app do I use?", "Where do I track my spending now?", "What's my go-to budgeting app again?"},
		behavior: []string{"Log this expense for me — which budgeting app are you using?", "Check how much I've spent this month; what app are you opening?", "Set me a savings goal — where are you setting it up?"},
	},
	{
		write: []string{
			"For language learning I use %s now — never %s, the lessons got repetitive.",
			"I do my daily practice on %s these days. I quit %s.",
			"Standing preference: my language app is %s, not %s.",
		},
		read:     []string{"Which language app do I use?", "Where do I practice my language now?", "What's my go-to language-learning app again?"},
		behavior: []string{"Start my daily lesson — which language app are you opening?", "Add fifteen minutes of practice; what app are you using?", "Set up a new language for me — where are you setting it up?"},
	},
	{
		write: []string{
			"For rental cars I book %s now — never %s, they hit me with junk fees.",
			"I rent through %s these days. I'm done with %s.",
			"Standing choice: my car-rental service is %s, not %s.",
		},
		read:     []string{"Which car-rental service do I use?", "Where do I rent my cars now?", "What's my go-to rental-car company again?"},
		behavior: []string{"Reserve a car for my trip — which rental service are you booking?", "Grab me an SUV for the weekend; what company are you using?", "Sort out a rental at the airport — who are you going with?"},
	},
	{
		write: []string{
			"For hotels I book through %s now — never %s, the listings were misleading.",
			"I sort my stays on %s these days. I stopped using %s.",
			"Standing preference: my hotel-booking service is %s, not %s.",
		},
		read:     []string{"Which hotel-booking service do I use?", "Where do I book my hotels now?", "What's my go-to hotel service again?"},
		behavior: []string{"Book me a room for the conference — which service are you using?", "Find a place near the venue; what booking service are you on?", "Set up a two-night stay — where are you reserving it?"},
	},
	{
		write: []string{
			"For parking I use %s now — never %s, it double-charged me twice.",
			"I pay for parking through %s these days. I ditched %s.",
			"Standing choice: my parking app is %s, not %s.",
		},
		read:     []string{"Which parking app do I use?", "Where do I pay for parking now?", "What's my go-to parking app again?"},
		behavior: []string{"Pay for my spot downtown — which parking app are you using?", "Extend my meter for an hour; what app are you on?", "Find me parking near the stadium — where are you booking it?"},
	},
	{
		write: []string{
			"For event tickets I use %s now — never %s, their fees were outrageous.",
			"I buy my tickets through %s these days. I avoid %s.",
			"Standing preference: my ticketing service is %s, not %s.",
		},
		read:     []string{"Which ticketing service do I use?", "Where do I buy my event tickets now?", "What's my go-to ticketing service again?"},
		behavior: []string{"Grab me two seats for the show — which ticketing service are you using?", "Find tickets to the game; what service are you on?", "Book me into that concert — where are you buying them?"},
	},
}
