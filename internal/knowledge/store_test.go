package knowledge

import (
	"context"
	"testing"

	"github.com/MaxsysLab/herald/internal/embed"
)

func TestRetrieverRanksRelevantPassageFirst(t *testing.T) {
	ctx := context.Background()
	r := &Retriever{Embedder: embed.FakeEmbedder{}, Store: NewMemStore()}
	err := r.Index(ctx, []Passage{
		{ID: "wifi", Source: "kb:amenities", Text: "free guest wifi is available for all customers"},
		{ID: "parking", Source: "kb:amenities", Text: "street parking is free after six pm nearby"},
		{ID: "hours", Source: "kb:hours", Text: "we open early every weekday morning"},
	})
	if err != nil {
		t.Fatalf("index: %v", err)
	}

	got, err := r.Retrieve(ctx, "is there free wifi for guests", 3)
	if err != nil {
		t.Fatalf("retrieve: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 results, got %d", len(got))
	}
	if got[0].Passage.ID != "wifi" {
		t.Fatalf("top result = %q, want wifi", got[0].Passage.ID)
	}
	if got[0].Score <= got[1].Score {
		t.Fatalf("results not sorted by score: %v", got)
	}
}

func TestSearchEmptyStore(t *testing.T) {
	got, err := NewMemStore().Search(context.Background(), []float32{1, 0}, 5)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil for empty store, got %v", got)
	}
}
