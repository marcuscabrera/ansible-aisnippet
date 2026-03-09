// Package similarity provides a lightweight native Go TF-IDF engine for
// finding the most similar snippet description to a user query.
//
// This replaces the Python Gensim + Jieba dependency with a pure-Go
// Bag-of-Words / Cosine-Similarity implementation that is sufficient for
// the small, English-language snippets.json corpus.
package similarity

import (
	"math"
	"regexp"
	"strings"
)

// nonAlphaNum matches any character that is not a letter or digit.
var nonAlphaNum = regexp.MustCompile(`[^a-z0-9]+`)

// tokenize lowercases a string and splits it into word tokens, discarding
// punctuation and very short tokens.
func tokenize(text string) []string {
	lower := strings.ToLower(text)
	parts := nonAlphaNum.Split(lower, -1)
	tokens := parts[:0]
	for _, p := range parts {
		if len(p) > 1 {
			tokens = append(tokens, p)
		}
	}
	return tokens
}

// Engine holds the pre-built TF-IDF index for a set of documents (snippet keys).
type Engine struct {
	docs    []string    // original document strings (snippet keys)
	idf     map[string]float64
	tfidfVectors []map[string]float64
}

// New creates an Engine from the provided list of document strings.
// The index is built once at construction time.
func New(docs []string) *Engine {
	e := &Engine{docs: docs}
	e.build()
	return e
}

// build computes TF-IDF vectors for all documents.
func (e *Engine) build() {
	n := len(e.docs)
	if n == 0 {
		return
	}

	tokenizedDocs := make([][]string, n)
	for i, doc := range e.docs {
		tokenizedDocs[i] = tokenize(doc)
	}

	// Compute document frequency (DF) for each term.
	df := make(map[string]int)
	for _, tokens := range tokenizedDocs {
		seen := make(map[string]bool)
		for _, t := range tokens {
			if !seen[t] {
				df[t]++
				seen[t] = true
			}
		}
	}

	// Compute IDF: log((N + 1) / (df + 1)) + 1  (smooth IDF to avoid zero division)
	e.idf = make(map[string]float64, len(df))
	for term, freq := range df {
		e.idf[term] = math.Log(float64(n+1)/float64(freq+1)) + 1
	}

	// Compute TF-IDF vectors.
	e.tfidfVectors = make([]map[string]float64, n)
	for i, tokens := range tokenizedDocs {
		tf := make(map[string]int)
		for _, t := range tokens {
			tf[t]++
		}
		vec := make(map[string]float64, len(tf))
		for term, count := range tf {
			rawTF := float64(count) / float64(len(tokens))
			vec[term] = rawTF * e.idf[term]
		}
		e.tfidfVectors[i] = normalize(vec)
	}
}

// FindMostSimilar returns the document string that is most similar to the query.
// If the corpus is empty, an empty string is returned.
func (e *Engine) FindMostSimilar(query string) string {
	if len(e.docs) == 0 {
		return ""
	}

	// Build TF-IDF vector for the query.
	tokens := tokenize(query)
	if len(tokens) == 0 {
		return e.docs[0]
	}
	tf := make(map[string]int)
	for _, t := range tokens {
		tf[t]++
	}
	queryVec := make(map[string]float64, len(tf))
	for term, count := range tf {
		rawTF := float64(count) / float64(len(tokens))
		idfScore := e.idf[term] // zero for unknown terms
		queryVec[term] = rawTF * idfScore
	}
	queryVec = normalize(queryVec)

	// Find the document with maximum cosine similarity.
	bestIdx := 0
	bestScore := -1.0
	for i, docVec := range e.tfidfVectors {
		score := cosineSimilarity(queryVec, docVec)
		if score > bestScore {
			bestScore = score
			bestIdx = i
		}
	}
	return e.docs[bestIdx]
}

// normalize returns a unit-length version of the vector.
func normalize(vec map[string]float64) map[string]float64 {
	var norm float64
	for _, v := range vec {
		norm += v * v
	}
	if norm == 0 {
		return vec
	}
	norm = math.Sqrt(norm)
	result := make(map[string]float64, len(vec))
	for k, v := range vec {
		result[k] = v / norm
	}
	return result
}

// cosineSimilarity computes the dot product of two unit-length vectors
// (equivalent to cosine similarity when both are normalized).
func cosineSimilarity(a, b map[string]float64) float64 {
	var dot float64
	for term, va := range a {
		if vb, ok := b[term]; ok {
			dot += va * vb
		}
	}
	return dot
}
