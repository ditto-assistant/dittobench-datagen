// Package gen produces fresh, anti-overfit benchmark datasets for a single
// submission. The permutation space is large enough that a miner cannot
// precompute a lookup table or overfit to a fixed dataset.
//
// # Anti-cheat combinatorics
//
// Each submission draws a fresh crypto-random seed (see FreshSeed, logged with
// the run). Every layer below multiplies the entropy:
//
// Tool cases:
//   - category × template × entity are sampled per case from large pools
//     (datagen); the prompt is one of the category's hand-written phrasing
//     variants, chosen by the seeded rng, so the surface wording varies per seed
//     while the deterministic ground truth (expected_tools + expected_behavior)
//     is preserved: scoring stays exact and the dataset is reproducible.
//
// Memory cases (persona engine):
//   - a procedural persona universe is built from the seed; its beats render to a
//     haystack from deterministic template surfaces (seeded phrasing variants),
//     and questions are derived from ground truth with per-(seed, question)
//     phrasing variants.
//   - every pair gets a FRESH timestamp anchored to the dataset epoch (per-session
//     day offsets + per-pair spacing), so temporal structure differs each run.
//
// Generation is fully non-LLM and byte-reproducible from (seed, bench_version):
// the anti-cheat entropy is the seed-driven combinatorics (persona facts,
// template-variant selection, timestamps, sampling) plus a fresh per-submission
// crypto-random seed, not a post-hoc paraphrase. Even the small profile is far
// beyond any feasible lookup table; medium and full are far larger.
package gen

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math/rand"

	"github.com/ditto-assistant/dittobench-datagen/protocol"
)

// Profile is the size knobs for a run_size.
type Profile struct {
	Tools int // number of tool cases
	Mem   int // number of memory cases
	// Waves is the number of staged Tier-C seeding waves (1 = single seed).
	// RawPairsFrac is the fraction of memory questions seeded pairs-only (Tier B).
	// small stays a single Tier-A wave so the smoke path is cheap + simple.
	Waves        int
	RawPairsFrac float64
	// IsoCases is the number of multi-graph isolation cases: a second
	// persona is seeded under a different user_id and these cases query one user
	// while the other holds a conflicting value. 0 keeps the small/smoke path
	// single-user and cheap.
	IsoCases int
}

// Profiles maps run_size → counts (v2/v3/v4). Frozen: changing these changes
// those contracts' dataset bytes. v5 uses profilesV5 via ProfileForVersion.
var Profiles = map[string]Profile{
	"small":  {Tools: 6, Mem: 6, Waves: 1, RawPairsFrac: 0, IsoCases: 0},
	"medium": {Tools: 20, Mem: 20, Waves: 2, RawPairsFrac: 0.3, IsoCases: 2},
	"full":   {Tools: 60, Mem: 50, Waves: 2, RawPairsFrac: 0.35, IsoCases: 4},
}

// profilesV5 scales the bench_version 5 run sizes up for comprehensiveness (v5
// plan 4.3: scale the store so retrieval recall is the bottleneck, not parsing).
// Relative to v4: more memory cases and tool cases, DEEPER multi-session staging
// (more waves, so knowledge-update / point-in-time / longitudinal chains span
// more of the timeline), a larger raw-pairs (Tier-B) share, and more multi-graph
// isolation personas. The bigger memory-case count also grows the procedural
// persona (personaOptsFor), so the haystack and distractor density rise with it.
// small stays a cheap single-wave smoke path. Only v5 uses these, so v2/v3/v4
// bytes are untouched.
// Counts are chosen with the between-seed DIFFICULTY variance in mind: more cases
// at the FIXED stratified per-category mix lowers tool_mean variance (measured with
// cmd/benchcal --version 5: tool σ falls from ~0.036 at n=80 to ~0.026 at n=120),
// so scaling up adds coverage AND tightens difficulty rather than widening it.
var profilesV5 = map[string]Profile{
	"small":  {Tools: 6, Mem: 6, Waves: 1, RawPairsFrac: 0, IsoCases: 0},
	"medium": {Tools: 40, Mem: 40, Waves: 3, RawPairsFrac: 0.35, IsoCases: 3},
	"full":   {Tools: 110, Mem: 85, Waves: 4, RawPairsFrac: 0.4, IsoCases: 6},
}

// profilesV8 is the bench_version 8 DEEP-HISTORY sizing. The change relative to
// v5-v7 is deliberately asymmetric: the number of GRADED cases barely moves,
// while the history those cases must be answered from grows by roughly 25x to
// LongMemEval_S parity (~115k tokens; arXiv:2410.10813). Through v7 the scored
// haystack was ~4.5k tokens, small enough to hold in context whole, so a harness
// could skip retrieval and still score; the volume added here is what makes
// retrieval recall the binding constraint. Because the case count is flat, the
// per-submission INFERENCE cost is roughly unchanged: what grows is ingest and
// retrieval, which is exactly the capability under test.
//
// Waves rises with the session count so staged ingestion still lands in
// meaningful increments across a much longer timeline. small stays a cheap
// single-wave smoke path with no deep history at all.
var profilesV8 = map[string]Profile{
	"small":  {Tools: 6, Mem: 6, Waves: 1, RawPairsFrac: 0, IsoCases: 0},
	"medium": {Tools: 40, Mem: 45, Waves: 4, RawPairsFrac: 0.35, IsoCases: 3},
	"full":   {Tools: 110, Mem: 100, Waves: 6, RawPairsFrac: 0.4, IsoCases: 6},
}

// ProfileFor returns the Profile for a run_size, defaulting to small. Uses the
// historical (v2/v3/v4) sizes; canonical versioned callers use ProfileForVersion.
func ProfileFor(runSize string) (Profile, bool) {
	p, ok := Profiles[runSize]
	if !ok {
		return Profiles["small"], false
	}
	return p, true
}

// ProfileForVersion returns the Profile for a (run_size, bench_version). v5 uses
// the scaled-up profilesV5; earlier versions use the frozen historical sizes, so
// their dataset bytes are unchanged. Deterministic, so any third party holding
// (seed, run_size, bench_version) regenerates the identical dataset.
func ProfileForVersion(runSize string, benchVersion int) (Profile, bool) {
	if benchVersion >= protocol.BenchVersionV8 {
		if p, ok := profilesV8[runSize]; ok {
			return p, true
		}
		return profilesV8["small"], false
	}
	if benchVersion >= protocol.BenchVersionV5 {
		if p, ok := profilesV5[runSize]; ok {
			return p, true
		}
		return profilesV5["small"], false
	}
	return ProfileFor(runSize)
}

// FreshSeed returns a cryptographically-random int64 seed for a submission. The
// caller logs it with the run so a dispute can reproduce the exact dataset.
func FreshSeed() int64 {
	var b [8]byte
	if _, err := cryptorand.Read(b[:]); err != nil {
		// Extremely unlikely; fall back to a non-zero constant-ish value.
		return int64(1)
	}
	// Mask to a non-negative int64 for clean logging/JSON.
	return int64(binary.LittleEndian.Uint64(b[:]) >> 1)
}

// NewRNG returns a math/rand source for a run. It seeds from the version-rotated
// seed (protocol.RotateSeed) so the generated surface rotates per bench_version;
// determinism per (seed, bench_version) is what lets a logged run be reproduced
// exactly.
func NewRNG(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(protocol.RotateSeed(seed)))
}

// NewRNGForVersion returns the generation stream for an explicit immutable
// benchmark contract.
func NewRNGForVersion(seed int64, benchVersion int) (*rand.Rand, error) {
	rotated, err := protocol.RotateSeedForVersion(seed, benchVersion)
	if err != nil {
		return nil, err
	}
	return rand.New(rand.NewSource(rotated)), nil
}

// fmtErr wraps a stage error with package context.
func fmtErr(stage string, err error) error {
	return fmt.Errorf("gen %s: %w", stage, err)
}
