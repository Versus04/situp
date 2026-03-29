package index

import (
	"strings"
	"sync"
)

var defaultStopWords = map[string]bool{
	"the": true, "is": true, "a": true, "and": true, "of": true, "to": true,
}

type Index struct {
	mu        sync.RWMutex
	TF        map[string]map[string]int // term -> url -> tf
	DF        map[string]int            // term -> document frequency
	StopWords map[string]bool
}

func NewDefault() *Index {
	stop := make(map[string]bool, len(defaultStopWords))
	for k, v := range defaultStopWords {
		stop[k] = v
	}
	return &Index{
		TF:        make(map[string]map[string]int),
		DF:        make(map[string]int),
		StopWords: stop,
	}
}

func NormalizeToken(s string) string {
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)
	s = strings.Trim(s, ".,!?;:\"()[]{}")
	return s
}

// AddToken updates TF/DF for one token.
// `seen` is per-document and is used to ensure DF increments once per doc.
func (idx *Index) AddToken(url string, rawToken string, seen map[string]bool) {
	token := NormalizeToken(rawToken)
	if token == "" {
		return
	}
	if idx.StopWords[token] {
		return
	}

	idx.mu.Lock()
	defer idx.mu.Unlock()

	if !seen[token] {
		idx.DF[token]++
		seen[token] = true
	}
	if idx.TF[token] == nil {
		idx.TF[token] = make(map[string]int)
	}
	idx.TF[token][url]++
}

func (idx *Index) DocFreq(term string) int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.DF[term]
}

// TermFreqs returns a copy of the postings map for safe iteration.
func (idx *Index) TermFreqs(term string) map[string]int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	postings := idx.TF[term]
	if postings == nil {
		return nil
	}
	cp := make(map[string]int, len(postings))
	for url, tf := range postings {
		cp[url] = tf
	}
	return cp
}
