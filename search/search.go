package search

import (
	"math"
	"sort"
	"strings"

	"github.com/Versus04/situp/index"
)

type Result struct {
	URL   string
	Score float64
}

func Query(idx *index.Index, query string, maxPages int, topN int) []Result {
	words := strings.Fields(strings.ToLower(query))
	scores := make(map[string]float64)

	for _, w := range words {
		term := index.NormalizeToken(w)
		if term == "" {
			continue
		}

		df := idx.DocFreq(term)
		if df == 0 {
			continue
		}

		idf := math.Log(float64(maxPages) / float64(df))
		for url, tf := range idx.TermFreqs(term) {
			scores[url] += float64(tf) * idf
		}
	}

	results := make([]Result, 0, len(scores))
	for url, score := range scores {
		results = append(results, Result{URL: url, Score: score})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if topN > 0 && len(results) > topN {
		results = results[:topN]
	}
	return results
}
