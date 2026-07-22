package hermes

type ActionResult struct {
	CommandID string `json:"commandId"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

type Client interface {
	SendAction(workspaceID, command string) (ActionResult, error)
}

type MockClient struct{}

func (m *MockClient) SendAction(workspaceID, command string) (ActionResult, error) {
	return ActionResult{
		CommandID: "mock-" + workspaceID,
		Status:    "accepted",
		Message:   command,
	}, nil
}
