package gen

import (
	"fmt"
	"sort"

	"github.com/ditto-assistant/dittobench-datagen/protocol"
	"github.com/ditto-assistant/dittobench-datagen/universe"
)

// ReviewAnnotation records generation-time evidence for local human review.
// It is deliberately separate from DatasetArtifact: these fields never change
// the canonical dataset bytes and never cross the validator-to-harness wire.
type ReviewAnnotation struct {
	CaseID          string   `json:"case_id"`
	RequiredPairIDs []string `json:"required_pair_ids"`
	Facts           []string `json:"facts"`
	Constraints     []string `json:"constraints"`
	Operations      []string `json:"operations"`
}

// DatasetReview combines the canonical artifact with reviewer-only provenance.
// The artifact is exactly the value produced by GenerateDataset.
type DatasetReview struct {
	Artifact          DatasetArtifact    `json:"artifact"`
	MemoryAnnotations []ReviewAnnotation `json:"memory_annotations,omitempty"`
}

// GenerateDatasetReview generates the real benchmark artifact and, for V8
// shared-world cases, the planted-evidence plan used to prove answerability.
// Review metadata is additive and local-only; it is not included in SHA256Hex.
func GenerateDatasetReview(seed int64, prof Profile, benchVersion int) (DatasetReview, error) {
	artifact, err := GenerateDataset(seed, prof, benchVersion)
	if err != nil {
		return DatasetReview{}, err
	}
	review := DatasetReview{Artifact: artifact}
	if benchVersion < protocol.BenchVersionV8 {
		return review, nil
	}

	scale, count := v8WorldProfile(prof.Mem)
	if count == 0 {
		return review, nil
	}
	world := universe.Generate(seed, scale)
	plans, err := world.QuestionPlans(count)
	if err != nil {
		return DatasetReview{}, fmt.Errorf("generate review annotations: %w", err)
	}
	artifactCases := make(map[string]bool, len(artifact.MemoryCases))
	for _, c := range artifact.MemoryCases {
		artifactCases[c.ID] = true
	}
	for _, plan := range plans {
		if !artifactCases[plan.Case.ID] {
			continue
		}
		review.MemoryAnnotations = append(review.MemoryAnnotations, ReviewAnnotation{
			CaseID:          plan.Case.ID,
			RequiredPairIDs: plan.SortedEvidenceIDs(),
			Facts:           append([]string(nil), plan.Facts...),
			Constraints:     append([]string(nil), plan.Constraints...),
			Operations:      append([]string(nil), plan.Operations...),
		})
	}
	sort.Slice(review.MemoryAnnotations, func(i, j int) bool {
		return review.MemoryAnnotations[i].CaseID < review.MemoryAnnotations[j].CaseID
	})
	return review, nil
}
