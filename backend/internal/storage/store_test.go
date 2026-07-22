package storage

import (
	"testing"

	"wealthos-backend/internal/domain"
)

func TestInMemoryGetWorkspaceRulesPrioritizesAccountScopedRule(t *testing.T) {
	store := NewInMemoryStore()

	ws := domain.ID("ws-1")
	acc := domain.ID("acc-1")

	workspaceRule, err := store.CreateAutomationRule(domain.AutomationRule{
		WorkspaceID: ws,
		Name:        "workspace-any",
		Priority:    1,
		Direction:   "out",
		Enabled:     true,
	})
	if err != nil {
		t.Fatalf("create workspace rule: %v", err)
	}

	accountRule, err := store.CreateAutomationRule(domain.AutomationRule{
		WorkspaceID: ws,
		AccountID:   acc,
		Name:        "account-specific",
		Priority:    10,
		Direction:   "out",
		Enabled:     true,
	})
	if err != nil {
		t.Fatalf("create account rule: %v", err)
	}

	rules := store.GetWorkspaceRules(ws, acc, "out")
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
	if rules[0].ID != accountRule.ID {
		t.Fatalf("expected account-scoped rule first, got %s then %s", rules[0].Name, rules[1].Name)
	}
	if rules[1].ID != workspaceRule.ID {
		t.Fatalf("expected workspace rule second, got %s", rules[1].Name)
	}
}

func TestInMemoryGetWorkspaceRulesFiltersByDirection(t *testing.T) {
	store := NewInMemoryStore()
	ws := domain.ID("ws-2")

	_, err := store.CreateAutomationRule(domain.AutomationRule{
		WorkspaceID: ws,
		Name:        "in-only",
		Priority:    1,
		Direction:   "in",
		Enabled:     true,
	})
	if err != nil {
		t.Fatalf("create in rule: %v", err)
	}
	_, err = store.CreateAutomationRule(domain.AutomationRule{
		WorkspaceID: ws,
		Name:        "out-only",
		Priority:    1,
		Direction:   "out",
		Enabled:     true,
	})
	if err != nil {
		t.Fatalf("create out rule: %v", err)
	}

	inRules := store.GetWorkspaceRules(ws, "", "in")
	if len(inRules) != 1 {
		t.Fatalf("expected 1 in rule, got %d", len(inRules))
	}
	if inRules[0].Direction != "in" {
		t.Fatalf("expected in-direction rule, got %q", inRules[0].Direction)
	}
}

