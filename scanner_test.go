package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPerformAuthTest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "admin" || password != "password" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := server.Client()
	endpoint := APIEndpoint{URL: server.URL, Method: "GET"}
	auth := Auth{Username: "admin", Password: "password"}

	err := performAuthTest(client, endpoint, auth)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	auth.Password = "wrongpassword"
	err = performAuthTest(client, endpoint, auth)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestPerformHTTPMethodTest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := server.Client()
	endpoint := APIEndpoint{URL: server.URL, Method: "POST"}

	err := performHTTPMethodTest(client, endpoint)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	endpoint.Method = "GET"
	err = performHTTPMethodTest(client, endpoint)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestPerformInjectionTest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		if strings.Contains(string(body), "' OR '1'='1") {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := server.Client()
	endpoint := APIEndpoint{URL: server.URL, Method: "POST", Body: "key=%s"}
	payload := "' OR '1'='1"

	err := performInjectionTest(client, endpoint, payload)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	payload = "safe_payload"
	err = performInjectionTest(client, endpoint, payload)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
