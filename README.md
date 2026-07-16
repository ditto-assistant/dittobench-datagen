# dittobench-datagen

The deterministic dataset generator for [DittoBench](https://github.com/ditto-assistant/dittobench-starter-kit),
the benchmark for Bittensor Subnet 118. This is the exact code the validators use
to build the dataset every submission is scored against. It is published so the
benchmark is fully auditable.

## Why this is public

DittoBench grades AI-agent submissions against a freshly generated dataset. The
generator is fully deterministic and non-LLM. A given `(seed, bench_version)`
always produces the identical dataset, byte for byte. The seed for each scored
submission is derived from an on-chain block hash fixed after the miner commits
their submission, so it cannot be predicted or pre-computed against. That is what
makes it safe to open the generator without enabling overfitting.

With this repo you can:

- Regenerate any scored submission's dataset. Take the seed the platform
  published for a submission and reproduce the exact dataset, including answer
  keys, it was graded against.
- Re-grade independently, verifying a leaderboard score yourself instead of
  trusting the validators.
- Inspect the anti-overfit machinery: how cases, distractors, memory graphs, and
  tool fixtures are constructed.

Grading is deterministic and judge-free. The memory grader (`grade/`) is included
so you can re-score a published transcript yourself; the platform's
composite-scoring service and leaderboard are separate and not part of this repo.

## Install

```bash
go install github.com/ditto-assistant/dittobench-datagen/cmd/generate@latest
```

or build from source:

```bash
git clone https://github.com/ditto-assistant/dittobench-datagen
cd dittobench-datagen
go build ./cmd/generate
```

## Usage

```bash
# regenerate a submission's dataset from its published seed
generate -seed 123456789 -run-size full -out dataset.json

# print only the dataset_sha256 (the hash the platform pins and validators verify)
generate -seed 123456789 -run-size full -sha

# a fresh random seed (printed on stderr) for exploration
generate -run-size small
```

`generate` writes the canonical JSON artifact to stdout (or `-out`) and prints
`seed`, `run_size`, `bench_version`, and `dataset_sha256` to stderr. That
`dataset_sha256` is the same hash the platform commits to and a validator
re-derives. If it matches the one published for a submission, you are looking at
the identical dataset that submission was scored against.

Run sizes: `small` (smoke), `medium`, `full` (the scored profile).

## Determinism guarantee

The same seed always yields the same bytes. The CI determinism test generates a
fixed seed twice and asserts the hashes match, and a known-vector test pins the
canonical hash for a fixed seed. The module has no external dependencies, only the
Go standard library, so a build from source reproduces the validators' bytes
exactly. The validators run this same module, so their bytes match yours by
construction.

The output is a function of `(seed, bench_version)`, not the seed alone. Each
version folds its number into the generation stream (`protocol.RotateSeed`), so
the rendered surface rotates when `bench_version` bumps. A harness that only
pattern-matches one version's rendered templates degrades on the next, while a
harness that actually reasons is unaffected. The rotation is public and
deterministic, so reproducibility is preserved: pin the seed and the version and
you get the same bytes every time.

## Layout

Every package is importable, so this module is the single source of truth for the
generator. The scoring and generation services depend on it rather
than keeping their own copies.

- `cmd/generate`: the CLI entry point.
- `cmd/generate-service`: the same generation behind HTTP
  (`POST /generate?seed=&run_size=` → DatasetArtifact JSON + `X-Dataset-SHA256`
  header), with the `Dockerfile`/`cloudbuild.yaml` the SN118 platform deploys it
  from. The deployment is private (IAM-gated) so platform infrastructure cannot
  be farmed for generation, but there is no secret in the code — it computes
  exactly what `cmd/generate` computes for the same `(seed, run_size)`.
- `cmd/graderaudit`: the grader false-negative audit. Given an artifact and a
  JSONL transcript dump it emits a labeling sheet of every memory case that
  survived the disqualifying scans but failed the typed answer check, plus
  per-answer-kind counts, so the grader's measured false-negative rate can be
  published per bench version.
- `cmd/gstudy`: the offline reliability analyzer. Given a JSONL of scored runs
  it reports a G-study variance decomposition (seed vs. item vs. residual) and
  per-category difficulty/discrimination estimates, flagging saturated and floor
  categories; pure analysis, no LLM.
- `gen`: the generation pipeline (tool cases, memory suite, write-then-read
  lifecycle chains, isolation graphs, artifact assembly and hashing).
  `gen.GenerateDataset(seed, profile)` is the one entry point.
- `persona`, `datagen`: case content builders.
- `protocol`: the wire shapes, including the `DatasetArtifact` schema validators
  score.
- `toolexec`, `catalog`: the tool fixtures and tool catalog.

## License

MIT. See [LICENSE](LICENSE).
