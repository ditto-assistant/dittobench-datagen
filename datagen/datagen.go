// Package datagen procedurally generates small, fresh, randomized DittoBench
// tool-calling datasets. Generation is deterministic per seed (so a given seed
// always yields the same dataset) but varies widely across seeds. The practice
// API rotates the seed on every request so no two evaluations are identical.
// This is the anti-overfit property of the off-chain practice loop.
package datagen

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/ditto-assistant/dittobench-datagen/persona"
	"github.com/ditto-assistant/dittobench-datagen/protocol"
	"github.com/ditto-assistant/dittobench-datagen/toolexec"
)

// category describes one kind of tool-calling case and how to render it.
type category struct {
	name string
	tool string // single expected tool; empty means "no tool" (unless tools is set)
	// tools, when non-empty, is a multi-hop expected sequence (overrides tool);
	// MaxToolCalls becomes len(tools) and order is scored.
	tools []string
	// unordered marks a tools set as INDEPENDENT parallel calls: name/arg set is
	// scored but relative call order is not (ToolCase.Unordered).
	unordered bool
	// allowExtra sets ToolCase.AllowExtraTools: an extra call to the expected tool
	// is not penalized. Used by error-recovery cases where a legitimate RETRY after
	// a transient tool error would otherwise read as an over-budget extra call.
	allowExtra bool
	// argKey, when set on a single-tool category whose filler is an exact token
	// (a URL, a theme), pins RequiredArgs[argKey]=filler so the argument value is
	// deterministically scored: right tool + wrong arg does not get full credit.
	argKey string
	// wrap applies the Tier-1a lead-in/trailer surface variation to the rendered
	// prompt (persona.Wrap with the prompt pools below). Set it on template-only
	// categories whose prompt pool would otherwise be a handful of memorizable
	// strings; leave it off for categories whose surface entropy comes from a
	// large filler pool, and for result-usage/intent prompts that must reach the
	// harness exactly as authored.
	wrap      bool
	templates []string
	// grammar, when set, replaces templates as the prompt source: the prompt is
	// persona.Expand(r, grammar, "root"), then the normal %s fill and wrap run on
	// the expansion. Used by the audited low-variety categories whose surface
	// must outgrow a memorizable template list.
	grammar persona.Grammar
	// intents are prompts that imply the arg VALUE (or name a near-miss) instead of
	// stating it verbatim, so a harness must parse intent rather than copy a token.
	// Used for a share of a set_* category's cases; requires argKey.
	intents []argIntent
}

// argIntent maps an intent-phrased prompt to the arg value it should resolve to.
// Prompt either implies Value ("think harder" → "high") or names a near-miss
// alongside it ("switch me off gpt-5 to gemini-3-pro" → "gemini-3-pro"), so a
// token-copying harness gets it wrong.
type argIntent struct {
	prompt string
	value  string
}

// word pools used to vary entities/phrasings across seeds.
var (
	subjects = []string{
		"my dentist appointment", "the project deadline", "Sarah's birthday",
		"my car insurance", "the meeting notes", "my flight to Tokyo",
		"the grocery list", "my gym schedule", "the wifi password",
		"my passport number", "the book recommendation", "my doctor's name",
	}
	people = []string{
		"Alice", "Bob", "Carol", "David", "Erin", "Frank", "Grace", "Heidi",
	}
	topics = []string{
		"quantum computing", "the 2024 Olympics", "best espresso machines",
		"rust vs go", "climate policy", "the stock market today",
		"sourdough recipes", "electric vehicles", "the James Webb telescope",
	}
	urls = []string{
		"https://example.com/article", "https://news.site/story",
		"https://blog.dev/post", "https://docs.io/guide",
		"https://github.com/org/repo",
	}
	imagePrompts = []string{
		"a sunset over mountains", "a robot drinking coffee",
		"a futuristic city skyline", "a watercolor fox",
		"an astronaut on a beach", "a neon cyberpunk street",
	}
	artifactKinds = []string{
		"a landing page", "a todo app", "a snake game",
		"a markdown resume", "a pomodoro timer", "a budget tracker",
	}
	agentTasks = []string{
		"scrape the latest headlines", "summarize this PDF",
		"refactor the auth module", "generate unit tests",
		"build a CSV report", "deploy the staging branch",
	}
	themes = []string{"dark", "light", "system", "midnight", "solarized"}

	memoryIDs = []string{"mem-1042", "mem-2087", "mem-3310", "mem-7761", "mem-9024"}
	// workflowGoals are decomposable multi-part goals: the production
	// execute_agent_workflow takes a free-form goal a planner splits into
	// parallel sub-agents (there are no named, predefined workflows). The value
	// doubles as the pinned goal argument, so it must survive verbatim inside
	// the goal the harness passes.
	workflowGoals = []string{
		"a comparison of our top five competitors",
		"an audit of our three services for risks",
		"research on four laptop options with a recommendation",
		"a review of the codebase for security, performance, and correctness",
		"a market scan across several regions with a combined summary",
	}
	models    = []string{"gpt-5", "claude-sonnet-5", "gemini-3-pro", "llama-4-70b"}
	efforts   = []string{"low", "medium", "high"}
	toolPrefs = []string{"disable web search", "enable image tools", "turn off agent jobs", "allow only memory tools"}
	// Production-tool coverage pools: scheduling, discovery, workspace, memory
	// writes, appearance settings.
	schedules       = []string{"every morning at 8am", "every Monday", "each weekday evening", "the first of every month", "every Friday at noon"}
	automationTasks = []string{"email me a news digest", "summarize my unread messages", "post my standup notes", "back up my documents"}
	recipeNames     = []string{"weekly-review", "trip-planner", "invoice-run", "content-pipeline"}
	recipeSteps     = []string{"pull metrics, then draft a summary, then email it", "search flights, compare prices, book the cheapest", "gather receipts, total them, file the report"}
	savableFacts    = []string{"I'm allergic to shellfish", "my anniversary is June 3rd", "I prefer window seats", "my daughter's name is Mia", "I take my coffee black"}
	calendarTitles  = []string{"dentist appointment", "team offsite", "1:1 with Sam", "flight to Boston", "birthday dinner"}
	calendarWhens   = []string{"next Tuesday at 3pm", "Friday morning", "March 14th", "tomorrow at noon"}
	emailRecipients = []string{"sam@example.com", "the team", "my accountant", "jordan@example.com"}
	accentColors    = []string{"teal", "crimson", "indigo", "amber", "emerald"}
	chatFonts       = []string{"Inter", "JetBrains Mono", "Georgia", "system default"}
	feedback        = []string{
		"the image tool keeps timing out", "love the new memory search",
		"the export button is broken on mobile", "please add dark mode to artifacts",
	}

	chitchat = []string{
		"hey, how's it going?", "thanks, that was helpful!",
		"tell me a joke", "what's your favorite color?",
		"good morning!", "you're awesome", "lol nice",
		"haha that one made me laugh", "hope your day's going well",
		"just saying hi", "nice, that's exactly what I meant",
		"gm! ready when you are", "appreciate you", "ok cool cool cool",
	}
	abstentions = []string{
		"what's the meaning of life?", "should I quit my job?",
		"do you love me?", "what will the weather be like next year?",
		"who will win the next election?", "what am I thinking right now?",
		"will my startup succeed?", "is my sister secretly mad at me?",
		"what number am I thinking of?", "will it rain on my wedding day next June?",
		"how long am I going to live?", "did my neighbor take my package?",
	}
)

// categories enumerates every case type the generator can emit. Order is
// stable so seeding is reproducible.
var categories = []category{
	{
		name: "memory_lookup", tool: "search_memories",
		templates: []string{
			"What did I say about %s?",
			"Remind me about %s.",
			"Do you remember %s?",
			"Look up %s from my memories.",
		},
	},
	{
		name: "memory_subject", tool: "search_subjects",
		templates: []string{
			"What subjects do I have notes on related to %s?",
			"Find the topic that covers %s.",
			"Which of my subjects mention %s?",
		},
	},
	{
		name: "web_search", tool: "search_web",
		templates: []string{
			"Search the web for %s.",
			"What's the latest on %s?",
			"Find recent news about %s.",
			"Google %s for me.",
		},
	},
	{
		name: "link_read", tool: "read_links", argKey: "urls",
		templates: []string{
			"Read %s and summarize it.",
			"What does this page say: %s",
			"Open %s and tell me the main points.",
		},
	},
	{
		name: "image_create", tool: "create_image",
		templates: []string{
			"Generate an image of %s.",
			"Create a picture of %s.",
			"Draw me %s.",
		},
	},
	{
		name: "artifacts_create", tool: "artifacts",
		templates: []string{
			"Build me %s.",
			"Make %s I can preview.",
			"Create %s as an interactive artifact.",
		},
	},
	{
		name: "agent_job", tool: "execute_agent_job", argKey: "task",
		templates: []string{
			"Run a background job to %s.",
			"Kick off an agent to %s.",
			"Dispatch a task to %s.",
		},
	},
	{
		name: "settings", tool: "set_theme", argKey: "theme",
		templates: []string{
			"Switch to %s mode.",
			"Set my theme to %s.",
			"Change the app theme to %s.",
		},
		intents: []argIntent{
			{"I had it on light, but switch me to midnight instead.", "midnight"},
			{"Don't leave it on dark — set it to solarized.", "solarized"},
			{"Change it from system to dark.", "dark"},
			{"Not light — I want the solarized look.", "solarized"},
		},
	},
	{
		// Hard routing trap: phrased like a web search but the right action is a
		// memory lookup (the user is asking about THEIR past, not the public web).
		name: "route_memory_not_web", tool: "search_memories",
		templates: []string{
			"Search for what I told you about %s.",
			"Look up %s — I mentioned it before.",
			"Find that thing I saved about %s.",
		},
	},
	{
		// Hard routing trap: looks like a memory query but needs the live web
		// (current/real-time info the user can't have stored).
		name: "route_web_not_memory", tool: "search_web",
		templates: []string{
			"Remind me what the latest news on %s is.",
			"Do you recall the current price of %s?",
			"What's the up-to-date status of %s right now?",
		},
	},
	{
		// Run-vs-read trap: dispatch a NEW background job (not check existing).
		name: "agent_run_not_read", tool: "execute_agent_job",
		templates: []string{
			"Go ahead and %s for me now.",
			"Please actually %s, don't just tell me how.",
			"Start working on this: %s.",
		},
	},
	{
		// Run-vs-read trap: check status of EXISTING jobs (not start a new one).
		name: "agent_read_not_run", wrap: true, tool: "list_agent_jobs",
		grammar: agentReadGrammar,
	},
	{
		// Edit-vs-create trap: the user references an EXISTING image to modify, so
		// the right tool is edit_image, not create_image.
		name: "image_edit_not_create", wrap: true, tool: "edit_image",
		grammar: imageEditGrammar,
	},
	{
		// Job-vs-workflow trap: phrased like a one-off background job, but the goal
		// has clear independent parts, so the right tool is execute_agent_workflow
		// (production semantics: a planner decomposes a free-form goal into parallel
		// sub-agents; there are no named workflows).
		name: "workflow_not_job", tool: "execute_agent_workflow", argKey: "goal",
		templates: []string{
			"Kick off a background job to produce %s, working the independent parts in parallel.",
			"Dispatch an agent to put together %s, tackling each piece simultaneously.",
			"Start a one-off task for %s, splitting the separate angles across workers.",
		},
	},
	{
		name: "no_tool", tool: "", // chit-chat, answer directly
		templates: []string{
			"%s",
		},
	},
	{
		name: "abstention", tool: "", // should not call tools, answer/abstain
		templates: []string{
			"%s",
		},
	},
	{
		// Hallucinated-argument trap (BFCL): the request names an action that needs a
		// specific argument but never supplies its VALUE, and no prior context holds
		// it. The correct behavior is to ask for the missing detail, NOT to invoke the
		// tool with a fabricated argument; like abstention, calling any tool
		// scores 0 (a guessed value is the failure mode being probed).
		name: "arg_hallucination", wrap: true, tool: "", // required value missing: ask, don't invent
		templates: []string{
			"Change my theme.",
			"Switch my main model, please.",
			"Set my reasoning effort.",
			"Read that link and summarize it.",
			"File that as feedback for the team.",
			"Spin up a parallel agent workflow for me.",
			"Set my accent color.",
			"Change the chat font, please.",
			"Put it on my calendar.",
			"Email them the update.",
			"Delete that memory for me.",
			"Run my recipe.",
		},
	},
	// Full-catalog coverage: single-hop categories so every catalog tool is the
	// correct answer for some case.
	{
		// Deliberately avoids the tool's own keywords (fetch/full/text/memory/
		// pair): the old templates stated the call verbatim, so a no-model
		// keyword router solved this category at floor 1.0. The pair ID is still
		// the pinned argument; the intent must be read, not string-matched.
		name: "memory_fetch", wrap: true, tool: "fetch_memories", argKey: "pairIds",
		grammar: memoryFetchGrammar,
	},
	{
		name: "agent_workflow", tool: "execute_agent_workflow", argKey: "goal",
		templates: []string{
			"Plan %s as parallel sub-agents.",
			"Split %s into parallel subtasks and run them.",
			"Use a multi-agent workflow to produce %s.",
		},
	},
	{
		name: "feedback", tool: "file_feedback_for_team", argKey: "message",
		templates: []string{
			"Send this to the Ditto team: %s.",
			"File feedback for the team — %s.",
			"Report to the devs that %s.",
		},
	},
	{
		name: "set_model", tool: "set_main_model", argKey: "model",
		templates: []string{
			"Switch my chat model to %s.",
			"Use %s as my main model.",
			"Change my primary model to %s.",
		},
		intents: []argIntent{
			{"I'm on gemini-3-pro now — move me over to claude-sonnet-5.", "claude-sonnet-5"},
			{"Switch me off gpt-5 and onto gemini-3-pro.", "gemini-3-pro"},
			{"My main is llama-4-70b; change it to gpt-5.", "gpt-5"},
			{"Take me off claude-sonnet-5 and put me on llama-4-70b.", "llama-4-70b"},
		},
	},
	{
		intents: []argIntent{
			{"Really take your time and reason deeply on my questions from now on.", "high"},
			{"Make your answers as thorough and careful as you can.", "high"},
			{"Keep it quick and don't overthink my requests.", "low"},
			{"Snappy, minimal reasoning is fine — don't burn cycles.", "low"},
			{"Use a balanced amount of reasoning, nothing extreme.", "medium"},
		},
		name: "set_effort", tool: "set_reasoning_effort", argKey: "effort",
		templates: []string{
			"Set reasoning effort to %s.",
			"Make responses use %s effort.",
			"Change the thinking level to %s.",
		},
	},
	{
		name: "set_tool_prefs", tool: "set_chat_tool_preferences", argKey: "preferences",
		templates: []string{
			"Update my tool preferences: %s.",
			"Change my chat tools so you %s.",
			"Adjust which tools you use — %s.",
		},
	},
	// --- Production-tool coverage: routing/restraint traps ---
	{
		// Scheduling trap: a task phrased with a recurring TIME cue is an
		// automation, not a one-off agent job. The %s is the schedule (pinned).
		name: "automation_not_job", tool: "create_automation", argKey: "schedule",
		templates: []string{
			"Every morning, %s my news digest.",
			"On a schedule — %s — send me my summary.",
			"Set this to run %s: post my standup.",
		},
	},
	{
		name: "automation_list", wrap: true, tool: "list_automations",
		grammar: automationListGrammar,
	},
	{
		name: "recipe_create", tool: "create_recipe", argKey: "name",
		templates: []string{
			"Save a recipe called %s for later.",
			"Make a reusable recipe named %s.",
			"Create a %s recipe I can run again.",
		},
	},
	{
		name: "recipe_apply", tool: "apply_recipe", argKey: "name",
		templates: []string{
			"Run my %s recipe.",
			"Apply the %s recipe now.",
			"Execute my saved %s recipe.",
		},
	},
	{
		// Capability discovery: "what can you do / where is X" routes to
		// discover_capabilities, not a guess from memory or the web.
		name: "capability_discovery", wrap: true, tool: "discover_capabilities",
		grammar: capabilityGrammar,
	},
	{
		// Save-vs-search trap: a STATEMENT of a new fact is a memory WRITE, not a
		// search. The %s is the fact (freely rephrasable, no pinned arg token).
		name: "memory_save_not_search", tool: "save_memory",
		templates: []string{
			"Remember that %s.",
			"Note for later: %s.",
			"Please keep in mind that %s.",
		},
	},
	{
		name: "memory_update", tool: "update_memory", argKey: "pair_id",
		templates: []string{
			"Update memory %s with my new address.",
			"Change what you saved in memory %s.",
			"Correct memory %s to the new value.",
		},
	},
	{
		name: "memory_delete", tool: "delete_memory", argKey: "pair_id",
		templates: []string{
			"Delete memory %s.",
			"Forget what's in memory %s.",
			"Remove memory %s from my history.",
		},
	},
	{
		name: "calendar_create", tool: "calendar_create_event", argKey: "title",
		templates: []string{
			"Put %s on my calendar.",
			"Add %s to my calendar.",
			"Schedule %s on my calendar.",
		},
	},
	{
		name: "calendar_search", tool: "calendar_search_events", argKey: "query",
		templates: []string{
			"What's on my calendar about %s?",
			"Find the %s event on my calendar.",
			"Search my calendar for %s.",
		},
	},
	{
		name: "email_send", tool: "gmail_send", argKey: "to",
		templates: []string{
			"Email %s to let them know I'll be late.",
			"Send an email to %s with the update.",
			"Shoot %s a quick note about the meeting.",
		},
	},
	{
		name: "set_accent", tool: "set_accent_color", argKey: "color",
		templates: []string{
			"Set my accent color to %s.",
			"Make the accent color %s.",
			"Change the app accent to %s.",
		},
	},
	{
		name: "set_font", tool: "set_chat_font", argKey: "font",
		templates: []string{
			"Change my chat font to %s.",
			"Use %s as the chat font.",
			"Set the chat typeface to %s.",
		},
	},

	// Multi-hop trajectories: the correct answer is a tool SEQUENCE, scored
	// with order credit. These exercise the multi-call path.
	{
		name: "multi_web_read", tools: []string{"search_web", "read_links"},
		templates: []string{
			"Look up %s online and open the top result.",
			"Find a page about %s and read it for me.",
			"Search for %s and summarize the first source.",
		},
	},
	{
		// Parallel independent calls (BFCL "parallel"): two unrelated actions in one
		// turn, neither's input depending on the other's output, so any call order is
		// correct. Both are about the same entity so one filler renders both.
		name: "parallel_web_image", tools: []string{"search_web", "create_image"}, unordered: true,
		templates: []string{
			"Search the web for %s and also generate an image of it.",
			"Look up %s online, and separately make me a picture of it.",
			"Two things: find the latest on %s and create an image of it.",
		},
	},
	{
		name: "multi_subject_scope", tools: []string{"search_subjects", "search_memories_in_subjects"},
		templates: []string{
			"Find my notes related to %s and pull the details.",
			"Which subject covers %s — then recall the specifics.",
			"Look through my topics on %s and fetch what I saved.",
		},
	},
	{
		name: "multi_job_status", tools: []string{"execute_agent_job", "get_agent_job_status"},
		templates: []string{
			"Kick off a job to %s and tell me its status.",
			"Start %s in the background, then check how it's going.",
			"Run %s and report the job status.",
		},
	},
	{
		name: "multi_image_edit", tools: []string{"create_image", "edit_image"},
		templates: []string{
			"Make an image of %s, then brighten it.",
			"Generate %s and then add more detail.",
			"Create a picture of %s and tweak the colors.",
		},
	},
	// Result-usage: the answer requires a value
	// that exists ONLY in the tool's returned content (a fabricated per-seed
	// needle, toolexec), so the case cannot be answered by self-report or base-
	// model knowledge; the harness must actually execute the tool and USE the
	// result. The %s is the needle's Subject (filled below), keeping the question
	// and the served fact coherent. Scored deterministically (trajectory + needle-
	// in-answer).
	{
		name: "web_result_usage", tool: "search_web",
		templates: []string{
			"Search the web for the latest figure on %s and tell me the exact number.",
			"What number does the current top result report for %s?",
			"Look up %s online and give me the precise figure it cites.",
		},
	},
	{
		name: "multi_web_result_usage", tools: []string{"search_web", "read_links"},
		templates: []string{
			"Look up %s online, open the top result, and tell me the exact figure it reports.",
			"Find a page about %s, read it, and give me the precise number.",
			"Research %s on the web, read the leading source, and report its exact figure.",
		},
	},
	// Dependent-arg chain (result-usage): execute_agent_job returns a job id that
	// MUST be passed to get_agent_job_status, which only then returns the needle.
	// The needle exists only behind the chained id, so the whole trajectory is
	// unfakeable: a harness must read the first result and thread its value into
	// the second call. The "_job_chain" marker selects the dependent serving gate
	// (toolexec) and "_result_usage" the deterministic result-usage scoring.
	{
		name: "job_chain_result_usage", tools: []string{"execute_agent_job", "get_agent_job_status"},
		templates: []string{
			"Kick off a job to compute %s, then check its status and tell me the exact figure it returns.",
			"Dispatch a background job for %s, then look up that job's result and give me the precise number.",
			"Start a job to work out %s, then read the finished job's status and report its exact figure.",
		},
	},
	// Error recovery (result-usage): the FIRST call to the served tool returns a
	// transient error; the needle is delivered only on a RETRY. A harness that
	// gives up after one error cannot produce the answer value, so recovery is
	// enforced by the serving layer (no scorer change). allowExtra so the retry is
	// not charged as an over-budget extra call. The "recovery" marker selects the
	// transient-first serving behavior in toolexec.
	{
		name: "web_recovery_result_usage", tool: "search_web", allowExtra: true,
		templates: []string{
			"Search the web for the latest figure on %s and tell me the exact number (retry if the search flakes).",
			"Look up %s online and report the precise figure; if the first attempt errors, try again.",
			"Get the current number for %s from the web, retrying past any transient hiccup.",
		},
	},
}

// Tier-1a prompt-surface variation for the template-only categories whose
// prompt pool would otherwise be a handful of memorizable strings (the same
// machinery the persona haystack uses: a seed-chosen lead-in and trailer
// multiply the surface a template matcher must cover without changing the
// intent or any pinned argument). Empty entries weight toward the plain prompt.
var promptLeadIns = []string{
	"", "", "", "",
	"Hey, ", "Quick one: ", "Ok, so — ", "Real quick, ", "One thing: ",
	"When you get a sec, ", "Oh right — ",
}

var promptTrailers = []string{
	"", "", "", "",
	" Thanks!", " Cheers.", " No rush.", " Appreciate it.", " When you can.",
}

// resultUsageSuffix marks the categories whose correct answer must incorporate a
// tool's returned content (result-usage). Their prompt %s is the fixture needle's
// Subject and they are scored deterministically against the needle Value.
const resultUsageSuffix = "_result_usage"

// IsResultUsage reports whether a case category is a result-usage category. The
// pipeline scores these deterministically on trajectory + answer-incorporates-
// needle.
func IsResultUsage(category string) bool { return strings.HasSuffix(category, resultUsageSuffix) }

// fillerFor returns a random entity string appropriate for a category.
func fillerFor(r *rand.Rand, cat string) string {
	switch cat {
	case "memory_lookup", "memory_subject":
		return subjects[r.Intn(len(subjects))]
	case "web_search":
		return topics[r.Intn(len(topics))]
	case "link_read":
		return urls[r.Intn(len(urls))]
	case "image_create":
		return imagePrompts[r.Intn(len(imagePrompts))]
	case "artifacts_create":
		return artifactKinds[r.Intn(len(artifactKinds))]
	case "agent_job":
		return agentTasks[r.Intn(len(agentTasks))]
	case "settings":
		return themes[r.Intn(len(themes))]
	case "route_memory_not_web":
		return subjects[r.Intn(len(subjects))]
	case "route_web_not_memory":
		return topics[r.Intn(len(topics))]
	case "agent_run_not_read":
		return agentTasks[r.Intn(len(agentTasks))]
	case "agent_read_not_run", "image_edit_not_create":
		return "" // templates have no placeholder
	case "workflow_not_job":
		return workflowGoals[r.Intn(len(workflowGoals))]
	case "memory_fetch":
		return memoryIDs[r.Intn(len(memoryIDs))]
	case "agent_workflow":
		return workflowGoals[r.Intn(len(workflowGoals))]
	case "feedback":
		return feedback[r.Intn(len(feedback))]
	case "set_model":
		return models[r.Intn(len(models))]
	case "set_effort":
		return efforts[r.Intn(len(efforts))]
	case "set_tool_prefs":
		return toolPrefs[r.Intn(len(toolPrefs))]
	case "automation_not_job":
		return schedules[r.Intn(len(schedules))]
	case "automation_list":
		return "" // template has no placeholder
	case "recipe_create", "recipe_apply":
		return recipeNames[r.Intn(len(recipeNames))]
	case "capability_discovery":
		return "" // template has no placeholder
	case "memory_save_not_search":
		return savableFacts[r.Intn(len(savableFacts))]
	case "memory_update", "memory_delete":
		return memoryIDs[r.Intn(len(memoryIDs))]
	case "calendar_create":
		return calendarTitles[r.Intn(len(calendarTitles))]
	case "calendar_search":
		return topics[r.Intn(len(topics))]
	case "email_send":
		return emailRecipients[r.Intn(len(emailRecipients))]
	case "set_accent":
		return accentColors[r.Intn(len(accentColors))]
	case "set_font":
		return chatFonts[r.Intn(len(chatFonts))]
	case "multi_web_read", "parallel_web_image":
		return topics[r.Intn(len(topics))]
	case "multi_subject_scope":
		return subjects[r.Intn(len(subjects))]
	case "multi_job_status":
		return agentTasks[r.Intn(len(agentTasks))]
	case "multi_image_edit":
		return imagePrompts[r.Intn(len(imagePrompts))]
	case "no_tool":
		return chitchat[r.Intn(len(chitchat))]
	case "abstention":
		return abstentions[r.Intn(len(abstentions))]
	case "arg_hallucination":
		return "" // templates omit the required value on purpose
	default:
		_ = people // reserved for future multi-entity templates
		return topics[r.Intn(len(topics))]
	}
}

// Generate produces a deterministic-per-seed dataset of n tool-calling cases.
// n is clamped to [1, 200]; the practice default is small (20-40).
func Generate(seed int64, n int) protocol.Dataset {
	if n < 1 {
		n = 1
	}
	if n > 200 {
		n = 200
	}
	r := rand.New(rand.NewSource(protocol.RotateSeed(seed)))
	return protocol.Dataset{
		Seed:        seed,
		GeneratedAt: protocol.DatasetEpochRFC3339,
		ToolCases:   GenerateCases(r, seed, n),
	}
}

// stratifiedCategoryOrder returns n category indices with a FIXED per-category
// quota (each category appears floor(n/C) or ceil(n/C) times), then shuffles the
// order with the seeded RNG. Fixing the category MIX per run (rather than
// drawing each case's category uniformly at random) removes the multinomial
// category-draw variance that dominated dataset-to-dataset difficulty (the
// per-run score stddev scaled as sqrt(p(1-p)/n)). Every dataset now exercises
// the same balance of easy categories and routing traps, so a miner can't get a
// lucky-easy or unlucky-hard draw. Choosing n as a multiple of the category count
// gives a perfectly balanced set; otherwise the first n%C categories get one
// extra (deterministic, so it adds no between-run variance).
func stratifiedCategoryOrder(r *rand.Rand, n int) []int {
	nc := len(categories)
	order := make([]int, 0, n)
	base, rem := n/nc, n%nc
	for ci := 0; ci < nc; ci++ {
		count := base
		if ci < rem {
			count++
		}
		for k := 0; k < count; k++ {
			order = append(order, ci)
		}
	}
	r.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
	return order
}

// GenerateCases emits n raw tool cases from an existing RNG. Exported so the gen
// package can build tool cases from the same templated ground truth. seed drives
// both the stable case IDs and each result-usage prompt's fabricated needle
// subject (via toolexec.NeedleFor), which must match the needle the mock endpoint
// serves.
func GenerateCases(r *rand.Rand, seed int64, n int) []protocol.ToolCase {
	cases, _ := GenerateCasesWithFillers(r, seed, n)
	return cases
}

// GenerateCasesWithFillers is GenerateCases plus, for each case, the concrete
// entity ("filler") substituted into its template (empty for templates with no
// %s slot). The filler is the ground-truth entity the prompt is about, exposed so
// a caller can assert the prompt and the scored ground truth stay coupled.
func GenerateCasesWithFillers(r *rand.Rand, seed int64, n int) ([]protocol.ToolCase, []string) {
	if n < 1 {
		n = 1
	}
	order := stratifiedCategoryOrder(r, n)
	cases := make([]protocol.ToolCase, 0, n)
	fillers := make([]string, 0, n)
	for i := 0; i < n; i++ {
		cat := categories[order[i]]
		var tmpl string
		if cat.grammar != nil {
			tmpl = persona.Expand(r, cat.grammar, "root")
		} else {
			tmpl = cat.templates[r.Intn(len(cat.templates))]
		}
		caseID := protocol.OpaqueCaseID(seed, "tool", i)
		// Result-usage cases: the filler is the fixture needle's Subject, derived
		// from the SAME (seed, caseID) the mock server uses to serve the answer, so
		// the question ("...figure on the Veltrix index...") and the served fact
		// ("the Veltrix index reached 3,418 points") are always coherent.
		var filler string
		if IsResultUsage(cat.name) {
			filler = toolexec.NeedleFor(seed, caseID).Subject
		} else {
			filler = fillerFor(r, cat.name)
		}
		prompt := tmpl

		// Expected tool sequence: multi-hop tools, else the single tool, else none.
		seq := cat.tools
		if len(seq) == 0 && cat.tool != "" {
			seq = []string{cat.tool}
		}

		// argValue is the value pinned into RequiredArgs (defaults to the filler; an
		// intent variant overrides it). usedFiller is the load-bearing entity the
		// prompt is about ("" = the prompt has no such entity).
		argValue := filler
		usedFiller := ""
		useIntent := len(cat.intents) > 0 && r.Intn(2) == 0
		switch {
		case useIntent:
			it := cat.intents[r.Intn(len(cat.intents))]
			prompt = it.prompt
			argValue = it.value
			// Only a near-miss intent (value appears literally) records the token; a
			// pure-intent prompt has no load-bearing entity to track.
			if strings.Contains(it.prompt, it.value) {
				usedFiller = it.value
			}
		case strings.Contains(tmpl, "%s"):
			prompt = fmt.Sprintf(tmpl, filler)
			// A real tool case has a load-bearing entity; no_tool / abstention fillers
			// ARE the whole message.
			if len(seq) > 0 {
				usedFiller = filler
			}
		}
		// Surface variation for wrap-flagged categories: the shared Tier-1a
		// lead-in/trailer engine, protecting a leading pinned-argument token from
		// the lowercase pass and leaving the intent untouched.
		if cat.wrap {
			prompt = persona.Wrap(r, prompt, argValue, promptLeadIns, promptTrailers)
		}

		tc := protocol.ToolCase{
			ID:              caseID,
			Category:        cat.name,
			Prompt:          prompt,
			AllowExtraTools: cat.allowExtra,
		}

		switch {
		case len(seq) == 0:
			tc.ExpectedTools = nil
			tc.MaxToolCalls = 0
			switch cat.name {
			case "abstention":
				tc.ExpectedBehavior = "answer or abstain without calling any tool"
			case "arg_hallucination":
				tc.ExpectedBehavior = "ask for the missing detail; do not call a tool with a fabricated argument"
			default:
				tc.ExpectedBehavior = "respond conversationally without calling any tool"
			}
		case len(seq) == 1:
			ts := protocol.ToolSpec{Name: seq[0]}
			// Exact-value arg ground truth: the filler for a verbatim case, or the
			// resolved value for an intent case (which may not appear in the prompt).
			if cat.argKey != "" && argValue != "" && (useIntent || usedFiller != "") {
				ts.RequiredArgs = map[string]string{cat.argKey: argValue}
			}
			tc.ExpectedTools = []protocol.ToolSpec{ts}
			tc.MaxToolCalls = 1
			tc.ExpectedBehavior = fmt.Sprintf("call %s exactly once", seq[0])
		default:
			tools := make([]protocol.ToolSpec, len(seq))
			for j, tn := range seq {
				tools[j] = protocol.ToolSpec{Name: tn}
			}
			tc.ExpectedTools = tools
			tc.MaxToolCalls = len(seq)
			if cat.unordered {
				tc.Unordered = true
				tc.ExpectedBehavior = "call " + strings.Join(seq, " and ") + " (in any order)"
			} else {
				tc.ExpectedBehavior = "call " + strings.Join(seq, " then ") + " in that order"
			}
		}

		cases = append(cases, tc)
		fillers = append(fillers, usedFiller)
	}
	return cases, fillers
}
