package storage

import (
	"testing"

	"wealthos-backend/internal/domain"
)

func TestNilUUIDTreatsBlankAndNonUUIDValuesAsNull(t *testing.T) {
	if got := nilUUID(""); got != nil {
		t.Fatalf("expected blank ID to map to nil, got %#v", got)
	}
	if got := nilUUID("uncategorized"); got != nil {
		t.Fatalf("expected non-UUID sentinel to map to nil, got %#v", got)
	}

	valid := domain.ID("1a7a0094-74dd-4ca8-879e-9cd782da7f66")
	if got := nilUUID(valid); got != valid {
		t.Fatalf("expected valid UUID to be preserved, got %#v", got)
	}
}

func TestInMemoryGetUserRulesPrioritizesAccountScopedRule(t *testing.T) {
	store := NewInMemoryStore()

	ws := domain.ID("ws-1")
	acc := domain.ID("acc-1")

	userRule, err := store.CreateAutomationRule(domain.AutomationRule{
		UserID:    ws,
		Name:      "user-any",
		Priority:  1,
		Direction: "out",
		Enabled:   true,
	})
	if err != nil {
		t.Fatalf("create user rule: %v", err)
	}

	accountRule, err := store.CreateAutomationRule(domain.AutomationRule{
		UserID:    ws,
		AccountID: acc,
		Name:      "account-specific",
		Priority:  10,
		Direction: "out",
		Enabled:   true,
	})
	if err != nil {
		t.Fatalf("create account rule: %v", err)
	}

	rules := store.GetUserRules(ws, acc, "out")
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
	if rules[0].ID != accountRule.ID {
		t.Fatalf("expected account-scoped rule first, got %s then %s", rules[0].Name, rules[1].Name)
	}
	if rules[1].ID != userRule.ID {
		t.Fatalf("expected user rule second, got %s", rules[1].Name)
	}
}

func TestInMemoryGetUserRulesFiltersByDirection(t *testing.T) {
	store := NewInMemoryStore()
	ws := domain.ID("ws-2")

	_, err := store.CreateAutomationRule(domain.AutomationRule{
		UserID:    ws,
		Name:      "in-only",
		Priority:  1,
		Direction: "in",
		Enabled:   true,
	})
	if err != nil {
		t.Fatalf("create in rule: %v", err)
	}
	_, err = store.CreateAutomationRule(domain.AutomationRule{
		UserID:    ws,
		Name:      "out-only",
		Priority:  1,
		Direction: "out",
		Enabled:   true,
	})
	if err != nil {
		t.Fatalf("create out rule: %v", err)
	}

	inRules := store.GetUserRules(ws, "", "in")
	if len(inRules) != 1 {
		t.Fatalf("expected 1 in rule, got %d", len(inRules))
	}
	if inRules[0].Direction != "in" {
		t.Fatalf("expected in-direction rule, got %q", inRules[0].Direction)
	}
}

func TestDeleteLoanRemovesItsAutoDisbursementTransaction(t *testing.T) {
	store := NewInMemoryStore()
	userID := domain.ID("ws-loan-delete")
	loan, err := store.CreateLoan(domain.Loan{
		UserID:           userID,
		PrincipalInitial: "1000000",
		PrincipalBalance: "1000000",
		AnnualRate:       "109.5",
	})
	if err != nil {
		t.Fatalf("create loan: %v", err)
	}
	if _, err := store.CreateTransaction(domain.Transaction{
		UserID:    userID,
		AccountID: "account-1",
		Type:      domain.TransactionTypeLoanDisbursement,
		Amount:    "1000000",
		Currency:  "VND",
		Note:      "loan principal: " + string(loan.ID),
		Source:    "loan_disbursement",
	}); err != nil {
		t.Fatalf("create loan disbursement: %v", err)
	}

	if err := store.DeleteLoan(userID, loan.ID); err != nil {
		t.Fatalf("delete loan: %v", err)
	}
	if _, ok := store.GetLoan(loan.ID); ok {
		t.Fatal("expected loan to be deleted")
	}
	if transactions := store.ListTransactions(userID, ""); len(transactions) != 0 {
		t.Fatalf("expected auto disbursement to be deleted, got %#v", transactions)
	}
}
