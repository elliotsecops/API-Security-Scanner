package discovery

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"api-security-scanner/types"
)

func TestDiscoverEndpoints(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "<html><body><a href=\"/api/users\">Users</a><a href=\"/static/logo.png\">Logo</a></body></html>")
	})
	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		fmt.Fprint(w, `{"users":[]}`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	d := NewAPIDiscovery(DiscoveryConfig{
		Enabled:     true,
		MaxDepth:    1,
		FollowLinks: true,
	})
	// speed up HTTP client (no need but ensure)
	d.client.Timeout = 2 * time.Second

	endpoints, err := d.DiscoverEndpoints(server.URL + "/")
	if err != nil {
		t.Fatalf("unexpected error discovering endpoints: %v", err)
	}

	target := types.APIEndpoint{URL: server.URL + "/api/users", Method: http.MethodGet}
	if !containsEndpoint(endpoints, target) {
		t.Fatalf("expected discovered endpoints to contain %v, got %#v", target, endpoints)
	}

	for _, ep := range endpoints {
		if ep.Method == http.MethodPost || ep.Method == http.MethodPut || ep.Method == http.MethodDelete || ep.Method == http.MethodPatch {
			t.Fatalf("did not expect mutating methods for endpoint, got %s", ep.Method)
		}
	}
}

func TestDiscoverParameters(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.RawQuery {
		case "id=1", "test=value":
			fmt.Fprint(w, `{"ok":true}`)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	d := NewAPIDiscovery(DiscoveryConfig{
		Enabled:        true,
		DiscoverParams: true,
	})
	d.client.Timeout = 2 * time.Second

	endpoint := types.APIEndpoint{URL: server.URL + "/api/search"}
	params, err := d.DiscoverParameters(endpoint)
	if err != nil {
		t.Fatalf("unexpected error discovering params: %v", err)
	}

	if len(params) == 0 {
		t.Fatal("expected at least one discovered parameter")
	}

	if !containsString(params, "id=1") {
		t.Fatalf("expected discovered params to include id=1, got %v", params)
	}
}

func containsEndpoint(list []types.APIEndpoint, target types.APIEndpoint) bool {
	for _, ep := range list {
		if ep.URL == target.URL && ep.Method == target.Method {
			return true
		}
	}
	return false
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
