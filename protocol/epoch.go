package protocol

import (
	"fmt"
	"time"
)

// Benchmark versions are immutable generation contracts. CurrentBenchVersion is
// used only for display and release metadata; canonical generation must always
// receive an explicit version so historical datasets remain reproducible after
// a release advances.
//
// v2 is frozen: it is what on-chain scoring has been running since 2026-07-13,
// and its bytes must keep regenerating identically so any already-scored run
// stays auditable. v3 is the ANTI-GAMING release -- dump-guard grading, needle
// gating, adversarial distractors, composed injection framings, the cross-user
// lifecycle probe, and the reproduce-under-transform audit. Every one of those
// is gated on the version, so adding hardening never disturbs v2's bytes.
// v4 is the FALSE-POSITIVE release. v3's anti-gaming machinery penalized a
// number of legitimate behaviours: the canary could be audit-duplicated so one
// breach was charged twice, delete INSTRUCTIONS were graded on echoing a noun
// phrase, and several composite factors taxed a transparent memory agent for how
// it retrieves. Those are corrected here rather than in v3, because a bench
// version is an immutable contract: v3's bytes and its already-recorded scores
// stay exactly as they were, and re-scoring happens by moving to a new contract
// rather than by rewriting an old one.
// v5 is the CONVERSATIONAL-GROUNDING, COVERAGE, AND EFFICIENCY release (see
// dittobench-api docs/BENCHMARK-V5-PLAN.md). v4 rewarded a phrase-list router that
// recognizes memory writes only through a closed cue family and dumps retrieved
// memory on any unmatched turn (the "Aurora-9" failure): a greeting could surface
// an off-topic seeded value and still score near the top, because no case graded
// conversational relevance and no case tested a plain declarative statement as a
// durable write. v5 adds, all version-gated so v4's bytes and already-recorded
// scores are untouched:
//   - a conversational-sanity gate (greeting non-leak, declarative acknowledgement,
//     abstention-over-confabulation) and ordinary no-save-verb declarative writes
//     with their persistence and behavior-change proofs;
//   - the accept-set grading primitive (non-verbatim answers), Code Mode tool
//     coverage, and harder capability dimensions (multi-hop relational KG joins,
//     temporal-depth), all high-entropy and metamorphic-twinned to resist overfit;
//   - a token-efficiency (waste-penalty) contract the validator applies from
//     trusted relay telemetry, binding generation, model/provider profile, and the
//     starter baseline to one immutable contract.
const (
	BenchVersionV2      = 2
	BenchVersionV3      = 3
	BenchVersionV4      = 4
	BenchVersionV5      = 5
	BenchVersionV6      = 6
	CurrentBenchVersion = BenchVersionV6

	// BenchVersion is retained as a source-compatible alias for consumers that
	// report the active release. Generation code must not use it to select a
	// dataset contract.
	BenchVersion = CurrentBenchVersion
)

var (
	datasetEpochV2 = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	datasetEpochV3 = time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	datasetEpochV4 = time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	datasetEpochV5 = time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	datasetEpochV6 = time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

	// DatasetEpoch and DatasetEpochRFC3339 retain the v2 values for legacy
	// package callers. Canonical versioned generation uses DatasetEpochForVersion.
	DatasetEpoch        = datasetEpochV2
	DatasetEpochRFC3339 = datasetEpochV2.Format(time.RFC3339)
)

// SupportedBenchVersion reports whether this module can reproduce a version.
func SupportedBenchVersion(version int) bool {
	return version == BenchVersionV2 || version == BenchVersionV3 ||
		version == BenchVersionV4 || version == BenchVersionV5 ||
		version == BenchVersionV6
}

// DatasetEpochForVersion returns the immutable reference instant for version.
func DatasetEpochForVersion(version int) (time.Time, error) {
	switch version {
	case BenchVersionV2:
		return datasetEpochV2, nil
	case BenchVersionV3:
		return datasetEpochV3, nil
	case BenchVersionV4:
		return datasetEpochV4, nil
	case BenchVersionV5:
		return datasetEpochV5, nil
	case BenchVersionV6:
		return datasetEpochV6, nil
	default:
		return time.Time{}, fmt.Errorf("unsupported bench_version %d (supported: 2, 3, 4, 5, 6)", version)
	}
}

// RotateSeedForVersion folds an explicit benchmark version into a run seed.
// It is deterministic and retains the exact historical v2 mixing function.
func RotateSeedForVersion(seed int64, version int) (int64, error) {
	if !SupportedBenchVersion(version) {
		return 0, fmt.Errorf("unsupported bench_version %d (supported: 2, 3, 4, 5, 6)", version)
	}
	v := uint64(version)
	x := uint64(seed) ^ (v * 0x9E3779B97F4A7C15)
	x ^= x >> 30
	x *= 0xBF58476D1CE4E5B9
	x ^= x >> 27
	x *= 0x94D049BB133111EB
	x ^= x >> 31
	return int64(x), nil
}

// RotateSeed retains the v2 behavior for legacy package callers. New canonical
// generation paths must call RotateSeedForVersion with an explicit version.
func RotateSeed(seed int64) int64 {
	rotated, _ := RotateSeedForVersion(seed, BenchVersionV2)
	return rotated
}
