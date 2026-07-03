package knowledge

import "testing"

func sample() *BusinessProfile {
	return &BusinessProfile{
		Name: "Bloom Café",
		Services: []Service{
			{Name: "Cappuccino", Price: &Money{Amount: "4.50", Currency: "USD"}, Aliases: []string{"cap"}},
			{Name: "Catering", Price: nil},
		},
		Policies: []Policy{{Topic: "refund", Text: "Full refund within 30 days."}},
		Contacts: []Contact{
			{Kind: "email", Value: "hi@bloom.example"},
			{Kind: "escalation", Value: "manager@bloom.example"},
		},
	}
}

func TestFindService(t *testing.T) {
	p := sample()
	if s := p.FindService("Cappuccino"); s == nil || s.Name != "Cappuccino" {
		t.Fatalf("exact match failed: %+v", s)
	}
	if s := p.FindService("cap"); s == nil || s.Name != "Cappuccino" {
		t.Fatalf("alias match failed: %+v", s)
	}
	if s := p.FindService("how much is a cappuccino"); s == nil || s.Name != "Cappuccino" {
		t.Fatalf("substring match failed: %+v", s)
	}
	if s := p.FindService("espresso martini"); s != nil {
		t.Fatalf("expected no match, got %+v", s)
	}
}

func TestPriceOf(t *testing.T) {
	p := sample()
	if svc, ok := p.PriceOf("cappuccino"); !ok || svc.Price.String() != "USD 4.50" {
		t.Fatalf("cappuccino price lookup failed: ok=%v svc=%+v", ok, svc)
	}
	// Catering has no listed price → must report not-found so the officer defers.
	if _, ok := p.PriceOf("catering"); ok {
		t.Fatal("catering has no price; PriceOf must return false")
	}
	if _, ok := p.PriceOf("unicorn"); ok {
		t.Fatal("unknown service must return false")
	}
}

func TestPolicyOn(t *testing.T) {
	p := sample()
	if pol := p.PolicyOn("refund"); pol == nil || pol.Text == "" {
		t.Fatalf("refund policy lookup failed: %+v", pol)
	}
	if pol := p.PolicyOn("nonexistent"); pol != nil {
		t.Fatalf("expected nil policy, got %+v", pol)
	}
}

func TestEscalationContactPrefersEscalationKind(t *testing.T) {
	p := sample()
	if got := p.EscalationContact(); got != "manager@bloom.example" {
		t.Fatalf("escalation contact = %q, want manager@bloom.example", got)
	}
}
