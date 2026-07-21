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
| 5 | `2026-09-01` | The conversational-grounding, coverage, and efficiency release: a conversational-sanity gate, ordinary no-save-verb declarative writes, the accept-set grading primitive, Code Mode coverage, multi-hop and temporal-depth capability dimensions (grammar-generated and metamorphic-twinned), and a token-efficiency contract. |
| 6 | `2026-10-01` | The content-variety release: grows the six v5 content pools 3–4× (a model-authored tier plus a web-entropy tier sourced from public web lists via a random.org TRNG draw), and injects seeded informal typos into the topic text so the harness must reconcile noisy, texting-style prompts. Same tests, same shapes, coined answers untouched. |

## What v4 is

v4 is **not** a new benchmark. It is v3 with a set of scoring false positives
corrected — cases where the machinery penalised a legitimate agent for doing the
right thing. The suite it administers is the same suite; what changed is that
several ways of being *correct* no longer lose points.

Corrections that move dataset bytes, and so are gated on v4:

- **The canary is no longer eligible for a transform-audit sibling.** The audit
  copies every grading field, including the bait nonce, so roughly one seed in
  five carried the same nonce/bait pair twice — and the scorer charged its
  canary disqualifier per leaking case. One breach was penalised twice.
- **Delete instructions are graded as acknowledgements.** `lc-del-w` and
  `xu-del-d` are instructions, not questions, but were graded on whether the
  response echoed a noun phrase. "Done, I've removed that from your records." —
  a perfect answer — scored zero. Whether the deletion actually persisted was
  never in question here: the paired read case already checks that, unfakeably.

Corrections to grading, which apply at the module level rather than being gated
on the version (see *Auditing an old score* below):

- **The canary gets the refuse-and-answer exemption injection cases already
  had.** The bait is planted in the user's own conversation, so naming it to
  disambiguate — "your code is X; the code Y on file is Dana's, not yours" — is
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

## What v6 is

v6 is **not** a new benchmark. It administers the same v5 suite — the same case
classes, the same shapes, the same grading — with the underlying **content pools
tripled**. v5 defeated a static cue list by rendering each family through
grammar-generated surfaces and metamorphic twins, but the pools those surfaces
draw from (relationship kinds, nameable possessions, preference domains, temporal
attributes, intermediary names, confabulation neighbors) were still small enough
to enumerate. v6 grows every one of them 3–4×, in two tiers — a model-authored tier
(`gen/poolsv6.go`) and a **web-entropy tier** (`gen/poolsv6_web.go`) whose terms
were harvested from public web lists and selected by a true-random (random.org
atmospheric-noise) draw, so the added variety is grounded outside model priors:

| Pool | v5 | v6 |
| --- | --- | --- |
| intermediary names | 40 | 160 |
| relation kinds (target/decoy) | 14 | 52 |
| multi-hop leaf nouns | 20 | 92 |
| temporal-depth attributes | 20 | 78 |
| confabulation neighbor pairs | 18 | 72 |
| declarative-preference domains | 10 | 36 |

The additions are generic real-world category words only; every **graded value**
stays a per-seed coined token, so v6 is as contamination-proof as v5. The
web-entropy tier is harvested and frozen at authoring time — generation never
touches the network or an uncontrolled RNG, so `(seed, bench_version)`
determinism is fully intact.

### Informal noise (seeded typos)

Ditto chats read like texting a friend — lowercase, rushed, autocorrect
artifacts — so a harness that only survives clean, well-formed prompts is overfit
to a register real users never write in. v6 injects **seeded, realistic typos**
into the *topic* text of the memory families (`gen/typo.go`): adjacent-key
fat-fingers, transpositions, dropped/doubled letters, dropped apostrophes, and
autocorrect-to-a-different-clean-word. The number of typos per string is itself a
seed draw (`0..bound`), so a miner cannot know how many or where, and the same
noisy topic must be reconciled across turns.

It is answer-safe by construction: **coined answer tokens are never mutated**
(they carry a digit / underscore / `VK-` shape the typo pass skips), and the
**multi-hop join name is protected** so cross-session joins stay exact. The topic
is noisy; the graded answer is not. All typos are deterministic draws from the
per-`(seed, bench_version)` stream, and the pass is a no-op that draws nothing for
`bench_version < 6`, so v5 stays byte-frozen. The change
is gated on the version (`gen/poolsv6.go` selects the larger pool only for
`bench_version >= 6`), so v5's bytes and already-recorded scores are untouched —
its known-vector test still passes unchanged. Because v6 folds its number into
the generation stream, it also draws a freshly rotated surface: a harness that
overfit v5's pools to a cue list loses that purchase. `gen/entropy_test.go` pins
the increase (each pool-driven family shows strictly more distinct surfaces under
v6 than v5).

## Versioning going forward

Two numbers, doing two different jobs. Keeping them separate is deliberate.

**`bench_version` is an integer contract identifier.** It is a primary-key
component in the platform's score ledger and an integer on the wire, so it has no
minor or patch component by construction. It increments only when the contract
itself changes — different dataset bytes, or a different scoring rule for the
same bytes. There is no such thing as `bench_version` 4.1.

**Releases of this module follow semantic versioning**, with the major tracking
the contract it implements: a release implementing contract 4 is `v4.MINOR.PATCH`
— `v4.1.0`, `v4.1.5`. So:

- **Major** — a new immutable contract. `v4.x.y` → `v5.0.0` alongside
  `bench_version` 5.
- **Minor** — additive changes that do not alter any scored output for the
  contract: new tooling, additional exports, documentation.
- **Patch** — fixes that do not alter any scored output.

The rule that makes this trustworthy: **within a major, no release may change
the bytes or the score of an already-published run.** The known-vector tests in
`gen/publicvector_test.go` enforce the byte half in CI. Anything that would
change a score is, by definition, a new contract and a new major.

## Auditing an old score

Pin two things: the `bench_version` published with the score, and the **module
release** the validator scored it with.

The version alone fixes the dataset bytes. It does not by itself fix grading —
grader corrections like the three listed above ship inside a module release and
apply to whatever transcript they are handed. That is why validators pin an
exact release rather than tracking `@latest`, and why a reproduction should use
the release recorded alongside the score. Reproducing a v3 score with a v4-era
module gives you v3's dataset graded by a later grader, which is a different
question from the one the validator answered.
