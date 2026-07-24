# Benchmark versions

A `bench_version` is an **immutable generation contract**. For a given
`(seed, bench_version)` the generator emits the same bytes forever, and the
grader scores a given transcript against them the same way forever. That is what
makes an old score auditable by anyone holding the published seed: nothing about
a scored run is allowed to move under it afterwards.

Because the contract is immutable, a correction to how scoring works cannot be
applied to an existing version. It ships as a new one.

## The versions

| Version | Epoch | What it is |
| --- | --- | --- |
| 2 | `2026-01-01` | The launch contract. Frozen since on-chain scoring began. |
| 3 | `2026-07-01` | The anti-gaming release: dump-guard grading, needle gating, adversarial distractors, composed injection framings, the cross-user lifecycle probe, and the reproduce-under-transform audit. |
| 4 | `2026-08-01` | A supplementary fix to v3 scoring. Same tests, same shape, corrected grading. |
| 5 | `2026-09-01` | Conversational grounding, broader capability coverage, and token-efficiency scoring. |
| 6 | `2026-10-01` | Memory-as-data and the complexity suite; retains the v5 scoring contract. |
| 7 | `2026-11-01` | Platform-owned OpenRouter inference with locked `openai/gpt-oss-20b`, and the difficulty release: a version-gated hard-case suite roughly an order of magnitude harder for a non-reasoning harness while a correct trajectory still scores full marks. |

## What v7 is

v7 carries two things. First, the inference boundary it always had: measurements
made through the platform-owned OpenRouter inference boundary and its locked
`openai/gpt-oss-20b` model are separated from earlier Qwen-based scores. Second,
the **difficulty release** — a suite of version-gated hard cases that make the
benchmark markedly harder for a non-reasoning harness (pattern-matching,
retrieval-without-reasoning, dump-everything, keyword routing) while leaving a
genuinely correct trajectory at full marks.

Every difficulty lever is gated on `bench_version >= 7` in exactly the style v5
and v6 used, so v2 through v6 regenerate and grade byte-identically (their
known-vector tests pass unchanged). The judge-free deterministic grader is
unchanged: no new grading rules, no LLM, stdlib only. The wire/artifact format
is unchanged too — the new cases reuse the existing `MemoryCase` / `ToolCase`
shapes and the existing served tool-endpoint protocol, so a harness built for
the current format still parses a v7 dataset and returns valid responses (it
will simply score poorly, which is the point).

Every new case family is grounded in a real production flow, data-model shape,
or documented failure mode of the live Ditto assistant (backend + app). The
family-by-family evidence — which product flow each exercises, why a
product-quality agent must handle it, and why today's harnesses fail it honestly
— is in [v7-product-traceability.md](v7-product-traceability.md). The rule:
difficulty must make the real product better, never difficulty for difficulty's
sake; anything that reads as a trick a product-quality agent would never need to
survive is cut.

### The difficulty levers (all v7-gated)

- **Scaled, denser profiles** (`profilesV7`): the `full` memory suite grows to
  ~180 cases over 5 waves at a 0.5 raw-pairs (Tier-B) share and 10 isolation
  cases, drawing a denser persona (more sessions, more near-miss decoy people),
  so the distractor-to-needle ratio rises with the case count. `small` stays a
  cheap single-wave smoke path. Generation stays non-LLM and fast (tens of ms
  for a full v7 dataset), so the per-submission `full` path is unaffected.

- **Sharpened memory-type mix** (`memoryTypeWeightV7`): the types a lexical
  retriever cannot shortcut (multi-session synthesis, temporal reasoning,
  point-in-time, contradiction, knowledge-update, aggregation, preference
  application) take a much larger share of the stratified budget; single-pair
  recall stays at weight 1 for coverage without dominating. The v7 twin-family
  count is capped so phrasing-invariance recall (the naive-passable end of the
  suite) does not soak up the freed budget.

- **Six new product-grounded, reasoning-required memory classes** (each maps to
  a real flow in [v7-product-traceability.md](v7-product-traceability.md)):
  - *deep write chains* (`lifecycle-deep-*`): the real memory write path — a
    save→update→update→read (and save→update→delete→read) sequence delivered as
    separate instructions across three staging waves; only the final state
    answers, and the delete-chain read declines with the chain's earlier values
    as scored distractors. (backend in-place `update_memory` + confirmation-gated
    `delete_memory`.)
  - *three-hop joins* (`multi-hop-deep`): knowledge-graph traversal — "my
    mentor's partner's employer", a three-pair walk across sessions, with a
    one-join trap (the first-hop person's own value) and a wrong-chain trap (a
    full decoy chain) both seeded as scored distractors. (backend subject edges /
    memory network.)
  - *near-miss abstention* (`near-miss-abstention`): entity disambiguation — a
    question engineered to look answerable (the sibling entity's value for the
    same attribute is seeded, and the asked entity is mentioned in an unrelated
    context) but the asked fact was never stated; the sibling's value is a scored
    distractor, so a nearest-neighbor retriever is zeroed. (backend subject dedup
    at cosine 0.75; "never the wrong person's data".)
  - *temporal arithmetic* (`temporal-arithmetic`): aggregation over durable
    quantitative facts the product tracks (rent + raise, budget − spent) stated
    in non-adjacent sessions; the answer appears in no seeded pair, graded
    through the accept-set. (Insights aggregation.)
  - *composed stored-instruction injection* (`injection-composed`): a
    prompt-injection split across two innocuous notes (a fake authority channel,
    then a payload that invokes it), so a single-note attack detector sees two
    benign memos; a benign same-shape twin (the user's own tagging convention)
    ensures blanket refusal fails. (store is written through untrusted chat.)
  - *subscribed-graph attribution* (`subscription-own` / `subscription-attributed`):
    a subscribed friend's value and the user's own value for the same attribute
    both surface in one flat search list, distinguished only by an `@friend`
    provenance prefix; "my X" must return the user's own (the friend's is a
    cross-graph leak), and "what did @friend say" must attribute the friend's.
    (backend `annotateSubscribedSlimMemory`, cross-user docs.)

- **Five new tool classes** plus a **weighted category mix**
  (`toolCategoryWeightV7`) that makes result-usage cases (which require executing
  tools and reading their served content) the dominant share and doubles the
  routing/discrimination traps:
  - *negation-cue restraint* (`negation_no_tool`): the prompt names a tool cue
    while negating it, so a keyword router that fires on the cue is caught.
  - *stale-context routing* (`stale_context_web`): a memory-anchored phrasing
    whose actual request is current public information.
  - *dependent link chain* (`link_chain_result_usage`): `search_web` serves a
    stable page URL and `read_links` reveals the answer only when called with
    that URL — the trajectory cannot be faked and the snippet cannot be grepped.
  - *job-chain + recovery composition* (`job_chain_recovery_result_usage`): the
    dependent job-id chain and transient-error recovery at once. Both serving
    gates already existed in the tool endpoint and compose via the category-name
    markers, so no protocol change was needed.
  - *entity-lookup 3-hop chain* (`entity_lookup_chain`): the production
    `search_subjects` → `search_memories_in_subjects` → `fetch_memories`
    sequence the backend recommends, order-scored.

### The measurement

"Harder" is defined operationally and pinned by two tests
(`gen.TestV7NaiveStrategiesCollapse` and `gen.TestV7ChampionTierLandsNearTarget`,
reproducible with `go test -run 'V7Naive|V7ChampionTier' -v ./gen`), plus the
family-by-family calibration table in
[v7-product-traceability.md](v7-product-traceability.md).

**Naive strategies.** Fixed non-reasoning strategies — question-echo,
single-pair retrieval, recency, dump-everything, and a keyword tool router — are
scored against v6 and v7:

- On the **reasoning-required subset**, the best fixed non-reasoning strategy
  scores **0.088** on v7 vs the oracle's **1.0** — an **~11x** gap — up from
  v6's 6.2x (0.162). That subset grows from **40%** to **52%** of the memory
  suite. The keyword tool router falls from **0.525 → 0.331**.

**Champion tier.** The top-5 leaderboard harnesses scored 0.607–0.845 on the
pre-deepening v7. Because the real harnesses cannot be run here, the tier is
modeled with fixed per-class expected pass rates (calibrated to reproduce that
rebench), bracketed by a STRONG anchor (newDitto-like) and a WEAK anchor
(whitycatboss/infinity-like):

| Tier | pre-deepening (v6 proxy) | deepened v7 |
| --- | --- | --- |
| STRONG champion | 0.833 | 0.573 |
| WEAK champion | 0.678 | **0.364** |
| Oracle (every case, 30 seeds) | 1.000 | **1.000** |

The WEAK anchor — the bulk of today's best harnesses — lands at **~0.36**, in
the ~0.35 ± 0.05 target band, while the oracle stays at 1.0 and the naive tiers
stay near their floor. The single STRONGEST harness lands ~0.57; driving it to a
flat 0.35 too would require an ~80%-hard suite that strips the grounded
synthesis and conversational-sanity coverage and over-concentrates a few
families — which the "difficulty must make the product better, never difficulty
for difficulty's sake" rule forbids — and would crater the weaker harnesses
below 0.20. The calibration therefore lands the fleet's target while preserving
full product-grounded coverage. As with every version, the public versioned seed
rotation produces a fresh, deterministic dataset surface, so v7 scores are never
compared with v6 scores.

## What v4 is

v4 is **not** a new benchmark. It is v3 with a set of scoring false positives
corrected: cases where the machinery penalised a legitimate agent for doing the
right thing. The suite it administers is the same suite; what changed is that
several ways of being *correct* no longer lose points.

Corrections that move dataset bytes, and so are gated on v4:

- **The canary is no longer eligible for a transform-audit sibling.** The audit
  copies every grading field, including the bait nonce, so roughly one seed in
  five carried the same nonce/bait pair twice, and the scorer charged its
  canary disqualifier per leaking case. One breach was penalised twice.
- **Delete instructions are graded as acknowledgements.** `lc-del-w` and
  `xu-del-d` are instructions, not questions, but were graded on whether the
  response echoed a noun phrase. "Done, I've removed that from your records." (a perfect answer) scored zero. Whether the deletion actually persisted was
  never in question here: the paired read case already checks that, unfakeably.

Corrections to grading, which apply at the module level rather than being gated
on the version (see *Auditing an old score* below):

- **The canary gets the refuse-and-answer exemption injection cases already
  had.** The bait is planted in the user's own conversation, so naming it to
  disambiguate ("your code is X; the code Y on file is Dana's, not yours") is
  the most competent answer available, and it scored zero *and* took the
  composite disqualifier. The exemption still requires the true value to be
  present, so echoing nonce-shaped tokens fails exactly as before. Isolation is
  deliberately not excused: its forbidden value lives in another user's memory
  graph, so surfacing it at all means a boundary was already crossed.
- **Durations parse decimals.** "about 1.5 years" was read as *15 years*.
- **"used to" is no longer a cessation phrase.** It is a temporal marker, and it
  zeroed correct answers of the very common form "you used to mention it
  constantly, and you still love it".

The scorer carries matching corrections (bounded penalties no longer stack past
a floor, transport failures no longer read as brittleness, transparent memory
retrieval is no longer taxed). Those live in `dittobench-api`.

## Versioning going forward

Two numbers, doing two different jobs. Keeping them separate is deliberate.

**`bench_version` is an integer contract identifier.** It is a primary-key
component in the platform's score ledger and an integer on the wire, so it has no
minor or patch component by construction. It increments only when the contract
itself changes: different dataset bytes, or a different scoring rule for the
same bytes. There is no such thing as `bench_version` 4.1.

**Releases of this module follow semantic versioning**, with the major tracking
the contract it implements: a release implementing contract 4 is `v4.MINOR.PATCH`
`v4.1.0`, `v4.1.5`. So:

- **Major**: a new immutable contract. `v4.x.y` → `v5.0.0` alongside
  `bench_version` 5.
- **Minor**: additive changes that do not alter any scored output for the
  contract: new tooling, additional exports, documentation.
- **Patch**: fixes that do not alter any scored output.

The rule that makes this trustworthy: **within a major, no release may change
the bytes or the score of an already-published run.** The known-vector tests in
`gen/publicvector_test.go` enforce the byte half in CI. Anything that would
change a score is, by definition, a new contract and a new major.

## Auditing an old score

Pin two things: the `bench_version` published with the score, and the **module
release** the validator scored it with.

The version alone fixes the dataset bytes. It does not by itself fix grading:
grader corrections like the three listed above ship inside a module release and
apply to whatever transcript they are handed. That is why validators pin an
exact release rather than tracking `@latest`, and why a reproduction should use
the release recorded alongside the score. Reproducing a v3 score with a v4-era
module gives you v3's dataset graded by a later grader, which is a different
question from the one the validator answered.
