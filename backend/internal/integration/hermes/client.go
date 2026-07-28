package hermes

type ActionResult struct {
	CommandID string `json:"commandId"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

type Client interface {
	SendAction(userID, command string) (ActionResult, error)
}

type MockClient struct{}

func (m *MockClient) SendAction(userID, command string) (ActionResult, error) {
	return ActionResult{
		CommandID: "mock-" + userID,
		Status:    "accepted",
		Message:   command,
	}, nil
}
