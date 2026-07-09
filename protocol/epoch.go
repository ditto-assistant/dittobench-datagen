package protocol

import "time"

// BenchVersion is the scoring benchmark version stamped into every run's
// details. Bump it with every scoring-affecting change so ledger scores from
// different versions are never silently compared, and re-score the ledger on a
// bump. It stays at 2 until the first live scoring run: before then there is no
// ledger to keep comparable, so every change ships under version 2.
const BenchVersion = 2

// DatasetEpoch is the pinned reference "as-of" instant for all generated
// datasets. Benchmark generation must be a pure function of the run seed and
// bench_version (the reproducibility contract: same (seed, bench_version) =>
// byte-identical plan-layer dataset). Wall-clock time (time.Now) is therefore
// banned from the generation
// path; haystack base dates are drawn *backward* from this fixed epoch instead,
// and the GeneratedAt envelope fields carry it verbatim so two runs of the same
// seed diff clean.
//
// It is NOT a wire timestamp of when a run physically executed: the subnet
// validator stamps its own wall-clock generated_at on the ScoreReport it
// forwards to the platform, and generated_at is not part of the platform's
// DB/signature contract. Pinned per bench_version; bump it only alongside a
// bench_version bump.
var DatasetEpoch = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// DatasetEpochRFC3339 is DatasetEpoch pre-rendered for the string GeneratedAt
// envelope fields (protocol.Dataset, protocol.ScoreReport). Derived from
// DatasetEpoch so the two can never drift.
var DatasetEpochRFC3339 = DatasetEpoch.UTC().Format(time.RFC3339)
