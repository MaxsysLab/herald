// Command herald is the entry point for the Herald virtual front-office officer.
//
// The `demo` subcommand runs the full answer pipeline entirely offline (fake LLM and
// embedder, no API keys) so the core behavior — deterministic facts, grounded
// retrieval, and safe escalation — can be seen immediately after a clone.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/MaxsysLab/herald/internal/core"
	"github.com/MaxsysLab/herald/internal/embed"
	"github.com/MaxsysLab/herald/internal/knowledge"
	"github.com/MaxsysLab/herald/internal/llm"
	"github.com/MaxsysLab/herald/internal/officer"
	"github.com/MaxsysLab/herald/internal/server"
)

const version = "0.1.0-dev"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "version", "-v", "--version":
		fmt.Println("herald " + version)
	case "demo":
		runDemo()
	case "serve":
		runServe(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: herald <command>\n\ncommands:\n  serve     start the web-chat server + management console\n  demo      run the offline answer-pipeline demo (no API keys needed)\n  version   print the version\n\nserve flags:\n  -addr string   listen address (default \":8080\")\n  -config string config file path (created from a sample if missing)")
}

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := fs.String("addr", ":8080", "listen address")
	cfgPath := fs.String("config", "", "config file path (created from a sample if missing)")
	_ = fs.Parse(args)

	srv, err := server.New(*cfgPath)
	if err != nil {
		log.Fatalf("herald: %v", err)
	}
	log.Printf("herald serving on %s  (console: http://localhost%s/admin)", *addr, normalizeAddr(*addr))
	if err := srv.Run(*addr); err != nil {
		log.Fatalf("herald: %v", err)
	}
}

func normalizeAddr(addr string) string {
	if len(addr) > 0 && addr[0] == ':' {
		return addr
	}
	return ":" + addr
}

func runDemo() {
	ctx := context.Background()

	profile := &knowledge.BusinessProfile{
		Name:       "Bloom Café",
		Tagline:    "Neighborhood coffee, done right",
		Industry:   "coffee shop",
		BrandVoice: "warm, concise, welcoming",
		Services: []knowledge.Service{
			{Name: "Cappuccino", Price: &knowledge.Money{Amount: "4.50", Currency: "USD"}, Aliases: []string{"cap"}},
			{Name: "Cold Brew", Price: &knowledge.Money{Amount: "5.25", Currency: "USD"}},
			{Name: "Catering", Description: "Event catering", Price: nil}, // price not listed → must defer
		},
		Hours: []knowledge.Hour{
			{Day: "Mon-Fri", Open: "07:00", Close: "18:00"},
			{Day: "Sat-Sun", Open: "08:00", Close: "16:00"},
		},
		Policies: []knowledge.Policy{
			{Topic: "refund", Text: "We happily remake or refund any drink you're not satisfied with, no receipt needed."},
		},
		Contacts: []knowledge.Contact{
			{Kind: "escalation", Value: "manager@bloomcafe.example", Label: "Café manager"},
		},
	}

	retriever := &knowledge.Retriever{Embedder: embed.FakeEmbedder{}, Store: knowledge.NewMemStore()}
	_ = retriever.Index(ctx, []knowledge.Passage{
		{ID: "wifi", Source: "kb:amenities.md", Text: "Free guest wifi is available; the network is BloomGuest and the password is on your receipt."},
		{ID: "parking", Source: "kb:amenities.md", Text: "Street parking is free after 6pm; a paid lot is next door during the day."},
		{ID: "allergens", Source: "kb:menu.md", Text: "Oat, almond, and soy milk are available. Ask staff about allergens in pastries."},
	})

	// A fake generator stands in for the swappable LLM so the demo needs no key.
	fake := &llm.FakeClient{Reply: "Yes! Free guest wifi is available — the network is BloomGuest and the password is printed on your receipt."}

	// The fake embedder has a high noise floor, so the grounding threshold sits well
	// above zero — junk queries must fall below it and escalate rather than get a
	// confident wrong answer. A real embedder (Voyage/local) has a cleaner separation.
	off := &officer.Officer{Profile: profile, Retriever: retriever, LLM: fake, MinScore: 0.25}

	queries := []string{
		"How much is a cappuccino?",    // structured fact → exact price
		"What are your hours?",         // structured fact → hours
		"Do you have wifi?",            // grounded retrieval → LLM draft
		"How much does catering cost?", // in scope, price not listed → escalate
		"What's the meaning of life?",  // out of knowledge → escalate
	}

	fmt.Printf("Herald demo — %s\n\n", profile.Name)
	for _, q := range queries {
		reply, err := off.Answer(ctx, core.Query{Text: q, Channel: core.ChannelWebChat})
		if err != nil {
			fmt.Printf("Q: %s\n  ERROR: %v\n\n", q, err)
			continue
		}
		fmt.Printf("Q: %s\n  [%s] %s\n", q, reply.Verdict, reply.Text)
		for _, c := range reply.Citations {
			fmt.Printf("    ↳ cite %s (%.2f)\n", c.Source, c.Score)
		}
		if reply.Escalation != nil {
			fmt.Printf("    ↳ handoff → %s (%s)\n", reply.Escalation.Contact, reply.Escalation.Reason)
		}
		fmt.Println()
	}
}
