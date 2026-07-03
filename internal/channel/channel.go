// Package channel is the pluggable channel seam. Web chat is the default and only
// always-on channel; email, third-party messaging (WhatsApp/SMS), and voice are opt-in
// modules that adapt their transport to the same inbound/outbound shape and route
// through the officer. Concrete transports (an HTTP chat endpoint, an IMAP poller) land
// on top of this interface in later increments.
package channel

import (
	"context"

	"github.com/MaxsysLab/herald/internal/core"
)

// Inbound is a normalized customer message arriving on some channel.
type Inbound struct {
	ConversationID string
	Query          core.Query
}

// Outbound is the officer's reply, normalized for delivery back out the channel.
type Outbound struct {
	ConversationID string
	Reply          core.Reply
}

// Handler answers an inbound message. The officer satisfies this shape via an adapter;
// channels depend on the interface, never on the officer package directly.
type Handler interface {
	Handle(ctx context.Context, in Inbound) (Outbound, error)
}

// Channel is a transport that receives inbound messages and delivers outbound replies.
type Channel interface {
	Name() string
}
