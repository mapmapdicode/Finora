package sepay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client interface {
	StartConnectUrl(userID string) (string, error)
	RevokeConnection(connectionID string) error
	VerifyWebhook(r *http.Request) error
}

type MockClient struct {
	BaseURL string
}

// BankHubLinkClient only exposes the server-side operations needed to start a
// hosted-link session. Client credentials never leave the backend.
type BankHubLinkClient interface {
	Configured() bool
	CreateLink(ctx context.Context, companyXID, completionRedirectURI string) (LinkSession, error)
	ListBankAccounts(ctx context.Context, companyXID, query string) ([]BankHubAccount, error)
}

type LinkSession struct {
	XID           string `json:"xid"`
	HostedLinkURL string `json:"hosted_link_url"`
	ExpiresAt     string `json:"expires_at"`
}

type BankHubTransaction struct {
	TransactionID   string `json:"transaction_id"`
	TransactionDate string `json:"transaction_date"`
	BankAccountXID  string `json:"bank_account_xid"`
	ReferenceNumber string `json:"reference_number"`
	TransferType    string `json:"transfer_type"`
	Amount          any    `json:"amount"`
	Content         string `json:"transaction_content"`
}

// BankHubAccount is the provider-owned representation returned by the Bank
// Hub bank-account endpoint. Its fields are intentionally not accepted from
// the mobile app: the backend fetches them with its own bearer token.
type BankHubAccount struct {
	XID              string `json:"xid"`
	BrandName        string `json:"brand_name"`
	BankCode         string `json:"bank_code"`
	AccountNumber    string `json:"account_number"`
	BankAPIConnected any    `json:"bank_api_connected"`
	Active           any    `json:"active"`
}

type BankHubClient struct {
	baseURL      string
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

func NewBankHubClient(baseURL, clientID, clientSecret string) *BankHubClient {
	return &BankHubClient{
		baseURL:      strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		clientID:     strings.TrimSpace(clientID),
		clientSecret: strings.TrimSpace(clientSecret),
		httpClient:   &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *BankHubClient) Configured() bool {
	return c != nil && c.baseURL != "" && c.clientID != "" && c.clientSecret != ""
}

func (c *BankHubClient) CreateLink(ctx context.Context, companyXID, completionRedirectURI string) (LinkSession, error) {
	if !c.Configured() {
		return LinkSession{}, fmt.Errorf("SePay Bank Hub is not configured")
	}
	if strings.TrimSpace(companyXID) == "" {
		return LinkSession{}, fmt.Errorf("SePay Bank Hub company xid is required")
	}

	accessToken, err := c.accessToken(ctx)
	if err != nil {
		return LinkSession{}, err
	}

	payload := map[string]string{
		"company_xid": companyXID,
		"purpose":     "LINK_BANK_ACCOUNT",
	}
	if redirect := strings.TrimSpace(completionRedirectURI); redirect != "" {
		payload["completion_redirect_uri"] = redirect
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return LinkSession{}, err
	}
	linkReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/link-token/create", bytes.NewReader(body))
	if err != nil {
		return LinkSession{}, err
	}
	linkReq.Header.Set("Content-Type", "application/json")
	linkReq.Header.Set("Authorization", "Bearer "+accessToken)

	var session LinkSession
	if err := c.doJSON(linkReq, &session); err != nil {
		return LinkSession{}, fmt.Errorf("create Bank Hub link token: %w", err)
	}
	if session.XID == "" || session.HostedLinkURL == "" {
		return LinkSession{}, fmt.Errorf("create Bank Hub link token: incomplete response")
	}
	return session, nil
}

// ListTransactions is deliberately used only by the reconciliation worker,
// never directly from a mobile request. It closes webhook delivery gaps while
// retaining provider transaction IDs for idempotency.
func (c *BankHubClient) ListTransactions(ctx context.Context, companyXID, bankAccountXID, startDate, endDate string) ([]BankHubTransaction, error) {
	if !c.Configured() {
		return nil, fmt.Errorf("SePay Bank Hub is not configured")
	}
	accessToken, err := c.accessToken(ctx)
	if err != nil {
		return nil, err
	}
	page := 1
	var all []BankHubTransaction
	for {
		q := url.Values{}
		q.Set("company_xid", companyXID)
		q.Set("bank_account_xid", bankAccountXID)
		q.Set("start_transaction_date", startDate)
		q.Set("end_transaction_date", endDate)
		q.Set("per_page", "50")
		q.Set("page", fmt.Sprintf("%d", page))
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/transaction?"+q.Encode(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		var response struct {
			Data []BankHubTransaction `json:"data"`
			Meta struct {
				Page       int  `json:"page"`
				TotalPages int  `json:"total_pages"`
				HasMore    bool `json:"has_more"`
			} `json:"meta"`
		}
		if err := c.doJSON(req, &response); err != nil {
			return nil, fmt.Errorf("list Bank Hub transactions: %w", err)
		}
		all = append(all, response.Data...)
		if len(response.Data) == 0 || (!response.Meta.HasMore && (response.Meta.TotalPages == 0 || response.Meta.TotalPages <= page)) {
			return all, nil
		}
		page++
	}
}

// ListBankAccounts reads the authoritative provider list after Hosted Link
// reports completion. query is normally the account number supplied by the
// Hosted Link event, rather than an arbitrary provider XID from mobile.
func (c *BankHubClient) ListBankAccounts(ctx context.Context, companyXID, query string) ([]BankHubAccount, error) {
	if !c.Configured() {
		return nil, fmt.Errorf("SePay Bank Hub is not configured")
	}
	if strings.TrimSpace(companyXID) == "" {
		return nil, fmt.Errorf("SePay Bank Hub company xid is required")
	}
	accessToken, err := c.accessToken(ctx)
	if err != nil {
		return nil, err
	}
	page := 1
	var all []BankHubAccount
	for {
		q := url.Values{}
		q.Set("company_xid", companyXID)
		q.Set("per_page", "50")
		q.Set("page", fmt.Sprintf("%d", page))
		if value := strings.TrimSpace(query); value != "" {
			q.Set("q", value)
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/bank-account?"+q.Encode(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		var response struct {
			Data []BankHubAccount `json:"data"`
			Meta struct {
				HasMore    bool `json:"has_more"`
				Page       int  `json:"page"`
				TotalPages int  `json:"total_pages"`
			} `json:"meta"`
		}
		if err := c.doJSON(req, &response); err != nil {
			return nil, fmt.Errorf("list Bank Hub accounts: %w", err)
		}
		all = append(all, response.Data...)
		if (!response.Meta.HasMore && (response.Meta.TotalPages == 0 || page >= response.Meta.TotalPages)) || len(response.Data) == 0 {
			return all, nil
		}
		page++
	}
}

func (c *BankHubClient) accessToken(ctx context.Context) (string, error) {
	form := url.Values{}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.clientID, c.clientSecret)
	var tokenResponse struct {
		AccessToken string `json:"access_token"`
	}
	if err := c.doJSON(req, &tokenResponse); err != nil {
		return "", fmt.Errorf("get Bank Hub access token: %w", err)
	}
	if tokenResponse.AccessToken == "" {
		return "", fmt.Errorf("get Bank Hub access token: response has no access_token")
	}
	return tokenResponse.AccessToken, nil
}

func (c *BankHubClient) doJSON(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("unexpected HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func (m *MockClient) StartConnectUrl(userID string) (string, error) {
	return m.BaseURL + "/integrations/sepay/mock?user=" + userID, nil
}

func (m *MockClient) RevokeConnection(connectionID string) error {
	return nil
}

func (m *MockClient) VerifyWebhook(r *http.Request) error {
	_ = r
	return nil
}
