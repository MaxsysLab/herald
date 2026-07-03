package knowledge

import (
	"context"

	"github.com/MaxsysLab/herald/internal/embed"
)

// Retriever embeds a query and searches a VectorStore for grounding passages.
type Retriever struct {
	Embedder embed.Embedder
	Store    VectorStore
}

// Index embeds and adds raw text chunks to the store in one step. Each chunk keeps
// its source label for later citation.
func (r *Retriever) Index(ctx context.Context, chunks []Passage) error {
	if len(chunks) == 0 {
		return nil
	}
	texts := make([]string, len(chunks))
	for i, c := range chunks {
		texts[i] = c.Text
	}
	vecs, err := r.Embedder.Embed(ctx, texts)
	if err != nil {
		return err
	}
	for i := range chunks {
		chunks[i].Vector = vecs[i]
	}
	return r.Store.Add(ctx, chunks)
}

// Retrieve returns the top-k passages most similar to query.
func (r *Retriever) Retrieve(ctx context.Context, query string, k int) ([]Scored, error) {
	vecs, err := r.Embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	if len(vecs) == 0 {
		return nil, nil
	}
	return r.Store.Search(ctx, vecs[0], k)
}
