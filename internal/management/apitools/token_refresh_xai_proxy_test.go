package apitools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
	coreauth "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/auth"
)

func TestRefreshXAIOAuthAccessTokenUsesProxyPoolEntry(t *testing.T) {
	connectHosts := make(chan string, 1)
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodConnect {
			select {
			case connectHosts <- r.Host:
			default:
			}
		}
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer proxy.Close()

	svc := &Service{cfg: &config.Config{
		ProxyPool: []config.ProxyPoolEntry{{ID: "pool-a", Name: "A", URL: proxy.URL, Enabled: true}},
	}}
	auth := &coreauth.Auth{
		Provider: "xai",
		ProxyID:  "pool-a",
		Metadata: map[string]any{
			"type":           "xai",
			"auth_kind":      "oauth",
			"refresh_token":  "rt",
			"token_endpoint": "https://auth.x.ai/oauth2/token",
		},
	}

	if _, err := svc.refreshXAIOAuthAccessToken(context.Background(), auth); err == nil {
		t.Fatal("expected refresh to fail through the stub proxy")
	}
	select {
	case host := <-connectHosts:
		if host != "auth.x.ai:443" {
			t.Fatalf("proxy CONNECT host = %q, want auth.x.ai:443", host)
		}
	default:
		t.Fatal("refresh request did not go through the proxy pool entry")
	}
}
