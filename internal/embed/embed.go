// Package embed defines the text-embedding seam used by the knowledge layer's
// retrieval (RAG). The interface is provider-agnostic so the economy tier can use a
// small local CPU model (no GPU, nothing leaves the box) while a premium tier can use
// a hosted embedder (e.g. Voyage). Anthropic has no embeddings endpoint, so Claude is
// never an embedder here — only a generation provider (see internal/llm).
package embed

import (
	"context"
	"hash/fnv"
	"math"
)

// Embedder turns text into vectors for similarity search.
type Embedder interface {
	// Embed returns one vector per input string, each of length Dim().
	Embed(ctx context.Context, texts []string) ([][]float32, error)
	// Dim is the embedding dimensionality.
	Dim() int
	// Name identifies the provider for logging/diagnostics.
	Name() string
}

// FakeEmbedder is a deterministic, dependency-free embedder for offline tests and
// the demo. It hashes whitespace-delimited tokens into a fixed number of buckets and
// L2-normalizes, so texts that share words land near each other under cosine
// similarity — enough to exercise real retrieval logic without a model or network.
type FakeEmbedder struct {
	D int // dimensionality; defaults to 64 when zero
}

// Dim returns the embedding dimensionality. The default (256) is wide enough to keep
// hash collisions between unrelated tokens rare for the small vocabularies a single
// business has, without a real embedding model.
func (f FakeEmbedder) Dim() int {
	if f.D <= 0 {
		return 256
	}
	return f.D
}

// Name identifies this embedder.
func (f FakeEmbedder) Name() string { return "fake" }

// Embed hashes tokens into buckets and L2-normalizes each vector.
func (f FakeEmbedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	dim := f.Dim()
	out := make([][]float32, len(texts))
	for i, t := range texts {
		vec := make([]float32, dim)
		for _, tok := range tokenize(t) {
			if stopwords[tok] {
				continue // function words create spurious similarity; skip them
			}
			h := fnv.New32a()
			_, _ = h.Write([]byte(tok))
			vec[h.Sum32()%uint32(dim)]++
		}
		out[i] = normalize(vec)
	}
	return out, nil
}

// stopwords are ignored so shared function words ("the", "of", "is") can't ground an
// otherwise-unrelated query. A real embedder downweights these implicitly; the fake
// one must do it explicitly.
var stopwords = map[string]bool{
	"a": true, "an": true, "and": true, "are": true, "as": true, "at": true,
	"be": true, "but": true, "by": true, "can": true, "do": true, "does": true,
	"for": true, "from": true, "have": true, "how": true, "i": true, "in": true,
	"is": true, "it": true, "me": true, "my": true, "of": true, "on": true,
	"or": true, "our": true, "s": true, "the": true, "there": true, "this": true,
	"to": true, "was": true, "we": true, "what": true, "whats": true, "when": true,
	"where": true, "which": true, "who": true, "why": true, "with": true, "you": true,
	"your": true,
}

func tokenize(s string) []string {
	var toks []string
	cur := make([]rune, 0, 16)
	flush := func() {
		if len(cur) > 0 {
			toks = append(toks, string(cur))
			cur = cur[:0]
		}
	}
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
			cur = append(cur, r+('a'-'A'))
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			cur = append(cur, r)
		default:
			flush()
		}
	}
	flush()
	return toks
}

func normalize(v []float32) []float32 {
	var sum float64
	for _, x := range v {
		sum += float64(x) * float64(x)
	}
	if sum == 0 {
		return v
	}
	inv := float32(1 / math.Sqrt(sum))
	for i := range v {
		v[i] *= inv
	}
	return v
}
