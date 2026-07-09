// Command gstudy is an offline reliability analyzer for DittoBench scored runs
// (BENCHMARK-V3-IDEAS #4/#5). It reads a JSONL of scored runs and reports:
//
//   - a crossed generalizability (G-study) variance decomposition of per-case
//     scores into seed, item (category), and residual components, so you can see
//     whether the benchmark is seed-dominated ("buy seeds") or item-dominated
//     ("buy items") — the dominant facet flips with benchmark size, and an
//     agent+memory bench is usually the small-N, seed-dominated regime;
//   - a lightweight 2PL-style per-category difficulty (mean pass) and
//     discrimination (spread) estimate, flagging saturated (everyone-passes) and
//     floor (everyone-fails) categories that carry ~0 Fisher information at the
//     champion boundary and should be retired or rebalanced.
//
// It is pure analysis over already-scored data — no LLM, no chain, stdlib only.
// Input JSONL, one run per line:
//
//	{"seed":123,"per_case":[{"category":"web_search","score":1.0}, ...]}
//
// Usage: gstudy < runs.jsonl   (or gstudy -in runs.jsonl)
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"sort"
)

type perCase struct {
	Category string  `json:"category"`
	Score    float64 `json:"score"`
}

type run struct {
	Seed    int64     `json:"seed"`
	PerCase []perCase `json:"per_case"`
}

func main() {
	in := flag.String("in", "", "input JSONL path (default: stdin)")
	flag.Parse()

	r := os.Stdin
	if *in != "" {
		f, err := os.Open(*in)
		if err != nil {
			fmt.Fprintln(os.Stderr, "open:", err)
			os.Exit(1)
		}
		defer f.Close()
		r = f
	}

	runs, err := readRuns(r)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read:", err)
		os.Exit(1)
	}
	if len(runs) == 0 {
		fmt.Fprintln(os.Stderr, "no runs read")
		os.Exit(1)
	}

	rep := analyze(runs)
	b, _ := json.MarshalIndent(rep, "", "  ")
	fmt.Println(string(b))
}

func readRuns(r *os.File) ([]run, error) {
	var out []run
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 1<<20), 64<<20)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var rn run
		if err := json.Unmarshal(line, &rn); err != nil {
			return nil, err
		}
		out = append(out, rn)
	}
	return out, sc.Err()
}

// Report is the gstudy output.
type Report struct {
	Runs          int                   `json:"runs"`
	Cases         int                   `json:"cases"`
	GrandMean     float64               `json:"grand_mean"`
	Variance      VarianceComponents    `json:"variance_components"`
	DominantFacet string                `json:"dominant_facet"`
	Categories    []CategoryReliability `json:"categories"`
	Advice        string                `json:"advice"`
}

// VarianceComponents is the crossed decomposition of per-case score variance.
type VarianceComponents struct {
	Total     float64 `json:"total"`
	Seed      float64 `json:"seed"`     // between-seed (run) component
	Item      float64 `json:"item"`     // between-category (item) component
	Residual  float64 `json:"residual"` // seed×item interaction + noise
	SeedFrac  float64 `json:"seed_frac"`
	ItemFrac  float64 `json:"item_frac"`
	ResidFrac float64 `json:"residual_frac"`
}

// CategoryReliability is a 2PL-style per-category summary.
type CategoryReliability struct {
	Category       string  `json:"category"`
	N              int     `json:"n"`
	Difficulty     float64 `json:"difficulty"`     // mean pass (higher = easier)
	Discrimination float64 `json:"discrimination"` // std of scores (0 = no signal)
	Flag           string  `json:"flag,omitempty"` // "saturated" | "floor" | ""
}

const (
	saturatedAbove = 0.98
	floorBelow     = 0.02
)

func analyze(runs []run) Report {
	// Collect all scores, grand mean, and group by seed and by category.
	var all []float64
	seedScores := map[int64][]float64{}
	catScores := map[string][]float64{}
	for _, rn := range runs {
		for _, c := range rn.PerCase {
			all = append(all, c.Score)
			seedScores[rn.Seed] = append(seedScores[rn.Seed], c.Score)
			catScores[c.Category] = append(catScores[c.Category], c.Score)
		}
	}
	grand := mean(all)
	total := variance(all, grand)

	// Between-seed / between-item components: size-weighted variance of the group
	// means around the grand mean (a one-way ANOVA between-group estimate).
	seedGroups := make([][]float64, 0, len(seedScores))
	for _, v := range seedScores {
		seedGroups = append(seedGroups, v)
	}
	catGroups := make([][]float64, 0, len(catScores))
	for _, v := range catScores {
		catGroups = append(catGroups, v)
	}
	seedComp := betweenGroup(seedGroups, grand)
	itemComp := betweenGroup(catGroups, grand)
	resid := total - seedComp - itemComp
	if resid < 0 {
		resid = 0 // crossed components can slightly over-explain on unbalanced data
	}

	vc := VarianceComponents{Total: r6(total), Seed: r6(seedComp), Item: r6(itemComp), Residual: r6(resid)}
	if total > 0 {
		vc.SeedFrac = r6(seedComp / total)
		vc.ItemFrac = r6(itemComp / total)
		vc.ResidFrac = r6(resid / total)
	}

	facet := "residual"
	if seedComp >= itemComp && seedComp >= resid {
		facet = "seed"
	} else if itemComp >= seedComp && itemComp >= resid {
		facet = "item"
	}

	cats := make([]CategoryReliability, 0, len(catScores))
	for cat, ss := range catScores {
		m := mean(ss)
		cr := CategoryReliability{
			Category:       cat,
			N:              len(ss),
			Difficulty:     r6(m),
			Discrimination: r6(math.Sqrt(variance(ss, m))),
		}
		switch {
		case m >= saturatedAbove:
			cr.Flag = "saturated"
		case m <= floorBelow:
			cr.Flag = "floor"
		}
		cats = append(cats, cr)
	}
	sort.Slice(cats, func(i, j int) bool { return cats[i].Category < cats[j].Category })

	advice := adviceFor(facet)

	return Report{
		Runs:          len(runs),
		Cases:         len(all),
		GrandMean:     r6(grand),
		Variance:      vc,
		DominantFacet: facet,
		Categories:    cats,
		Advice:        advice,
	}
}

func adviceFor(facet string) string {
	switch facet {
	case "seed":
		return "seed-dominated: increase the number of seeds per comparison (buy seeds); CRN paired scoring helps most here"
	case "item":
		return "item-dominated: increase cases per run or drop low-information (saturated/floor) categories (buy items)"
	default:
		return "residual-dominated: seed×item interaction or grading noise dominates; check judge reliability and per-category stability"
	}
}

// betweenGroup returns the size-weighted variance of group means around the grand
// mean — the between-group (facet) variance component of a one-way decomposition.
func betweenGroup(groups [][]float64, grand float64) float64 {
	var num float64
	var n int
	for _, g := range groups {
		if len(g) == 0 {
			continue
		}
		gm := mean(g)
		num += float64(len(g)) * (gm - grand) * (gm - grand)
		n += len(g)
	}
	if n == 0 {
		return 0
	}
	return num / float64(n)
}

func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	var s float64
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

func variance(xs []float64, m float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	var s float64
	for _, x := range xs {
		d := x - m
		s += d * d
	}
	return s / float64(len(xs))
}

func r6(x float64) float64 { return math.Round(x*1e6) / 1e6 }
