package main

import "testing"

func TestConsoleURL(t *testing.T) {
	cases := []struct {
		addr string
		want string
	}{
		{":8080", "http://localhost:8080/admin"},
		{"127.0.0.1:8080", "http://127.0.0.1:8080/admin"},
		{"example.org:8080", "http://example.org:8080/admin"},
		{"[::1]:8080", "http://[::1]:8080/admin"},
		{"8080", "http://localhost:8080/admin"},
	}
	for _, c := range cases {
		if got := consoleURL(c.addr); got != c.want {
			t.Errorf("consoleURL(%q) = %q, want %q", c.addr, got, c.want)
		}
	}
}
