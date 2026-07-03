package knowledge

import (
	"context"
	"math"
	"sort"
)

// Passage is one indexed chunk of the business knowledge base. The Vector is derived
// at index time and never serialized (it is recomputed on load).
type Passage struct {
	ID     string    `json:"id"`
	Source string    `json:"source"` // e.g. "kb:refund-policy.md#2"
	Text   string    `json:"text"`
	Vector []float32 `json:"-"`
}

// Scored is a passage with its similarity to a query.
type Scored struct {
	Passage Passage
	Score   float64
}

// VectorStore indexes passages and retrieves the nearest to a query vector. The P0
// implementation is an in-memory cosine store (MemStore); a sqlite-vec backend that
// keeps the index on disk without a separate service is the documented upgrade seam.
type VectorStore interface {
	Add(ctx context.Context, passages []Passage) error
	Search(ctx context.Context, vector []float32, k int) ([]Scored, error)
}

// MemStore is a dependency-free, in-memory VectorStore using cosine similarity.
// Vectors are assumed L2-normalized by the embedder, so cosine reduces to a dot
// product; MemStore normalizes defensively regardless.
type MemStore struct {
	passages []Passage
}

// NewMemStore returns an empty in-memory store.
func NewMemStore() *MemStore { return &MemStore{} }

// Add appends passages to the index.
func (m *MemStore) Add(_ context.Context, passages []Passage) error {
	m.passages = append(m.passages, passages...)
	return nil
}

// Len reports how many passages are indexed.
func (m *MemStore) Len() int { return len(m.passages) }

// Search returns the top-k passages by cosine similarity, highest first.
func (m *MemStore) Search(_ context.Context, vector []float32, k int) ([]Scored, error) {
	if k <= 0 || len(m.passages) == 0 {
		return nil, nil
	}
	scored := make([]Scored, 0, len(m.passages))
	for _, p := range m.passages {
		scored = append(scored, Scored{Passage: p, Score: cosine(vector, p.Vector)})
	}
	sort.SliceStable(scored, func(i, j int) bool { return scored[i].Score > scored[j].Score })
	if k > len(scored) {
		k = len(scored)
	}
	return scored[:k], nil
}

// cosine returns the cosine similarity of two vectors, or 0 if either is degenerate
// or lengths differ.
func cosine(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		na += float64(a[i]) * float64(a[i])
		nb += float64(b[i]) * float64(b[i])
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}
