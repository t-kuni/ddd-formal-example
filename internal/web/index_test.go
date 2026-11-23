package web

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIndexShowsBuildTime(t *testing.T) {
	previous := builtAt
	t.Cleanup(func() { builtAt = previous })

	for _, tc := range []struct {
		name, timestamp, want string
	}{
		{"Docker build", "2026-11-22T20:30:45Z", `<time datetime="2026-11-22T20:30:45Z">2026-11-23 05:30:45 JST</time>`},
		{"local Go build", "", "ビルド 日時未記録"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			builtAt = tc.timestamp
			response := httptest.NewRecorder()
			(&API{}).Handler().ServeHTTP(response, httptest.NewRequest("GET", "/", nil))
			if response.Code != 200 || !strings.Contains(response.Body.String(), tc.want) {
				t.Fatalf("expected %q in successful response, got status %d and body %s", tc.want, response.Code, response.Body.String())
			}
			if got := response.Header().Get("Cache-Control"); got != "no-store" {
				t.Errorf("Cache-Control = %q, want no-store", got)
			}
		})
	}
}
