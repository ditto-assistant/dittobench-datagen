package persona

import (
	"math/rand"
	"strings"
)

// Deep history (bench_version 8).
//
// Through v7 the scored haystack was ~4.5k tokens over 12 sessions and ~95 days.
// That is small enough to paste into a context window whole, which is the one
// thing a memory benchmark must not permit: a harness could skip retrieval
// entirely and still score. v8 scales the history to LongMemEval_S parity
// (~115k tokens over 30-50 sessions; arXiv:2410.10813) using BACKGROUND THREADS
// -- multi-turn, topically-coherent narratives that run across non-adjacent
// sessions and carry no ground truth at all.
//
// Three properties are what make the added volume load-bearing rather than
// padding, and each is enforced by a test in filler_test.go:
//
//   - Coherence. A thread is an ARC over one mundane subject (kickoff ->
//     complication -> grind -> turn -> coda), with later stages referring back
//     through a short coreferent instead of restating the subject. Semantically
//     disjoint padding inflates lexical-retrieval scores and is a well-attested
//     memory-benchmark validity problem; an on-topic narrative that never states
//     a fact does not.
//   - Length parity with graded turns. If fact beats stayed one-liners while
//     filler ran to paragraphs, "index only the short turns" would be a free
//     exploit worth more than retrieval. v8 elaborates fact beats from the SAME
//     clause pools this file uses (see Elaborate), so the two are drawn from one
//     length and vocabulary distribution.
//   - Answer disjointness. The filler vocabulary is deliberately mundane --
//     appliances, paperwork, gutters, permits -- and shares no token with any
//     answer pool in pools.go. TestFillerNeverStatesAFact asserts that under the
//     grader's own matcher across many seeds.
//
// Everything here is reached only through the v8 gate in BuildPlanForVersion, so
// no earlier contract's RNG stream or bytes are touched.

// fillerSubject is one background-thread subject. The slots are what let a small
// hand-reviewed set of arc templates produce a large surface: the templates
// supply narrative shape, the subject supplies every concrete noun, so surface
// count is (templates x subjects x notes) rather than a flat template list.
//
// No slot value may appear in any answer pool (pools.go). The subjects are
// chosen from domestic and bureaucratic life precisely because no persona fact
// family covers it.
type fillerSubject struct {
	label  string   // spoken in full at kickoff only: "the dishwasher that keeps stopping mid-cycle"
	coref  string   // short referent for every later stage: "the dishwasher"
	actor  string   // the counterpart in the story: "the repair engineer"
	thing  string   // the focal part or document: "the drain pump"
	action string   // gerund for the grind stage: "bailing it out by hand"
	notes  []string // concrete clauses, no ground truth, no answer-pool tokens
}

// fillerSubjects is the background-thread registry. Ordered (a slice, never a
// map range) so thread selection is reproducible from the seed.
var fillerSubjects = []fillerSubject{
	{
		label: "the dishwasher that keeps stopping mid-cycle", coref: "the dishwasher",
		actor: "the repair engineer", thing: "the drain pump", action: "bailing it out by hand",
		notes: []string{
			"it gets about twenty minutes in and then just sits there humming",
			"there is standing water at the bottom every single time",
			"the error light blinks four times, pauses, then does it again",
			"I have taken the filter out twice now and it was barely dirty",
			"the manual is no help, it just says call an approved technician",
			"the warranty ran out about a month before all this started",
		},
	},
	{
		label: "the gutters that overflow every time it rains hard", coref: "the gutters",
		actor: "the roofer", thing: "the downpipe", action: "scooping wet leaves out of the brackets",
		notes: []string{
			"water comes over the front edge in a sheet instead of going down",
			"there is a damp patch spreading on the wall underneath",
			"the ladder only just reaches, which is not a great feeling",
			"someone suggested a mesh guard but that seems like a whole project",
			"the moss on the roof edge has got noticeably thicker this year",
			"it only really shows up in properly heavy rain, never in a drizzle",
		},
	},
	{
		label: "renewing my passport before the backlog gets worse", coref: "the passport",
		actor: "the counter clerk", thing: "the application form", action: "chasing the reference number",
		notes: []string{
			"the online form timed out and I had to start the whole thing again",
			"the photo got rejected twice for shadow in the background",
			"the tracking page has said the same thing for eleven days",
			"the helpline queue was fifty minutes and then it cut off",
			"I had to dig out an old document I was fairly sure I had thrown away",
			"the fee went up between me starting the form and finishing it",
		},
	},
	{
		label: "the boiler pressure that keeps dropping overnight", coref: "the boiler",
		actor: "the heating engineer", thing: "the pressure gauge", action: "topping it back up every few days",
		notes: []string{
			"it sits fine all day and then reads low again by morning",
			"there is no visible leak anywhere I can actually get to",
			"the radiators upstairs are cold at the top and warm at the bottom",
			"bleeding them helped for about a week and then it came back",
			"the engineer wants to leave a gauge on it for a fortnight",
			"apparently a slow leak under the floor is the expensive possibility",
		},
	},
	{
		label: "the broadband that drops out every evening", coref: "the broadband",
		actor: "the support line", thing: "the router", action: "power-cycling it and hoping",
		notes: []string{
			"it goes at almost exactly the same time each night",
			"the line test they run always comes back clean, of course",
			"they sent a replacement router and it made no difference at all",
			"the engineer visit window was four hours and nobody came",
			"the socket by the door is loose, which may or may not be relevant",
			"speeds are fine at midday and unusable by eight",
		},
	},
	{
		label: "the fence panel that came down in the storm", coref: "the fence",
		actor: "the neighbour", thing: "the post", action: "propping it up with a spare batten",
		notes: []string{
			"the post snapped clean off at ground level, rotted right through",
			"we cannot work out whose side of the boundary it actually is",
			"the deeds are vague in a way that seems almost deliberate",
			"there is a second post that looks like it is going the same way",
			"the quote to do all three panels was more than I expected",
			"concrete spurs came up as the fix that lasts, apparently",
		},
	},
	{
		label: "the parking permit renewal that went wrong", coref: "the permit",
		actor: "the council office", thing: "the proof of address", action: "resubmitting the same documents",
		notes: []string{
			"they say they never received the second document",
			"the permit expired while the appeal was still open",
			"I got a ticket in the window on the third day of all this",
			"the appeal form asks for a reference the letter does not include",
			"two different departments have given me two different answers",
			"the automated line does not have an option for this situation",
		},
	},
	{
		label: "the leak under the kitchen sink", coref: "the leak",
		actor: "the plumber", thing: "the trap", action: "putting a towel down twice a day",
		notes: []string{
			"it is slow enough that you only notice from the smell",
			"the cupboard floor has swollen and gone soft in one corner",
			"tightening the joint helped for a day and a half",
			"the fitting looks like it has been repaired badly once before",
			"everything has to come out of the cupboard to get near it",
			"the plumber can come a week on Tuesday, which is the earliest",
		},
	},
	{
		label: "the flat-pack wardrobe I have been putting off", coref: "the wardrobe",
		actor: "the helpline", thing: "the instructions", action: "sorting screws into little piles",
		notes: []string{
			"one of the side panels arrived with a chip out of the corner",
			"step fourteen shows a bracket that is not in any of the bags",
			"the diagram is drawn from an angle nothing is ever seen from",
			"it needs two people for one step and exactly one for the rest",
			"the replacement panel is a four-week wait",
			"the box was too big to get up the stairs in one piece",
		},
	},
	{
		label: "the insurance claim for the water damage", coref: "the claim",
		actor: "the loss adjuster", thing: "the photographs", action: "digging out receipts from years ago",
		notes: []string{
			"they want proof of purchase for things bought long before smartphones",
			"the adjuster visit was rescheduled twice in one week",
			"the policy wording has an exclusion nobody mentioned at the start",
			"the first offer was about a third of the replacement cost",
			"everything has to stay exactly where it is until they have seen it",
			"the claim reference changes format depending on who I speak to",
		},
	},
	{
		label: "the car service that turned into a bigger job", coref: "the service",
		actor: "the mechanic", thing: "the brake discs", action: "getting the bus in the meantime",
		notes: []string{
			"it went in for an oil change and came back with a list",
			"two of the items are advisories and two are apparently urgent",
			"the courtesy car was gone by the time I got there",
			"the second opinion came back cheaper but a week later",
			"there is a noise on the left turn that nobody can reproduce",
			"the part has to come from somewhere else and takes four days",
		},
	},
	{
		label: "the allotment waiting list I finally got to the top of", coref: "the allotment",
		actor: "the site secretary", thing: "the shed", action: "clearing bindweed by the barrowful",
		notes: []string{
			"the plot has not been touched in at least two years",
			"there is a shed on it that is technically standing up",
			"the bindweed roots reach further down than the fork does",
			"the water butt at that end of the site is cracked",
			"the tenancy rules say it has to be half cleared within three months",
			"the plot next door is immaculate, which is quietly intimidating",
		},
	},
	{
		label: "the printer that will not talk to anything anymore", coref: "the printer",
		actor: "the support chat", thing: "the driver", action: "reinstalling it from scratch",
		notes: []string{
			"it worked fine until an update landed and then it simply did not",
			"it shows up on the network and then vanishes ten minutes later",
			"the cartridges it wants are somehow more expensive than a new one",
			"scanning works, printing does not, which makes no sense to me",
			"the support chat asked me to check the cable four separate times",
			"the queue fills up with jobs that will never come out",
		},
	},
	{
		label: "the damp patch in the back bedroom", coref: "the damp",
		actor: "the surveyor", thing: "the meter reading", action: "running a dehumidifier every night",
		notes: []string{
			"it comes and goes with the weather, which makes it hard to pin down",
			"the meter reads high in one corner and normal a foot away",
			"the surveyor thinks it is coming in rather than rising",
			"the paint has started lifting in a line along the skirting",
			"there is a smell in that room that never quite clears",
			"three trades have given three completely different diagnoses",
		},
	},
	{
		label: "the phone contract I have been trying to leave", coref: "the contract",
		actor: "the retentions team", thing: "the cancellation code", action: "sitting on hold in the evenings",
		notes: []string{
			"the code they gave me expired before the transfer went through",
			"the retentions offer was better than anything advertised to new customers",
			"the final bill included a month I had already paid for",
			"the transfer failed silently and nobody told either end",
			"the online account still shows the old plan a fortnight later",
			"each call starts from the beginning with a different person",
		},
	},
	{
		label: "the tree at the end of the garden that has to come down", coref: "the tree",
		actor: "the tree surgeon", thing: "the stump", action: "raking up what is left of it",
		notes: []string{
			"half of it is dead and the dead half leans over the shed",
			"there is a protection order question that has to be checked first",
			"the quote doubles if they have to take the stump out properly",
			"access is through the side gate, which is not wide enough",
			"the neighbour is keen to see it gone and is not offering to help",
			"it drops something sticky on everything underneath it all summer",
		},
	},
	{
		label: "the freezer that has iced itself solid", coref: "the freezer",
		actor: "the appliance shop", thing: "the door seal", action: "chipping ice out with a wooden spoon",
		notes: []string{
			"the top drawer will not open at all anymore",
			"the seal has gone hard and does not sit flat along one edge",
			"defrosting it means finding somewhere to put everything for a day",
			"it has started running almost constantly, which cannot be cheap",
			"there is a pool of water underneath it every few mornings",
			"a replacement seal is apparently a special order item",
		},
	},
	{
		label: "the recycling collection that keeps getting missed", coref: "the collection",
		actor: "the depot", thing: "the report form", action: "dragging the bin back in again",
		notes: []string{
			"it has been missed three weeks running now",
			"the online report form closes the case without anyone coming",
			"the schedule changed and nobody on the street was told",
			"the bin is out by six and it makes no difference at all",
			"the crew took the general waste and left the rest",
			"the depot number rings out after five in the afternoon",
		},
	},
	{
		label: "the loft that needs clearing before the insulation goes in", coref: "the loft",
		actor: "the installer", thing: "the hatch", action: "carrying boxes down the ladder one at a time",
		notes: []string{
			"there is more up there than I remembered putting up there",
			"the hatch is too small for half of what needs to come out",
			"the installer will not work around stored things, understandably",
			"two of the boxes have got damp at some point and are ruined",
			"the ladder wobbles in a way I have decided to ignore",
			"the grant deadline is what is making all of this urgent",
		},
	},
	{
		label: "the bank account they froze for no reason I can see", coref: "the account",
		actor: "the branch manager", thing: "the identity check", action: "waiting for a callback that never comes",
		notes: []string{
			"a direct debit bounced before I even knew anything was wrong",
			"the branch says it is a back-office matter and they cannot see it",
			"the identity documents were verified twice, in person, already",
			"the app just shows a generic error with a code that means nothing",
			"nobody will say how long the review normally takes",
			"the complaint reference has not been updated since it was opened",
		},
	},
	{
		label: "the roof tiles that shifted in the wind", coref: "the roof",
		actor: "the roofer", thing: "the ridge", action: "picking fragments off the path",
		notes: []string{
			"two came off completely and a few more have slid",
			"the ridge looks uneven from the far side of the road",
			"scaffolding is most of the cost, the tiles themselves are nothing",
			"nobody wants to quote until the wind drops",
			"there is no sign of anything coming through the ceiling yet",
			"the same thing apparently happened to the house opposite",
		},
	},
	{
		label: "the tax return I keep opening and closing", coref: "the return",
		actor: "the accountant", thing: "the spreadsheet", action: "matching receipts against statements",
		notes: []string{
			"one figure is out by an amount too small to find and too big to ignore",
			"the online system logs me out while I am reading the guidance",
			"two statements for the same period disagree by a few entries",
			"the deadline is far enough away to be dangerous",
			"the accountant needs a document I have to request and wait for",
			"last year's version is somehow no help at all this year",
		},
	},
	{
		label: "the doorbell wiring that stopped working", coref: "the doorbell",
		actor: "the electrician", thing: "the transformer", action: "testing it with the front door open",
		notes: []string{
			"the chime unit clicks but does not actually ring",
			"the button is fine, I swapped it to be sure",
			"the transformer is in the meter cupboard behind everything else",
			"deliveries have been left outside for a fortnight because of it",
			"a wireless one would solve it but the wiring is already there",
			"the electrician says it is a twenty minute job whenever he can get here",
		},
	},
	{
		label: "the gym membership that will not cancel", coref: "the membership",
		actor: "the front desk", thing: "the cancellation form", action: "emailing them again about it",
		notes: []string{
			"cancelling has to be done in person, apparently, in writing",
			"the form has to be handed in a full month before the billing date",
			"the payment went out again the day after I handed it in",
			"the front desk and the head office give different rules",
			"nobody can find the form I signed and dated in front of them",
			"the app still shows an active membership either way",
		},
	},
	{
		label: "the shed roof that has started letting water in", coref: "the shed",
		actor: "the builder's merchant", thing: "the felt", action: "moving everything off the floor",
		notes: []string{
			"the felt has torn along one whole edge",
			"whatever is stored in there now needs to be somewhere else first",
			"the timber underneath is soft in two places",
			"a whole new roof is not much more than patching it properly",
			"it has to be dry for two days running, which is the real problem",
			"the door has swollen and does not shut cleanly anymore",
		},
	},
	{
		label: "the delivery that has been rescheduled four times", coref: "the delivery",
		actor: "the driver", thing: "the tracking page", action: "staying in for another window",
		notes: []string{
			"the tracking page has said out for delivery twice with nothing arriving",
			"the driver marked it as refused, which is simply not true",
			"the depot is an hour away and only open when I am at work",
			"the slot is a four hour window given the evening before",
			"the seller says it is with the courier and the courier says the opposite",
			"there is a card through the door for a day I was in all day",
		},
	},
	{
		label: "the hedge that has got completely out of hand", coref: "the hedge",
		actor: "the gardener", thing: "the trimmer", action: "filling bag after bag with clippings",
		notes: []string{
			"it has grown over the path enough that you have to walk around it",
			"there is a nest in it, so nothing can happen for a few weeks",
			"the garden waste bin holds about a tenth of what comes off",
			"cutting it back hard risks it not coming back at all",
			"the far end is over the boundary and technically not mine",
			"the trimmer gave up halfway through and now smells of burning",
		},
	},
	{
		label: "the smoke alarm that chirps at three in the morning", coref: "the alarm",
		actor: "the landlord's contractor", thing: "the backup battery", action: "standing on a chair with a torch",
		notes: []string{
			"it only ever does it in the middle of the night",
			"changing the battery bought about six days of quiet",
			"it is wired in, so taking it down is not really an option",
			"there are three of them and I cannot tell which one it is",
			"the chirp stops the moment anyone is standing under it",
			"the whole set is apparently near the end of its rated life",
		},
	},
	{
		label: "the window that will not close properly since the frame moved", coref: "the window",
		actor: "the glazier", thing: "the hinge", action: "shouldering it shut every night",
		notes: []string{
			"there is a draught you can feel from the other side of the room",
			"the hinge has dropped and the whole thing sits crooked in the frame",
			"the lock lines up if you lift it, which cannot be right",
			"condensation has started forming between the panes",
			"the glazier says the unit is fine and the frame is the problem",
			"the frame is the expensive half, naturally",
		},
	},
	{
		label: "the shared driveway resurfacing everyone has to agree on", coref: "the driveway",
		actor: "the contractor", thing: "the quote", action: "knocking on doors about it again",
		notes: []string{
			"four households have to agree and one will not answer the door",
			"the potholes at the entrance have got genuinely bad",
			"the quote splits four ways but only if everyone signs",
			"one neighbour wants gravel and everyone else wants tarmac",
			"the work would block the whole thing for two days",
			"nobody can find the original agreement about who maintains it",
		},
	},
	{
		label: "the library card that stopped working mid-loan", coref: "the card",
		actor: "the librarian", thing: "the account record", action: "queueing at the desk about it",
		notes: []string{
			"the system shows a fine for a book I definitely returned",
			"the return was scanned, I watched it happen",
			"the card works for the door and not for the machines",
			"the account merged with someone else's at some point",
			"the librarian can see the problem and cannot fix it from that terminal",
			"the renewal went through and the block stayed on anyway",
		},
	},
	{
		label: "the extractor fan that stopped extracting", coref: "the fan",
		actor: "the electrician", thing: "the ducting", action: "leaving the window open instead",
		notes: []string{
			"it spins but nothing actually moves through it",
			"the ducting behind it has come loose somewhere in the wall",
			"the bathroom ceiling has started to mark in the corner",
			"the outside vent flap does not open when it runs",
			"getting to it means taking a panel off that was tiled over",
			"a stronger unit would need a bigger hole, which is the whole issue",
		},
	},
}

// fillerStage indexes the arc. Threads run stages in order and hold at the last
// stage if they outlive their arc, so a long-running thread stays coherent.
const (
	fillerKickoff = iota
	fillerComplication
	fillerGrind
	fillerTurn
	fillerCoda
	fillerStages
)

// fillerUserTmpls are the user turns per arc stage. Slots: {label} (kickoff
// only, so the full subject phrase appears once per thread), {coref}, {actor},
// {thing}, {action}, {note}. Every template renders two clauses so the turn has
// the length of a real chat message rather than a benchmark stub.
var fillerUserTmpls = [fillerStages][]string{
	fillerKickoff: {
		"I have finally had to start dealing with {label}. {note}, so it is not something I can keep ignoring.",
		"Something I have been meaning to sort out: {label}. The state of it is that {note}.",
		"New thing on my list this week, {label}. {note}, which is the part that bothers me.",
		"I want to talk through {label} if you have a minute. {note}, and I am not sure where to start.",
		"Adding {label} to the pile. {note}, and it has clearly been building up for a while.",
	},
	fillerComplication: {
		"{coref} has got more complicated. {actor} looked at it and reckons {note}.",
		"Bit of a setback with {coref}. It turns out {note}, which changes the plan somewhat.",
		"So {coref} is not as simple as I hoped. {note}, and {actor} was fairly blunt about it.",
		"Update on {coref}, and not a good one. {note}, so {thing} is now the thing to worry about.",
		"{coref} took a turn today. {note}, which nobody warned me about at the start.",
	},
	fillerGrind: {
		"Spent most of the day {action}. {note}, so it is slow going.",
		"Still {action} with {coref}. {note}, and I am starting to lose patience with it.",
		"More of the same on {coref}. I was {action} for a couple of hours and {note}.",
		"{coref} is eating my evenings. Mostly {action}, mostly because {note}.",
		"Another round of {action}. {note}, which is about where I expected to be by now.",
	},
	fillerTurn: {
		"Some progress on {coref} at last. {actor} finally sorted {thing}, though {note}.",
		"{coref} is looking better. The fix was {thing} in the end, but {note}.",
		"Small win with {coref} today. {actor} got somewhere with it, although {note}.",
		"Things have moved on {coref}. It is mostly handled now, except that {note}.",
		"{coref} is nearly behind me. {thing} was the answer, more or less, and {note}.",
	},
	fillerCoda: {
		"{coref} is finally done. Worth remembering that {note}, if it ever comes up again.",
		"Closing the book on {coref}. The whole thing took far longer than it should have, and {note}.",
		"{coref} is sorted at last. In hindsight the giveaway was that {note}.",
		"That is {coref} finished. I will not miss it, and I will not forget that {note}.",
		"{coref} is behind me now. The one lesson is that {note}, which I did not appreciate at the start.",
	},
}

// fillerAsstTmpls are the assistant turns per arc stage: substantive replies of
// the length a real assistant writes, so filler is not separable from a graded
// turn by response length either. None carries a recoverable fact.
var fillerAsstTmpls = [fillerStages][]string{
	fillerKickoff: {
		"That does sound like it has been brewing for a while. Given that {note}, I would start by ruling out the simple causes before anyone gets paid to look at it. Worth writing down when it happens and what you were doing at the time, because that pattern is usually what tells you the answer.",
		"Thanks for the detail, it helps. If {note}, then the useful first move is establishing whether it is getting worse or just staying annoying, since those point in quite different directions. Either way I would not rush into paying for {thing} yet.",
		"Understood, and that is a fair thing to be irritated by. The fact that {note} narrows it down more than you might think. Before you call anyone out, it is worth checking the obvious things yourself so you are not paying someone to do it for you.",
		"Noted. Something like this usually looks worse than it is, though I would not assume that. Since {note}, keep an eye on whether it changes with the weather or the time of day, because that is often the clue that makes it obvious.",
		"That is a frustrating one to inherit. Given {note}, the sensible order is: work out what it is, work out whether it is urgent, then work out who fixes it. Skipping the middle step is how these turn expensive.",
	},
	fillerComplication: {
		"That is a real complication, and unfortunately a fairly common one. If {actor} is right that {note}, then you are into a different kind of job than the one you thought you had. I would get that in writing before agreeing to anything.",
		"Sorry, that is not what you wanted to hear. When {note}, the cost usually sits in the access rather than the work itself, which is worth knowing before you see a number. A second opinion here is not paranoid, it is just sensible.",
		"That changes things. The detail that {note} is the part I would push back on, because it is the sort of claim that is either obviously true or completely unverifiable. Ask {actor} what exactly leads them to that conclusion.",
		"Understood, and I would take that seriously. Given {note}, {thing} is now the thing that decides how big this gets. Try to get a straight answer on that before any more work happens.",
		"That is a genuine setback rather than a delay. If {note}, then the timeline you had in mind is probably optimistic. I would plan around the worse version and be pleasantly surprised if it is not.",
	},
	fillerGrind: {
		"That sounds thoroughly tedious. There is not much to say except that {action} is usually the part nobody quotes for and everybody underestimates. If {note}, then it is worth pacing it rather than trying to finish it in a single push.",
		"Grinding work like that is hard to stay motivated through. Given {note}, I would set a limit on how much of it you take on yourself before deciding it is worth paying someone. There is a point where your time is the expensive part.",
		"Fair enough, and it does not sound like there is a shortcut here. Since {note}, the only real advice is to keep going in chunks and not let it become the thing you dread. Small sessions beat one heroic weekend, generally.",
		"That is a lot of effort for something so unglamorous. If {note}, at least you will know it has been done properly, which counts for something later. Do keep track of what you have already covered so you are not repeating yourself.",
		"Understandable, and honestly it does not sound like there is a faster way. Given {note}, I would just make sure you are not making the underlying problem worse while you deal with the symptom.",
	},
	fillerTurn: {
		"That is good to hear, and it sounds like the right diagnosis in the end. The remaining issue, that {note}, is the kind of thing that either settles down on its own or comes straight back, so give it a fortnight before deciding it is fixed.",
		"Progress at last. If {thing} really was the cause, then you should see a clear difference rather than a subtle one. The fact that {note} is worth keeping an eye on, but I would not read too much into it yet.",
		"Glad {actor} got somewhere. I would still keep whatever paperwork came out of it, because the caveat that {note} is exactly the sort of detail that matters if this recurs. Otherwise, that sounds close to done.",
		"That is a reasonable outcome given where it started. The one thing I would check is whether {note} is a side effect of the fix or something separate that happened to show up at the same time.",
		"Good. It usually is something like {thing} in the end, however elaborate the theories get first. Since {note}, I would give it a little while before declaring victory, but that does sound like the corner turned.",
	},
	fillerCoda: {
		"That is a relief, and it took long enough. The thing worth remembering is that {note}, because that is the detail that would have saved you weeks if you had known it at the start. Worth noting somewhere for next time.",
		"Good to have it closed off. In hindsight the pattern was there fairly early, and {note} was the giveaway. These things are always obvious in retrospect and never at the time.",
		"Well, that is done. If it ever recurs, the shortcut is knowing that {note}, which is not something anyone tells you up front. Glad it is off your list.",
		"That sounds properly finished rather than temporarily quiet, which is the important distinction. Since {note}, I would expect it to stay sorted. Nice to have one of these actually end.",
		"Closing that off is worth something in itself. The lesson that {note} generalises fairly well to the next thing like it, for whatever that is worth. Onwards.",
	},
}

// fillerTerseUser / fillerTerseAsst are the SHORT tail of the background-thread
// length distribution: a one-line check-in on a running concern. Real chat
// histories are a mixture of long turns and terse ones, and the mixture is also
// load-bearing for anti-gaming: if every filler turn were a paragraph, the short
// end of the haystack would be pure evidence and "read the short turns first"
// would be a free retrieval prior worth more than actually retrieving. Used only
// from the complication stage onward, so a thread still introduces itself in full.
var fillerTerseUser = []string{
	"{coref} again, briefly: {note}.",
	"Quick one on {coref}. {note}.",
	"No real change on {coref}, {note}.",
	"{coref} update: {note}.",
	"Still nothing useful on {coref}. {note}.",
}

var fillerTerseAsst = []string{
	"Noted, thanks for the update.",
	"Understood. Keep me posted on that one.",
	"Got it, nothing to do at this end then.",
	"Right, that is worth keeping track of.",
	"Noted, hopefully the next update is a better one.",
}

// fillerSlots renders one arc template against a subject and a chosen note.
// Unknown slots are impossible here (the template set is closed and covered by
// TestFillerTemplatesHaveNoUnfilledSlots), so no fallback is needed.
func fillerSlots(tmpl string, s fillerSubject, note string) string {
	return strings.NewReplacer(
		"{label}", s.label,
		"{coref}", s.coref,
		"{actor}", s.actor,
		"{thing}", s.thing,
		"{action}", s.action,
		"{note}", note,
	).Replace(tmpl)
}

// fillerBeat renders one turn of a thread at the given stage. It always consumes
// exactly three draws from r regardless of stage, so the RNG stream stays
// independent of thread shape.
//
// used records every user turn already emitted. A thread that revisits the same
// stage (a long arc, or two turns in one session) would otherwise re-roll the
// same (note, template) pair often enough to repeat itself verbatim, and a
// haystack with repeated turns is one a deduplicating retriever gets to shrink
// for free. On a collision the note index rotates forward deterministically
// until the rendered turn is new; the rotation consumes no extra draws, so
// determinism is unaffected.
func fillerBeat(r *rand.Rand, s fillerSubject, stage int, used map[string]bool) Beat {
	if stage >= fillerStages {
		stage = fillerStages - 1
	}
	ni := r.Intn(len(s.notes))
	u := fillerUserTmpls[stage][r.Intn(len(fillerUserTmpls[stage]))]
	a := fillerAsstTmpls[stage][r.Intn(len(fillerAsstTmpls[stage]))]
	// A quarter of post-kickoff turns are terse check-ins, so the filler spans
	// the same length range as the elaborated fact beats it sits among.
	if terse := r.Intn(4); terse == 0 && stage > fillerKickoff {
		u = fillerTerseUser[ni%len(fillerTerseUser)]
		a = fillerTerseAsst[ni%len(fillerTerseAsst)]
	}
	note := s.notes[ni]
	text := capitalize(fillerSlots(u, s, note))
	for tries := 1; used[text] && tries < len(s.notes); tries++ {
		note = s.notes[(ni+tries)%len(s.notes)]
		text = capitalize(fillerSlots(u, s, note))
	}
	used[text] = true
	return Beat{
		Kind:     BeatNoise,
		Topic:    s.coref,
		UserText: text,
		AsstText: fillerSlots(a, s, note),
	}
}

// thread is one live background narrative: which subject it is about and how far
// through its arc it has got.
//
// The arc advances PER TURN, not per session. An earlier design advanced it per
// session, which let a thread picked several times in one session emit that many
// KICKOFF turns, restating its full subject phrase each time. That is both
// incoherent and a gift to a lexical retriever, which gets to cluster the filler
// on the repeated phrase. A thread now spends one or two turns at a stage and
// then moves on, so the full subject phrase is spoken once and every turn after
// it is a callback through the short coreferent.
type thread struct {
	subj  fillerSubject
	stage int
	left  int // turns still to emit at the current stage
	picks int // picks so far in the current session (reset each session)
}

// buildFillerBeats lays background threads across the session grid and returns,
// per session, the filler beats to interleave with that session's fact beats.
//
// The model is a live thread POOL rather than one arc at a time: eight to twelve
// narratives are in flight at any point in the timeline, each contributing a turn
// or two per session, so a session reads as a real conversation touching several
// running concerns rather than a single monologue. Threads start and retire
// continuously, which is also what puts the evidence for a graded question at an
// unpredictable depth in the history rather than at a fixed offset.
//
// Deterministic: every draw comes from r and the layout is a pure function of
// (r, nSessions, targetBeats).
func buildFillerBeats(r *rand.Rand, nSessions, targetBeats int) [][]Beat {
	out := make([][]Beat, nSessions)
	if nSessions <= 0 || targetBeats <= 0 {
		return out
	}
	perSession := targetBeats / nSessions
	if perSession < 1 {
		perSession = 1
	}

	var live []*thread
	used := make(map[string]bool)
	emitted := 0
	for si := 0; si < nSessions; si++ {
		// Session budget jitters around the mean so filler density is not a
		// constant a harness could threshold on to segment the history.
		budget := perSession - 2 + r.Intn(5)
		if budget < 1 {
			budget = 1
		}
		// Enough arcs in flight that the per-session cap below can always be
		// honoured for the session budget.
		want := 8 + r.Intn(5)
		for len(live) < want {
			subj := fillerSubjects[r.Intn(len(fillerSubjects))]
			live = append(live, &thread{subj: subj, left: 1 + r.Intn(2)})
		}
		for _, t := range live {
			t.picks = 0
		}
		for i := 0; i < budget; i++ {
			t := pickThread(r, live)
			if t == nil { // every live arc has had its turn this session
				break
			}
			out[si] = append(out[si], fillerBeat(r, t.subj, t.stage, used))
			emitted++
			t.picks++
			t.left--
			if t.left <= 0 {
				t.stage++
				t.left = 1 + r.Intn(2)
			}
		}
		// Retire the arcs that have run to the end of their story.
		kept := live[:0]
		for _, t := range live {
			if t.stage < fillerStages {
				kept = append(kept, t)
			}
		}
		live = kept
		if emitted >= targetBeats && si < nSessions-1 {
			// Budget met early: hold a floor of one beat per remaining session so
			// no session is conspicuously empty, but stop growing the history.
			perSession = 1
		}
	}
	return out
}

// maxThreadTurnsPerSession caps how often one arc may speak in a single session.
// Without a cap the random pick clusters, and a session of six consecutive turns
// about the same subject reads as padding rather than as a life.
const maxThreadTurnsPerSession = 2

// pickThread returns a live arc that has not used up its turns this session,
// starting from a seeded offset and scanning forward so the choice stays uniform
// without an unbounded retry loop. nil when every arc is capped.
func pickThread(r *rand.Rand, live []*thread) *thread {
	if len(live) == 0 {
		return nil
	}
	start := r.Intn(len(live))
	for i := 0; i < len(live); i++ {
		t := live[(start+i)%len(live)]
		if t.picks < maxThreadTurnsPerSession {
			return t
		}
	}
	return nil
}

// elaborationClauses close out a fact beat's user turn with an inconsequential
// remark of the same register and length as a filler turn. Without this, graded
// turns would be one-liners in a haystack of paragraphs and "index the short
// turns" would beat retrieval. None carries a recoverable fact.
var elaborationClauses = []string{
	" Sorry, that was a long way round to a short point. It has been one of those weeks where everything needs doing at once and none of it is interesting.",
	" Anyway, that is the useful part. The rest of the day was mostly admin and a queue I did not expect to be in.",
	" I keep meaning to write these things down as they happen rather than remembering them a fortnight later. It never quite becomes a habit.",
	" That is probably more detail than you needed. I have got into the habit of over-explaining because half the time the detail turns out to matter.",
	" Not urgent, just something I wanted on the record before it slips my mind entirely. It has been that sort of month.",
	" I will spare you the rest of it. The short version is that everything took two steps more than it should have.",
	" Filing that away mostly so I stop turning it over in my head. It is not a problem, it is just noise I would rather put somewhere.",
	" Sorry for the ramble. There is a lot going on at the moment and I am clearly using you as somewhere to put it.",
}

// elaborationAcks extend a fact beat's assistant turn to the length of a filler
// reply. They acknowledge without restating any value, so the canonical answer
// token still appears exactly where it did before.
var elaborationAcks = []string{
	" No need to apologise for the detail, it is genuinely easier to keep track of things when they arrive in context rather than on their own. I will keep it in mind for anything related that comes up later.",
	" That is worth having written down. Weeks like that are exactly when things get forgotten, so putting it somewhere is the sensible instinct rather than an overreaction.",
	" Understood, and thanks for the context around it. It is usually the surrounding detail that makes the difference later, so I would rather have too much of it than too little.",
	" Noted. There is no such thing as too much background here, and it is often the offhand part that turns out to be the useful one further down the line.",
	" That is fine, and it is a reasonable use of me. Getting it out of your head and into somewhere it will keep is most of the value.",
	" Got it. I will hold onto the surrounding detail as well as the main point, since the two usually only make sense together.",
}

// elaborationBriefs are the short end of the fact-beat elaboration range. Every
// fact beat gets SOMETHING appended: leaving a third of them bare put them in the
// short tail of the haystack on their own, which made "read the shortest turns
// first" a retrieval prior five times better than chance (caught by
// TestV8FactBeatsAreNotSeparableByLength).
var elaborationBriefs = []string{
	" Nothing that needs doing about it.",
	" Just so it is on the record.",
	" No action needed, just noting it.",
	" Wanted that written down somewhere.",
	" Small thing, but worth a mention.",
}

// Elaborate extends a fact beat so its length and register match the background
// threads around it. The canonical value is untouched: clauses are APPENDED, so
// containment grading and every value-verbatim invariant hold exactly as before.
// Always consumes four draws from r regardless of which branch lands, so the RNG
// stream is independent of the length that comes out.
func Elaborate(r *rand.Rand, b Beat) Beat {
	clause := elaborationClauses[r.Intn(len(elaborationClauses))]
	ack := elaborationAcks[r.Intn(len(elaborationAcks))]
	brief := elaborationBriefs[r.Intn(len(elaborationBriefs))]
	switch r.Intn(4) {
	case 0: // terse, to sit alongside the filler's one-line check-ins
		b.UserText += brief
	case 1:
		b.UserText += brief
		b.AsstText += ack
	default: // full length, to sit alongside a filler arc turn
		b.UserText += clause
		b.AsstText += ack
	}
	return b
}
