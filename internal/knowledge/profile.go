// Package knowledge is Herald's business-awareness layer — how the officer "knows"
// the business it fronts. It has two halves:
//
//   - The structured BusinessProfile (this file): facts of record — services, prices,
//     hours, policies, contacts. Volatile, must-be-exact facts (especially prices) are
//     answered by deterministic lookup here, never by fuzzy retrieval. This is the core
//     defense against the "price of apples → color of oranges" failure.
//   - The knowledge base (store.go / retrieve.go): unstructured docs/FAQ/site content,
//     retrieved by vector similarity and always cited.
package knowledge

import "strings"

// Money keeps a price as an exact string plus currency. Prices are never floats —
// a receptionist quoting a rounded price is a real-world failure.
type Money struct {
	Amount   string // exact, e.g. "4.99"
	Currency string // ISO 4217, e.g. "USD"
}

// String renders a price like "USD 4.99".
func (m Money) String() string {
	if m.Amount == "" {
		return ""
	}
	if m.Currency == "" {
		return m.Amount
	}
	return m.Currency + " " + m.Amount
}

// Service is an offering the business sells or provides.
type Service struct {
	Name        string
	Description string
	Price       *Money   // nil means "not listed / contact us"
	Aliases     []string // alternate names customers might use
}

// Hour is one row of operating hours.
type Hour struct {
	Day   string // "Mon", "Mon-Fri", "Sat", ...
	Open  string // "09:00"
	Close string // "17:00"
	Note  string // e.g. "by appointment"
}

// Policy is a stated business policy on a topic.
type Policy struct {
	Topic string // "returns", "booking", "payment", ...
	Text  string
}

// Contact is a way to reach the business or a human. Kind "escalation" marks the
// human hand-off target the officer uses when it must defer.
type Contact struct {
	Kind  string // "phone", "email", "address", "escalation"
	Value string
	Label string
}

// BusinessProfile is the structured identity of one business — the "facts of record"
// an onboarded officer speaks from.
type BusinessProfile struct {
	Name        string
	Tagline     string
	Industry    string
	Description string
	OfficerName string // the persona the officer introduces itself as; defaults to "Herald"
	BrandVoice  string // tone/persona guidance for the officer
	Services    []Service
	Hours       []Hour
	Policies    []Policy
	Contacts    []Contact
}

// FindService returns the service whose name or an alias matches query
// (case-insensitive), preferring exact matches over substring matches. Returns nil
// when nothing matches — the officer must then defer rather than invent.
func (p *BusinessProfile) FindService(query string) *Service {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil
	}
	var substr *Service
	for i := range p.Services {
		s := &p.Services[i]
		for _, name := range append([]string{s.Name}, s.Aliases...) {
			n := strings.ToLower(strings.TrimSpace(name))
			if n == "" {
				continue
			}
			if n == q {
				return s
			}
			if substr == nil && (strings.Contains(q, n) || strings.Contains(n, q)) {
				substr = s
			}
		}
	}
	return substr
}

// PriceOf returns the service and true only when a matching service has a listed
// price. A match without a price, or no match, returns false — signalling the officer
// to defer instead of guessing a number.
func (p *BusinessProfile) PriceOf(query string) (*Service, bool) {
	s := p.FindService(query)
	if s == nil || s.Price == nil || s.Price.Amount == "" {
		return s, false
	}
	return s, true
}

// PolicyOn returns the policy for a topic (case-insensitive), or nil.
func (p *BusinessProfile) PolicyOn(topic string) *Policy {
	t := strings.ToLower(strings.TrimSpace(topic))
	for i := range p.Policies {
		if strings.ToLower(p.Policies[i].Topic) == t {
			return &p.Policies[i]
		}
	}
	return nil
}

// EscalationContact returns the value of the first "escalation" contact, falling back
// to the first phone, then the first email, then empty.
func (p *BusinessProfile) EscalationContact() string {
	var phone, email string
	for _, c := range p.Contacts {
		switch strings.ToLower(c.Kind) {
		case "escalation":
			if c.Value != "" {
				return c.Value
			}
		case "phone":
			if phone == "" {
				phone = c.Value
			}
		case "email":
			if email == "" {
				email = c.Value
			}
		}
	}
	if phone != "" {
		return phone
	}
	return email
}
