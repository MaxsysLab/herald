package llm

import "context"

// FakeClient is an offline, deterministic Client for tests and the demo. It records
// the last request it saw and returns a canned reply (or error).
type FakeClient struct {
	Reply   string
	Err     error
	LastReq Request
}

// Name identifies the fake provider.
func (f *FakeClient) Name() string { return "fake" }

// Complete records the request and returns the canned reply/error.
func (f *FakeClient) Complete(_ context.Context, req Request) (Response, error) {
	f.LastReq = req
	if f.Err != nil {
		return Response{}, f.Err
	}
	return Response{Text: f.Reply, Model: "fake"}, nil
}
