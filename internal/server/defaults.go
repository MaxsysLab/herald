package server

import (
	"github.com/MaxsysLab/herald/internal/config"
	"github.com/MaxsysLab/herald/internal/embed"
	"github.com/MaxsysLab/herald/internal/knowledge"
	"github.com/MaxsysLab/herald/internal/llm"
)

// DefaultConfig is a ready-to-run sample so `herald serve` works with zero setup. It
// mirrors the offline demo's café: exact-fact services, hours, a policy, an escalation
// contact, and a few knowledge passages. Swap it out via the console or a config file.
func DefaultConfig() *config.ProjectConfig {
	return &config.ProjectConfig{
		Name: "sample",
		Profile: knowledge.BusinessProfile{
			Name:        "Bloom Café",
			Tagline:     "Neighborhood coffee, done right",
			Industry:    "coffee shop",
			OfficerName: "Herald",
			BrandVoice:  "warm, concise, welcoming",
			Services: []knowledge.Service{
				{Name: "Cappuccino", Price: &knowledge.Money{Amount: "4.50", Currency: "USD"}, Aliases: []string{"cap"}},
				{Name: "Cold Brew", Price: &knowledge.Money{Amount: "5.25", Currency: "USD"}},
				{Name: "Catering", Description: "Event catering", Price: nil},
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
		},
		Knowledge: []knowledge.Passage{
			{ID: "wifi", Source: "kb:amenities.md", Text: "free guest wifi is available; the network is BloomGuest and the password is on your receipt"},
			{ID: "parking", Source: "kb:amenities.md", Text: "street parking is free after six pm; a paid lot is next door during the day"},
			{ID: "allergens", Source: "kb:menu.md", Text: "oat, almond, and soy milk are available; ask staff about allergens in pastries"},
		},
		LLM:      llm.ProviderConfig{Provider: llm.DefaultProvider},
		Embed:    embed.ProviderConfig{Provider: "fake"},
		MinScore: 0.25,
	}
}
