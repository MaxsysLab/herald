package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	s, err := New("") // in-memory default config; no keys → verbatim mode
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return httptest.NewServer(s.Handler())
}

func postChat(t *testing.T, base, msg string) chatReplyDTO {
	t.Helper()
	body, _ := json.Marshal(chatRequest{Message: msg})
	resp, err := http.Post(base+"/api/chat", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("chat post: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("chat status = %d", resp.StatusCode)
	}
	var dto chatReplyDTO
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return dto
}

func TestChatAPIStructuredPrice(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()
	r := postChat(t, ts.URL, "how much is a cappuccino?")
	if r.Verdict != "answer" || !strings.Contains(r.Text, "4.50") {
		t.Fatalf("price reply wrong: %+v", r)
	}
	if len(r.Citations) == 0 || !strings.HasPrefix(r.Citations[0].Source, "profile:pricing") {
		t.Fatalf("expected profile citation: %+v", r.Citations)
	}
}

func TestChatAPIEscalates(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()
	r := postChat(t, ts.URL, "explain quantum chromodynamics please")
	if r.Verdict != "escalate" || r.Escalation == nil {
		t.Fatalf("expected escalation: %+v", r)
	}
}

func TestGreetingNamesTheOfficer(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/api/greeting")
	if err != nil {
		t.Fatalf("greeting: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	var g map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&g)
	if !strings.Contains(g["greeting"], "Herald") {
		t.Fatalf("greeting should name Herald: %q", g["greeting"])
	}
}

func TestAddKnowledgeThenAnswer(t *testing.T) {
	s, err := New("")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	// A distinctive passage the default KB doesn't contain.
	kb, _ := json.Marshal(map[string]string{
		"source": "kb:loyalty",
		"text":   "loyalty members earn a free pastry after ten stamped visits to the cafe",
	})
	resp, err := http.Post(ts.URL+"/admin/knowledge", "application/json", bytes.NewReader(kb))
	if err != nil || resp.StatusCode != http.StatusNoContent {
		t.Fatalf("add knowledge: err=%v status=%v", err, resp.StatusCode)
	}
	_ = resp.Body.Close()

	r := postChat(t, ts.URL, "do loyalty members earn a free pastry")
	if r.Verdict != "answer" {
		t.Fatalf("expected grounded answer after indexing, got %+v", r)
	}
	if len(r.Citations) == 0 || r.Citations[0].Source != "kb:loyalty" {
		t.Fatalf("expected kb:loyalty citation: %+v", r.Citations)
	}
}

func TestWidgetAndAdminServe(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	wj, err := http.Get(ts.URL + "/widget.js")
	if err != nil {
		t.Fatalf("widget.js: %v", err)
	}
	body, _ := io.ReadAll(wj.Body)
	_ = wj.Body.Close()
	if wj.StatusCode != 200 || !strings.Contains(string(body), "BASE=") {
		t.Fatalf("widget.js not served correctly (status %d)", wj.StatusCode)
	}

	ad, err := http.Get(ts.URL + "/admin")
	if err != nil {
		t.Fatalf("admin: %v", err)
	}
	_ = ad.Body.Close()
	if ad.StatusCode != 200 {
		t.Fatalf("admin status = %d", ad.StatusCode)
	}
}
