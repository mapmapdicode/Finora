package sepay

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBankHubClientCreateLinkUsesServerSideCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/token":
			user, password, ok := r.BasicAuth()
			if !ok || user != "client-id" || password != "client-secret" {
				t.Fatalf("missing or invalid basic credentials")
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"access-token","ttl":60000}`))
		case "/v1/link-token/create":
			if got := r.Header.Get("Authorization"); got != "Bearer access-token" {
				t.Fatalf("expected bearer token, got %q", got)
			}
			body, _ := io.ReadAll(r.Body)
			for _, want := range []string{`"company_xid":"company-xid"`, `"purpose":"LINK_BANK_ACCOUNT"`, `"completion_redirect_uri":"https://app.example/callback"`} {
				if !strings.Contains(string(body), want) {
					t.Fatalf("request body %s does not contain %s", body, want)
				}
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"xid":"link-xid","hosted_link_url":"https://bankhub.example/link","expires_at":"2026-01-01 00:00:00"}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewBankHubClient(server.URL, "client-id", "client-secret")
	link, err := client.CreateLink(context.Background(), "company-xid", "https://app.example/callback")
	if err != nil {
		t.Fatalf("CreateLink: %v", err)
	}
	if link.XID != "link-xid" || link.HostedLinkURL != "https://bankhub.example/link" {
		t.Fatalf("unexpected link response: %+v", link)
	}
}

func TestBankHubClientRequiresConfiguration(t *testing.T) {
	client := NewBankHubClient("", "", "")
	if client.Configured() {
		t.Fatal("expected unconfigured client")
	}
	if _, err := client.CreateLink(context.Background(), "company-xid", ""); err == nil {
		t.Fatal("expected unconfigured client error")
	}
}

func TestBankHubClientListBankAccountsUsesProviderFilterAndPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/token":
			_, _ = w.Write([]byte(`{"access_token":"access-token"}`))
		case "/v1/bank-account":
			if got := r.Header.Get("Authorization"); got != "Bearer access-token" {
				t.Fatalf("missing bearer token: %q", got)
			}
			if got := r.URL.Query().Get("company_xid"); got != "company-xid" {
				t.Fatalf("company_xid = %q", got)
			}
			if got := r.URL.Query().Get("q"); got != "123456" {
				t.Fatalf("q = %q", got)
			}
			if r.URL.Query().Get("page") == "1" {
				_, _ = w.Write([]byte(`{"data":[{"xid":"account-1","brand_name":"MBBank","account_number":"123456","bank_api_connected":true,"active":true}],"meta":{"has_more":true}}`))
				return
			}
			_, _ = w.Write([]byte(`{"data":[{"xid":"account-2","brand_name":"MBBank","account_number":"123456","bank_api_connected":"true","active":"active"}],"meta":{"has_more":false}}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	accounts, err := NewBankHubClient(server.URL, "id", "secret").ListBankAccounts(context.Background(), "company-xid", "123456")
	if err != nil {
		t.Fatalf("ListBankAccounts: %v", err)
	}
	if len(accounts) != 2 || accounts[0].XID != "account-1" || accounts[1].XID != "account-2" {
		t.Fatalf("unexpected accounts: %+v", accounts)
	}
}
