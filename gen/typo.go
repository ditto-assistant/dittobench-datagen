package gen

import (
	"math/rand"
	"strings"

	"github.com/ditto-assistant/dittobench-datagen/protocol"
)

// bench_version 6 INFORMAL-NOISE layer (seeded typos). Ditto chats read like
// texting a friend — lowercase, rushed, autocorrect artifacts — so a harness that
// only survives clean, well-formed prompts is overfit to a register real users
// never write in. v6 injects realistic, seeded typos into the TOPIC text (nouns,
// relations, filler) of the memory families, forcing a harness to reconcile the
// same noisy topic across turns.
//
// Contract-preserving by construction:
//   - Deterministic: every typo is a draw from the shared per-(seed,version) rng
//     at a fixed generation point, so (seed, bench_version) still reproduces
//     identical bytes forever.
//   - v6-gated: typoText is a no-op (and draws NOTHING) for bench_version < 6, so
//     v5's rng stream and bytes are untouched.
//   - Answers stay clean: the number of typos per string is itself a seed draw
//     (0..bound), so a miner cannot know how many or where — but coined tokens
//     (every graded value/distractor/sentinel: they all carry a digit, an
//     underscore, or a VK-/all-caps-hyphen shape) are NEVER mutated, and callers
//     additionally protect the multi-hop join names. So the graded answer is
//     always exactly recoverable; only the surrounding topic gets noisy.
//
// The categories model what a phone actually does: adjacent-key fat-fingers,
// transpositions, dropped/doubled letters, dropped apostrophes, and
// autocorrect-to-a-different-clean-word (the same mechanism behind the infamous
// curse->clean-word swaps, applied to a clean corpus).

// qwertyNeighbors maps a lowercase letter to physically adjacent keys, for
// fat-finger substitutions.
var qwertyNeighbors = map[byte]string{
	'a': "qwsz", 'b': "vghn", 'c': "xdfv", 'd': "serfcx", 'e': "wsdr", 'f': "drtgvc",
	'g': "ftyhbv", 'h': "gyujnb", 'i': "ujko", 'j': "huikmn", 'k': "jiolm", 'l': "kop",
	'm': "njk", 'n': "bhjm", 'o': "iklp", 'p': "ol", 'q': "wa", 'r': "edft",
	's': "awedxz", 't': "rfgy", 'u': "yhji", 'v': "cfgb", 'w': "qase", 'x': "zsdc",
	'y': "tghu", 'z': "asx",
}

// contractionsNoApostrophe are the dropped-apostrophe forms a phone keyboard
// produces when the apostrophe is skipped.
var contractionsNoApostrophe = map[string]string{
	"don't": "dont", "won't": "wont", "can't": "cant", "i'm": "im", "it's": "its",
	"what's": "whats", "i've": "ive", "you're": "youre", "we'll": "well", "i'd": "id",
	"that's": "thats", "there's": "theres", "let's": "lets", "i'll": "ill", "he's": "hes",
	"she's": "shes", "they're": "theyre", "we're": "were", "you've": "youve", "isn't": "isnt",
	"doesn't": "doesnt", "didn't": "didnt", "wasn't": "wasnt", "haven't": "havent",
}

// autocorrectSwaps replaces a word with a DIFFERENT real (or near-real) word the
// phone "corrected" it to. Curated to avoid negation/meaning inversion (never
// no<->not etc.) so the topic a harness must reconcile is still the same topic.
var autocorrectSwaps = map[string]string{
	"than": "then", "then": "than", "lose": "loose", "your": "you're",
	"really": "realy", "definitely": "definately", "tonight": "tonite",
	"tomorrow": "tomorow", "favourite": "favorite", "probably": "probly",
	"though": "tho", "through": "thru", "because": "becuase", "receive": "recieve",
	"weird": "wierd", "restaurant": "restaraunt",
}

type tok struct {
	text string
	word bool
}

// tokenize splits s into alternating word / separator runs. A word run is
// [A-Za-z0-9'_-] so contractions ("don't"), hyphenated nouns ("e-bike"), AND
// coined tokens ("VK-9BM...", "gavotu_8841", "84-GAVO-TUKE") each stay a SINGLE
// token — critical so coinedShaped can protect a coined value whole rather than
// letting a letter-only fragment of it slip through as eligible.
func tokenize(s string) []tok {
	var out []tok
	var b strings.Builder
	inWord := false
	flush := func(w bool) {
		if b.Len() > 0 {
			out = append(out, tok{text: b.String(), word: w})
			b.Reset()
		}
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		isWordChar := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '\'' || c == '_' || c == '-'
		if isWordChar != inWord {
			flush(inWord)
			inWord = isWordChar
		}
		b.WriteByte(c)
	}
	flush(inWord)
	return out
}

// coinedShaped reports whether a token looks like a coined value (so it is never
// mutated). Every CoinShaped variant carries a digit, an underscore, or a VK- /
// all-caps-hyphen shape; this catches all of them without an explicit list.
func coinedShaped(w string) bool {
	if strings.HasPrefix(w, "VK-") {
		return true
	}
	caps := 0
	for i := 0; i < len(w); i++ {
		c := w[i]
		if c >= '0' && c <= '9' || c == '_' {
			return true
		}
		if c >= 'A' && c <= 'Z' {
			caps++
		}
	}
	return caps >= 4 // an all-caps run like GAVOTU
}

func lettersLen(w string) int {
	n := 0
	for i := 0; i < len(w); i++ {
		if w[i] >= 'a' && w[i] <= 'z' || w[i] >= 'A' && w[i] <= 'Z' {
			n++
		}
	}
	return n
}

// eligible reports whether a word token may be typo'd: real word, long enough,
// not coined-shaped, and not containing any protected substring (e.g. a join name).
func eligible(w string, protect []string) bool {
	if lettersLen(w) < 3 || coinedShaped(w) {
		return false
	}
	for _, p := range protect {
		if p != "" && strings.Contains(w, p) {
			return false
		}
	}
	return true
}

// letterIndices returns the byte offsets of the ASCII letters in w.
func letterIndices(w string) []int {
	var idx []int
	for i := 0; i < len(w); i++ {
		if w[i] >= 'a' && w[i] <= 'z' || w[i] >= 'A' && w[i] <= 'Z' {
			idx = append(idx, i)
		}
	}
	return idx
}

// mutateWord applies one realistic typo category to w, chosen from those that
// apply. Returns w unchanged if none apply.
func mutateWord(r *rand.Rand, w string) string {
	lower := strings.ToLower(w)
	var cats []int // 0 fat-finger, 1 transpose, 2 drop, 3 double, 4 apostrophe, 5 autocorrect
	if _, ok := contractionsNoApostrophe[lower]; ok {
		cats = append(cats, 4)
	}
	if _, ok := autocorrectSwaps[lower]; ok {
		cats = append(cats, 5)
	}
	if lettersLen(w) >= 3 {
		cats = append(cats, 0, 1, 2, 3)
	}
	if len(cats) == 0 {
		return w
	}
	switch cats[r.Intn(len(cats))] {
	case 4: // MissingApostrophe
		return contractionsNoApostrophe[lower]
	case 5: // AutocorrectSwap
		return matchCase(w, autocorrectSwaps[lower])
	case 0: // FatFinger: adjacent-key substitution
		li := letterIndices(w)
		pos := li[r.Intn(len(li))]
		c := w[pos]
		lc := c | 0x20 // to lower
		nb := qwertyNeighbors[lc]
		if nb == "" {
			return w
		}
		repl := nb[r.Intn(len(nb))]
		if c >= 'A' && c <= 'Z' {
			repl -= 0x20
		}
		return w[:pos] + string(repl) + w[pos+1:]
	case 1: // Transpose two string-adjacent letters (never across a hyphen)
		li := letterIndices(w)
		var adj []int // positions p where p and p+1 are both letters
		for _, p := range li {
			if p+1 < len(w) && (w[p+1] >= 'a' && w[p+1] <= 'z' || w[p+1] >= 'A' && w[p+1] <= 'Z') {
				adj = append(adj, p)
			}
		}
		if len(adj) == 0 {
			return w
		}
		a := adj[r.Intn(len(adj))]
		return w[:a] + string(w[a+1]) + string(w[a]) + w[a+2:]
	case 2: // Drop a letter (keep at least 3)
		li := letterIndices(w)
		if len(li) <= 3 {
			return w
		}
		pos := li[r.Intn(len(li))]
		return w[:pos] + w[pos+1:]
	case 3: // Double a letter
		li := letterIndices(w)
		pos := li[r.Intn(len(li))]
		return w[:pos+1] + string(w[pos]) + w[pos+1:]
	}
	return w
}

// matchCase gives repl the leading-capitalization of src (so "Really"->"Realy",
// not "realy").
func matchCase(src, repl string) string {
	if len(src) > 0 && len(repl) > 0 && src[0] >= 'A' && src[0] <= 'Z' && repl[0] >= 'a' && repl[0] <= 'z' {
		return string(repl[0]-0x20) + repl[1:]
	}
	return repl
}

// applyTypos injects a seed-bounded, unpredictable number of realistic typos into
// s, leaving protected and coined-shaped tokens untouched. The count draw happens
// unconditionally (so the rng stream is stable) and may be zero.
func applyTypos(r *rand.Rand, s string, protect []string) string {
	tokens := tokenize(s)
	var elig []int
	for i, tk := range tokens {
		if tk.word && eligible(tk.text, protect) {
			elig = append(elig, i)
		}
	}
	if len(elig) == 0 {
		return s
	}
	maxK := 1 + len(elig)/4 // ~a quarter of eligible words, at most
	k := r.Intn(maxK + 1)   // 0..maxK — miners cannot predict how many
	if k == 0 {
		return s
	}
	perm := r.Perm(len(elig))
	for j := 0; j < k; j++ {
		idx := elig[perm[j]]
		tokens[idx].text = mutateWord(r, tokens[idx].text)
	}
	var b strings.Builder
	for _, tk := range tokens {
		b.WriteString(tk.text)
	}
	return b.String()
}

// typoText is the caller entry point: for bench_version >= 6 it injects seeded
// typos into s (protecting the given substrings); for earlier versions it returns
// s unchanged AND draws nothing, so pre-v6 bytes are byte-identical.
func typoText(r *rand.Rand, benchVersion int, s string, protect ...string) string {
	if benchVersion < protocol.BenchVersionV6 {
		return s
	}
	return applyTypos(r, s, protect)
}
