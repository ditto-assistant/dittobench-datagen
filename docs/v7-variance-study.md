# DittoBench v7 variance study (seed-to-seed spread under the ~10x hardening)

**Question.** Bench_version 7 made the suite ~10x harder (new memory modules:
deepchain, deepjoin, near-miss abstention, tempcalc, composedinj; new tool
categories: negation-cue restraint, stale-context routing, link_chain,
job_chain_recovery; full profile scaled to 120 tool + ~131 memory cases). Does
that added difficulty widen the seed-to-seed composite spread enough to
threaten the leaderboard's fairness machinery (hysteresis bands, the ~0.007
protection margin, the `1.64*sqrt(SE_c^2 + SE_ch^2)` band, confirmation
seeds)?

**Answer (verdict up front): GO.** v7 widens the top-tier between-seed SD by
~6% (0.0179 → 0.0190 composite points; paired CRN noise floor 0.0259 →
0.0279, +7.7%). The pure dataset-difficulty (structural) component of that
spread is 0.0003 composite points in BOTH versions — negligible against the
0.007 margin — so essentially all of the spread is the agent's own per-case
error realization, which the confirmation-seed machinery already averages
away. No single case type dominates the variance (max ~7% share), no v7
category is saturated or floored for a competent agent, and v7 *lowers*
per-case composite leverage relative to v6 because the suite grew. The only
operational change recommended is sizing confirmation-seed counts ~15–20%
higher than their v6-calibrated values (tables below). No generator change is
needed and **no byte-vector change was made**: the v7 dataset vector pinned by
the difficulty release is untouched.

---

## 1. Methodology

Everything is deterministic and LLM-free, using the repo's own machinery:

- **Driver:** `cmd/vstudy` (added by this study; `go run ./cmd/vstudy -seeds
  300 -out <dir>`). For each bench version (6 and 7) it generates the `full`
  profile dataset for seeds 1..300 via `gen.GenerateDataset` and scores fixed
  strategies against every case with the real grader (`grade.Memory`).
- **Analyzer:** `cmd/gstudy` (pre-existing) consumes the JSONL that vstudy
  emits (`runs_v{6,7}_{strategy}.jsonl`) and reports the G-study variance
  decomposition and per-category difficulty/discrimination flags.
- **N = 300 seeds per version** (600 generated full datasets total; ~1 minute
  wall clock). Noise floor is zero by construction — re-running a seed
  reproduces the identical scores (`TestEvalSeedDeterministic`).

**Composite.** `0.5*tool_mean + 0.5*memory_mean`, the v2+ accuracy composite.
Validator-side gates (efficiency, metamorphic, transform audit) multiply this
downstream; for the fixed honest strategies simulated here they are 1.0, so
they are variance-neutral for this comparison.

**Strategies.**

- *Naive tiers* (replicated from `gen/v7difficulty_test.go`, the difficulty
  release's measurement): `parrot`, `overlap`, `recency`, `dump`, `abstain`
  memory strategies, each paired with the fixed keyword tool router.
- *`strong` (mid tier, the decision-boundary model):* the oracle response
  (`AnswerKind`-keyed canonical answer, verified score 1.0) downgraded by a
  deterministic per-case error draw — a hash of (version, seed, case id, salt)
  compared against a per-case-type error rate (near-zero on plain recall,
  ~0.10 on routine reasoning, 0.18 on v5/v6 hard classes, 0.20–0.30 on the v7
  deep classes and dependent tool chains; full table in `cmd/vstudy/main.go`).
  This stands in for a realistic top miner near the champion boundary
  (composite ~0.92 on v6, ~0.89 on v7). Two independent salts ("A"/"B") give
  two *equal-skill* agents for the paired CRN comparison.
- *`uniform` (control):* every case fails with the same p = 0.10, removing
  the per-type rate assumptions entirely.
- *`oracle` (ceiling):* canonical answers everywhere. Scored 1.0 on **every
  memory case of all 600 datasets** (0 oracle failures) — the
  answer-key-sufficiency invariant of `TestV7OracleSolvable` holds at 10x the
  seed count it pins.

**What is not modeled.** Real model stochasticity (temperature, retries) adds
run-to-run variance on top of these numbers for both versions equally; the
validator's v7 scoring levers (result-usage gate, mismatch penalties) act
per-case and are consistent with the 0/1 per-case scoring used here.

## 2. Composite spread per strategy (300 seeds each)

Quantiles are (p5, p50, p95). SD is the between-seed SD of the composite —
i.e. the SE of a single-seed score.

| strategy | v6 mean | v6 SD | v6 p5/p50/p95 | v7 mean | v7 SD | v7 p5/p50/p95 | SD change |
|---|---|---|---|---|---|---|---|
| parrot + router | 0.3326 | 0.0123 | .3125/.3337/.3511 | 0.2726 | 0.0094 | .2560/.2727/.2887 | −24% |
| overlap + router | 0.4354 | 0.0251 | .3945/.4353/.4759 | 0.3440 | 0.0178 | .3149/.3445/.3722 | −29% |
| recency + router | 0.3625 | 0.0246 | .3229/.3623/.4023 | 0.2731 | 0.0154 | .2493/.2730/.2990 | −37% |
| dump + router | 0.3331 | 0.0163 | .3066/.3317/.3599 | 0.2835 | 0.0140 | .2609/.2835/.3077 | −14% |
| abstain + router | 0.3583 | 0.0121 | .3385/.3577/.3749 | 0.3023 | 0.0094 | .2857/.3023/.3170 | −22% |
| uniform p=0.10 | 0.8985 | 0.0222 | .8615/.8980/.9328 | 0.8999 | 0.0191 | .8674/.9004/.9296 | −14% |
| **strong (top tier)** | **0.9187** | **0.0179** | .8888/.9184/.9474 | **0.8860** | **0.0190** | .8526/.8853/.9155 | **+6%** |
| oracle | 1.0000 | 0.0000 | — | 1.0000 | 0.0000 | — | — |

Readings:

- **The top-tier SD rises only 0.0179 → 0.0190 (+6% relative, +0.0011
  absolute).** The added per-case difficulty is offset by the larger suite
  (205 → ~252 cases): the `uniform` control — same skill, no difficulty-mix
  assumptions — actually *tightens* under v7 (0.0222 → 0.0191), exactly the
  `sqrt(N)` prediction. The +6% for `strong` is the difficulty *mix* effect
  (more weight on high-error case types), not dataset instability.
- Every naive tier scores lower AND spreads *less* under v7 — the hardening
  pushed them toward a floor, which compresses their variance. No fairness
  concern arrives from below.
- Mean drop for the fixed top tier (0.9187 → 0.8860) is the intended
  difficulty, not variance.

## 3. G-study decomposition (cmd/gstudy)

Per-case scores, seed facet vs item (category) facet vs residual:

| runs file | grand mean | seed frac | item frac | residual frac | dominant |
|---|---|---|---|---|---|
| v6 strong | 0.919 | 0.43% | 3.2% | 96.4% | residual |
| v7 strong | 0.886 | 0.35% | 5.7% | 93.9% | residual |
| v6 overlap (naive) | 0.443 | 0.24% | 61.1% | 38.7% | item |
| v7 overlap (naive) | 0.341 | 0.15% | 66.4% | 33.5% | item |

- For a competent agent both versions are **residual-dominated with a tiny
  seed component** — the classic "per-case Bernoulli noise" regime where
  averaging over cases/seeds works and CRN pairing helps. The seed facet
  *shrinks* under v7 (0.43% → 0.35%).
- For the naive tier the item facet dominates (and grows under v7): the
  benchmark discriminates by *category*, which is precisely the design intent
  of the difficulty release (parser-like harnesses collapse on the reasoning
  categories — see the `parser-like` trap-gap telemetry, 0.31 → 0.28 gap with
  plain-recall accuracy falling 0.48 → 0.43).

**Structural vs Bernoulli split** (strong tier; structural = SD of the
*expected* composite per seed, i.e. pure case-mix difficulty drift):

| | v6 | v7 |
|---|---|---|
| structural SD | 0.0003 | 0.0003 |
| Bernoulli SD (paired diff / √2) | 0.0183 | 0.0197 |
| total SD | 0.0179 | 0.0190 |

The stratified per-category quotas hold the dataset's expected difficulty
essentially constant across seeds in both versions: **structural spread is
~4% of the protection margin.** The memory mix does drift slightly per seed
(e.g. point-in-time 2–6 cases, multi-session 3–6, isolation 5–8 across seeds)
but its composite effect is the 0.0003 above — not worth a generator change.

## 4. Per-category flags and variance offenders

`cmd/gstudy` flags on the strong-tier runs: **no v7 category floors, and
nothing saturates beyond the plain-recall categories that are near-1.0 by
construction** (declarative-write 0.98 etc. — coverage/floor-catching cases,
intentionally easy). On the naive runs the new v7 classes floor
(lifecycle-deep-read, multi-hop-deep, near-miss-abstention,
temporal-arithmetic, link_chain, job_chain_recovery, stale-context…) — that
is the difficulty design working, not a reliability defect: the same
categories discriminate strongly for the mid tier.

Top shares of the strong tier's composite variance (Var of each category's
weighted contribution / total across categories):

| v6 | share | v7 | share |
|---|---|---|---|
| mem temporal-reasoning | 7.0% | mem temporal-reasoning | 6.9% |
| mem preference-application | 6.9% | mem multi-session | 5.2% |
| mem multi-session | 5.4% | mem preference-application | 5.1% |
| mem multi-hop-relational | 3.9% | mem isolation | 4.4% |
| mem temporal-depth | 3.5% | mem point-in-time | 4.1% |
| — | | tool job_chain_recovery_result_usage | 3.6% |
| — | | tool link_chain_result_usage | 3.3% |

- **No single case type dominates** (max ~7%, same as v6, and the top
  offenders are *pre-existing* v5/v6 types, not the new v7 classes; the
  largest new-class share is 3.6%).
- The five all-or-nothing result-usage tool chains jointly carry ~17% of
  category-level variance on ~21% of the tool suite — proportionate, not a
  design smell.
- Per-case composite leverage (one case flipping): v6 = 0.0052 per memory
  case / 0.0045 per tool case; **v7 = 0.0038 / 0.0042 — lower than v6**
  because the suite grew. No case is worth "multiple margins": the worst
  correlated block (a deep-chain write miss cascading into its read) swings
  ~2 cases ≈ 0.008, about one margin, and that is a *skill* signal, not
  noise.

## 5. Decision-boundary impact

Live-fold machinery assumed: protection margin **m = 0.007** composite
points; challenger band `1.64*sqrt(SE_ch^2 + SE_c^2)`; confirmation seeds run
champion and challenger on the same fresh seeds (CRN pairing).

**Equal-skill false-separation probability (single seed).** Two equal-skill
top-tier agents:

| comparison | v6 | v7 |
|---|---|---|
| same seed (CRN), P(\|Δ\| > m), empirical | 0.750 | 0.793 |
| same seed (CRN), normal approx (σ_diff = 0.0259 / 0.0279) | 0.787 | 0.802 |
| different single seeds (σ√2 = 0.0253 / 0.0269 per side) | 0.782 | 0.794 |

Single-seed comparisons at the 0.007 margin were already ~75–80% likely to
spuriously exceed the margin under **v6**; v7 moves that by ~1–4 points. The
margin only functions because of the SE band + confirmation seeds, and that
was equally true before the hardening.

**Confirmation seeds to resolve a true gap Δ** (paired CRN, one-sided
α = 0.05; "detect" = band exactly equals the gap, 50% power; "95% power" =
`n = ((1.645+1.645)·σ_diff/Δ)²`):

| true gap Δ | v6 detect | v6 95% power | v7 detect | v7 95% power |
|---|---|---|---|---|
| 0.005 | 73 | 292 | 85 | 337 |
| 0.010 | 19 | 73 | 22 | 85 |
| 0.020 | 5 | 19 | 6 | 22 |
| = margin 0.007 | 38 | 149 | 43 | 172 |

**Recommendation:** scale whatever confirmation-seed count was calibrated for
v6 by `(0.0279/0.0259)² ≈ 1.16` — i.e. **+15–20%** (e.g. 19 → 22 seeds to
detect a 0.01 gap; 73 → 85 for 95% power). SE-based bands that are computed
from the observed per-seed spread (rather than a hard-coded σ) need no change
at all — they will simply measure the slightly larger σ. If any component
still assumes the historical σ ≤ 0.01 design target (`cmd/benchcal`'s
`targetSigma`, dittobench-api), note that a *single-seed* top-tier σ of
~0.019 was already the v6 reality in this model; the 0.01 target only ever
described the structural half, which remains ~0.0003.

## 6. Fix assessment (none required, none made)

The trigger for a variance-reducing change — "a single case type dominates,
or one case flips the composite by multiple points" — is not met:

- max category share ~7% (v6-parity), largest *new* v7 class 3.6%;
- structural (dataset difficulty) SD 0.0003 in both versions, so pinning the
  drifting memory-mix counts (the only candidate generator change, e.g.
  fixed quotas for point-in-time/multi-session/isolation) would buy at most
  0.0003 composite points of SD while re-pinning the v7 byte vector;
- per-case leverage went *down* v6 → v7;
- partial-credit restructuring of the deep chains would reduce the intended
  difficulty (the all-or-nothing final-value contract is the anti-grep
  design) for at most a 2-case-correlation gain.

Accordingly this study ships **measurement tooling only** (`cmd/vstudy` +
tests). **No dataset-generation bytes were touched; the v7 known vector is
unchanged.**

## 7. Reproduction

```
go test ./cmd/vstudy ./cmd/gstudy          # driver invariants
go run ./cmd/vstudy -seeds 300 -out /tmp/vstudy-runs > /tmp/vstudy.json
go run ./cmd/gstudy -in /tmp/vstudy-runs/runs_v7_strong.jsonl
```

Caveats: the strong tier's per-type error rates are documented modeling
assumptions (see `memErrRates`/`toolErrRates`); the `uniform` control shows
the headline conclusion (v7 does not meaningfully widen top-tier spread) is
insensitive to them. Real-model temperature noise is additive to both
versions and is exactly what CRN confirmation seeds cancel.

**Verdict: GO** — ship v7 at current variance; bump confirmation-seed counts
~15–20% and leave the 0.007 margin and SE-band machinery as they are.
