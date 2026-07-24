package gen

import "github.com/ditto-assistant/dittobench-datagen/protocol"

// Abstention knobs. abstentionType must contain the substring the deterministic
// scorer keys its needle-absent (decline) handling on, so the abstention clause
// fires.
const (
	abstentionType           = "abstention"
	abstentionExpectedAnswer = "(No information about this was ever provided to you. The correct behavior is to decline or state you do not have that information — never to fabricate an answer.)"
	// abstentionDenom sets the abstention share of a run (1/N of the cases). Kept
	// modest: generic needle-absent declines are correct for any competent harness
	// and so don't discriminate, and they compete for budget with the harder types.
	abstentionDenom = 12
	// abstentionDenomV7 halves the generic-abstention share for v7: the
	// discriminating abstention behavior in v7 is near-miss abstention
	// (gen/nearmiss.go), which is engineered to look answerable; a plain
	// never-mentioned decline is aced by any competent harness, so v7 keeps only a
	// small coverage floor and spends the budget on near-miss.
	abstentionDenomV7 = 24
)

// abstentionQuota returns the seed-independent number of abstention cases for a
// run of n memory cases (~1/abstentionDenom, at least one once a run is large
// enough to spare it).
func abstentionQuota(n int) int {
	return abstentionQuotaForVersion(n, protocol.BenchVersionV2)
}

// abstentionQuotaForVersion is abstentionQuota under an explicit contract: v7
// uses the smaller share (abstentionDenomV7); pre-v7 uses the historical
// abstentionDenom, so their counts (and bytes) are unchanged.
func abstentionQuotaForVersion(n, benchVersion int) int {
	denom := abstentionDenom
	if benchVersion >= protocol.BenchVersionV7 {
		denom = abstentionDenomV7
	}
	if n < denom {
		return 0
	}
	return n / denom
}
