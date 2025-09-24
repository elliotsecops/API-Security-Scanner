package scanner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"api-security-scanner/types"
)

func TestTestAuth(t *testing.T) {
	const (
		username = "user"
		password = "pass"
	)

	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			if !ok || user != username || pass != password {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if r.Method != http.MethodGet {
				t.Fatalf("unexpected method: %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		endpoint := types.APIEndpoint{URL: server.URL, Method: http.MethodGet}
		err := testAuth(endpoint, Auth{Username: username, Password: password})
		if err != nil {
			t.Fatalf("expected success, got error: %v", err)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		endpoint := types.APIEndpoint{URL: server.URL, Method: http.MethodGet}
		err := testAuth(endpoint, Auth{Username: username, Password: password})
		if err == nil {
			t.Fatal("expected authentication error, got nil")
		}
		var authErr AuthError
		if !errors.As(err, &authErr) {
			t.Fatalf("expected AuthError, got %T", err)
		}
	})

	t.Run("unexpected status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		}))
		defer server.Close()

		endpoint := types.APIEndpoint{URL: server.URL, Method: http.MethodGet}
		err := testAuth(endpoint, Auth{Username: username, Password: password})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var authErr AuthError
		if errors.As(err, &authErr) {
			t.Fatalf("expected generic error, got AuthError: %v", err)
		}
	})
}

func TestTestHTTPMethod(t *testing.T) {
	const (
		username = "user"
		password = "pass"
	)

	t.Run("method allowed", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		endpoint := types.APIEndpoint{URL: server.URL, Method: http.MethodPost}
		err := testHTTPMethod(endpoint, Auth{Username: username, Password: password})
		if err != nil {
			t.Fatalf("expected success, got error: %v", err)
		}
	})

	t.Run("method disallowed", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}))
		defer server.Close()

		endpoint := types.APIEndpoint{URL: server.URL, Method: http.MethodPost}
		err := testHTTPMethod(endpoint, Auth{Username: username, Password: password})
		if err == nil {
			t.Fatal("expected HTTPMethodError, got nil")
		}
		var methodErr HTTPMethodError
		if !errors.As(err, &methodErr) {
			t.Fatalf("expected HTTPMethodError, got %T", err)
		}
	})

	t.Run("unexpected status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		}))
		defer server.Close()

		endpoint := types.APIEndpoint{URL: server.URL, Method: http.MethodPost}
		err := testHTTPMethod(endpoint, Auth{Username: username, Password: password})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var methodErr HTTPMethodError
		if errors.As(err, &methodErr) {
			t.Fatalf("expected generic error, got HTTPMethodError: %v", err)
		}
	})
}

func TestTestInjection(t *testing.T) {
	const (
		username = "user"
		password = "pass"
	)

	t.Run("no injection detected", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			user, pass, ok := r.BasicAuth()
			if !ok || user != username || pass != password {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed reading body: %v", err)
			}
			switch string(body) {
			case "%s", "safe":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("{\"status\":\"ok\"}"))
			default:
				t.Fatalf("unexpected body: %s", string(body))
			}
		}))
		defer server.Close()

		endpoint := types.APIEndpoint{URL: server.URL, Method: http.MethodPost, Body: "%s"}
		err := testInjection(endpoint, Auth{Username: username, Password: password}, []string{"safe"})
		if err != nil {
			t.Fatalf("expected no injection, got error: %v", err)
		}
	})

	t.Run("injection detected", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			user, pass, ok := r.BasicAuth()
			if !ok || user != username || pass != password {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed reading body: %v", err)
			}
			switch string(body) {
			case "%s":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("baseline"))
			case "' OR '1'='1":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("You have an error in your SQL syntax"))
			default:
				t.Fatalf("unexpected body: %s", string(body))
			}
		}))
		defer server.Close()

		endpoint := types.APIEndpoint{URL: server.URL, Method: http.MethodPost, Body: "%s"}
		err := testInjection(endpoint, Auth{Username: username, Password: password}, []string{"' OR '1'='1"})
		if err == nil {
			t.Fatal("expected injection error, got nil")
		}
		var injErr InjectionError
		if !errors.As(err, &injErr) {
			t.Fatalf("expected InjectionError, got %T", err)
		}
	})
}

func TestTestXSS(t *testing.T) {
	const (
		username = "user"
		password = "pass"
	)

	t.Run("no xss detected", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			user, pass, ok := r.BasicAuth()
			if !ok || user != username || pass != password {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed reading body: %v", err)
			}
			switch string(body) {
			case `{"payload":"value"}`:
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"ok"}`))
			case `{"payload":"<script>alert('XSS')</script>"}`:
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"ok"}`))
			default:
				t.Fatalf("unexpected body: %s", string(body))
			}
		}))
		defer server.Close()

		endpoint := types.APIEndpoint{URL: server.URL, Method: http.MethodPost, Body: `{"payload":"value"}`}
		err := testXSS(endpoint, Auth{Username: username, Password: password}, []string{"<script>alert('XSS')</script>"})
		if err != nil {
			t.Fatalf("expected no xss, got error: %v", err)
		}
	})

	t.Run("xss detected", func(t *testing.T) {
		payload := "<script>alert('XSS')</script>"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			user, pass, ok := r.BasicAuth()
			if !ok || user != username || pass != password {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed reading body: %v", err)
			}
			switch string(body) {
			case `{"payload":"value"}`:
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"ok"}`))
			case `{"payload":"<script>alert('XSS')</script>"}`:
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(fmt.Sprintf("<script>%s</script>", payload)))
			default:
				t.Fatalf("unexpected body: %s", string(body))
			}
		}))
		defer server.Close()

		endpoint := types.APIEndpoint{URL: server.URL, Method: http.MethodPost, Body: `{"payload":"value"}`}
		err := testXSS(endpoint, Auth{Username: username, Password: password}, []string{payload})
		if err == nil {
			t.Fatal("expected xss error, got nil")
		}
		var xssErr XSSError
		if !errors.As(err, &xssErr) {
			t.Fatalf("expected XSSError, got %T", err)
		}
	})
}

func TestTestHeaderSecurity(t *testing.T) {
	const (
		username = "user"
		password = "pass"
	)

	t.Run("no issues", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			user, pass, ok := r.BasicAuth()
			if !ok || user != username || pass != password {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
			w.Header().Set("Content-Security-Policy", "default-src 'self'")
			w.Header().Set("Access-Control-Allow-Origin", "https://example.com")
			w.Header().Add("Set-Cookie", "session=abc; Secure; HttpOnly; SameSite=Strict")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		endpoint := types.APIEndpoint{URL: server.URL, Method: http.MethodGet}
		if err := testHeaderSecurity(endpoint, Auth{Username: username, Password: password}, map[string]string{"X-Custom": "value"}); err != nil {
			t.Fatalf("expected no header issues, got error: %v", err)
		}
	})

	t.Run("issues detected", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			user, pass, ok := r.BasicAuth()
			if !ok || user != username || pass != password {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("Server", "nginx")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Add("Set-Cookie", "session=abc")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		endpoint := types.APIEndpoint{URL: server.URL, Method: http.MethodGet}
		err := testHeaderSecurity(endpoint, Auth{Username: username, Password: password}, nil)
		if err == nil {
			t.Fatal("expected header security error, got nil")
		}
		var hdrErr HeaderSecurityError
		if !errors.As(err, &hdrErr) {
			t.Fatalf("expected HeaderSecurityError, got %T", err)
		}
	})

	t.Run("custom headers forwarded", func(t *testing.T) {
		received := make(chan string, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			received <- r.Header.Get("X-Test-Passthrough")
			user, pass, ok := r.BasicAuth()
			if !ok || user != username || pass != password {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
			w.Header().Set("Content-Security-Policy", "default-src 'self'")
			w.Header().Set("Access-Control-Allow-Origin", "https://example.com")
			w.Header().Add("Set-Cookie", "session=abc; Secure; HttpOnly; SameSite=Strict")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		custom := map[string]string{"X-Test-Passthrough": "expected"}
		endpoint := types.APIEndpoint{URL: server.URL, Method: http.MethodGet}
		if err := testHeaderSecurity(endpoint, Auth{Username: username, Password: password}, custom); err != nil {
			t.Fatalf("expected no header issues, got error: %v", err)
		}
		select {
		case seen := <-received:
			if seen != "expected" {
				t.Fatalf("expected custom header to be forwarded, got %q", seen)
			}
		default:
			t.Fatal("expected server to receive custom header")
		}
	})
}

func TestTestAuthBypass(t *testing.T) {
	const (
		username = "user"
		password = "pass"
	)

	t.Run("no bypass", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			user, pass, ok := r.BasicAuth()
			switch {
			case !ok:
				w.WriteHeader(http.StatusUnauthorized)
			case user != username || pass != password:
				w.WriteHeader(http.StatusUnauthorized)
			default:
				w.WriteHeader(http.StatusOK)
			}
		}))
		defer server.Close()

		endpoint := types.APIEndpoint{URL: server.URL + "/secure", Method: http.MethodGet}
		if err := testAuthBypass(endpoint, Auth{Username: username, Password: password}); err != nil {
			t.Fatalf("expected no bypass detection, got error: %v", err)
		}
	})

	t.Run("missing auth bypass", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			if _, _, ok := r.BasicAuth(); !ok {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		endpoint := types.APIEndpoint{URL: server.URL + "/secure", Method: http.MethodGet}
		err := testAuthBypass(endpoint, Auth{Username: username, Password: password})
		if err == nil {
			t.Fatal("expected bypass detection when auth not enforced")
		}
		var bypassErr AuthBypassError
		if !errors.As(err, &bypassErr) {
			t.Fatalf("expected AuthBypassError, got %T", err)
		}
	})

	t.Run("header bypass", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			if r.Header.Get("X-Forwarded-For") == "127.0.0.1" {
				w.WriteHeader(http.StatusOK)
				return
			}
			if _, _, ok := r.BasicAuth(); ok {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		endpoint := types.APIEndpoint{URL: server.URL + "/secure", Method: http.MethodGet}
		err := testAuthBypass(endpoint, Auth{Username: username, Password: password})
		if err == nil {
			t.Fatal("expected bypass detection with spoofed headers")
		}
		var bypassErr AuthBypassError
		if !errors.As(err, &bypassErr) {
			t.Fatalf("expected AuthBypassError, got %T", err)
		}
	})
}

func TestTestParameterTampering(t *testing.T) {
	const (
		username = "user"
		password = "pass"
	)

	t.Run("no idor", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			user, pass, ok := r.BasicAuth()
			if !ok || user != username || pass != password {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			switch r.URL.Path {
			case "/items/1":
				w.WriteHeader(http.StatusOK)
			case "/items/2":
				w.WriteHeader(http.StatusForbidden)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		realAddr := strings.TrimPrefix(strings.TrimPrefix(server.URL, "http://"), "https://")
		mutatedAddr := strings.ReplaceAll(realAddr, "1", "2")
		restore := patchTransportForAddr(t, mutatedAddr, realAddr)
		defer restore()

		endpoint := types.APIEndpoint{URL: server.URL + "/items/1", Method: http.MethodPost, Body: `{"key":"value"}`}
		if err := testParameterTampering(endpoint, Auth{Username: username, Password: password}); err != nil {
			t.Fatalf("expected no IDOR detection, got error: %v", err)
		}
	})

	t.Run("idor detected", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			user, pass, ok := r.BasicAuth()
			if !ok || user != username || pass != password {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			switch r.URL.Path {
			case "/items/1":
				w.WriteHeader(http.StatusOK)
			case "/items/2":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"id":2}`))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		realAddr := strings.TrimPrefix(strings.TrimPrefix(server.URL, "http://"), "https://")
		mutatedAddr := strings.ReplaceAll(realAddr, "1", "2")
		restore := patchTransportForAddr(t, mutatedAddr, realAddr)
		defer restore()

		endpoint := types.APIEndpoint{URL: server.URL + "/items/1", Method: http.MethodPost, Body: `{"key":"value"}`}
		err := testParameterTampering(endpoint, Auth{Username: username, Password: password})
		if err == nil {
			t.Fatal("expected parameter tampering error, got nil")
		}
		var paramErr ParameterTamperingError
		if !errors.As(err, &paramErr) {
			t.Fatalf("expected ParameterTamperingError, got %T", err)
		}
	})
}

func TestTestNoSQLInjection(t *testing.T) {
	const (
		username = "user"
		password = "pass"
	)

	t.Run("no nosql detected", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			user, pass, ok := r.BasicAuth()
			if !ok || user != username || pass != password {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		}))
		defer server.Close()

		endpoint := types.APIEndpoint{URL: server.URL + "/data", Method: http.MethodPost, Body: `{"payload":"value"}`}
		if err := testNoSQLInjection(endpoint, Auth{Username: username, Password: password}, []string{"{$ne: null}"}); err != nil {
			t.Fatalf("expected no NoSQL detection, got error: %v", err)
		}
	})

	t.Run("nosql detected", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			user, pass, ok := r.BasicAuth()
			if !ok || user != username || pass != password {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed reading body: %v", err)
			}
			if strings.Contains(string(body), "{$ne: null}") {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("MongoError: injection detected"))
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		}))
		defer server.Close()

		endpoint := types.APIEndpoint{URL: server.URL + "/data", Method: http.MethodPost, Body: `{"payload":"value"}`}
		err := testNoSQLInjection(endpoint, Auth{Username: username, Password: password}, []string{"{$ne: null}"})
		if err == nil {
			t.Fatal("expected NoSQL injection error, got nil")
		}
		var nosqlErr NoSQLInjectionError
		if !errors.As(err, &nosqlErr) {
			t.Fatalf("expected NoSQLInjectionError, got %T", err)
		}
	})
}

func TestIndicatorsOfSQLInjection(t *testing.T) {
	t.Run("sql error message", func(t *testing.T) {
		if !indicatorsOfSQLInjection("You have an error in your SQL syntax", "ok") {
			t.Fatal("expected detection when SQL error message present")
		}
	})

	t.Run("response length delta", func(t *testing.T) {
		if !indicatorsOfSQLInjection(strings.Repeat("a", 10), "abc") {
			t.Fatal("expected detection based on length delta")
		}
	})

	t.Run("structure change", func(t *testing.T) {
		if !indicatorsOfSQLInjection("{", "{}") {
			t.Fatal("expected detection based on structure change")
		}
	})

	t.Run("no indicator", func(t *testing.T) {
		if indicatorsOfSQLInjection("stable", "stable") {
			t.Fatal("expected no detection when responses match")
		}
	})
}

func TestIndicatorsOfNoSQLInjection(t *testing.T) {
	payload := "{$ne: null}"

	t.Run("nosql error message", func(t *testing.T) {
		if !indicatorsOfNoSQLInjection("MongoError: bad query", "{}", payload) {
			t.Fatal("expected detection when NoSQL error message present")
		}
	})

	t.Run("payload reflected", func(t *testing.T) {
		if !indicatorsOfNoSQLInjection(payload, "{}", payload) {
			t.Fatal("expected detection when payload reflected")
		}
	})

	t.Run("response expansion", func(t *testing.T) {
		if !indicatorsOfNoSQLInjection(strings.Repeat("a", 100), "short", payload) {
			t.Fatal("expected detection when response expands")
		}
	})

	t.Run("no indicator", func(t *testing.T) {
		if indicatorsOfNoSQLInjection("stable", "stable", payload) {
			t.Fatal("expected no detection when responses match")
		}
	})
}

func patchTransportForAddr(t *testing.T, mutated, real string) func() {
	t.Helper()
	if mutated == real {
		return func() {}
	}
	base, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		t.Fatalf("unexpected transport type %T", http.DefaultTransport)
	}
	clone := base.Clone()
	originalDial := clone.DialContext
	if originalDial == nil {
		dialer := &net.Dialer{}
		originalDial = dialer.DialContext
	}
	clone.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		if addr == mutated {
			addr = real
		}
		return originalDial(ctx, network, addr)
	}
	http.DefaultTransport = clone
	return func() {
		http.DefaultTransport = base
	}
}
