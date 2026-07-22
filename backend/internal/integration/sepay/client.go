package sepay

import "net/http"

type Client interface {
	StartConnectUrl(workspaceID string) (string, error)
	RevokeConnection(connectionID string) error
	VerifyWebhook(r *http.Request) error
}

type MockClient struct {
	BaseURL string
}

func (m *MockClient) StartConnectUrl(workspaceID string) (string, error) {
	return m.BaseURL + "/integrations/sepay/mock?workspace=" + workspaceID, nil
}

func (m *MockClient) RevokeConnection(connectionID string) error {
	return nil
}

func (m *MockClient) VerifyWebhook(r *http.Request) error {
	_ = r
	return nil
}
