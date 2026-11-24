package profile

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDiscover(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/openid-configuration" {
			http.NotFound(w, r)
			return
		}

		w.Write([]byte(`{"issuer":"x","userinfo_endpoint":"https://idp.example.com/userinfo"}`))
	}))
	defer server.Close()

	userInfoURL, err := Discover(t.Context(), server.URL+"/")
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	if userInfoURL != "https://idp.example.com/userinfo" {
		t.Errorf("userinfo URL = %q", userInfoURL)
	}
}

func TestDiscoverWithoutUserInfoEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"issuer":"x"}`))
	}))
	defer server.Close()

	if _, err := Discover(t.Context(), server.URL); err == nil {
		t.Fatal("Discover succeeded without a userinfo endpoint")
	}
}

func TestFetchUserInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		w.Write([]byte(`{"sub":"123","name":"Test User","picture":"https://idp.example.com/a.png"}`))
	}))
	defer server.Close()

	userInfo, err := FetchUserInfo(t.Context(), server.URL, "Bearer token")
	if err != nil {
		t.Fatalf("FetchUserInfo: %v", err)
	}

	if userInfo["name"] != "Test User" || userInfo["picture"] != "https://idp.example.com/a.png" {
		t.Errorf("user info = %v", userInfo)
	}

	if _, err = FetchUserInfo(t.Context(), server.URL, "Bearer wrong"); err == nil {
		t.Error("FetchUserInfo succeeded with a rejected token")
	}
}
