// Command generate-service is the HTTP wrapper the SN118 platform deploys to
// pin a per-submission dataset.
//
// POST /generate?seed=<int64>&run_size=<small|medium|full> returns the FULL
// canonical DatasetArtifact (JSON): the tool cases, the staged memory seeding
// waves, the memory cases (with their unlock wave + graph user_id), and the
// tool-fixture digests — everything a validator needs to replay the eval, plus
// the oracle (expected tools/answers) the validator scores against. The
// artifact's SHA-256 is surfaced in the X-Dataset-SHA256 response header so
// the platform can pin it without re-marshaling.
//
// The platform calls this once per submission at screening pass and pins
// (seed, dataset_sha256, run_size); validators regenerate the dataset from the
// pin and their engine fails a run whose regenerated hash mismatches.
//
// The DEPLOYMENT is private (an IAM-gated Cloud Run service only the platform
// may invoke, so platform infrastructure cannot be farmed as a generation
// oracle), but the code carries no secret: generation is fully DETERMINISTIC
// and non-LLM (gen.GenerateDataset), byte-reproducible from (seed, run_size) —
// exactly what the public cmd/generate CLI computes. The anti-precompute
// guarantee is the seed, which is fixed from an on-chain block hash only after
// the miner commits their submission.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/ditto-assistant/dittobench-datagen/gen"
)

// defaultRunSize is used when the caller omits run_size. small is the cheapest
// profile (mirrors the scoring engine's profile default).
const defaultRunSize = "small"

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("POST /generate", handleGenerate)

	addr := ":" + envOr("PORT", "8090")
	log.Printf("dittobench-datagen generate service listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

// handleGenerate produces one seeded dataset artifact. seed is required (the
// platform supplies a fresh one per submission); run_size selects the profile
// (defaults to small). The response is the canonical DatasetArtifact JSON.
func handleGenerate(w http.ResponseWriter, r *http.Request) {
	seed, err := strconv.ParseInt(r.URL.Query().Get("seed"), 10, 64)
	if err != nil {
		http.Error(w, "seed query param required (int64)", http.StatusBadRequest)
		return
	}
	runSize := r.URL.Query().Get("run_size")
	if runSize == "" {
		runSize = defaultRunSize
	}
	prof, ok := gen.ProfileFor(runSize)
	if !ok {
		http.Error(w, "unknown run_size (want small|medium|full)", http.StatusBadRequest)
		return
	}

	artifact := gen.GenerateDataset(seed, prof)
	w.Header().Set("Content-Type", "application/json")
	// Surface the content hash so the platform can record/pin it with the ticket
	// without re-marshaling.
	if h, _, herr := artifact.SHA256Hex(); herr == nil {
		w.Header().Set("X-Dataset-SHA256", h)
	}
	if err := json.NewEncoder(w).Encode(artifact); err != nil {
		log.Printf("encode dataset (seed=%d run_size=%s): %v", seed, runSize, err)
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
