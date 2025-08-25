package agent

// genspark: naive frequency-based summarizer (no external AI)

import (
	"regexp"
	"sort"
	"strings"
)

// genspark: basic english stopwords set
var stop = map[string]bool{
	"the": true, "a": true, "an": true, "and": true, "or": true, "to": true, "of": true,
	"in": true, "on": true, "for": true, "with": true, "is": true, "are": true, "was": true,
	"were": true, "be": true, "been": true, "by": true, "as": true, "at": true, "from": true,
	"that": true, "this": true, "it": true, "its": true, "we": true, "you": true, "your": true,
}

// genspark: Summarize selects top k sentences by term frequency score
func Summarize(text string, k int) []string {
	if k <= 0 {
		k = 3
	}
	sentences := splitSentences(text)
	if len(sentences) == 0 {
		return nil
	}
	freq := wordFreq(strings.Join(sentences, " "))
	type pair struct{ idx int; score float64 }
	var scores []pair
	for i, s := range sentences {
		s := strings.TrimSpace(s)
		if s == "" { continue }
		w := tokenize(s)
		var total float64
		for _, tok := range w {
			if stop[strings.ToLower(tok)] { continue }
			total += freq[strings.ToLower(tok)]
		}
		scores = append(scores, pair{i, total})
	}
	sort.Slice(scores, func(i, j int) bool { return scores[i].score > scores[j].score })
	if k > len(scores) { k = len(scores) }
	// genspark: keep original order for selected sentences
	selected := scores[:k]
	sort.Slice(selected, func(i, j int) bool { return selected[i].idx < selected[j].idx })
	out := make([]string, 0, k)
	for _, p := range selected {
		out = append(out, strings.TrimSpace(sentences[p.idx]))
	}
	return out
}

// genspark: tokenize into words
func tokenize(s string) []string {
	re := regexp.MustCompile(`[A-Za-z\d']+`)
	return re.FindAllString(s, -1)
}

// genspark: word frequency map
func wordFreq(s string) map[string]float64 {
	m := make(map[string]float64)
	for _, t := range tokenize(s) {
		l := strings.ToLower(t)
		if stop[l] { continue }
		m[l] += 1
	}
	return m
}

// genspark: naive sentence splitter
func splitSentences(s string) []string {
	re := regexp.MustCompile(`(?m)([^.!?]+[.!?])`)
	m := re.FindAllString(s, -1)
	if len(m) == 0 {
		return []string{s}
	}
	return m
}
