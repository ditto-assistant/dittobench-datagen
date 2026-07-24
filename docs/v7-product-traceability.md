# v7 difficulty suite — product traceability

Every v7 case family traces to a real Ditto product flow, data-model shape, or
documented failure mode in the production repos (`backend`, `ditto-app`). The
rule the operator set: difficulty must make the real product better, never
difficulty for difficulty's sake — anything that reads as a trick a
product-quality agent would never need to survive is cut. This table is the
evidence for each family.

Paths are relative to the sibling production repos
(`/Users/peyton/code/ditto/backend`, `/Users/peyton/code/ditto/ditto-app`).

## Memory families

| Case family (question_type) | Generator | Product flow / data shape it exercises | Evidence (repo:file) | Why a product-quality agent must handle it | Why today's harnesses fail it honestly |
| --- | --- | --- | --- | --- | --- |
| Deep write chains (`lifecycle-deep-write` / `lifecycle-deep-read`) | `gen/deepchain.go` | The real memory write path: `save_memory` then a series of `update_memory` mutations then `delete_memory`, each an in-place mutation with no version history; "forget that" must actually persist. | backend `pkg/mcp/tools.go` handleSaveMemory, `memory_update.go` (in-place `UpdateMemoryPairContent`, optimistic `revision`/`baseRevision`), `memory_delete.go` (confirmation-gated hard delete); ditto-app `src/hooks/useMemoryDeletion.tsx` (delete cascade), failure-mode #5 (delete no-op) | Users correct facts repeatedly across sessions and expect the latest state, and expect deletions to stick. | A shallow last-write-wins or single-step store passes single mutations but drops one mutation in a multi-step chain, or reports a delete it never persisted. |
| Three-hop relational joins (`multi-hop-deep`) | `gen/deepjoin.go` | Knowledge-graph traversal across subjects/edges: "my mentor's partner's employer" is a 3-pair walk no single memory contains. | backend `pkg/services/sync/graph.go` (subject_edges cooccur/semantic), `subject_edges.go` / `get_memory_network`; ditto-app Atlas/`MemoryMapView`; product doc "chaining up to 5 levels deep" | Ditto promises connected memory across people/places/things; multi-hop recall is the KG's whole point. | A single-hop retriever grabs the first-hop person's own value (one-join trap) or the wrong relative's leaf (wrong-chain trap); both are seeded as scored distractors. |
| Near-miss abstention (`near-miss-abstention`) | `gen/nearmiss.go` | Entity disambiguation: a sibling entity's value for the same attribute is stored, and the asked entity is mentioned in an unrelated context, so both attribute and entity words retrieve — but the asked fact was never stated. | backend failure-mode #3 (subject dedup at cosine 0.75 merges/forks near-duplicate entities), #9 (noise-gate recall miss); ditto-app "never show the wrong person's data" (`c10335a3`, cross-graph View 404 fixes) | Confidently declining when the specific fact is absent (rather than reporting the closest neighbor) is core to a trustworthy memory. | A nearest-neighbor retriever reports the sibling's value; it is a scored distractor, so the case is zeroed, not merely uncredited. |
| Temporal arithmetic (`temporal-arithmetic`) | `gen/tempcalc.go` | Aggregation/derivation over durable quantitative facts stated in different sessions (rent + raise, budget − spent, miles + earned) — the kind of running totals Insights derives. | backend recency is an MLP retrieval signal (`docs/learned-retrieval-weights.md`); ditto-app `InsightsView` ("insights always link back to source memories"); backend failure-mode #6 (date/tz) | Users ask "how much is left / what's the total now"; the answer is a function of several stored facts, not any single one. | The answer appears in no seeded pair; echoing either stored input fails the accept-set, so a retrieve-and-echo harness cannot pass. |
| Composed stored-instruction injection (`injection-composed` / `composed-note-benign`) | `gen/composedinj.go` | Memory-as-data vs memory-as-instructions when the store is attacker-influenceable: a fake authority channel note plus a payload note that invokes it, split across sessions. | backend `pkg/mcp/knowledge_graph_subscriptions.go` (deny-by-default so a third party can't flip the graph), `memory_publish.go` privacy scanner; the memory store is written through ordinary chat | The agent must treat retrieved memory as data and honor the user's own first-person statement over an unverified in-band "authority". | A single-note injection detector sees two individually-benign memos; the benign same-shape twin (the user's own tagging convention) catches a harness that blanket-refuses anything instruction-shaped. |
| Subscribed-graph attribution (`subscription-own` / `subscription-attributed`) | `gen/subscription.go` | Subscribed graphs are merged into one flat search result list, distinguished only by a rewritten `@<friend>` provenance prefix. | backend `pkg/tools/memory/tools.go` `annotateSubscribedSlimMemory` (`subscribed_graph:` / `@username / ...`), fan-out over own+subscribed KGs, failure-mode #7, `docs/CROSSUSER_EXAMPLES.md`; ditto-app `SubscribePanel` ("included in memory search"), graph-scoping fixes `8668a17b`/`37cce44d` | The agent must answer the user's own value for "my X", and attribute a friend's value when asked about the friend — never confuse the two. | A retriever that returns the highest-similarity hit regardless of ownership leaks the friend's value (scored as a cross-graph leak) or answers about the friend with the user's own value. |
| Entity-lookup 3-hop tool chain (`entity_lookup_chain`, tool suite) | `datagen/datagen.go` | The exact recommended production sequence: `search_subjects` → `search_memories_in_subjects` → `fetch_memories`. | backend `pkg/mcp/server.go` tool descriptions (explicit recommended sequence), `tool_catalog.go` | Entity-focused recall in production is a subject-first, then within-subject, then fetch-full traversal. | A harness that calls `search_memories` directly, or gets the 3-hop order wrong, fails the order-scored trajectory. |

The v5/v6 hard families this release grows (multi-hop-relational, temporal-depth,
multi-query-recall, non-verbatim-computed, passive-consolidation) and the
persona-derived synthesis classes it retains (multi-session, temporal-reasoning,
contradiction, knowledge-update, point-in-time, aggregation, computed-answer) are
the same flagship "builds understanding across sessions" behaviors the product
document leads with; v7 raises their share rather than inventing new gimmicks.

## Tool families

| Case family | Generator / serving | Product flow it exercises | Evidence | Why harnesses fail it |
| --- | --- | --- | --- | --- |
| Negation-cue restraint (`negation_no_tool`) | `datagen/datagen.go` | The user names a tool cue while negating it ("don't search the web — just from general knowledge"). | ditto-app composer recall toggle (`SendMessage.tsx`), `SettingsProposalCard` restraint | A keyword router that fires on the cue word calls a tool it was told not to. |
| Stale-context routing (`stale_context_web`) | `datagen/datagen.go` | Memory-anchored phrasing whose actual request is current public info. | backend route split search_memories vs search_web; ditto-app `about.recall` | A router that keys on "I told you / my notes" routes to memory instead of the live web. |
| Dependent link chain (`link_chain_result_usage`) | `toolexec/toolexec.go` link-chain gate | `search_web` returns a page URL; `read_links` must be called with that URL to get the answer. | backend result-usage / observed-execution model (`toolexec` doc) | The snippet carries a scored decoy; the harness must read the served URL, not grep the snippet. |
| Job-chain + recovery (`job_chain_recovery_result_usage`) | `toolexec/toolexec.go` (composed gates) | `execute_agent_job` → poll `get_agent_job_status` (transient error → retry) → answer. | backend `pkg/mcp/agent_tools.go` (job dispatch + status poll + workspace read) | The status call flakes once and gates the answer on the served job id; a harness that neither retries nor threads the id fails. |

## What was deliberately NOT added (would read as a trick)

- No riddles, trivia, or puzzles unrelated to a memory/tool flow.
- No adversarial unicode / prompt-parsing gotchas — a product agent never needs
  to survive them.
- No answers that depend on a hidden marker or fixed position a real agent could
  not infer from meaning (the served needle sentence already rotates its
  connective/position; the tool chains gate on real served values, not tells).
- Temporal-arithmetic domains are limited to durable personal quantities the
  product actually tracks (rent, PTO, savings, budgets, miles, pages, steps,
  course modules) — not arbitrary math word problems.

## Difficulty calibration (measured, not asserted)

The champion-equivalent tier is modeled with fixed per-class expected pass rates
calibrated to the operator's rebench (top-5 harnesses scored 0.607–0.845 on the
pre-deepening v7). Two anchors bracket the fleet: a STRONG anchor
(newDitto-like) and a WEAK anchor (whitycatboss/infinity-like). See
`gen.TestV7ChampionTierLandsNearTarget` (reproduce with
`go test -run V7ChampionTier -v ./gen`) and `gen.TestV7NaiveStrategiesCollapse`.

| Tier | pre-deepening (v6 proxy) | deepened v7 |
| --- | --- | --- |
| STRONG champion (newDitto-like) | 0.833 | 0.573 |
| WEAK champion (whitycatboss/infinity-like) | 0.678 | **0.364** |
| Best fixed non-reasoning strategy (whole suite) | 0.342 | 0.341 |
| Best fixed non-reasoning strategy (reasoning subset) | 0.162 | 0.088 |
| Oracle (canonical answer, every case, 30 seeds) | 1.000 | **1.000** |

Per-harness projection of the modeled strong-anchor drop (≈0.26) applied to the
rebench composites: newDitto 0.845→0.58, cliM@X 0.800→0.54, ditto-agent-v2
0.799→0.54, infinity 0.634→0.37, whitycatboss 0.607→0.35.

The WEAK anchor — the bulk of "today's best harnesses" — lands at ~0.36, in the
operator's ~0.35 ± 0.05 target band, while the oracle stays at 1.0 and the naive
tiers stay near their floor. The single STRONGEST harness lands ~0.57: driving it
to a flat 0.35 too would require an ~80%-hard suite that removes the grounded
synthesis and conversational-sanity coverage above and over-concentrates a few
families — which the "difficulty must make the product better, never difficulty
for difficulty's sake" directive forbids — and would crater the weaker harnesses
below 0.20, which is worse benchmark design. The calibration therefore lands the
fleet's target while preserving full product-grounded coverage.
