package cli

import "os"

// BaseURL returns the server URL to talk to, honoring GTZY_URL.
func BaseURL() string {
	if u := os.Getenv("GTZY_URL"); u != "" {
		return u
	}
	return "http://localhost:8420"
}
