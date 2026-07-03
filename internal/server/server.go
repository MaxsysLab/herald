// Package server exposes the officer over HTTP: a chat API a website calls, an
// embeddable chat widget, a standalone test page, and a management console. It runs
// out of the box with no API keys (structured facts + verbatim retrieval), and lights
// up natural drafting + hosted embeddings when the relevant keys are present.
package server

import (
	"context"
	"encoding/json"
	htmltemplate "html/template"
	"net/http"
	"strconv"
	"sync"
	texttemplate "text/template"

	"github.com/MaxsysLab/herald/internal/config"
	"github.com/MaxsysLab/herald/internal/core"
	embedpkg "github.com/MaxsysLab/herald/internal/embed"
	"github.com/MaxsysLab/herald/internal/knowledge"
	"github.com/MaxsysLab/herald/internal/llm"
	"github.com/MaxsysLab/herald/internal/officer"
)

// Server holds the live configuration and the officer built from it.
type Server struct {
	mu       sync.RWMutex
	cfgPath  string // where config is persisted; empty = in-memory only
	cfg      *config.ProjectConfig
	off      *officer.Officer
	embedNm  string
	llmReady bool

	dash   *htmltemplate.Template
	widget *texttemplate.Template
}

// New builds a Server from a config file. A missing/empty path uses a built-in sample
// config so the server is immediately runnable and demoable.
func New(cfgPath string) (*Server, error) {
	var cfg *config.ProjectConfig
	if cfgPath != "" {
		if c, err := config.Load(cfgPath); err == nil {
			cfg = c
		}
	}
	if cfg == nil {
		cfg = DefaultConfig()
		if cfgPath != "" {
			_ = cfg.Save(cfgPath)
		}
	}
	s := &Server{
		cfgPath: cfgPath,
		cfg:     cfg,
		dash:    htmltemplate.Must(htmltemplate.New("dash").Parse(dashboardHTML)),
		widget:  texttemplate.Must(texttemplate.New("widget").Parse(widgetJS)),
	}
	if err := s.rebuild(); err != nil {
		return nil, err
	}
	return s, nil
}

// rebuild reconstructs the officer from the current config. Callers hold s.mu.
func (s *Server) rebuild() error {
	emb, err := embedpkg.Resolve(s.cfg.Embed)
	if err != nil {
		return err
	}
	store := knowledge.NewMemStore()
	retr := &knowledge.Retriever{Embedder: emb, Store: store}

	ps := make([]knowledge.Passage, len(s.cfg.Knowledge))
	copy(ps, s.cfg.Knowledge)
	for i := range ps {
		if ps[i].ID == "" {
			ps[i].ID = "kb-" + strconv.Itoa(i)
		}
	}
	if len(ps) > 0 {
		if err := retr.Index(context.Background(), ps); err != nil {
			return err
		}
	}

	llmc := resolveLLM(s.cfg.LLM)
	s.off = &officer.Officer{Profile: &s.cfg.Profile, Retriever: retr, LLM: llmc, MinScore: s.cfg.MinScore}
	s.embedNm = emb.Name()
	s.llmReady = llmc != nil
	return nil
}

// resolveLLM returns a live client, or nil when a keyed provider has no key — in which
// case the officer answers from retrieved text verbatim rather than failing.
func resolveLLM(cfg llm.ProviderConfig) llm.Client {
	c, err := llm.Resolve(cfg)
	if err != nil {
		return nil
	}
	keyed := cfg.Provider == "" || cfg.Provider == "groq" || cfg.Provider == "openai" || cfg.Provider == "grok"
	if oc, ok := c.(*llm.OpenAICompatClient); ok && oc.APIKey == "" && keyed {
		return nil
	}
	return c
}

// Handler returns the HTTP router.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok")) })
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/admin", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/admin", s.handleAdmin)
	mux.HandleFunc("/admin/config", s.handleConfigPost)
	mux.HandleFunc("/admin/knowledge", s.handleKnowledgePost)
	mux.HandleFunc("/api/chat", s.cors(s.handleChat))
	mux.HandleFunc("/api/greeting", s.cors(s.handleGreeting))
	mux.HandleFunc("/widget.js", s.handleWidgetJS)
	mux.HandleFunc("/widget", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(chatPageHTML))
	})
	return mux
}

// Run starts the HTTP server.
func (s *Server) Run(addr string) error {
	srv := &http.Server{Addr: addr, Handler: s.Handler()}
	return srv.ListenAndServe()
}

// cors allows the chat API to be called from an embedded widget on another origin.
func (s *Server) cors(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h(w, r)
	}
}

type chatRequest struct {
	Message string `json:"message"`
}

type citationDTO struct {
	Source string  `json:"source"`
	Quote  string  `json:"quote"`
	Score  float64 `json:"score"`
}

type escalationDTO struct {
	Reason  string `json:"reason"`
	Contact string `json:"contact"`
	Summary string `json:"summary"`
}

type chatReplyDTO struct {
	Verdict    string         `json:"verdict"`
	Text       string         `json:"text"`
	Citations  []citationDTO  `json:"citations,omitempty"`
	Escalation *escalationDTO `json:"escalation,omitempty"`
}

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var req chatRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	s.mu.RLock()
	off := s.off
	s.mu.RUnlock()

	reply, err := off.Answer(r.Context(), core.Query{Text: req.Message, Channel: core.ChannelWebChat})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, toDTO(reply))
}

func (s *Server) handleGreeting(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	off := s.off
	s.mu.RUnlock()
	writeJSON(w, map[string]string{"greeting": off.Greeting()})
}

func (s *Server) handleAdmin(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	cfg := s.cfg
	pretty, _ := json.MarshalIndent(cfg, "", "  ")
	data := map[string]any{
		"BusinessName": cfg.Profile.Name,
		"OfficerName":  s.off.Name(),
		"ConfigJSON":   string(pretty),
		"KBCount":      len(cfg.Knowledge),
		"MinScore":     cfg.MinScore,
		"EmbedName":    s.embedNm,
		"LLMReady":     s.llmReady,
		"LLMMode":      llmMode(s.llmReady),
		"Base":         baseURL(r),
	}
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.dash.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleConfigPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var cfg config.ProjectConfig
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&cfg); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	prev := s.cfg
	s.cfg = &cfg
	if err := s.rebuild(); err != nil {
		s.cfg = prev
		_ = s.rebuild()
		http.Error(w, "rebuild failed: "+err.Error(), http.StatusBadRequest)
		return
	}
	s.persist()
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleKnowledgePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var p knowledge.Passage
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&p); err != nil || p.Text == "" {
		http.Error(w, "need {source, text}", http.StatusBadRequest)
		return
	}
	if p.Source == "" {
		p.Source = "kb:added"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg.Knowledge = append(s.cfg.Knowledge, p)
	if err := s.rebuild(); err != nil {
		http.Error(w, "index failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	s.persist()
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleWidgetJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	// {{.Base}} is emitted as a quoted JS string literal (var BASE={{.Base}};).
	_ = s.widget.Execute(w, map[string]any{"Base": jsString(baseURL(r))})
}

// persist writes the current config to disk if a path is configured. Callers hold s.mu.
func (s *Server) persist() {
	if s.cfgPath != "" {
		_ = s.cfg.Save(s.cfgPath)
	}
}

func toDTO(r core.Reply) chatReplyDTO {
	dto := chatReplyDTO{Verdict: string(r.Verdict), Text: r.Text}
	for _, c := range r.Citations {
		dto.Citations = append(dto.Citations, citationDTO{Source: c.Source, Quote: c.Quote, Score: c.Score})
	}
	if r.Escalation != nil {
		dto.Escalation = &escalationDTO{Reason: r.Escalation.Reason, Contact: r.Escalation.Contact, Summary: r.Escalation.Summary}
	}
	return dto
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func baseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if p := r.Header.Get("X-Forwarded-Proto"); p != "" {
		scheme = p
	}
	return scheme + "://" + r.Host
}

func llmMode(ready bool) string {
	if ready {
		return "drafting"
	}
	return "verbatim (no key)"
}

// jsString returns s as a safely-quoted JavaScript string literal.
func jsString(s string) string { b, _ := json.Marshal(s); return string(b) }
