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

The judge (the LLM that grades open-ended response quality) and the scoring engine
are separate components and are not part of this repo.

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

## Layout

Every package is importable, so this module is the single source of truth for the
generator. The private scoring service and generate service depend on it rather
than keeping their own copies.

- `cmd/generate`: the CLI entry point.
- `gen`: the generation pipeline (tool cases, memory suite, isolation graphs,
  artifact assembly and hashing). `gen.GenerateDataset(seed, profile)` is the one
  entry point.
- `persona`, `datagen`: case content builders.
- `protocol`: the wire shapes, including the `DatasetArtifact` schema validators
  score.
- `toolexec`, `catalog`: the tool fixtures and tool catalog.

## License

MIT. See [LICENSE](LICENSE).
