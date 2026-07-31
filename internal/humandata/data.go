// Package humandata exposes frozen, deterministic human-name data for V8.
//
// The benchmark never calls a live name service. The checked-in snapshots and
// their provenance are part of G(seed, version), so a public seed remains byte
// reproducible even when an upstream dataset changes.
package humandata

import (
	"bufio"
	_ "embed"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

//go:embed data/given_names.tsv
var givenNamesTSV string

//go:embed data/surnames.tsv
var surnamesTSV string

//go:embed data/nicknames.tsv
var nicknamesTSV string

type weightedName struct {
	value  string
	weight int64
}

var (
	givenNames = parseWeighted(givenNamesTSV)
	surnames   = parseWeighted(surnamesTSV)
	nicknames  = parseNicknames(nicknamesTSV)
)

// GivenName returns a deterministic mixture of common and long-tail U.S.
// given names. Four of five draws follow observed frequency; every fifth draw
// samples uniformly from ranks 1,001-10,000 so uncommon and international
// names do not disappear beneath the head of the distribution.
func GivenName(r *rand.Rand, ordinal int) string {
	if ordinal%5 == 4 && len(givenNames) > 1000 {
		return givenNames[1000+r.Intn(len(givenNames)-1000)].value
	}
	limit := 1000
	if limit > len(givenNames) {
		limit = len(givenNames)
	}
	return weightedPick(r, givenNames[:limit])
}

// Surname returns a deterministic U.S. Census surname sample. Most draws use
// observed frequency, with a bounded long-tail stratum for broader coverage.
func Surname(r *rand.Rand, ordinal int) string {
	if ordinal%7 == 6 && len(surnames) > 1000 {
		return surnames[1000+r.Intn(len(surnames)-1000)].value
	}
	limit := 2000
	if limit > len(surnames) {
		limit = len(surnames)
	}
	return weightedPick(r, surnames[:limit])
}

// PreferredName returns a real, frozen diminutive when one is known. Names
// without a documented mapping use the given name unchanged; arbitrary
// truncation and token-shaped "nicknames" are never invented.
func PreferredName(given string, r *rand.Rand) string {
	options := nicknames[strings.ToLower(given)]
	if len(options) == 0 {
		return given
	}
	return options[r.Intn(len(options))]
}

// AllGivenNames returns the frozen 10,000-name vocabulary for V8-only persona
// attributes that need a broad, non-enumerable human-name surface.
func AllGivenNames() []string {
	out := make([]string, len(givenNames))
	for i, entry := range givenNames {
		out[i] = entry.value
	}
	return out
}

func weightedPick(r *rand.Rand, entries []weightedName) string {
	if len(entries) == 0 {
		panic("humandata: empty weighted corpus")
	}
	var total int64
	for _, entry := range entries {
		total += entry.weight
	}
	draw := r.Int63n(total)
	for _, entry := range entries {
		if draw < entry.weight {
			return entry.value
		}
		draw -= entry.weight
	}
	return entries[len(entries)-1].value
}

func parseWeighted(raw string) []weightedName {
	scanner := bufio.NewScanner(strings.NewReader(raw))
	if !scanner.Scan() || scanner.Text() != "name\tcount" {
		panic("humandata: invalid weighted-corpus header")
	}
	var out []weightedName
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "\t")
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" {
			panic(fmt.Sprintf("humandata: invalid weighted row %q", scanner.Text()))
		}
		weight, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil || weight <= 0 {
			panic(fmt.Sprintf("humandata: invalid weight in %q", scanner.Text()))
		}
		out = append(out, weightedName{value: parts[0], weight: weight})
	}
	if err := scanner.Err(); err != nil {
		panic(fmt.Sprintf("humandata: scan weighted corpus: %v", err))
	}
	return out
}

func parseNicknames(raw string) map[string][]string {
	scanner := bufio.NewScanner(strings.NewReader(raw))
	if !scanner.Scan() || scanner.Text() != "name\tnickname" {
		panic("humandata: invalid nickname-corpus header")
	}
	out := map[string][]string{}
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "\t")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			panic(fmt.Sprintf("humandata: invalid nickname row %q", scanner.Text()))
		}
		key := strings.ToLower(parts[0])
		out[key] = append(out[key], parts[1])
	}
	if err := scanner.Err(); err != nil {
		panic(fmt.Sprintf("humandata: scan nickname corpus: %v", err))
	}
	return out
}
